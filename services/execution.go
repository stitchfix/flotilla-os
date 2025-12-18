package services

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/stitchfix/flotilla-os/utils"

	"github.com/aws/aws-sdk-go/aws"

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
	CreateDefinitionRunByDefinitionID(ctx context.Context, definitionID string, req *state.DefinitionExecutionRequest) (state.Run, error)
	CreateDefinitionRunByAlias(ctx context.Context, alias string, req *state.DefinitionExecutionRequest) (state.Run, error)
	List(
		ctx context.Context,
		limit int,
		offset int,
		sortOrder string,
		sortField string,
		filters map[string][]string,
		envFilters map[string]string) (state.RunList, error)
	Get(ctx context.Context, runID string) (state.Run, error)
	UpdateStatus(ctx context.Context, runID string, status string, exitCode *int64, runExceptions *state.RunExceptions, exitReason *string) error
	Terminate(ctx context.Context, runID string, userInfo state.UserInfo) error
	ReservedVariables() []string
	ListClusters(ctx context.Context) ([]state.ClusterMetadata, error)
	GetDefaultCluster() string
	GetEvents(ctx context.Context, run state.Run) (state.PodEventList, error)
	CreateTemplateRunByTemplateID(ctx context.Context, templateID string, req *state.TemplateExecutionRequest) (state.Run, error)
	CreateTemplateRunByTemplateName(ctx context.Context, templateName string, templateVersion string, req *state.TemplateExecutionRequest) (state.Run, error)
	UpdateClusterMetadata(ctx context.Context, cluster state.ClusterMetadata) error
	DeleteClusterMetadata(ctx context.Context, clusterID string) error
	GetClusterByID(ctx context.Context, clusterID string) (state.ClusterMetadata, error)
	GetRunStatus(ctx context.Context, runID string) (state.RunStatus, error)
}

type executionService struct {
	stateManager          state.Manager
	eksClusterClient      cluster.Client
	eksExecutionEngine    engine.Engine
	emrExecutionEngine    engine.Engine
	reservedEnv           map[string]func(run state.Run) string
	eksClusterOverride    string
	eksClusterDefault     string
	eksTierDefault        string
	eksGPUClusterOverride string
	eksGPUClusterDefault  string
	checkImageValidity    bool
	baseUri               string
	spotReAttemptOverride float32
	eksSpotOverride       bool
	spotThresholdMinutes  float64
	terminateJobChannel   chan state.TerminateJob
	validEksClusters      []string
	//validEksClusterTiers  string
}

func (es *executionService) GetEvents(ctx context.Context, run state.Run) (state.PodEventList, error) {
	ctx, span := utils.TraceJob(ctx, "flotilla.get_events", run.RunID)
	defer span.Finish()
	utils.TagJobRun(span, run)
	return es.eksExecutionEngine.GetEvents(ctx, run)
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

	es.validEksClusters = strings.Split(conf.GetString("eks_clusters"), ",")
	for k, _ := range es.validEksClusters {
		es.validEksClusters[k] = strings.TrimSpace(es.validEksClusters[k])
	}
	es.eksClusterOverride = conf.GetString("eks_cluster_override")
	es.eksGPUClusterOverride = conf.GetString("eks_gpu_cluster_override")
	es.eksClusterDefault = conf.GetString("eks_cluster_default")
	es.eksGPUClusterDefault = conf.GetString("eks_gpu_cluster_default")
	es.eksTierDefault = conf.GetString("eks_tier_default")
	//es.validEksClusterTiers = conf.GetString("eks_cluster_tiers")

	if !slices.Contains(es.validEksClusters, es.eksClusterDefault) || !slices.Contains(es.validEksClusters, es.eksGPUClusterDefault) {
		return nil, fmt.Errorf("an invalid cluster has been set as a default\nvalid_clusters:%s\neks_cluster_default:%s\neks_gpu_cluster_default:%s", es.validEksClusters, es.eksClusterDefault, es.eksGPUClusterDefault)
	}

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
func (es *executionService) CreateDefinitionRunByDefinitionID(ctx context.Context, definitionID string, req *state.DefinitionExecutionRequest) (state.Run, error) {
	ctx, span := utils.TraceJob(ctx, "flotilla.definition.create_run", "")
	defer span.Finish()
	span.SetTag("definition_id", definitionID)

	// Ensure definition exists
	definition, err := es.stateManager.GetDefinition(ctx, definitionID)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return state.Run{}, err
	}
	return es.createFromDefinition(ctx, definition, req)
}

// Create constructs and queues a new Run on the cluster specified, based on an alias
func (es *executionService) CreateDefinitionRunByAlias(ctx context.Context, alias string, req *state.DefinitionExecutionRequest) (state.Run, error) {
	ctx, span := utils.TraceJob(ctx, "flotilla.alias.create_run", "")
	defer span.Finish()
	span.SetTag("alias", alias)

	// Ensure definition exists
	definition, err := es.stateManager.GetDefinitionByAlias(ctx, alias)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return state.Run{}, err
	}

	return es.createFromDefinition(ctx, definition, req)
}

func (es *executionService) createFromDefinition(ctx context.Context, definition state.Definition, req *state.DefinitionExecutionRequest) (state.Run, error) {
	var (
		run state.Run
		err error
	)
	ctx, span := utils.TraceJob(ctx, "flotilla.definition.create_run", run.RunID)
	defer span.Finish()

	fields := req.GetExecutionRequestCommon()
	rand.Seed(time.Now().Unix())

	/*
		cluster is set based on the following precedence (low to high):
			1. Cluster is passed in from request
			2. Cluster from cluster metadata and active
			3. Cluster from task definition
			3. Default cluster from config

		cluster is then checked for validity.

		if required, cluster overrides should be introduced and set here
	*/
	clusterMetadata, err := es.ListClusters(ctx)
	var activeClusters []string
	if len(clusterMetadata) > 0 {
		for _, cluster := range clusterMetadata {
			if cluster.Status == state.StatusActive {
				if es.clusterSupportsTier(cluster, req.Tier) {
					activeClusters = append(activeClusters, cluster.Name)
				}
			}
		}
	}

	if req.ClusterName != "" {
		fields.ClusterName = req.ClusterName
	} else if len(activeClusters) > 0 {
		fields.ClusterName = activeClusters[rand.Intn(len(activeClusters))]
	} else if definition.TargetCluster != "" {
		fields.ClusterName = definition.TargetCluster
	} else if fields.Gpu != nil && *fields.Gpu > 0 {
		fields.ClusterName = es.eksGPUClusterDefault
	} else {
		fields.ClusterName = es.eksClusterDefault
	}

	for _, c := range clusterMetadata {
		es.validEksClusters = append(es.validEksClusters, c.Name)
	}
	if !es.isClusterValid(fields.ClusterName) {
		return run, fmt.Errorf("%s was not found in the list of valid clusters: %s", fields.ClusterName, es.validEksClusters)
	}
	span.SetTag("clusterName", fields.ClusterName)
	run.User = req.OwnerID
	es.sanitizeExecutionRequestCommonFields(fields)
	// Construct run object with StatusQueued and new UUID4 run id
	run, err = es.constructRunFromDefinition(ctx, definition, req)
	if err != nil {
		return run, err
	}
	return es.createAndEnqueueRun(ctx, run)
}

func (es *executionService) constructRunFromDefinition(ctx context.Context, definition state.Definition, req *state.DefinitionExecutionRequest) (state.Run, error) {
	run, err := es.constructBaseRunFromExecutable(ctx, definition, req)

	if err != nil {
		return run, err
	}

	run.DefinitionID = definition.DefinitionID
	run.Alias = definition.Alias
	queuedAt := time.Now()
	run.QueuedAt = &queuedAt
	run.GroupName = definition.GroupName
	run.RequiresDocker = definition.RequiresDocker

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

func (es *executionService) constructBaseRunFromExecutable(ctx context.Context, executable state.Executable, req state.ExecutionRequest) (state.Run, error) {
	resources := executable.GetExecutableResources()
	fields := req.GetExecutionRequestCommon()
	var (
		run state.Run
		err error
	)

	fields.Engine = req.GetExecutionRequestCommon().Engine
	fields.Tier = es.resolveRequestTier(req.GetExecutionRequestCommon().Tier)
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

		taskExecutionMinutes, _ := es.stateManager.GetTaskHistoricalRuntime(ctx, *executableID, runID)
		reAttemptRate, _ := es.stateManager.GetPodReAttemptRate(ctx)
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
		reAttemptRate, _ := es.stateManager.GetPodReAttemptRate(ctx)
		if reAttemptRate >= es.spotReAttemptOverride {
			fields.NodeLifecycle = &state.OndemandLifecycle
		}
	}

	if fields.NodeLifecycle == nil {
		fields.NodeLifecycle = &state.SpotLifecycle
	}

	// Calculate command_hash from actual command (FIX for ARA bug)
	// This ensures jobs with different commands have different hashes,
	// even if they share the same description.
	if fields.Command != nil && len(*fields.Command) > 0 {
		// Regular EKS jobs: Hash the command
		fields.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*fields.Command))))
	} else if *fields.Engine == state.EKSSparkEngine && fields.Description != nil && len(*fields.Description) > 0 {
		// Spark jobs: Fall back to description (Spark jobs don't have commands)
		// The Spark "command" is in spark_extension, not the command field
		// Description uniquely identifies the Spark job type for ARA tracking
		fields.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*fields.Description))))
	}
	// If both command and description are NULL, command_hash remains NULL (malformed job)

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
		ServiceAccount:        fields.ServiceAccount,
		Tier:                  fields.Tier,
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
	ctx context.Context,
	limit int,
	offset int,
	sortOrder string,
	sortField string,
	filters map[string][]string,
	envFilters map[string]string) (state.RunList, error) {
	ctx, span := utils.TraceJob(ctx, "flotilla.list_runs", "")
	defer span.Finish()
	span.SetTag("limit", limit)
	span.SetTag("offset", offset)

	// If definition_id is present in filters, validate its
	// existence first
	definitionID, ok := filters["definition_id"]
	if ok {
		_, err := es.stateManager.GetDefinition(ctx, definitionID[0])
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
	return es.stateManager.ListRuns(ctx, limit, offset, sortField, sortOrder, filters, envFilters, []string{state.EKSEngine, state.EKSSparkEngine})
}

// Get returns the run with the given runID
func (es *executionService) Get(ctx context.Context, runID string) (state.Run, error) {
	ctx, span := utils.TraceJob(ctx, "flotilla.get_run", runID)
	defer span.Finish()
	span.SetTag("run_id", runID)
	run, err := es.stateManager.GetRun(ctx, runID)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
	}
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
	}
	return run, err
}

// UpdateStatus is for supporting some legacy runs that still manually update their status
func (es *executionService) UpdateStatus(ctx context.Context, runID string, status string, exitCode *int64, runExceptions *state.RunExceptions, exitReason *string) error {
	ctx, span := utils.TraceJob(ctx, "flotilla.update_status", runID)
	defer span.Finish()
	span.SetTag("run_id", runID)
	span.SetTag("status", status)
	if !state.IsValidStatus(status) {
		return exceptions.MalformedInput{ErrorString: fmt.Sprintf("status %s is invalid", status)}
	}
	run, err := es.stateManager.GetRun(ctx, runID)
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

	_, err = es.stateManager.UpdateRun(ctx, runID, state.Run{Status: status, ExitCode: exitCode, ExitReason: exitReason, RunExceptions: runExceptions, FinishedAt: &finishedAt, StartedAt: startedAt})
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
	ctx := context.Background()
	for job := range jobChan {
		runID := job.RunID
		userInfo := job.UserInfo
		ctx, span := utils.TraceJob(ctx, "flotilla.job.terminate_worker", runID)
		defer span.Finish()
		run, err := es.stateManager.GetRun(ctx, runID)
		if err != nil {
			span.SetTag("error", true)
			span.SetTag("error.msg", err.Error())
			break
		}
		utils.TagJobRun(span, run)
		if err != nil {
			break
		}

		subRuns, err := es.stateManager.ListRuns(ctx, 1000, 0, "status", "desc", nil, map[string]string{"PARENT_FLOTILLA_RUN_ID": run.RunID}, state.Engines)
		if err == nil && subRuns.Total > 0 {
			for _, subRun := range subRuns.Runs {
				es.terminateJobChannel <- state.TerminateJob{
					RunID:    subRun.RunID,
					UserInfo: job.UserInfo,
				}
			}
		}

		if run.Engine == nil {
			run.Engine = &state.EKSEngine
		}

		if run.Status != state.StatusStopped {
			if *run.Engine == state.EKSSparkEngine {
				err = es.emrExecutionEngine.Terminate(ctx, run)
			} else {
				err = es.eksExecutionEngine.Terminate(ctx, run)
			}
			exitReason := "Task terminated by user"
			if len(userInfo.Email) > 0 {
				exitReason = fmt.Sprintf("Task terminated by - %s", userInfo.Email)
			}

			exitCode := int64(1)
			finishedAt := time.Now()
			_, err = es.stateManager.UpdateRun(ctx, run.RunID, state.Run{
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
func (es *executionService) Terminate(ctx context.Context, runID string, userInfo state.UserInfo) error {
	ctx, span := utils.TraceJob(ctx, "flotilla.terminate_run", runID)
	defer span.Finish()
	span.SetTag("run_id", runID)
	if userInfo.Email != "" {
		span.SetTag("user.email", userInfo.Email)
	}
	es.terminateJobChannel <- state.TerminateJob{RunID: runID, UserInfo: userInfo}
	go es.terminateWorker(es.terminateJobChannel)
	return nil
}

// ListClusters returns a list of all execution clusters available with their metadata
func (es *executionService) ListClusters(ctx context.Context) ([]state.ClusterMetadata, error) {
	ctx, span := utils.TraceJob(ctx, "flotilla.list_clusters", "")
	defer span.Finish()
	clusters, err := es.stateManager.ListClusterStates(ctx)
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

func (es *executionService) GetDefaultCluster() string {
	return es.eksClusterDefault
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
func (es *executionService) createAndEnqueueRun(ctx context.Context, run state.Run) (state.Run, error) {
	var err error
	ctx, span := utils.TraceJob(ctx, "flotilla.job.create_and_enqueue", "")
	defer span.Finish()
	span.SetTag("job.run_id", run.RunID)
	utils.TagJobRun(span, run)
	if run.IdempotenceKey != nil {
		priorRunId, err := es.stateManager.CheckIdempotenceKey(ctx, *run.IdempotenceKey)
		if err == nil && len(priorRunId) > 0 {
			priorRun, err := es.Get(ctx, priorRunId)
			if err == nil {
				return priorRun, nil
			}
		}
	}

	// Save run to source of state - it is *CRITICAL* to do this
	// -before- queuing to avoid processing unsaved runs
	if err = es.stateManager.CreateRun(ctx, run); err != nil {
		return run, err
	}

	if *run.Engine == state.EKSEngine {
		err = es.eksExecutionEngine.Enqueue(ctx, run)
	} else {
		err = es.emrExecutionEngine.Enqueue(ctx, run)
	}
	queuedAt := time.Now()

	if err != nil {
		return run, err
	}

	// UpdateStatus the run's QueuedAt field
	if run, err = es.stateManager.UpdateRun(ctx, run.RunID, state.Run{QueuedAt: &queuedAt}); err != nil {
		return run, err
	}
	return run, nil
}
func (es *executionService) CreateTemplateRunByTemplateName(ctx context.Context, templateName string, templateVersion string, req *state.TemplateExecutionRequest) (state.Run, error) {
	ctx, span := utils.TraceJob(ctx, "flotilla.template.create_run_by_name", "")
	defer span.Finish()
	span.SetTag("template_name", templateName)
	span.SetTag("template_version", templateVersion)
	version, err := strconv.Atoi(templateVersion)

	if err != nil {
		//use the "latest" template - version not a integer
		fetch, template, err := es.stateManager.GetLatestTemplateByTemplateName(ctx, templateName)
		if fetch && err == nil {
			return es.CreateTemplateRunByTemplateID(ctx, template.TemplateID, req)
		}
	} else {
		fetch, template, err := es.stateManager.GetTemplateByVersion(ctx, templateName, int64(version))
		if fetch && err == nil {
			return es.CreateTemplateRunByTemplateID(ctx, template.TemplateID, req)
		}
	}
	return state.Run{},
		errors.New(fmt.Sprintf("invalid template name or version, template_name: %s, template_version: %s", templateName, templateVersion))
}

// Create constructs and queues a new Run on the cluster specified.
func (es *executionService) CreateTemplateRunByTemplateID(ctx context.Context, templateID string, req *state.TemplateExecutionRequest) (state.Run, error) {
	ctx, span := utils.TraceJob(ctx, "flotilla.template.create_run_by_id", "")
	defer span.Finish()
	span.SetTag("template_id", templateID)
	// Ensure template exists
	template, err := es.stateManager.GetTemplateByID(ctx, templateID)
	if err != nil {
		return state.Run{}, err
	}

	return es.createFromTemplate(ctx, template, req)
}

func (es *executionService) createFromTemplate(ctx context.Context, template state.Template, req *state.TemplateExecutionRequest) (state.Run, error) {
	var (
		run state.Run
		err error
	)

	fields := req.GetExecutionRequestCommon()
	es.sanitizeExecutionRequestCommonFields(fields)

	// Construct run object with StatusQueued and new UUID4 run id
	run, err = es.constructRunFromTemplate(ctx, template, req)
	if err != nil {
		return run, err
	}
	if !req.DryRun {
		return es.createAndEnqueueRun(ctx, run)
	}
	return run, nil
}

func (es *executionService) constructRunFromTemplate(ctx context.Context, template state.Template, req *state.TemplateExecutionRequest) (state.Run, error) {
	run, err := es.constructBaseRunFromExecutable(ctx, template, req)

	if err != nil {
		return run, err
	}

	run.DefinitionID = template.TemplateID
	run.Alias = template.TemplateID
	run.GroupName = "template_group_name"
	run.ExecutionRequestCustom = req.GetExecutionRequestCustom()

	return run, nil
}

// resolveRequestTier returns the requested tier or default tier if empty
func (es *executionService) resolveRequestTier(requestedTier state.Tier) state.Tier {
	if requestedTier == "" {
		return state.Tier(es.eksTierDefault)
	}
	return requestedTier
}

// clusterSupportsTier checks if a cluster supports the specified tier
func (es *executionService) clusterSupportsTier(cluster state.ClusterMetadata, requestedTier state.Tier) bool {
	resolvedTier := es.resolveRequestTier(requestedTier)
	for _, allowedTier := range cluster.AllowedTiers {
		if allowedTier == string(resolvedTier) {
			return true
		}
	}

	return false
}

func (es *executionService) isClusterValid(clusterName string) bool {
	return slices.Contains(es.validEksClusters, clusterName)
}

func (es *executionService) UpdateClusterMetadata(ctx context.Context, cluster state.ClusterMetadata) error {
	ctx, span := utils.TraceJob(ctx, "flotilla.update_cluster_metadata", cluster.Name)
	defer span.Finish()
	span.SetTag("cluster_name", cluster.Name)
	return es.stateManager.UpdateClusterMetadata(ctx, cluster)
}

func (es *executionService) DeleteClusterMetadata(ctx context.Context, clusterID string) error {
	ctx, span := utils.TraceJob(ctx, "flotilla.delete_cluster_metadata", clusterID)
	defer span.Finish()
	span.SetTag("cluster_id", clusterID)
	return es.stateManager.DeleteClusterMetadata(ctx, clusterID)
}

func (es *executionService) GetClusterByID(ctx context.Context, clusterID string) (state.ClusterMetadata, error) {
	ctx, span := utils.TraceJob(ctx, "flotilla.get_cluster_by_id", clusterID)
	defer span.Finish()
	span.SetTag("cluster_id", clusterID)
	return es.stateManager.GetClusterByID(ctx, clusterID)
}

// GetRunStatus fetches only the essential status information for a run
func (es *executionService) GetRunStatus(ctx context.Context, runID string) (state.RunStatus, error) {
	ctx, span := utils.TraceJob(ctx, "flotilla.get_run_status", runID)
	defer span.Finish()
	span.SetTag("run_id", runID)
	return es.stateManager.GetRunStatus(ctx, runID)
}
