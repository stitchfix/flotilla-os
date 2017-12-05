package engine

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"testing"
)

type mockQueueManager struct {
	statusUpdates []string
}

func (mqm *mockQueueManager) Name() string {
	return "mock"
}

func (mqm *mockQueueManager) QurlFor(name string, prefixed bool) (string, error) {
	return "", nil
}

func (mqm *mockQueueManager) Initialize(config.Config) error {
	return nil
}

func (mqm *mockQueueManager) Enqueue(qURL string, run state.Run) error {
	return nil
}

func (mqm *mockQueueManager) ReceiveRun(qURL string) (queue.RunReceipt, error) {
	return queue.RunReceipt{}, nil
}

func (mqm *mockQueueManager) ReceiveStatus(qURL string) (queue.StatusReceipt, error) {
	popped := mqm.statusUpdates[0]
	mqm.statusUpdates = mqm.statusUpdates[1:]

	return queue.StatusReceipt{StatusUpdate: popped}, nil
}

func (mqm *mockQueueManager) List() ([]string, error) {
	return nil, nil
}

func setUp(t *testing.T) ECSExecutionEngine {
	confDir := "../../conf"
	conf, _ := config.NewConfig(&confDir)

	eng := ECSExecutionEngine{}
	eng.Initialize(conf)

	eng.qm = &mockQueueManager{
		statusUpdates: []string{
			`{"version":"0","id":"cupcke","detail-type":"ECS Task State Change","source":"aws.ecs","account":"acctid","time":"2017-12-02T15:06:37Z","region":"us-east-1","resources":["taskarn"],"detail":{"clusterArn":"clusterarn","containerInstanceArn":"ciarn","containers":[{"containerArn":"containerarn","exitCode":0,"lastStatus":"STOPPED","name":"name","taskArn":"arn1"}],"createdAt":"2017-12-02T15:00:06.261Z","desiredStatus":"STOPPED","group":"grp","lastStatus":"STOPPED","overrides":{"containerOverrides":[{"environment":[{"name":"FLOTILLA_SERVER_MODE","value":"prod"},{"name":"FLOTILLA_RUN_ID","value":"runid1"}],"name":"name"}]},"startedAt":"2017-12-02T15:00:08.922Z","startedBy":"starterthing","stoppedAt":"2017-12-02T15:06:37.429Z","stoppedReason":"Essential container in task exited","updatedAt":"2017-12-02T15:06:37.429Z","taskArn":"arn1","taskDefinitionArn":"def1","version":3}}`,
		},
	}

	return eng
}

func TestECSExecutionEngine_PollStatus(t *testing.T) {
	eng := setUp(t)

	r, err := eng.PollStatus()
	if err != nil {
		t.Error(err)
	}

	run := r.Run
	if run.Status != state.StatusStopped {
		t.Errorf("Expected status: %s but wa %s", state.StatusStopped, run.Status)
	}

	if run.ExitCode == nil {
		t.Errorf("Expected non nil exit code")
	}

	fmt.Println(r.Run)
}
