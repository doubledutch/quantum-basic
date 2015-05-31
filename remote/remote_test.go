package remote

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"testing"

	"github.com/doubledutch/quantum/agent"
	"github.com/doubledutch/quantum/client"
)

const port = ":8877"

const testCommand = `
#! /bin/sh
echo hello world
`

type littleServer struct{}

func (littleServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(testCommand))
}
func TestRun(t *testing.T) {
	ls, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ls.Close()

	la, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer la.Close()

	go func() {
		h := littleServer{}

		http.Serve(ls, h)
	}()

	go func() {
		srv := agent.New(&agent.Config{
			Port: port,
		})
		srv.Add(&Job{})

		srv.Accept(la)
	}()

	remoteRequest := []byte("{\"source\": \"http://" + ls.Addr().String() + "/hello.sh\"," +
		"\"timeout\":200}")

	request := NewRequest(remoteRequest)
	client := client.New(nil)
	conn, err := client.Dial(la.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		// First line should be Running ... hello.sh
		running := <-conn.Logs()
		cmdReg := regexp.MustCompile(`^Running (/tmp/|C:\\Windows\\Temp\\)hello\.sh`)
		if !cmdReg.MatchString(running) {
			fmt.Println(running)
			t.Fatal("didn't get running line")
			t.FailNow()
		}
		// Next line should be the output: echo hello world
		output := <-conn.Logs()
		if output != "hello world\n" {
			fmt.Println(output)
			t.Fatal("didn't get running line")
			t.FailNow()
		}

		// Consume the rest
		for _ = range conn.Logs() {
		}
	}()

	if err := conn.Run(request); err != nil {
		t.Fatal(err)
	}
}

func TestRemoteRunHttpErr(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		srv := agent.New(&agent.Config{
			Port: port,
		})
		srv.Add(&Job{})

		srv.Accept(l)
	}()

	remoteRequest := []byte("{\"source\": \"http://this-should-not-resolve/hello.sh\"," +
		"\"timeout\":200}")

	request := NewRequest(remoteRequest)

	client := client.New(nil)
	conn, err := client.Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for _ = range conn.Logs() {

		}
	}()

	if err := conn.Run(request); err == nil {
		t.Fatal("expected command to fail")
	}
}
