package main

import (
	"log"
	"os"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/mux/gob"
	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum-basic/basic"
	"github.com/doubledutch/quantum-basic/remote"
	"github.com/doubledutch/quantum/agent"
)

const (
	defaultPort      = ":8877"
	defaultLogLevels = "IE"
)

func main() {
	port := os.Getenv("BASIC_PORT")
	if port == "" {
		port = defaultPort
	}

	logLevels := os.Getenv("BASIC_LOG_LEVELS")
	if logLevels == "" {
		logLevels = defaultLogLevels
	}

	lgr := lager.NewLogLager(&lager.LogConfig{
		Levels: lager.LevelsFromString(logLevels),
		Output: os.Stdout,
	})

	cc := &quantum.ConnConfig{
		Config: &quantum.Config{
			Pool:  new(gob.Pool),
			Lager: lgr,
		},
	}

	config := &agent.Config{
		Port:       port,
		ConnConfig: cc,
	}

	srv := agent.New(config)
	srv.Add(&basic.Job{})
	srv.Add(&remote.Job{})
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
