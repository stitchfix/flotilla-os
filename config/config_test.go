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
	if !c.IsSet("queue_namespace") || c.GetString("queue_namespace") != "dev-flotilla" {
		t.Errorf("Expected to read from conf dir [queue_namespace]:[dev-flotilla], was: %s",
			c.GetString("queue_namespace"))
	}
}

func TestConfig_GetPriorityClassForTier(t *testing.T) {
	confDir := "../conf"
	c, err := NewConfig(&confDir)
	if err != nil {
		t.Fatalf("Error creating config: %s", err.Error())
	}

	testCases := []struct {
		name          string
		tier          string
		expectedClass string
	}{
		{"tier 1", "1", "flotilla-tier-1"},
		{"tier 2", "2", "flotilla-tier-2"},
		{"tier 3", "3", "flotilla-tier-3"},
		{"tier 4", "4", "flotilla-tier-4"},
		{"unknown tier", "99", "flotilla-tier-4"}, // Should return default
		{"empty tier", "", "flotilla-tier-4"},     // Should return default
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := c.GetPriorityClassForTier(tc.tier)
			if result != tc.expectedClass {
				t.Errorf("Expected %s, got %s", tc.expectedClass, result)
			}
		})
	}
}

func TestConfig_GetPriorityClassesEnabled(t *testing.T) {
	confDir := "../conf"
	c, err := NewConfig(&confDir)
	if err != nil {
		t.Fatalf("Error creating config: %s", err.Error())
	}

	enabled := c.GetPriorityClassesEnabled()
	// Should be false by default
	if enabled {
		t.Error("Priority classes should be disabled by default")
	}
}

func TestConfig_GetAllConfiguredPriorityClasses(t *testing.T) {
	confDir := "../conf"
	c, err := NewConfig(&confDir)
	if err != nil {
		t.Fatalf("Error creating config: %s", err.Error())
	}

	classes := c.GetAllConfiguredPriorityClasses()
	// Should have 4 tier classes + 1 default = 5 total (default might be duplicate)
	if len(classes) < 4 {
		t.Errorf("Expected at least 4 priority classes, got %d", len(classes))
	}
}
