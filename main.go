package main

import (
	"log"
	"os"

	"github.com/doubledutch/go-env"
	"github.com/doubledutch/lager"
	"github.com/doubledutch/mux/gob"
	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum-basic/basic"
	"github.com/doubledutch/quantum-basic/remote"
	"github.com/doubledutch/quantum/agent"
	"github.com/doubledutch/quantum/consul"
	"github.com/doubledutch/quantum/inmemory"
)

const (
	defaultPort      = ":8877"
	defaultLogLevels = "IE"
)

var (
	port       string
	logLevels  string
	consulHTTP string
)

func main() {
	env.StringVar(&port, "LISTEN_PORT", defaultPort, "port to listen on")
	env.StringVar(&logLevels, "LOG_LEVELS", defaultLogLevels, "levels to log")
	env.StringVar(&consulHTTP, "CONSUL_HTTP", "", "consul http address")
	env.Parse()

	lgr := lager.NewLogLager(&lager.LogConfig{
		Levels: lager.LevelsFromString(logLevels),
		Output: os.Stdout,
	})

	registrators := []quantum.Registrator{inmemory.NewRegistrator()}

	if consulHTTP != "" {
		crg := consul.NewRegistrator(consulHTTP, lgr)
		registrators = append(registrators, crg)
		lgr.Debugf("Registered consul registrator")
	}

	mrg := &quantum.MultiRegistrator{
		Registrators: registrators,
	}

	cc := &quantum.ConnConfig{
		Config: &quantum.Config{
			Pool:  new(gob.Pool),
			Lager: lgr,
		},
	}

	registry := inmemory.NewRegistry(lgr)

	config := &agent.Config{
		ConnConfig:  cc,
		Port:        port,
		Registry:    registry,
		Registrator: mrg,
	}

	srv := agent.New(config)
	srv.Add(&basic.Job{})
	srv.Add(&remote.Job{})
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
