package metrics

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
)

type DatadogStatsdMetricsClient struct {
	client *statsd.Client
}

func (dd DatadogStatsdMetricsClient) Init(conf config.Config) error {
	if !conf.IsSet("metrics.datadog.address") {
		return errors.Errorf("Unable to initialize DatadogMetricsClient: metrics.datadog.address must be set in the config.")
	}

	addr := conf.GetString("metrics.datadog.address")
	client, err := statsd.New(addr)
	if err != nil {
		err = errors.Errorf("Unable to initialize DatadogMetricsClient.")
	}

	dd.client = client

	return nil
}

func (dd DatadogStatsdMetricsClient) Measure(metricType string) error {
	return nil
}
