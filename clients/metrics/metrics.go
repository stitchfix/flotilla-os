package metrics

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
)

type Client interface {
	Initialize(conf config.Config) error
}

//
// NewMetricsClient returns a cluster client
//
func NewMetricsClient(conf config.Config, name string) (Client, error) {
	switch name {
	case "datadog":
		m := &DatadogMetricsClient{}
		if err := m.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "DatadogMetricsClient")
		}
		return m, nil
	default:
		return nil, fmt.Errorf("No metrics client named [%s] was found", name)
	}
}
