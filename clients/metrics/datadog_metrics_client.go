package metrics

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/stitchfix/flotilla-os/config"
)

type DatadogStatsdMetricsClient struct {
	client *statsd.Client
}

// Initialize the client. Assumes the following keys are passed in:
// *metrics.dogstatsd.address* -- localhost:8125
// *metrics.dogstatsd.namespace* -- fixed key you want to prefix to all the metrics
func (dd *DatadogStatsdMetricsClient) Init(conf config.Config) error {
	host := os.Getenv("DD_AGENT_HOST")
	var addr string
	// If the host contains a colon and does not contain a square bracket, then the address is ipv6
	if strings.Contains(host, ":") && !strings.Contains(host, "[") {
		addr = fmt.Sprintf("[%s]:8125", host)
	} else {
		addr = fmt.Sprintf("%s:8125", host)
	}
	client, err := statsd.New(addr)
	if err != nil {
		return err
	}

	dd.client = client

	return nil
}

// Decrement metric value, tags associated with the metric, and rate corresponds to the value
func (dd *DatadogStatsdMetricsClient) Decrement(name Metric, tags []string, rate float64) error {
	return dd.client.Decr(string(name), tags, rate)
}

// Increment metric value, tags associated with the metric, and rate corresponds to the value
func (dd *DatadogStatsdMetricsClient) Increment(name Metric, tags []string, rate float64) error {
	return dd.client.Incr(string(name), tags, rate)
}

// Histogram tracks the statistical distribution of a set of values
func (dd *DatadogStatsdMetricsClient) Histogram(name Metric, value float64, tags []string, rate float64) error {
	return dd.client.Histogram(string(name), value, tags, rate)
}

// Distribution tracks the statistical distribution of a set of values
func (dd *DatadogStatsdMetricsClient) Distribution(name Metric, value float64, tags []string, rate float64) error {
	return dd.client.Distribution(string(name), value, tags, rate)
}

// Timing sends timing information, it is an alias for TimeInMilliseconds
func (dd *DatadogStatsdMetricsClient) Timing(name Metric, value time.Duration, tags []string, rate float64) error {
	return dd.client.Timing(string(name), value, tags, rate)
}

// Set counts the number of unique elements in a group
func (dd *DatadogStatsdMetricsClient) Set(name Metric, value string, tags []string, rate float64) error {
	return dd.client.Set(string(name), value, tags, rate)
}

// NewEvent creates a new event with the given title and text.
func (dd *DatadogStatsdMetricsClient) Event(e event) error {
	se := statsd.NewEvent(e.Title, e.Text)
	se.Tags = e.Tags
	return dd.client.Event(se)
}
