package main

import (
	"log"
	"os"
	"time"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/mux/gob"
	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum-basic/basic"
	"github.com/doubledutch/quantum-basic/remote"
	"github.com/doubledutch/quantum/agent"
)

const (
	defaultAddr      = "localhost:8877"
	defaultLogLevels = "IE"
)

func main() {
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	logLevels := os.Getenv("LOG_LEVELS")
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

	tlsConfig, err := agent.NewTLSConfigFromEnv()
	if err == nil {
		lgr.Debugf("Using the provided TLS configuration")
		cc.TLSConfig = tlsConfig
		cc.Timeout = 10 * time.Second
	} else if err != agent.ErrProvideAllCertFiles {
		log.Fatalf("Err creating TLS config: %s", err)
	}

	config := &agent.Config{
		Addr:       addr,
		ConnConfig: cc,
	}

	srv := agent.New(config)
	srv.Add(&basic.Job{})
	srv.Add(&remote.Job{})
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
