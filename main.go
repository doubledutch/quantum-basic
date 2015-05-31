package main

import (
	"flag"
	"log"

	"github.com/doubledutch/quantum-basic/basic"
	"github.com/doubledutch/quantum-basic/remote"
	"github.com/doubledutch/quantum/agent"
)

const (
	defaultPort = ":8877"
)

func main() {
	config := agent.RegisterFlags(&agent.Config{
		Port: defaultPort,
	})
	flag.Parse()

	srv := agent.New(config)
	srv.Add(&basic.Job{})
	srv.Add(&remote.Job{})
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
