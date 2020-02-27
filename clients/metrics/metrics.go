package metrics

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"sync"
)

type Metric string

const (
	// Metric associated to submission of jobs to EKS
	EngineEKSExecute Metric = "engine.eks.execute"
	// Metric associated to submission of jobs to SQS queue, before EKS submission.
	EngineEKSEnqueue Metric = "engine.eks.enqueue"
	// Metric associated to termination of jobs via the API.
	EngineEKSTerminate Metric = "engine.eks.terminate"
	// Metric associated to termination of pods hopping between hosts.
	EngineEKSRunPodnameChange Metric = "engine.eks.run_podname_changed"
	// Metric associated to pod events where there was a Cluster Autoscale event.
	EngineEKSNodeTriggeredScaledUp Metric = "engine.eks.triggered_scale_up"
)

type MetricTag string

const (
	// Metric tag for job success.
	StatusSuccess MetricTag = "status:success"
	// Metric tag for job failure.
	StatusFailure MetricTag = "status:failure"
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

// Instantiating the Metrics Client.
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

// Decr is just Count of -1
func Decrement(name Metric, tags []string, rate float64) error {
	if instance != nil {
		return instance.Decrement(name, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Decrement metric.")
}

// Incr is just Count of -1
func Increment(name Metric, tags []string, rate float64) error {
	if instance != nil {
		return instance.Increment(name, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Increment metric.")
}

//
// Histogram tracks the statistical distribution of a set of values
//
func Histogram(name Metric, value float64, tags []string, rate float64) error {
	if instance != nil {
		return instance.Histogram(name, value, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Histogram metric.")
}

//
// Distribution tracks the statistical distribution of a set of values
//
func Distribution(name Metric, value float64, tags []string, rate float64) error {
	if instance != nil {
		return instance.Distribution(name, value, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Distribution metric.")
}

// Set counts the number of unique elements in a group
func Set(name Metric, value string, tags []string, rate float64) error {
	if instance != nil {
		return instance.Set(name, value, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Set metric.")
}

// NewEvent creates a new event with the given title and text.
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
