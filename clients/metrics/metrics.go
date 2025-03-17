package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
)

type Metric string

const (
	// Metric associated to submission of jobs to EKS
	EngineEKSExecute Metric = "engine.eks.execute"
	// Metric associated to submission of jobs to SQS queue, before EKS submission.
	EngineEKSEnqueue Metric = "engine.eks.enqueue"
	// Metric associated to submission of jobs to EMR
	EngineEMRExecute Metric = "engine.emr.execute"
	// Metric associated to submission of jobs to SQS queue, before EMR submission.
	EngineEMREnqueue Metric = "engine.emr.enqueue"
	// Metric associated to termination of jobs via the API.
	EngineEKSTerminate Metric = "engine.eks.terminate"
	// Metric associated to termination of jobs via the API.
	EngineEMRTerminate Metric = "engine.emr.terminate"
	// Metric associated to termination of pods hopping between hosts.
	EngineEKSRunPodnameChange Metric = "engine.eks.run_podname_changed"
	// Metric associated to pod events where there was a Cluster Autoscale event.
	EngineEKSNodeTriggeredScaledUp Metric = "engine.eks.triggered_scale_up"
	// Timing for status worker processEKSRun
	StatusWorkerProcessEKSRun Metric = "status_worker.timing.process_eks_run"
	// Timing for acquire lock
	StatusWorkerAcquireLock Metric = "status_worker.timing.acquire_lock"
	// Timing for fetch_pod_metrics
	StatusWorkerFetchPodMetrics Metric = "status_worker.timing.fetch_pod_metrics"
	// Timing for fetch_update_status
	StatusWorkerFetchUpdateStatus Metric = "status_worker.timing.fetch_update_status"
	// Metric for locked runs
	StatusWorkerLockedRuns Metric = "status_worker.locked_runs"
	// Timing for fetch metrics
	StatusWorkerFetchMetrics Metric = "status_worker.fetch_metrics"
	// Timing for get pod list
	StatusWorkerGetPodList Metric = "status_worker.get_pod_list"
	// Timing for get events
	StatusWorkerGetEvents Metric = "status_worker.get_events"
	// Timing for get job
	StatusWorkerGetJob Metric = "status_worker.get_job"
	// Engine update run
	EngineUpdateRun Metric = "engine.update_run"
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
	Timing(name Metric, value time.Duration, tags []string, rate float64) error
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
		return fmt.Errorf("`metrics_client` not set in config, unable to instantiate metrics client")
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
			err = fmt.Errorf("no client named [%s] was found", name)
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

// Histogram tracks the statistical distribution of a set of values
func Histogram(name Metric, value float64, tags []string, rate float64) error {
	if instance != nil {
		return instance.Histogram(name, value, tags, rate)
	}

	return errors.Errorf("MetricsClient instance is nil, unable to send Histogram metric.")
}

// Distribution tracks the statistical distribution of a set of values
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

// Timing sends timing information, it is an alias for TimeInMilliseconds
func Timing(name Metric, value time.Duration, tags []string, rate float64) error {
	if instance != nil {
		return instance.Timing(name, value, tags, rate)
	}
	return errors.Errorf("MetricsClient instance is nil, unable to send Event metric.")
}
