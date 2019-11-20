package metrics

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"sync"
)

type MetricType string

const (
	MetricIncrement MetricType = "Increment"
	MetricDecrement MetricType = "Decrement"
)

type Client interface {
	Init(conf config.Config) error
	Decrement(name string, tags []string, rate float64)
	Increment(name string, tags []string, rate float64)
}

var once sync.Once
var instance Client

func InstantiateClient(conf config.Config) error {
	// Return an error if `metrics_client` isn't set in config.
	if !conf.IsSet("metrics_client") {
		return fmt.Errorf("`metrics_client` not set in config, unable to instantiate metrics client.")
	}

	var err error = nil
	name := conf.GetString("metrics_client")

	once.Do(func() {
		switch name {
		case "dogstatsd":
			client := &DatadogStatsdMetricsClient{}

			if err = client.Init(conf); err != nil {
				break
			}

			instance = client
		default:
			err = fmt.Errorf("No Client named [%s] was found", name)
		}
	})

	return err
}

func GetInstance() Client {
	return instance
}
