package metrics

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
)

type DatadogMetricsClient struct {
	client *statsd.Client
}

//
// Initialize DatadogMetricsClient.
//
func (dmc *DatadogMetricsClient) Initialize(conf config.Config) error {
	statsd, err := statsd.New("127.0.0.1:8125")
	if err != nil {
		return errors.Errorf("unabled to initialize DatadogMetricsClient")
	}

	dmc.client = statsd

	return nil
}
