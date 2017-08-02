package services

import (
	"github.com/nu7hatch/gouuid"
	"github.com/stitchfix/flotilla-os/clients/cluster"
	"github.com/stitchfix/flotilla-os/clients/registry"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
)

//
// ExecutionService interacts with the state manager and queue manager to queue runs, and perform
// CRUD operations on them
//
type ExecutionService interface {
	Create(definitionID string, clusterName string, env *state.EnvList) (state.Run, error)
	List(limit int, offset int, sortOrder string, sortField string, filters map[string]string) (state.RunList, error)
	Get(runID string) (state.Run, error)
	Terminate(runID string) error
	ReservedVariables() []string
}

type executionService struct {
	sm          state.Manager
	qm          queue.Manager
	cc          cluster.Client
	rc          registry.Client
	reservedEnv map[string]func(run state.Run) string
}

//
// NewExecutionService configures and returns an ExecutionService
//
func NewExecutionService(conf config.Config, sm state.Manager,
	qm queue.Manager,
	cc cluster.Client,
	rc registry.Client) (ExecutionService, error) {
	es := executionService{
		sm: sm,
		qm: qm,
		cc: cc,
		rc: rc,
	}
	//
	// Reserved environment variables dynamically generated
	// per run
	es.reservedEnv = map[string]func(run state.Run) string{
		"FLOTILLA_SERVER_MODE": func(run state.Run) string {
			return conf.GetString("flotilla_mode")
		},
		"FLOTILLA_RUN_ID": func(run state.Run) string {
			return run.RunID
		},
	}
	return &es, nil
}

func (es *executionService) ReservedVariables() []string {
	var keys []string
	for k, _ := range es.reservedEnv {
		keys = append(keys, k)
	}
	return keys
}

func (es *executionService) Create(definitionID string, clusterName string, env *state.EnvList) (state.Run, error) {
	var (
		run state.Run
		err error
	)

	// Ensure definition exists
	definition, err := es.sm.GetDefinition(definitionID)
	if err != nil {
		return run, err
	}

	// Validate that definition can be run (image exists, cluster has resources)
	if err = es.canBeRun(clusterName, definition, env); err != nil {
		return run, err
	}

	// Construct run object with StatusQueued and new UUID4 run id
	run, err = es.constructRun(clusterName, definition, env)
	if err != nil {
		return run, err
	}

	// Save run to source of state - it is *CRITICAL* to do this
	// -before- queuing to avoid processing unsaved runs
	if err = es.sm.CreateRun(run); err != nil {
		return run, err
	}

	// Get qurl
	qurl, err := es.qm.QurlFor(run.ClusterName)
	if err != nil {
		return run, err
	}

	// Queue run
	return run, es.qm.Enqueue(qurl, run)
}

func (es *executionService) constructRun(
	clusterName string, definition state.Definition, env *state.EnvList) (state.Run, error) {

	var (
		run state.Run
		err error
	)

	runID, err := es.newUUIDv4()
	if err != nil {
		return run, err
	}

	run = state.Run{
		RunID:        runID,
		ClusterName:  clusterName,
		DefinitionID: definition.DefinitionID,
		Status:       state.StatusQueued,
	}
	runEnv := es.constructEnviron(run, env)
	run.Env = &runEnv
	return run, nil
}

func (es *executionService) constructEnviron(run state.Run, env *state.EnvList) state.EnvList {
	size := len(es.reservedEnv)
	if env != nil {
		size += len(*env)
	}
	runEnv := make([]state.EnvVar, size)
	i := 0
	for k, f := range es.reservedEnv {
		runEnv[i] = state.EnvVar{
			Name:  k,
			Value: f(run),
		}
		i++
	}
	if env != nil {
		for j, e := range *env {
			runEnv[i+j] = e
		}
	}
	return state.EnvList(runEnv)
}

func (es *executionService) newUUIDv4() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (es *executionService) canBeRun(clusterName string, definition state.Definition, env *state.EnvList) error {
	if env != nil {
		for _, e := range *env {
			_, usingRestricted := es.reservedEnv[e.Name]
			if usingRestricted {
				return exceptions.ErrorReservedEnvironmentVariable
			}
		}
	}

	ok, err := es.rc.IsImageValid(definition.Image)
	if err != nil {
		return err
	}
	if !ok {
		return exceptions.ErrorImageNotFound
	}

	ok, err = es.cc.CanBeRun(clusterName, definition)
	if err != nil {
		return err
	}
	if !ok {
		return exceptions.ErrorClusterConfigurationIssue
	}
	return nil
}

func (es *executionService) List(
	limit int, offset int, sortOrder string, sortField string, filters map[string]string) (state.RunList, error) {
	return state.RunList{}, nil
}

func (es *executionService) Get(runID string) (state.Run, error) {
	return state.Run{}, nil
}

func (es *executionService) Terminate(runID string) error {
	return nil
}
