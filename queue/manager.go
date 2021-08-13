package queue

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

//
// Manager wraps operations on a queue
//
type Manager interface {
	Name() string
	QurlFor(name string, prefixed bool) (string, error)
	Initialize(config.Config, string) error
	Enqueue(qURL string, run state.Run) error
	ReceiveRun(qURL string) (RunReceipt, error)
	ReceiveStatus(qURL string) (StatusReceipt, error)
	ReceiveCloudTrail(qURL string) (state.CloudTrailS3File, error)
	ReceiveKubernetesEvent(qURL string) (state.KubernetesEvent, error)
	ReceiveKubernetesRun(queue string) (string, error)
	List() ([]string, error)
}

//
// RunReceipt wraps a Run and a callback to use
// when Run is finished processing
//
type RunReceipt struct {
	Run  *state.Run
	Done func() error
}

//
// StatusReceipt wraps a StatusUpdate and a callback to use
// when StatusUpdate is finished applying
//
type StatusReceipt struct {
	StatusUpdate *string
	Done         func() error
}

//
// NewQueueManager returns the Manager configured via `queue_manager`
//
func NewQueueManager(conf config.Config, name string) (Manager, error) {
	switch name {
	case state.EKSEngine:
		sqsEKS := &SQSManager{}
		if err := sqsEKS.Initialize(conf, state.EKSEngine); err != nil {
			return nil, errors.Wrap(err, "problem initializing SQSManager")
		}
		return sqsEKS, nil
	case state.EKSSparkEngine:
		sqsEKSSpark := &SQSManager{}
		if err := sqsEKSSpark.Initialize(conf, state.EKSSparkEngine); err != nil {
			return nil, errors.Wrap(err, "problem initializing SQSManager")
		}
		return sqsEKSSpark, nil
	default:
		return nil, fmt.Errorf("no QueueManager named [%s] was found", name)
	}
}
