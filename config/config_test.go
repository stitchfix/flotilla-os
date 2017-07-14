package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	var c Config
	c, _ = NewConfig(nil)

	toSet := "sprinkles"
	os.Setenv("CUPCAKE", toSet)

	if c.GetString("cupcake") != toSet {
		t.Errorf("Environment variables not set - expected %s but was %s", toSet, c.GetString("cupcake"))
	}

	confDir := "../conf"
	c, _ = NewConfig(&confDir)
	if !c.IsSet("queue.namespace") || c.GetString("queue.namespace") != "dev-flotilla" {
		t.Errorf(
			"Expected to read from conf dir [queue.namespace]:[dev-flotilla], was: %s",
			c.GetString("queue.namespace"))
	}
}
