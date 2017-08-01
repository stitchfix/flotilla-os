package main

import (
	"fmt"
	gklog "github.com/go-kit/kit/log"
	"github.com/stitchfix/flotilla-os/clients/cluster"
	"github.com/stitchfix/flotilla-os/clients/logs"
	"github.com/stitchfix/flotilla-os/clients/registry"
	"github.com/stitchfix/flotilla-os/config"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/services"
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

	rc, err := registry.NewRegistryClient(c)

	cc, err := cluster.NewClusterClient(c)

	lc, err := logs.NewLogsClient(c)

	logger.Log("queue_manager", qm.Name())
	logger.Log("state_manager", sm.Name())
	logger.Log("cluster_client", cc.Name())
	logger.Log("logs_client", lc.Name())
	logger.Log("registry_client", rc)
	logger.Log("msg", "initialized!")

	es, err := services.NewExecutionService(sm, qm, cc, rc)
	elist := []state.EnvVar{
		{Name: "FLOTILLA_SERVER_MODE", Value: "prod"},
	}
	env := state.EnvList(elist)
	fmt.Println(es.Create("testing-7658ce4d-5c7b-43b1-a8d5-17c7c4a77fcc", "flotilla-adhoc", &env))
}
