package main

import (
	"fmt"
	gklog "github.com/go-kit/kit/log"
	"github.com/pkg/errors"
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
	registryClient, err := registry.NewRegistryClient(c)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize registry client"))
		os.Exit(1)
	}

	//
	// Get cluster client for validating definitions
	// against execution clusters
	//
	ecsClusterClient, err := cluster.NewClusterClient(c, state.ECSEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize ECS cluster client"))
		os.Exit(1)
	}

	eksClusterClient, err := cluster.NewClusterClient(c, state.EKSEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize EKS cluster client"))
		//TODO
		//os.Exit(1)
	}

	//
	// Get logs client for reading run logs
	//
	ecsLogsClient, err := logs.NewLogsClient(c, logger, state.ECSEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize ECS logs client"))
		os.Exit(1)
	}

	eksLogsClient, err := logs.NewLogsClient(c, logger, state.EKSEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize EKS logs client"))
		os.Exit(1)
	}

	//
	// Get queue manager for queuing runs
	//
	ecsQueueManager, err := queue.NewQueueManager(c, state.ECSEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize ecs queue manager"))
		os.Exit(1)
	}

	eksQueueManager, err := queue.NewQueueManager(c, state.EKSEngine)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize eks queue manager"))
		os.Exit(1)
	}

	//
	// Get execution engine for interacting with backend
	// execution management framework (eg. ECS)
	//
	ecsExecutionEngine, err := engine.NewExecutionEngine(c, ecsQueueManager, state.ECSEngine, logger)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize ECS execution engine"))
		os.Exit(1)
	}

	eksExecutionEngine, err := engine.NewExecutionEngine(c, eksQueueManager, state.EKSEngine, logger)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize EKS execution engine"))
		os.Exit(1)
	}

	app, err := flotilla.NewApp(c, logger, ecsLogsClient, eksLogsClient, ecsExecutionEngine, eksExecutionEngine, stateManager, ecsClusterClient, eksClusterClient, registryClient)
	if err != nil {
		fmt.Printf("%+v\n", errors.Wrap(err, "unable to initialize app"))
		os.Exit(1)
	}

	log.Fatal(app.Run())
}
