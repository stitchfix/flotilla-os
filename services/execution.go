package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/stitchfix/flotilla-os/clients/cluster"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/execution/engine"
	"github.com/stitchfix/flotilla-os/state"
)

// ExecutionService interacts with the state manager and queue manager to queue runs, and perform
// CRUD operations on them
// * Acts as an intermediary layer between state and the execution engine
type ExecutionService interface {
	CreateDefinitionRunByDefinitionID(definitionID string, req *state.DefinitionExecutionRequest) (state.Run, error)
	CreateDefinitionRunByAlias(alias string, req *state.DefinitionExecutionRequest) (state.Run, error)
	List(
		limit int,
		offset int,
		sortOrder string,
		sortField string,
		filters map[string][]string,
		envFilters map[string]string) (state.RunList, error)
	Get(runID string) (state.Run, error)
	UpdateStatus(runID string, status string, exitCode *int64, runExceptions *state.RunExceptions, exitReason *string) error
	Terminate(runID string, userInfo state.UserInfo) error
	ReservedVariables() []string
	ListClusters() ([]string, error)
	GetEvents(run state.Run) (state.PodEventList, error)
	CreateTemplateRunByTemplateID(templateID string, req *state.TemplateExecutionRequest) (state.Run, error)
	CreateTemplateRunByTemplateName(templateName string, templateVersion string, req *state.TemplateExecutionRequest) (state.Run, error)
}

type executionService struct {
	stateManager          state.Manager
	eksClusterClient      cluster.Client
	eksExecutionEngine    engine.Engine
	emrExecutionEngine    engine.Engine
	reservedEnv           map[string]func(run state.Run) string
	eksClusterOverride    string
	eksGPUClusterOverride string
	checkImageValidity    bool
	baseUri               string
	spotReAttemptOverride float32
	eksSpotOverride       bool
	spotThresholdMinutes  float64
	terminateJobChannel   chan state.TerminateJob
}

func (es *executionService) GetEvents(run state.Run) (state.PodEventList, error) {
	return es.eksExecutionEngine.GetEvents(run)
}

// NewExecutionService configures and returns an ExecutionService
func NewExecutionService(conf config.Config, eksExecutionEngine engine.Engine, sm state.Manager, eksClusterClient cluster.Client, emrExecutionEngine engine.Engine) (ExecutionService, error) {
	es := executionService{
		stateManager:       sm,
		eksClusterClient:   eksClusterClient,
		eksExecutionEngine: eksExecutionEngine,
		emrExecutionEngine: emrExecutionEngine,
	}
	//
	// Reserved environment variables dynamically generated
	// per run

	ownerKey := conf.GetString("owner_id_var")
	if len(ownerKey) == 0 {
		ownerKey = "FLOTILLA_RUN_OWNER_ID"
	}

	es.eksClusterOverride = conf.GetString("eks_cluster_override")
	es.eksGPUClusterOverride = conf.GetString("eks_gpu_cluster_override")
	if conf.IsSet("check_image_validity") {
		es.checkImageValidity = conf.GetBool("check_image_validity")
	} else {
		es.checkImageValidity = true
	}

	if conf.IsSet("base_uri") {
		es.baseUri = conf.GetString("base_uri")
	}

	if conf.IsSet("eks_spot_reattempt_override") {
		es.spotReAttemptOverride = float32(conf.GetFloat64("eks_spot_reattempt_override"))
	} else {
		// defaults to 5% override.
		es.spotReAttemptOverride = float32(0.05)
	}

	if conf.IsSet("eks_spot_override") {
		es.eksSpotOverride = conf.GetBool("eks_spot_override")
	} else {
		es.eksSpotOverride = false
	}

	if conf.IsSet("eks_spot_threshold_minutes") {
		es.spotThresholdMinutes = conf.GetFloat64("eks_spot_threshold_minutes")
	} else {
		es.spotThresholdMinutes = 30.0
	}

	es.reservedEnv = map[string]func(run state.Run) string{
		"FLOTILLA_SERVER_MODE": func(run state.Run) string {
			return conf.GetString("flotilla_mode")
		},
		"FLOTILLA_RUN_ID": func(run state.Run) string {
			return run.RunID
		},
		"AWS_ROLE_SESSION_NAME": func(run state.Run) string {
			return run.RunID
		},
		ownerKey: func(run state.Run) string {
			return run.User
		},
	}

	es.terminateJobChannel = make(chan state.TerminateJob, 100)
	return &es, nil
}

// ReservedVariables returns the list of reserved run environment variable
// names
func (es *executionService) ReservedVariables() []string {
	var keys []string
	for k := range es.reservedEnv {
		keys = append(keys, k)
	}
	return keys
}

// Create constructs and queues a new Run on the cluster specified.
func (es *executionService) CreateDefinitionRunByDefinitionID(definitionID string, req *state.DefinitionExecutionRequest) (state.Run, error) {
	// Ensure definition exists
	definition, err := es.stateManager.GetDefinition(definitionID)
	if err != nil {
		return state.Run{}, err
	}

	return es.createFromDefinition(definition, req)
}

// Create constructs and queues a new Run on the cluster specified, based on an alias
func (es *executionService) CreateDefinitionRunByAlias(alias string, req *state.DefinitionExecutionRequest) (state.Run, error) {
	// Ensure definition exists
	definition, err := es.stateManager.GetDefinitionByAlias(alias)
	if err != nil {
		return state.Run{}, err
	}

	return es.createFromDefinition(definition, req)
}

func (es *executionService) createFromDefinition(definition state.Definition, req *state.DefinitionExecutionRequest) (state.Run, error) {
	var (
		run state.Run
		err error
	)
	fields := req.GetExecutionRequestCommon()
	rand.Seed(time.Now().Unix())
	fields.ClusterName = es.eksClusterOverride
	if fields.Gpu != nil && *fields.Gpu > 0 {
		fields.ClusterName = es.eksGPUClusterOverride
	}
	run.User = req.OwnerID
	es.sanitizeExecutionRequestCommonFields(fields)

	// Construct run object with StatusQueued and new UUID4 run id
	run, err = es.constructRunFromDefinition(definition, req)
	if err != nil {
		return run, err
	}

	return es.createAndEnqueueRun(run)
}

func (es *executionService) constructRunFromDefinition(definition state.Definition, req *state.DefinitionExecutionRequest) (state.Run, error) {
	run, err := es.constructBaseRunFromExecutable(definition, req)

	if err != nil {
		return run, err
	}

	run.DefinitionID = definition.DefinitionID
	run.Alias = definition.Alias
	queuedAt := time.Now()
	run.QueuedAt = &queuedAt
	run.GroupName = definition.GroupName
	if req.Description != nil {
		run.Description = req.Description
	}

	if req.IdempotenceKey != nil {
		run.IdempotenceKey = req.IdempotenceKey
	}

	if req.Arch != nil {
		run.Arch = req.Arch
	}

	if req.Labels != nil {
		run.Labels = *req.Labels
	}
	return run, nil
}

func (es *executionService) constructBaseRunFromExecutable(executable state.Executable, req state.ExecutionRequest) (state.Run, error) {
	resources := executable.GetExecutableResources()
	fields := req.GetExecutionRequestCommon()
	var (
		run state.Run
		err error
	)

	fields.Engine = req.GetExecutionRequestCommon().Engine

	// Compute the executable command based on the execution request. If the
	// execution request did not specify an overriding command, use the computed
	// `executableCmd` as the Run's Command.

	runID, err := state.NewRunID(fields.Engine)
	if err != nil {
		return run, err
	}

	if *fields.Engine == state.EKSEngine {
		executableCmd, err := executable.GetExecutableCommand(req)
		if err != nil {
			return run, err
		}

		if (fields.Command == nil || len(*fields.Command) == 0) && (len(executableCmd) > 0) {
			fields.Command = aws.String(executableCmd)
		}
		executableID := executable.GetExecutableID()

		taskExecutionMinutes, _ := es.stateManager.GetTaskHistoricalRuntime(*executableID, runID)
		reAttemptRate, _ := es.stateManager.GetPodReAttemptRate()
		if reAttemptRate >= es.spotReAttemptOverride &&
			fields.Engine != nil &&
			fields.NodeLifecycle != nil &&
			*fields.Engine == state.EKSEngine &&
			*fields.NodeLifecycle == state.SpotLifecycle {
			fields.NodeLifecycle = &state.OndemandLifecycle
		}

		if taskExecutionMinutes > float32(es.spotThresholdMinutes) {
			fields.NodeLifecycle = &state.OndemandLifecycle
		}
	}

	if *fields.Engine == state.EKSSparkEngine {
		if req.GetExecutionRequestCommon().SparkExtension == nil {
			return run, errors.New("spark_extension can't be nil, when using eks-spark engine type")
		}
		fields.SparkExtension = req.GetExecutionRequestCommon().SparkExtension
		reAttemptRate, _ := es.stateManager.GetPodReAttemptRate()
		if reAttemptRate >= es.spotReAttemptOverride {
			fields.NodeLifecycle = &state.OndemandLifecycle
		}
	}

	if fields.NodeLifecycle == nil {
		fields.NodeLifecycle = &state.SpotLifecycle
	}

	run = state.Run{
		RunID:                 runID,
		ClusterName:           fields.ClusterName,
		Image:                 resources.Image,
		Status:                state.StatusQueued,
		User:                  fields.OwnerID,
		Command:               fields.Command,
		Memory:                fields.Memory,
		Cpu:                   fields.Cpu,
		Gpu:                   fields.Gpu,
		Engine:                fields.Engine,
		NodeLifecycle:         fields.NodeLifecycle,
		EphemeralStorage:      fields.EphemeralStorage,
		ExecutableID:          executable.GetExecutableID(),
		ExecutableType:        executable.GetExecutableType(),
		ActiveDeadlineSeconds: fields.ActiveDeadlineSeconds,
		TaskType:              state.DefaultTaskType,
		SparkExtension:        fields.SparkExtension,
		CommandHash:           fields.CommandHash,
	}

	if fields.Labels != nil {
		run.Labels = *fields.Labels
	}

	runEnv := es.constructEnviron(run, fields.Env)
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

// List returns a list of Runs
// * validates definition_id and status filters
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
	return es.stateManager.ListRuns(limit, offset, sortField, sortOrder, filters, envFilters, []string{state.EKSEngine, state.EKSSparkEngine})
}

// Get returns the run with the given runID
func (es *executionService) Get(runID string) (state.Run, error) {
	return es.stateManager.GetRun(runID)
}

// UpdateStatus is for supporting some legacy runs that still manually update their status
func (es *executionService) UpdateStatus(runID string, status string, exitCode *int64, runExceptions *state.RunExceptions, exitReason *string) error {
	if !state.IsValidStatus(status) {
		return exceptions.MalformedInput{ErrorString: fmt.Sprintf("status %s is invalid", status)}
	}
	run, err := es.stateManager.GetRun(runID)
	if err != nil {
		return err
	}
	var startedAt *time.Time
	if run.StartedAt == nil {
		startedAt = run.QueuedAt
	} else {
		startedAt = run.StartedAt
	}
	finishedAt := time.Now()

	if exitReason == nil {
		extractedExitReason := es.extractExitReason(runExceptions)
		exitReason = &extractedExitReason
	}

	_, err = es.stateManager.UpdateRun(runID, state.Run{Status: status, ExitCode: exitCode, ExitReason: exitReason, RunExceptions: runExceptions, FinishedAt: &finishedAt, StartedAt: startedAt})
	return err
}

func (es *executionService) extractExitReason(runExceptions *state.RunExceptions) string {
	connectionError := regexp.MustCompile(`(?i).*(timeout|gatewayerror|socketerror|\s503\s|\s502\s|\s500\s|\s504\s|connectionerror).*`)
	pipError := regexp.MustCompile(`(?i).*(could\snot\sfind\sa\sversion|package\snot\sfound|ModuleNotFoundError|No\smatching\sdistribution\sfound).*`)
	yumError := regexp.MustCompile(`(?i).*(Nothing\sto\sdo).*`)
	gitError := regexp.MustCompile(`(?i).*(Could\snot\sread\sfrom\sremote\srepository|correct\saccess\srights|Repository\snot\sfound).*`)
	argumentError := regexp.MustCompile(`(?i).*(404|400|keyerror|column\smissing|RuntimeError).*`)
	syntaxError := regexp.MustCompile(`(?i).*(syntaxerror|typeerror|).*`)

	value, _ := json.Marshal(runExceptions)
	if value != nil {
		errorMsg := string(value)
		switch {
		case connectionError.MatchString(errorMsg):
			return "Connection error to downstream uri"
		case pipError.MatchString(errorMsg):
			return "Python pip package installation error"
		case yumError.MatchString(errorMsg):
			return "Yum installation error"
		case gitError.MatchString(errorMsg):
			return "Git clone error"
		case argumentError.MatchString(errorMsg):
			return "Data or argument error"
		case syntaxError.MatchString(errorMsg):
			return "Code or syntax error"
		default:
			return "Runtime exception encountered"
		}
	}
	return "Runtime exception encountered"
}

func (es *executionService) terminateWorker(jobChan <-chan state.TerminateJob) {
	for job := range jobChan {
		runID := job.RunID
		userInfo := job.UserInfo
		run, err := es.stateManager.GetRun(runID)
		if err != nil {
			break
		}

		subRuns, err := es.stateManager.ListRuns(1000, 0, "status", "desc", nil, map[string]string{"PARENT_FLOTILLA_RUN_ID": run.RunID}, state.Engines)
		if err == nil && subRuns.Total > 0 {
			for _, subRun := range subRuns.Runs {
				es.terminateJobChannel <- state.TerminateJob{
					RunID:    subRun.RunID,
					UserInfo: job.UserInfo,
				}
				go es.terminateWorker(es.terminateJobChannel)
			}

		}

		if run.Engine == nil {
			run.Engine = &state.EKSEngine
		}

		if run.Status != state.StatusStopped {
			if *run.Engine == state.EKSSparkEngine {
				err = es.emrExecutionEngine.Terminate(run)
			} else {
				err = es.eksExecutionEngine.Terminate(run)
			}
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
			break
		}
		break
	}
}

// Terminate stops the run with the given runID
func (es *executionService) Terminate(runID string, userInfo state.UserInfo) error {
	es.terminateJobChannel <- state.TerminateJob{RunID: runID, UserInfo: userInfo}
	go es.terminateWorker(es.terminateJobChannel)
	return nil
}

// ListClusters returns a list of all execution clusters available
func (es *executionService) ListClusters() ([]string, error) {
	return []string{}, nil
}

// sanitizeExecutionRequestCommonFields does what its name implies - sanitizes
func (es *executionService) sanitizeExecutionRequestCommonFields(fields *state.ExecutionRequestCommon) {
	if fields.Engine == nil {
		fields.Engine = &state.EKSEngine
	}

	if es.eksSpotOverride {
		fields.NodeLifecycle = &state.OndemandLifecycle
	}

	if fields.ActiveDeadlineSeconds == nil {
		if fields.NodeLifecycle == &state.OndemandLifecycle {
			fields.ActiveDeadlineSeconds = &state.OndemandActiveDeadlineSeconds
		} else {
			fields.ActiveDeadlineSeconds = &state.SpotActiveDeadlineSeconds
		}
	}
}

// createAndEnqueueRun creates a run object in the DB, enqueues it, then
// updates the db's run object with a new `queued_at` field.
func (es *executionService) createAndEnqueueRun(run state.Run) (state.Run, error) {
	var err error
	if run.IdempotenceKey != nil {
		priorRunId, err := es.stateManager.CheckIdempotenceKey(*run.IdempotenceKey)
		if err == nil && len(priorRunId) > 0 {
			priorRun, err := es.Get(priorRunId)
			if err == nil {
				return priorRun, nil
			}
		}
	}

	// Save run to source of state - it is *CRITICAL* to do this
	// -before- queuing to avoid processing unsaved runs
	if err = es.stateManager.CreateRun(run); err != nil {
		return run, err
	}

	if *run.Engine == state.EKSEngine {
		err = es.eksExecutionEngine.Enqueue(run)
	} else {
		err = es.emrExecutionEngine.Enqueue(run)
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
func (es *executionService) CreateTemplateRunByTemplateName(templateName string, templateVersion string, req *state.TemplateExecutionRequest) (state.Run, error) {
	version, err := strconv.Atoi(templateVersion)

	if err != nil {
		//use the "latest" template - version not a integer
		fetch, template, err := es.stateManager.GetLatestTemplateByTemplateName(templateName)
		if fetch && err == nil {
			return es.CreateTemplateRunByTemplateID(template.TemplateID, req)
		}
	} else {
		fetch, template, err := es.stateManager.GetTemplateByVersion(templateName, int64(version))
		if fetch && err == nil {
			return es.CreateTemplateRunByTemplateID(template.TemplateID, req)
		}
	}
	return state.Run{},
		errors.New(fmt.Sprintf("invalid template name or version, template_name: %s, template_version: %s", templateName, templateVersion))
}

// Create constructs and queues a new Run on the cluster specified.
func (es *executionService) CreateTemplateRunByTemplateID(templateID string, req *state.TemplateExecutionRequest) (state.Run, error) {
	// Ensure template exists
	template, err := es.stateManager.GetTemplateByID(templateID)
	if err != nil {
		return state.Run{}, err
	}

	return es.createFromTemplate(template, req)
}

func (es *executionService) createFromTemplate(template state.Template, req *state.TemplateExecutionRequest) (state.Run, error) {
	var (
		run state.Run
		err error
	)

	fields := req.GetExecutionRequestCommon()
	es.sanitizeExecutionRequestCommonFields(fields)

	// Construct run object with StatusQueued and new UUID4 run id
	run, err = es.constructRunFromTemplate(template, req)
	if err != nil {
		return run, err
	}
	if !req.DryRun {
		return es.createAndEnqueueRun(run)
	}
	return run, nil
}

func (es *executionService) constructRunFromTemplate(template state.Template, req *state.TemplateExecutionRequest) (state.Run, error) {
	run, err := es.constructBaseRunFromExecutable(template, req)

	if err != nil {
		return run, err
	}

	run.DefinitionID = template.TemplateID
	run.Alias = template.TemplateID
	run.GroupName = "template_group_name"
	run.ExecutionRequestCustom = req.GetExecutionRequestCustom()

	return run, nil
}
