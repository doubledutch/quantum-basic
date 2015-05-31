package basic

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/doubledutch/quantum/agent"
	"github.com/doubledutch/quantum/client"
)

const port = ":8877"

func TestRunType(t *testing.T) {
	job := &Job{}
	if job.Type() != runType {
		t.Fatalf("'%s' != '%s'", job.Type(), runType)
	}
}

func TestConfigure(t *testing.T) {
	request := Request{
		Command: "echo hello world",
	}

	d, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	job := &Job{}
	if err := job.Configure(d); err != nil {
		t.Fatal(err)
	}

	if job.r.Command != request.Command {
		t.Fatalf("'%s' != '%s'", job.r.Command, request.Command)
	}
}

func TestRun(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	expected := "hello world"
	basicRequest := []byte("{\"command\":\"echo " + expected + "\"}")

	go func() {
		srv := agent.New(&agent.Config{
			Port: port,
		})
		srv.Add(&Job{})

		srv.Accept(l)
	}()

	request := NewRequestJSON(basicRequest)

	client := client.New(nil)
	conn, err := client.Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if <-conn.Logs() != "Running echo "+expected+"\n" {
			t.Fatal("expected first line to be Running echo hello world\\n")
		}

		if <-conn.Logs() != expected+"\n" {
			t.Fatal("expected second line to be hello world\\n")
		}
	}()

	if err := conn.Run(request); err != nil {
		t.Fatal(err)
	}
}

func TestRunErr(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	basicRequest := []byte("{\"command\":\"asdf\"}")

	go func() {
		srv := agent.New(&agent.Config{
			Port: port,
		})
		srv.Add(&Job{})

		srv.Accept(l)
	}()

	request := NewRequestJSON(basicRequest)

	client := client.New(nil)
	conn, err := client.Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if <-conn.Logs() != "Running asdf\n" {
			t.Fatal("expected first line to be Running echo hello world\\n")
		}
		for _ = range conn.Logs() {
			// consume the rest
		}
	}()

	if err := conn.Run(request); err == nil {
		t.Fatal("expected command to fail")
	}
}
