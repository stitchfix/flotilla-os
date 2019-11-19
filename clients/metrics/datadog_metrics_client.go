package metrics

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"sync"
)

type DatadogStatsdClient struct {
	client *statsd.Client
}

var instance *DatadogStatsdClient
var once sync.Once

func GetInstance(conf config.Config) (*DatadogStatsdClient, error) {
	var err error = nil
	once.Do(func() {
		instance = &DatadogStatsdClient{}

		if !conf.IsSet("metrics.datadog.address") {
			err = errors.Errorf("Unable to initialize DatadogMetricsClient: metrics.datadog.address must be set in the config.")
		}

		addr := conf.GetString("metrics.datadog.address")
		statsd, err := statsd.New(addr)
		if err != nil {
			err = errors.Errorf("Unable to initialize DatadogMetricsClient.")
		}

		instance.client = statsd
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}
