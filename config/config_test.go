package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	var err error
	var c Config
	c, _ = NewConfig(nil)

	toSet := "sprinkles"
	os.Setenv("CUPCAKE", toSet)

	if c.GetString("cupcake") != toSet {
		t.Errorf("Environment variables not set - expected %s but was %s", toSet, c.GetString("cupcake"))
	}

	confDir := "../conf"
	c, err = NewConfig(&confDir)
	t.Log(err)
	if !c.IsSet("queue.prefix") || c.GetString("queue.prefix") != "flotilla-os-test" {
		t.Errorf(
			"Expected to read from conf dir [queue.prefix]:[flotilla-os-test], was: %s",
			c.GetString("queue.prefix"))
	}
}
