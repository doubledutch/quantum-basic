package basic

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/doubledutch/quantum"
)

const (
	runType = "basic"
)

// RunType is used to add Type to structs related to Job
type RunType struct{}

// Type for Job
func (bt RunType) Type() string {
	return runType
}

// Job defines a job for running commands
type Job struct {
	RunType
	*quantum.BasicJob
	r *Request
}

// Configure Job with Request
func (j *Job) Configure(d []byte) error {
	request := new(Request)
	err := json.Unmarshal(d, request)
	if err != nil {
		return err
	}

	j.BasicJob = quantum.NewBasicJob(j)

	j.r = request
	return nil
}

// Steps for Job
func (j *Job) Steps() []quantum.Step {
	return []quantum.Step{
		&RunStep{
			Command: j.r.Command,
		},
	}
}

// Request wraps the command for RunStep
type Request struct {
	RunType
	Command string
}

// NewRequest creates a quantum.Request from a Request
func NewRequest(command string) (qr quantum.Request, err error) {
	r := Request{
		Command: command,
	}
	d, err := json.Marshal(r)
	if err != nil {
		return
	}

	qr = NewRequestJSON(d)
	return
}

// NewRequestJSON creates a new basic request using json
func NewRequestJSON(data []byte) quantum.Request {
	return quantum.Request{
		Type: runType,
		Data: data,
	}
}

// RunStep for running command
type RunStep struct {
	Command string
}

// Run the command with quantum.Runner
func (b *RunStep) Run(state quantum.StateBag) error {
	runner := state.Get("runner").(quantum.Runner)
	conn := state.Get("conn").(quantum.AgentConn)
	outCh := conn.Logs()
	sigCh := conn.Signals()

	if b.Command == "" {
		commandRaw, ok := state.GetOk("command")
		if ok {
			b.Command = commandRaw.(string)
		}
	}

	log.Printf("Running command: %v", b.Command)

	err := runner.Run(b.Command, outCh, sigCh)
	if err != nil {
		return errors.New("Cmd: " + b.Command + " failed: " + err.Error())
	}
	return nil
}

// Cleanup the RunStep -- does nothing
func (b *RunStep) Cleanup(state quantum.StateBag) {
	// Nothing
}
