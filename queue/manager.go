package queue

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

//
// Manager wraps operations on a queue
//
type Manager interface {
	Name() string
	QurlFor(name string, prefixed bool) (string, error)
	Initialize(config.Config) error
	Enqueue(qURL string, run state.Run) error
	ReceiveRun(qURL string) (RunReceipt, error)
	ReceiveStatus(qURL string) (StatusReceipt, error)
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
	StatusUpdate string
	Done         func() error
}

//
// NewQueueManager returns the Manager configured via `queue_manager`
//
func NewQueueManager(conf config.Config) (Manager, error) {
	name := conf.GetString("queue_manager")
	if len(name) == 0 {
		name = "sqs"
	}

	switch name {
	case "sqs":
		sqsm := &SQSManager{}
		return sqsm, sqsm.Initialize(conf)
	default:
		return nil, fmt.Errorf("No QueueManager named [%s] was found", name)
	}
}
