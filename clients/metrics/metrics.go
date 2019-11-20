package metrics

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"sync"
)

type Metric string

const (
	EngineEKSExecuteSuccess Metric = "engine.eks.execute.success"
	EngineEKSExecuteFailure Metric = "engine.eks.execute.failure"
	EngineEKSEnqueueSuccess Metric = "engine.eks.enqueue.success"
	EngineEKSEnqueueFailure Metric = "engine.eks.enqueue.failure"
)

type Client interface {
	Init(conf config.Config) error
	Decrement(name Metric, tags []string, rate float64)
	Increment(name Metric, tags []string, rate float64)
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
			instance := &DatadogStatsdMetricsClient{}

			if err = instance.Init(conf); err != nil {
				instance = nil
				break
			}
		default:
			err = fmt.Errorf("No Client named [%s] was found", name)
		}
	})

	return err
}

func Decrement(name Metric, tags []string, rate float64) {
	instance.Decrement(name, tags, rate)
}

func Increment(name Metric, tags []string, rate float64) {
	instance.Increment(name, tags, rate)
}
