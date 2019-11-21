package metrics

import (
	"fmt"
	"github.com/pkg/errors"
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
	Decrement(name Metric, tags []string, rate float64) error
	Increment(name Metric, tags []string, rate float64) error
	Histogram(name Metric, value float64, tags []string, rate float64) error
	Distribution(name Metric, value float64, tags []string, rate float64) error
	Set(name Metric, value string, tags []string, rate float64) error
	Event(evt event) error
}

type event struct {
	Title string
	Text  string
	Tags  []string
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
			instance = &DatadogStatsdMetricsClient{}

			if err = instance.Init(conf); err != nil {
				err = errors.Errorf("Unable to initialize dogstatsd client.")
				instance = nil
				break
			}
		default:
			err = fmt.Errorf("No Client named [%s] was found", name)
		}
	})

	return err
}

func Decrement(name Metric, tags []string, rate float64) error {
	if instance != nil {
		return instance.Decrement(name, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Decrement metric.")
}

func Increment(name Metric, tags []string, rate float64) error {
	if instance != nil {
		return instance.Increment(name, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Increment metric.")
}

func Histogram(name Metric, value float64, tags []string, rate float64) error {
	if instance != nil {
		return instance.Histogram(name, value, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Histogram metric.")
}

func Distribution(name Metric, value float64, tags []string, rate float64) error {
	if instance != nil {
		return instance.Distribution(name, value, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Distribution metric.")
}

func Set(name Metric, value string, tags []string, rate float64) error {
	if instance != nil {
		return instance.Set(name, value, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Set metric.")
}

func Event(title string, text string, tags []string) error {
	if instance != nil {
		return instance.Event(event{
			Title: title,
			Text:  text,
			Tags:  tags,
		})
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Event metric.")
}
