package worker

import (
	"github.com/stitchfix/flotilla-os/config"
	"os"
	"testing"
	"time"
)

func TestGetPollInterval(t *testing.T) {
	conf, _ := config.NewConfig(nil)

	expected := time.Duration(500) * time.Millisecond
	os.Setenv("WORKER_RETRY_INTERVAL", "500ms")

	interval, err := GetPollInterval("retry", conf)
	if err != nil {
		t.Errorf(err.Error())
	}

	if interval != expected {
		t.Errorf("Expected interval: [%v] but was [%v]", expected, interval)
	}
}
