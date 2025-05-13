package worker

import (
	"os"
	"testing"
	"time"

	"github.com/stitchfix/flotilla-os/config"
)

func TestGetPollInterval(t *testing.T) {
	conf, _ := config.NewConfig(nil)

	expected := time.Duration(500) * time.Millisecond
	os.Setenv("WORKER_RETRY_INTERVAL", "500ms")

	interval, err := GetPollInterval("retry", conf)
	if err != nil {
		t.Error(err.Error())
	}

	if interval != expected {
		t.Errorf("Expected interval: [%v] but was [%v]", expected, interval)
	}
}
