package main

import (
	"fmt"
	gklog "github.com/go-kit/kit/log"
	"github.com/stitchfix/flotilla-os/clients/cluster"
	"github.com/stitchfix/flotilla-os/clients/logs"
	"github.com/stitchfix/flotilla-os/clients/registry"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	"github.com/stitchfix/flotilla-os/flotilla"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"log"
	"os"
)

func main() {

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: flotilla-os <conf_dir>")
		os.Exit(1)
	}

	//
	// Use go-kit for structured logging
	//
	l := gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stderr))
	eventSinks := []flotillaLog.EventSink{flotillaLog.NewLocalEventSink()}
	logger := flotillaLog.NewLogger(l, eventSinks)

	//
	// Wrap viper for configuration
	//
	confDir := args[1]
	c, err := config.NewConfig(&confDir)
	if err != nil {
		logger.Log("message", "Error initializing configuration")
		log.Fatal(err)
	}

	//
	// Get state manager for reading and writing
	// state about definitions and runs
	//
	sm, err := state.NewStateManager(c)
	if err != nil {
		logger.Log("message", "Error initializing state manager")
		log.Fatal(err)
	}

	//
	// Get queue manager for queuing runs
	//
	qm, err := queue.NewQueueManager(c)
	if err != nil {
		logger.Log("message", "Error initializing queue manager")
		log.Fatal(err)
	}

	//
	// Get registry client for validating images
	//
	rc, err := registry.NewRegistryClient(c)
	if err != nil {
		logger.Log("message", "Error initializing registry client")
		log.Fatal(err)
	}

	//
	// Get cluster client for validating definitions
	// against execution clusters
	//
	cc, err := cluster.NewClusterClient(c)
	if err != nil {
		logger.Log("message", "Error initializing cluster client")
		log.Fatal(err)
	}

	//
	// Get logs client for reading run logs
	//
	lc, err := logs.NewLogsClient(c, logger)
	if err != nil {
		logger.Log("message", "Error initializing logs client")
		log.Fatal(err)
	}

	//
	// Get execution engine for interacting with backend
	// execution management framework (eg. ECS)
	//
	ee, err := engine.NewExecutionEngine(c)
	if err != nil {
		logger.Log("message", "Error initializing execution engine")
		log.Fatal(err)
	}

	app, err := flotilla.NewApp(c, logger, lc, ee, sm, qm, cc, rc)
	log.Fatal(app.Run())
}
