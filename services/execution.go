package services

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stitchfix/flotilla-os/log"
	"math/rand"
	"text/template"
	"time"

	"github.com/stitchfix/flotilla-os/clients/cluster"
	"github.com/stitchfix/flotilla-os/clients/registry"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/execution/engine"
	"github.com/stitchfix/flotilla-os/state"
)

//
// ExecutionService interacts with the state manager and queue manager to queue runs, and perform
// CRUD operations on them
// * Acts as an intermediary layer between state and the execution engine
//
type ExecutionService interface {
	Create(definitionID string, clusterName string, env *state.EnvList, ownerID string, command *string, memory *int64, cpu *int64, engine *string, ephemeralStorage *int64, nodeLifecycle *string) (state.Run, error)
	CreateByAlias(alias string, clusterName string, env *state.EnvList, ownerID string, command *string, memory *int64, cpu *int64, engine *string, ephemeralStorage *int64, nodeLifecycle *string) (state.Run, error)
	List(
		limit int,
		offset int,
		sortOrder string,
		sortField string,
		filters map[string][]string,
		envFilters map[string]string) (state.RunList, error)
	Get(runID string) (state.Run, error)
	UpdateStatus(runID string, status string, exitCode *int64) error
	Terminate(runID string, userInfo state.UserInfo) error
	ReservedVariables() []string
	ListClusters() ([]string, error)
	GetEvents(run state.Run) (state.PodEventList, error)
}

type executionService struct {
	stateManager             state.Manager
	ecsClusterClient         cluster.Client
	eksClusterClient         cluster.Client
	registryClient           registry.Client
	ecsExecutionEngine       engine.Engine
	eksExecutionEngine       engine.Engine
	reservedEnv              map[string]func(run state.Run) string
	eksClusterOverride       string
	eksOverridePercent       int
	clusterOndemandWhitelist []string
	checkImageValidity       bool
}

func (es *executionService) GetEvents(run state.Run) (state.PodEventList, error) {
	return es.eksExecutionEngine.GetEvents(run)
}

//
// NewExecutionService configures and returns an ExecutionService
//
func NewExecutionService(conf config.Config,
	ecsExecutionEngine engine.Engine,
	eksExecutionEngine engine.Engine,
	sm state.Manager,
	ecsClusterClient cluster.Client,
	eksClusterClient cluster.Client,
	rc registry.Client,
	log log.Logger) (ExecutionService, error) {
	es := executionService{
		stateManager:       sm,
		ecsClusterClient:   ecsClusterClient,
		eksClusterClient:   eksClusterClient,
		registryClient:     rc,
		ecsExecutionEngine: ecsExecutionEngine,
		eksExecutionEngine: eksExecutionEngine,
	}
	//
	// Reserved environment variables dynamically generated
	// per run

	ownerKey := conf.GetString("owner_id_var")
	if len(ownerKey) == 0 {
		ownerKey = "FLOTILLA_RUN_OWNER_ID"
	}

	es.eksClusterOverride = conf.GetString("eks.cluster_override")
	es.eksOverridePercent = conf.GetInt("eks.cluster_override_percent")
	es.clusterOndemandWhitelist = conf.GetStringSlice("eks.cluster_ondemand_whitelist")
	if conf.IsSet("check_image_validity") {
		es.checkImageValidity = conf.GetBool("check_image_validity")
	} else {
		es.checkImageValidity = true
	}
	es.reservedEnv = map[string]func(run state.Run) string{
		"FLOTILLA_SERVER_MODE": func(run state.Run) string {
			return conf.GetString("flotilla_mode")
		},
		"FLOTILLA_RUN_ID": func(run state.Run) string {
			return run.RunID
		},
		ownerKey: func(run state.Run) string {
			return run.User
		},
	}
	// Warm cached cluster list
	_, _ = es.ecsClusterClient.ListClusters()
	return &es, nil
}

//
// ReservedVariables returns the list of reserved run environment variable
// names
//
func (es *executionService) ReservedVariables() []string {
	var keys []string
	for k := range es.reservedEnv {
		keys = append(keys, k)
	}
	return keys
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//
// Create constructs and queues a new Run on the cluster specified
//
func (es *executionService) Create(definitionID string, clusterName string, env *state.EnvList, ownerID string, command *string, memory *int64, cpu *int64, engine *string, ephemeralStorage *int64, nodeLifecycle *string) (state.Run, error) {

	// Ensure definition exists
	definition, err := es.stateManager.GetDefinition(definitionID)
	if err != nil {
		return state.Run{}, err
	}

	if engine == nil {
		engine = &state.DefaultEngine
	}

	// Handle the case the cluster name was of type EKS but engine was not set to EKS.
	if clusterName == es.eksClusterOverride {
		engine = &state.EKSEngine
	}

	if *engine == state.EKSEngine {
		clusterName = es.eksClusterOverride
	}

	// Added to facilitate migration of ECS jobs to EKS.
	if engine != &state.EKSEngine && es.eksOverridePercent > 0 && *definition.Privileged == false {
		modulo := 100 / es.eksOverridePercent
		if rand.Int()%modulo == 0 {
			engine = &state.EKSEngine
			if contains(es.clusterOndemandWhitelist, clusterName) {
				nodeLifecycle = &state.OndemandLifecycle
			} else {
				nodeLifecycle = &state.SpotLifecycle
			}
			clusterName = es.eksClusterOverride
		}
	}

	return es.createFromDefinition(definition, clusterName, env, ownerID, command, memory, cpu, engine, nodeLifecycle, ephemeralStorage)
}

//
// Create constructs and queues a new Run on the cluster specified, based on an alias
//
func (es *executionService) CreateByAlias(alias string, clusterName string, env *state.EnvList, ownerID string, command *string, memory *int64, cpu *int64, engine *string, ephemeralStorage *int64, nodeLifecycle *string) (state.Run, error) {

	// Ensure definition exists
	definition, err := es.stateManager.GetDefinitionByAlias(alias)
	if err != nil {
		return state.Run{}, err
	}

	if engine == nil {
		engine = &state.DefaultEngine
	}

	// Handle the case the cluster name was of type EKS but engine was not set to EKS.
	if clusterName == es.eksClusterOverride {
		engine = &state.EKSEngine
	}

	if *engine == state.EKSEngine {
		clusterName = es.eksClusterOverride
	}

	// Added to facilitate migration of ECS jobs to EKS.
	if engine != &state.EKSEngine && es.eksOverridePercent > 0 && *definition.Privileged == false {
		modulo := 100 / es.eksOverridePercent
		if rand.Int()%modulo == 0 {
			engine = &state.EKSEngine
			if contains(es.clusterOndemandWhitelist, clusterName) {
				nodeLifecycle = &state.OndemandLifecycle
			} else {
				nodeLifecycle = &state.SpotLifecycle
			}
			clusterName = es.eksClusterOverride
		}
	}
	return es.createFromDefinition(definition, clusterName, env, ownerID, command, memory, cpu, engine, nodeLifecycle, ephemeralStorage)
}

func (es *executionService) generateTaskTypeCommand(definition state.Definition) (string, error) {
	var taskType state.DefinitionTemplate
	taskType, err := es.stateManager.GetDefinitionTemplateByID(definition.TemplateID)

	if err != nil {
		return "", err
	}

	var CommandTemplate, _ = template.New("command").Parse(taskType.Template)
	var result bytes.Buffer
	if err := CommandTemplate.Execute(&result, definition.TemplatePayload); err != nil {
		return "", err
	}
	return result.String(), nil
}

func (es *executionService) createFromDefinition(definition state.Definition, clusterName string, env *state.EnvList, ownerID string, command *string, memory *int64, cpu *int64, engine *string, nodeLifecycle *string, ephemeralStorage *int64) (state.Run, error) {
	var (
		run state.Run
		err error
	)

	// Validate that definition can be run (image exists, cluster has resources)
	if err = es.canBeRun(clusterName, definition, env, *engine); err != nil {
		return run, err
	}

	if len(definition.TemplateID) > 0 {
		ptr, err := es.generateTaskTypeCommand(definition)
		command = &ptr
		if err != nil {
			return run, err
		}
	}

	// Construct run object with StatusQueued and new UUID4 run id
	run, err = es.constructRun(clusterName, definition, env, ownerID, command, memory, cpu, engine, nodeLifecycle, ephemeralStorage)
	if err != nil {
		return run, err
	}

	// Save run to source of state - it is *CRITICAL* to do this
	// -before- queuing to avoid processing unsaved runs
	if err = es.stateManager.CreateRun(run); err != nil {
		return run, err
	}

	// ECS Queue run
	if *engine == state.ECSEngine {
		err = es.ecsExecutionEngine.Enqueue(run)
	}
	if *engine == state.EKSEngine {
		err = es.eksExecutionEngine.Enqueue(run)
	}

	queuedAt := time.Now()

	if err != nil {
		return run, err
	}

	// UpdateStatus the run's QueuedAt field
	if run, err = es.stateManager.UpdateRun(run.RunID, state.Run{QueuedAt: &queuedAt}); err != nil {
		return run, err
	}

	return run, nil
}

func (es *executionService) constructRun(clusterName string, definition state.Definition, env *state.EnvList, ownerID string, command *string, memory *int64, cpu *int64, engine *string, nodeLifecycle *string, ephemeralStorage *int64) (state.Run, error) {

	var (
		run state.Run
		err error
	)

	if engine == nil {
		engine = &state.DefaultEngine
	}

	runID, err := state.NewRunID(engine)
	if err != nil {
		return run, err
	}

	run = state.Run{
		RunID:            runID,
		ClusterName:      clusterName,
		GroupName:        definition.GroupName,
		DefinitionID:     definition.DefinitionID,
		Alias:            definition.Alias,
		Image:            definition.Image,
		Status:           state.StatusQueued,
		User:             ownerID,
		Command:          command,
		Memory:           memory,
		Cpu:              cpu,
		Gpu:              definition.Gpu,
		Engine:           engine,
		NodeLifecycle:    nodeLifecycle,
		EphemeralStorage: ephemeralStorage,
	}

	if len(definition.TemplateID) > 0 {
		run.TemplateID = definition.TemplateID
		run.TemplatePayload = definition.TemplatePayload
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

func (es *executionService) canBeRun(clusterName string, definition state.Definition, env *state.EnvList, engine string) error {
	if env != nil {
		for _, e := range *env {
			_, usingRestricted := es.reservedEnv[e.Name]
			if usingRestricted {
				return exceptions.ConflictingResource{
					ErrorString: fmt.Sprintf("environment variable %s is reserved", e.Name)}
			}
		}
	}
	var ok bool
	var err error
	if es.checkImageValidity {
		ok, err = es.registryClient.IsImageValid(definition.Image)
		if err != nil {
			return err
		}
		if !ok {
			return exceptions.MissingResource{
				ErrorString: fmt.Sprintf(
					"image [%s] was not found in any of the configured repositories", definition.Image)}
		}
	}

	if engine == state.ECSEngine {
		ok, err = es.ecsClusterClient.CanBeRun(clusterName, definition)
	}
	if engine == state.EKSEngine {
		if *definition.Privileged == true {
			ok, err = false, errors.New("eks cannot run containers with privileged mode")
		} else {
			ok, err = true, nil
		}
	}

	if err != nil {
		return err
	}
	if !ok {
		return exceptions.MalformedInput{
			ErrorString: fmt.Sprintf(
				"definition [%s] cannot be run on cluster [%s]", definition.DefinitionID, clusterName)}
	}
	return nil
}

//
// List returns a list of Runs
// * validates definition_id and status filters
//
func (es *executionService) List(
	limit int,
	offset int,
	sortOrder string,
	sortField string,
	filters map[string][]string,
	envFilters map[string]string) (state.RunList, error) {

	// If definition_id is present in filters, validate its
	// existence first
	definitionID, ok := filters["definition_id"]
	if ok {
		_, err := es.stateManager.GetDefinition(definitionID[0])
		if err != nil {
			return state.RunList{}, err
		}
	}

	if statusFilters, ok := filters["status"]; ok {
		for _, status := range statusFilters {
			if !state.IsValidStatus(status) {
				// Status filter is invalid
				err := exceptions.MalformedInput{
					ErrorString: fmt.Sprintf("invalid status [%s]", status)}
				return state.RunList{}, err
			}
		}
	}
	return es.stateManager.ListRuns(limit, offset, sortField, sortOrder, filters, envFilters, []string{state.ECSEngine, state.EKSEngine})
}

//
// Get returns the run with the given runID
//
func (es *executionService) Get(runID string) (state.Run, error) {
	return es.stateManager.GetRun(runID)
}

//
// UpdateStatus is for supporting some legacy runs that still manually update their status
//
func (es *executionService) UpdateStatus(runID string, status string, exitCode *int64) error {
	if !state.IsValidStatus(status) {
		return exceptions.MalformedInput{ErrorString: fmt.Sprintf("status %s is invalid", status)}
	}
	_, err := es.stateManager.UpdateRun(runID, state.Run{Status: status, ExitCode: exitCode})
	return err
}

//
// Terminate stops the run with the given runID
//
func (es *executionService) Terminate(runID string, userInfo state.UserInfo) error {
	run, err := es.stateManager.GetRun(runID)
	if err != nil {
		return err
	}

	if run.Engine == nil {
		run.Engine = &state.ECSEngine
	}

	// If it's queued and not submitted, set status to stopped (checked by submit worker)
	if run.Status == state.StatusQueued && *run.Engine == state.ECSEngine {
		_, err = es.stateManager.UpdateRun(runID, state.Run{Status: state.StatusStopped})
		return err
	}

	if *run.Engine == state.ECSEngine {
		// If it's been submitted, let the status update workers handle setting it to stopped
		if run.Status != state.StatusStopped && len(run.TaskArn) > 0 && len(run.ClusterName) > 0 {
			return es.ecsExecutionEngine.Terminate(run)
		}
	}

	if *run.Engine == state.EKSEngine && run.Status != state.StatusStopped {
		err = es.eksExecutionEngine.Terminate(run)
		if err == nil || run.Status == state.StatusQueued {
			exitReason := "Task terminated by user"
			if len(userInfo.Email) > 0 {
				exitReason = fmt.Sprintf("Task terminated by - %s", userInfo.Email)
			}

			exitCode := int64(1)
			finishedAt := time.Now()
			_, err = es.stateManager.UpdateRun(run.RunID, state.Run{
				Status:     state.StatusStopped,
				ExitReason: &exitReason,
				ExitCode:   &exitCode,
				FinishedAt: &finishedAt,
			})
			return err
		}
		return nil
	}

	return exceptions.MalformedInput{
		ErrorString: fmt.Sprintf(
			"invalid run, state: %s, arn: %s, clusterName: %s", run.Status, run.TaskArn, run.ClusterName)}
}

//
// ListClusters returns a list of all execution clusters available
//
func (es *executionService) ListClusters() ([]string, error) {
	return es.ecsClusterClient.ListClusters()
}
