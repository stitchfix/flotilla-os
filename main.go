package main

import (
	"fmt"
	gklog "github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/clients/cluster"
	"github.com/stitchfix/flotilla-os/clients/logs"
	"github.com/stitchfix/flotilla-os/clients/metrics"
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
	l = gklog.With(l, "ts", gklog.DefaultTimestampUTC)
	eventSinks := []flotillaLog.EventSink{flotillaLog.NewLocalEventSink()}
	logger := flotillaLog.NewLogger(l, eventSinks)

	//
	// Wrap viper for configuration
	//
	confDir := args[1]
	c, err := config.NewConfig(&confDir)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize config"))
		os.Exit(1)
	}

	//
	// Instantiate metrics client.
	//
	if err = metrics.InstantiateClient(c); err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize metrics client"))
		os.Exit(1)
	}

	//
	// Get state manager for reading and writing
	// state about definitions and runs
	//
	stateManager, err := state.NewStateManager(c)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize state manager"))
		os.Exit(1)
	}

	//
	// Get registry client for validating images
	//
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize registry client"))
		os.Exit(1)
	}

	//
	// Get cluster client for validating definitions
	// against execution clusters
	//
	eksClusterClient, err := cluster.NewClusterClient(c, state.EKSEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize EKS cluster client"))
		//TODO
		//os.Exit(1)
	}

	eksLogsClient, err := logs.NewLogsClient(c, logger, state.EKSEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize EKS logs client"))
		//TODO
		//os.Exit(1)
	}

	//
	// Get queue manager for queuing runs
	//
	eksQueueManager, err := queue.NewQueueManager(c, state.EKSEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize eks queue manager"))
		os.Exit(1)
	}

	//
	// Get execution engine for interacting with backend
	// execution management framework (eg. EKS)
	//
	eksExecutionEngine, err := engine.NewExecutionEngine(c, eksQueueManager, state.EKSEngine, logger)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize EKS execution engine"))
		os.Exit(1)
	}

	emrExecutionEngine, err := engine.NewExecutionEngine(c, eksQueueManager, state.EKSSparkEngine, logger)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize EMR execution engine"))
		os.Exit(1)
	}
	app, err := flotilla.NewApp(c, logger, eksLogsClient, eksExecutionEngine, stateManager, eksClusterClient, eksQueueManager, emrExecutionEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize app"))
		os.Exit(1)
	}

	log.Fatal(app.Run())
}
