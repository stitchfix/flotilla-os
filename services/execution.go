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

//
// ExecutionService interacts with the state manager and queue manager to queue runs, and perform
// CRUD operations on them
// * Acts as an intermediary layer between state and the execution engine
//
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
	UpdateStatus(runID string, status string, exitCode *int64, runExceptions *state.RunExceptions) error
	Terminate(runID string, userInfo state.UserInfo) error
	ReservedVariables() []string
	ListClusters() ([]string, error)
	GetEvents(run state.Run) (state.PodEventList, error)
	CreateTemplateRunByTemplateID(templateID string, req *state.TemplateExecutionRequest) (state.Run, error)
	CreateTemplateRunByTemplateName(templateName string, templateVersion string, req *state.TemplateExecutionRequest) (state.Run, error)
}

type executionService struct {
	stateManager             state.Manager
	k8sClusterClient         cluster.Client
	k8sExecutionEngine       engine.Engine
	reservedEnv              map[string]func(run state.Run) string
	k8sClusterOverride       []string
	k8sOverridePercent       int
	clusterOndemandWhitelist []string
	checkImageValidity       bool
	baseUri                  string
	spotReAttemptOverride    float32
	k8sSpotOverride          bool
	spotThresholdMinutes     float64
	terminateJobChannel      chan state.TerminateJob
}

func (es *executionService) GetEvents(run state.Run) (state.PodEventList, error) {
	return es.k8sExecutionEngine.GetEvents(run)
}

//
// NewExecutionService configures and returns an ExecutionService
//
func NewExecutionService(conf config.Config, k8sExecutionEngine engine.Engine, sm state.Manager, k8sClusterClient cluster.Client) (ExecutionService, error) {
	es := executionService{
		stateManager:       sm,
		k8sClusterClient:   k8sClusterClient,
		k8sExecutionEngine: k8sExecutionEngine,
	}
	//
	// Reserved environment variables dynamically generated
	// per run

	ownerKey := conf.GetString("owner_id_var")
	if len(ownerKey) == 0 {
		ownerKey = "FLOTILLA_RUN_OWNER_ID"
	}

	es.k8sClusterOverride = conf.GetStringSlice("k8s.cluster_override")
	es.k8sOverridePercent = conf.GetInt("k8s.cluster_override_percent")
	es.clusterOndemandWhitelist = conf.GetStringSlice("k8s.cluster_ondemand_whitelist")
	if conf.IsSet("check_image_validity") {
		es.checkImageValidity = conf.GetBool("check_image_validity")
	} else {
		es.checkImageValidity = true
	}

	if conf.IsSet("base_uri") {
		es.baseUri = conf.GetString("base_uri")
	}

	if conf.IsSet("k8s.spot_reattempt_override") {
		es.spotReAttemptOverride = float32(conf.GetFloat64("k8s.spot_reattempt_override"))
	} else {
		// defaults to 5% override.
		es.spotReAttemptOverride = float32(0.05)
	}

	if conf.IsSet("k8s.spot_override") {
		es.k8sSpotOverride = conf.GetBool("k8s.spot_override")
	} else {
		es.k8sSpotOverride = false
	}

	if conf.IsSet("k8s.spot_threshold_minutes") {
		es.spotThresholdMinutes = conf.GetFloat64("k8s.spot_threshold_minutes")
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
		"FLOTILLA_RUN_PAYLOAD_URI": func(run state.Run) string {
			return fmt.Sprintf("%s/api/v6/history/%s/payload", es.baseUri, run.RunID)
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

//
// Create constructs and queues a new Run on the cluster specified.
//
func (es *executionService) CreateDefinitionRunByDefinitionID(definitionID string, req *state.DefinitionExecutionRequest) (state.Run, error) {
	// Ensure definition exists
	definition, err := es.stateManager.GetDefinition(definitionID)
	if err != nil {
		return state.Run{}, err
	}

	return es.createFromDefinition(definition, req)
}

//
// Create constructs and queues a new Run on the cluster specified, based on an alias
//
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
	fields.ClusterName = es.k8sClusterOverride[rand.Intn(len(es.k8sClusterOverride))]
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
	run.GroupName = definition.GroupName

	return run, nil
}

func (es *executionService) constructBaseRunFromExecutable(executable state.Executable, req state.ExecutionRequest) (state.Run, error) {
	resources := executable.GetExecutableResources()
	fields := req.GetExecutionRequestCommon()
	var (
		run state.Run
		err error
	)

	if fields.Engine == nil {
		fields.Engine = &state.DefaultEngine
	}

	// Compute the executable command based on the execution request. If the
	// execution request did not specify an overriding command, use the computed
	// `executableCmd` as the Run's Command.
	executableCmd, err := executable.GetExecutableCommand(req)

	if err != nil {
		return run, err
	}

	if (fields.Command == nil || len(*fields.Command) == 0) && (len(executableCmd) > 0) {
		fields.Command = aws.String(executableCmd)
	}

	runID, err := state.NewRunID(fields.Engine)
	if err != nil {
		return run, err
	}

	executableID := executable.GetExecutableID()
	taskExecutionMinutes, _ := es.stateManager.GetTaskHistoricalRuntime(*executableID, runID)
	reAttemptRate, _ := es.stateManager.GetPodReAttemptRate()
	if reAttemptRate >= es.spotReAttemptOverride &&
		fields.Engine != nil &&
		fields.NodeLifecycle != nil &&
		*fields.Engine == state.K8SEngine &&
		*fields.NodeLifecycle == state.SpotLifecycle {
		fields.NodeLifecycle = &state.OndemandLifecycle
	}

	if taskExecutionMinutes > float32(es.spotThresholdMinutes) {
		fields.NodeLifecycle = &state.OndemandLifecycle
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
	return es.stateManager.ListRuns(limit, offset, sortField, sortOrder, filters, envFilters, []string{state.K8SEngine})
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
func (es *executionService) UpdateStatus(runID string, status string, exitCode *int64, runExceptions *state.RunExceptions) error {
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

	exitReason := es.extractExitReason(runExceptions)
	_, err = es.stateManager.UpdateRun(runID, state.Run{Status: status, ExitCode: exitCode, ExitReason: &exitReason, RunExceptions: runExceptions, FinishedAt: &finishedAt, StartedAt: startedAt})
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
			run.Engine = &state.K8SEngine
		}

		if *run.Engine == state.K8SEngine && run.Status != state.StatusStopped {
			err = es.k8sExecutionEngine.Terminate(run)
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
				break
			}
			break
		}
		break
	}
}

//
// Terminate stops the run with the given runID
//
func (es *executionService) Terminate(runID string, userInfo state.UserInfo) error {
	es.terminateJobChannel <- state.TerminateJob{RunID: runID, UserInfo: userInfo}
	go es.terminateWorker(es.terminateJobChannel)
	return nil
}

//
// ListClusters returns a list of all execution clusters available
//
func (es *executionService) ListClusters() ([]string, error) {
	return []string{}, nil
}

//
// sanitizeExecutionRequestCommonFields does what its name implies - sanitizes
func (es *executionService) sanitizeExecutionRequestCommonFields(fields *state.ExecutionRequestCommon) {
	fields.Engine = &state.K8SEngine

	if es.k8sSpotOverride {
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

//
// createAndEnqueueRun creates a run object in the DB, enqueues it, then
// updates the db's run object with a new `queued_at` field.
//
func (es *executionService) createAndEnqueueRun(run state.Run) (state.Run, error) {
	var err error
	// Save run to source of state - it is *CRITICAL* to do this
	// -before- queuing to avoid processing unsaved runs
	if err = es.stateManager.CreateRun(run); err != nil {
		return run, err
	}

	err = es.k8sExecutionEngine.Enqueue(run)
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

//
// Create constructs and queues a new Run on the cluster specified.
//
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
