package main

import (
	gklog "github.com/go-kit/kit/log"
	"github.com/stitchfix/flotilla-os/config"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
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

	logger.Log("state_manager", sm)
	logger.Log("msg", "initialized!")
}
