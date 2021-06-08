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
	k8sClusterClient, err := cluster.NewClusterClient(c, state.K8SEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize K8S cluster client"))
		//TODO
		//os.Exit(1)
	}

	k8sLogsClient, err := logs.NewLogsClient(c, logger, state.K8SEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize K8S logs client"))
		//TODO
		//os.Exit(1)
	}

	//
	// Get queue manager for queuing runs
	//
	k8sQueueManager, err := queue.NewQueueManager(c, state.K8SEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize k8s queue manager"))
		os.Exit(1)
	}

	//
	// Get execution engine for interacting with backend
	// execution management framework (eg. K8S)
	//
	k8sExecutionEngine, err := engine.NewExecutionEngine(c, k8sQueueManager, state.K8SEngine, logger)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize K8S execution engine"))
		os.Exit(1)
	}

	app, err := flotilla.NewApp(c, logger, k8sLogsClient, k8sExecutionEngine, stateManager, k8sClusterClient, k8sQueueManager)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize app"))
		os.Exit(1)
	}

	log.Fatal(app.Run())
}
