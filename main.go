package main

import (
	"github.com/stitchfix/flotilla-os/state"
	gklog "github.com/go-kit/kit/log"
	"os"
	"log"
)

func main() {
	//
	// Use go-kit for structured logging
	//
	l := gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stderr))

	sm, err := state.NewStateManager(l, "postgres")
	if err != nil {
		log.Fatal(err)
	}

	l.Log("state_manager", sm)
	l.Log("msg", "initialized!")
}
