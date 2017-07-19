package main

import (
	"fmt"
	gklog "github.com/go-kit/kit/log"
	"github.com/stitchfix/flotilla-os/clients"
	"github.com/stitchfix/flotilla-os/config"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"log"
	"os"
)

func main() {
	//
	// Use go-kit for structured logging
	//
	l := gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stderr))
	logger := flotillaLog.NewLogger(l, nil)

	//
	// Wrap viper for configuration
	//
	confDir := "conf"
	c, err := config.NewConfig(&confDir)
	if err != nil {
		log.Fatal(err)
	}

	sm, err := state.NewStateManager(c)
	if err != nil {
		log.Fatal(err)
	}

	qm, err := queue.NewQueueManager(c)

	logger.Log("queue_manager", qm.Name())
	logger.Log("state_manager", sm.Name())
	logger.Log("msg", "initialized!")

	rc, err := clients.NewRegistryClient(c)
	fmt.Println(rc.IsImageValid("library/postgres:latest"))
}
