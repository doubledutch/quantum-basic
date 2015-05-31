package remote

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum-basic/basic"
)

const (
	runType = "remote"
)

// RunType of remote job
type RunType struct{}

// Type of remote job
func (rt RunType) Type() string {
	return runType
}

// Request defines a job for downloading and running scripts
type Request struct {
	Source string
	// Timeout in milliseconds
	Timeout int
}

// NewRequest defines a new remote request
func NewRequest(data []byte) quantum.Request {
	return quantum.Request{
		Type: runType,
		Data: data,
	}
}

// Job defines the job for download and running scripts
type Job struct {
	RunType
	*quantum.BasicJob
	r *Request
}

// Configure the job
func (j *Job) Configure(d []byte) error {
	request := new(Request)
	err := json.Unmarshal(d, request)
	if err != nil {
		return err
	}

	// Set it to a sane default
	if request.Timeout < 1 {
		request.Timeout = 5000
	}

	j.BasicJob = quantum.NewBasicJob(j)

	j.r = request
	return nil
}

// Steps for the RemoteJob
func (j *Job) Steps() []quantum.Step {
	return []quantum.Step{
		&DownloadStep{
			source:  j.r.Source,
			timeout: j.r.Timeout,
		},
		&basic.RunStep{},
	}
}

// DownloadStep downloads the source
type DownloadStep struct {
	source  string
	timeout int
}

// Run downloads the file
func (j *DownloadStep) Run(state quantum.StateBag) error {
	conn := state.Get("conn").(quantum.AgentConn)
	outCh := conn.Logs()

	// Download the file to tmp directory
	var tempDir string
	if runtime.GOOS == "windows" {
		tempDir = "C:\\Windows\\Temp\\"
	} else {
		tempDir = "/tmp/"
	}

	splitSource := strings.Split(j.source, "/")
	fileName := splitSource[len(splitSource)-1]

	// We probably need to do something special for powershell
	// With unix, we can use #! for now
	var prefix string
	splitFileName := strings.Split(fileName, ".")
	ext := splitFileName[len(splitFileName)-1]
	switch ext {
	case "ps1":
		prefix = "Powershell.exe -File "
	default:
		prefix = ""
	}

	filePath := tempDir + fileName

	f, err := os.Create(filePath)
	if err != nil {
		outCh <- fmt.Sprintf("Error creating tmp file %s: %s\n", filePath, err)
		return err
	}
	defer f.Close()

	client := http.Client{
		Timeout: time.Duration(j.timeout) * time.Millisecond,
	}

	r, err := client.Get(j.source)
	if err != nil {
		outCh <- fmt.Sprintf("Error downloading source %s: %s\n", filePath, err)
		return err
	}

	io.Copy(f, r.Body)
	os.Chmod(filePath, 0744)

	// Put the command and the directory in the statebag
	state.Put("filePath", filePath)
	state.Put("command", prefix+filePath)

	return nil
}

// Cleanup cleans up the DownloadStep
func (j *DownloadStep) Cleanup(state quantum.StateBag) {
	// Remove the tmp file
	filePathRaw, ok := state.GetOk("filePath")
	if !ok {
		return
	}

	filePath, ok := filePathRaw.(string)
	if !ok {
		return
	}

	os.Remove(filePath)
}
