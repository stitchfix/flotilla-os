package metrics

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
)

type DatadogStatsdMetricsClient struct {
	client *statsd.Client
}

func (dd *DatadogStatsdMetricsClient) Init(conf config.Config) error {
	if !conf.IsSet("metrics.dogstatsd.address") {
		return errors.Errorf("Unable to initialize DatadogMetricsClient: metrics.datadog.address must be set in the config.")
	}

	addr := conf.GetString("metrics.dogstatsd.address")
	client, err := statsd.New(addr)

	if err != nil {
		return err
	}

	dd.client = client

	return nil
}

func (dd *DatadogStatsdMetricsClient) Decrement(name Metric, tags []string, rate float64) error {
	return dd.client.Decr(string(name), tags, rate)
}

func (dd *DatadogStatsdMetricsClient) Increment(name Metric, tags []string, rate float64) error {
	fmt.Println("GOT IT IN DD")
	return dd.client.Incr(string(name), tags, rate)
}

func (dd *DatadogStatsdMetricsClient) Histogram(name Metric, value float64, tags []string, rate float64) error {
	return dd.client.Histogram(string(name), value, tags, rate)
}

func (dd *DatadogStatsdMetricsClient) Distribution(name Metric, value float64, tags []string, rate float64) error {
	return dd.client.Distribution(string(name), value, tags, rate)
}

func (dd *DatadogStatsdMetricsClient) Set(name Metric, value string, tags []string, rate float64) error {
	return dd.client.Set(string(name), value, tags, rate)
}

func (dd *DatadogStatsdMetricsClient) Event(e event) error {
	se := statsd.NewEvent(e.Title, e.Text)
	se.Tags = e.Tags
	return dd.client.Event(se)
}
