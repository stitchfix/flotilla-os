package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/aws/aws-sdk-go/aws"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/utils"
	"github.com/xeipuuv/gojsonschema"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var EKSEngine = "eks"

var EKSSparkEngine = "eks-spark"

var DefaultEngine = EKSEngine

var DefaultTaskType = "task"

var MinCPU = int64(256)

var MaxCPU = int64(128000)

var MinMem = int64(512)

var MaxMem = int64(500000)

var TTLSecondsAfterFinished = int32(3600)

var SpotActiveDeadlineSeconds = int64(172800)

var OndemandActiveDeadlineSeconds = int64(604800)

var SpotLifecycle = "spot"

var OndemandLifecycle = "ondemand"

var DefaultLifecycle = SpotLifecycle

var NodeLifeCycles = []string{OndemandLifecycle, SpotLifecycle}

var Engines = []string{EKSEngine, EKSSparkEngine}

// StatusRunning indicates the run is running
var StatusRunning = "RUNNING"

// StatusQueued indicates the run is queued
var StatusQueued = "QUEUED"

// StatusNeedsRetry indicates the run failed for infra reasons and needs retried
var StatusNeedsRetry = "NEEDS_RETRY"

// StatusPending indicates the run has been allocated to a host and is in the process of launching
var StatusPending = "PENDING"

// StatusStopped means the run is finished
var StatusStopped = "STOPPED"

var MaxLogLines = int64(256)

var EKSBackoffLimit = int32(0)

var GPUNodeTypes = []string{"p3.2xlarge", "p3.8xlarge", "p3.16xlarge", "g5.xlarge", "g5.2xlarge", "g5.4xlarge", "g5.8xlarge", "g5.12xlarge", "g5.16xlarge", "g5.24xlarge", "g5.48xlarge"}

var WorkerTypes = map[string]bool{
	"retry":  true,
	"submit": true,
	"status": true,
}

func IsValidWorkerType(workerType string) bool {
	return WorkerTypes[workerType]
}

// IsValidStatus checks that the given status
// string is one of the valid statuses
func IsValidStatus(status string) bool {
	return status == StatusRunning ||
		status == StatusQueued ||
		status == StatusNeedsRetry ||
		status == StatusPending ||
		status == StatusStopped
}

// NewRunID returns a new uuid for a Run
func NewRunID(engine *string) (string, error) {
	s, err := newUUIDv4()
	return fmt.Sprintf("%s-%s", *engine, s[len(*engine)+1:]), err
}

// NewDefinitionID returns a new uuid for a Definition
func NewDefinitionID(definition Definition) (string, error) {
	uuid4, err := newUUIDv4()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", definition.GroupName, uuid4), nil
}

func newUUIDv4() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// EnvList wraps a list of EnvVar
//   - abstraction to make it easier to read
//     and write to db
type EnvList []EnvVar

// PortsList wraps a list of int
//   - abstraction to make it easier to read
//     and write to db
type PortsList []int

// EnvVar represents a single environment variable
// for either a definition or a run
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type NodeList []string

// Tags wraps a list of strings
//   - abstraction to make it easier to read
//     and write to db
type Tags []string

// ExecutableResources define the resources and flags required to run an
// executable.
type ExecutableResources struct {
	Image                      string     `json:"image"`
	Memory                     *int64     `json:"memory,omitempty"`
	Gpu                        *int64     `json:"gpu,omitempty"`
	Cpu                        *int64     `json:"cpu,omitempty"`
	Env                        *EnvList   `json:"env"`
	AdaptiveResourceAllocation *bool      `json:"adaptive_resource_allocation,omitempty"`
	Ports                      *PortsList `json:"ports,omitempty"`
	Tags                       *Tags      `json:"tags,omitempty"`
}

type ExecutableType string

const (
	ExecutableTypeDefinition ExecutableType = "task_definition"
	ExecutableTypeTemplate   ExecutableType = "template"
)

type Executable interface {
	GetExecutableID() *string
	GetExecutableType() *ExecutableType
	GetExecutableResources() *ExecutableResources
	GetExecutableCommand(req ExecutionRequest) (string, error)
	GetExecutableResourceName() string // This will typically be an ARN.
}

func UnmarshalSparkExtension(data []byte) (SparkExtension, error) {
	var r SparkExtension
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *SparkExtension) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type SparkExtension struct {
	SparkSubmitJobDriver *SparkSubmitJobDriver `json:"spark_submit_job_driver,omitempty"`
	ApplicationConf      []Conf                `json:"application_conf,omitempty"`
	HiveConf             []Conf                `json:"hive_conf,omitempty"`
	EMRJobId             *string               `json:"emr_job_id,omitempty"`
	SparkAppId           *string               `json:"spark_app_id,omitempty"`
	EMRJobManifest       *string               `json:"emr_job_manifest,omitempty"`
	HistoryUri           *string               `json:"history_uri,omitempty"`
	MetricsUri           *string               `json:"metrics_uri,omitempty"`
	VirtualClusterId     *string               `json:"virtual_cluster_id,omitempty"`
	EMRReleaseLabel      *string               `json:"emr_release_label,omitempty"`
	ExecutorInitCommand  *string               `json:"executor_init_command,omitempty"`
	DriverInitCommand    *string               `json:"driver_init_command,omitempty"`
	AppUri               *string               `json:"app_uri,omitempty"`
	Executors            []string              `json:"executors,omitempty"`
	ExecutorOOM          *bool                 `json:"executor_oom,omitempty"`
	DriverOOM            *bool                 `json:"driver_oom,omitempty"`
}

type Conf struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

type SparkSubmitJobDriver struct {
	EntryPoint          *string   `json:"entry_point,omitempty"`
	EntryPointArguments []*string `json:"entry_point_arguments,omitempty"`
	SparkSubmitConf     []Conf    `json:"spark_submit_conf,omitempty"`
	Files               []string  `json:"files,omitempty"`
	PyFiles             []string  `json:"py_files,omitempty"`
	Jars                []string  `json:"jars,omitempty"`
	Class               *string   `json:"class,omitempty"`
	WorkingDir          *string   `json:"working_dir,omitempty"`
	NumExecutors        *int64    `json:"num_executors,omitempty"`
	ExecutorMemory      *int64    `json:"executor_memory,omitempty"`
}

type Labels map[string]string

// Common fields required to execute any Executable.
type ExecutionRequestCommon struct {
	ClusterName           string          `json:"cluster_name"`
	Env                   *EnvList        `json:"env"`
	OwnerID               string          `json:"owner_id"`
	Command               *string         `json:"command"`
	Memory                *int64          `json:"memory"`
	Cpu                   *int64          `json:"cpu"`
	Gpu                   *int64          `json:"gpu"`
	Engine                *string         `json:"engine"`
	EphemeralStorage      *int64          `json:"ephemeral_storage"`
	NodeLifecycle         *string         `json:"node_lifecycle"`
	ActiveDeadlineSeconds *int64          `json:"active_deadline_seconds,omitempty"`
	SparkExtension        *SparkExtension `json:"spark_extension,omitempty"`
	Description           *string         `json:"description,omitempty"`
	CommandHash           *string         `json:"command_hash,omitempty"`
	IdempotenceKey        *string         `json:"idempotence_key,omitempty"`
	Arch                  *string         `json:"arch,omitempty"`
	Labels                *Labels         `json:"labels,omitempty"`
}

type ExecutionRequestCustom map[string]interface{}
type ExecutionRequest interface {
	GetExecutionRequestCommon() *ExecutionRequestCommon
	GetExecutionRequestCustom() *ExecutionRequestCustom
}

type DefinitionExecutionRequest struct {
	*ExecutionRequestCommon
}

// Returns ExecutionRequestCommon, common between Template and Definition types
func (d *DefinitionExecutionRequest) GetExecutionRequestCommon() *ExecutionRequestCommon {
	return d.ExecutionRequestCommon
}

// Only relevant to the template type
func (d *DefinitionExecutionRequest) GetExecutionRequestCustom() *ExecutionRequestCustom {
	return nil
}

type TerminateJob struct {
	RunID    string
	UserInfo UserInfo
}

// task definition. It implements the `Executable` interface.
type Definition struct {
	DefinitionID string `json:"definition_id"`
	GroupName    string `json:"group_name,omitempty"`
	Alias        string `json:"alias"`
	Command      string `json:"command,omitempty"`
	TaskType     string `json:"task_type,omitempty"`
	ExecutableResources
}

// Return definition or template id
func (d Definition) GetExecutableID() *string {
	return &d.DefinitionID
}

// Returns definition or template
func (d Definition) GetExecutableType() *ExecutableType {
	t := ExecutableTypeDefinition
	return &t
}
func (d Definition) GetExecutableResources() *ExecutableResources {
	return &d.ExecutableResources
}

func (d Definition) GetExecutableCommand(req ExecutionRequest) (string, error) {
	return d.Command, nil
}

func (d Definition) GetExecutableResourceName() string {
	return d.DefinitionID
}

var commandWrapper = `
set -e
set -x

{{.Command}}
`
var CommandTemplate, _ = template.New("command").Parse(commandWrapper)

// WrappedCommand returns the wrapped command for the definition
// * wrapping ensures lines are logged and exit code is set
func (d *Definition) WrappedCommand() (string, error) {
	var result bytes.Buffer
	if err := CommandTemplate.Execute(&result, d); err != nil {
		return "", err
	}
	return result.String(), nil
}

type validationCondition struct {
	condition bool
	reason    string
}

// IsValid returns true only if this is a valid definition with all
// required information
func (d *Definition) IsValid() (bool, []string) {
	conditions := []validationCondition{
		{len(d.Image) == 0, "string [image] must be specified"},
		{len(d.Alias) == 0, "string [alias] must be specified"},
	}

	valid := true
	var reasons []string
	for _, cond := range conditions {
		if cond.condition {
			valid = false
			reasons = append(reasons, cond.reason)
		}
	}
	return valid, reasons
}

// UpdateWith updates this definition with information from another
func (d *Definition) UpdateWith(other Definition) {
	if len(other.DefinitionID) > 0 {
		d.DefinitionID = other.DefinitionID
	}
	if len(other.Image) > 0 {
		d.Image = other.Image
	}
	if len(other.GroupName) > 0 {
		d.GroupName = other.GroupName
	}
	if len(other.Alias) > 0 {
		d.Alias = other.Alias
	}
	if other.Memory != nil {
		d.Memory = other.Memory
	}
	if other.Gpu != nil {
		d.Gpu = other.Gpu
	}
	if other.Cpu != nil {
		d.Cpu = other.Cpu
	}
	if other.AdaptiveResourceAllocation != nil {
		d.AdaptiveResourceAllocation = other.AdaptiveResourceAllocation
	}
	if len(other.Command) > 0 {
		d.Command = other.Command
	}
	if len(other.TaskType) > 0 {
		d.TaskType = other.TaskType
	}
	if other.Env != nil {
		d.Env = other.Env
	}
	if other.Ports != nil {
		d.Ports = other.Ports
	}
	if other.Tags != nil {
		d.Tags = other.Tags
	}
}

func (d Definition) MarshalJSON() ([]byte, error) {
	type Alias Definition

	env := d.Env
	if env == nil {
		env = &EnvList{}
	}

	return json.Marshal(&struct {
		Env *EnvList `json:"env"`
		Alias
	}{
		Env:   env,
		Alias: (Alias)(d),
	})
}

// DefinitionList wraps a list of Definitions
type DefinitionList struct {
	Total       int          `json:"total"`
	Definitions []Definition `json:"definitions"`
}

func (dl *DefinitionList) MarshalJSON() ([]byte, error) {
	type Alias DefinitionList
	l := dl.Definitions
	if l == nil {
		l = []Definition{}
	}
	return json.Marshal(&struct {
		Definitions []Definition `json:"definitions"`
		*Alias
	}{
		Definitions: l,
		Alias:       (*Alias)(dl),
	})
}

// Run represents a single run of a Definition
//
// TODO:
//
//	Runs need to -copy- the run relevant information
//	from their associated definition when they are
//	created so they always have correct info. Currently
//	the definition can change during or after the run
//	is created and launched meaning the run is acting
//	on information that is no longer accessible.
type Run struct {
	RunID                   string                   `json:"run_id"`
	DefinitionID            string                   `json:"definition_id"`
	Alias                   string                   `json:"alias"`
	Image                   string                   `json:"image"`
	ClusterName             string                   `json:"cluster"`
	ExitCode                *int64                   `json:"exit_code,omitempty"`
	Status                  string                   `json:"status"`
	QueuedAt                *time.Time               `json:"queued_at,omitempty"`
	StartedAt               *time.Time               `json:"started_at,omitempty"`
	FinishedAt              *time.Time               `json:"finished_at,omitempty"`
	InstanceID              string                   `json:"-"`
	InstanceDNSName         string                   `json:"-"`
	GroupName               string                   `json:"group_name"`
	User                    string                   `json:"user,omitempty"`
	TaskType                string                   `json:"task_type,omitempty"`
	Env                     *EnvList                 `json:"env,omitempty"`
	Command                 *string                  `json:"command,omitempty"`
	CommandHash             *string                  `json:"command_hash,omitempty"`
	Memory                  *int64                   `json:"memory,omitempty"`
	MemoryLimit             *int64                   `json:"memory_limit,omitempty"`
	Cpu                     *int64                   `json:"cpu,omitempty"`
	CpuLimit                *int64                   `json:"cpu_limit,omitempty"`
	Gpu                     *int64                   `json:"gpu,omitempty"`
	ExitReason              *string                  `json:"exit_reason,omitempty"`
	Engine                  *string                  `json:"engine,omitempty"`
	NodeLifecycle           *string                  `json:"node_lifecycle,omitempty"`
	EphemeralStorage        *int64                   `json:"ephemeral_storage,omitempty"`
	PodName                 *string                  `json:"pod_name,omitempty"`
	Namespace               *string                  `json:"namespace,omitempty"`
	MaxMemoryUsed           *int64                   `json:"max_memory_used,omitempty"`
	MaxCpuUsed              *int64                   `json:"max_cpu_used,omitempty"`
	PodEvents               *PodEvents               `json:"pod_events,omitempty"`
	CloudTrailNotifications *CloudTrailNotifications `json:"cloudtrail_notifications,omitempty"`
	ExecutableID            *string                  `json:"executable_id,omitempty"`
	ExecutableType          *ExecutableType          `json:"executable_type,omitempty"`
	ExecutionRequestCustom  *ExecutionRequestCustom  `json:"execution_request_custom,omitempty"`
	AttemptCount            *int64                   `json:"attempt_count,omitempty"`
	SpawnedRuns             *SpawnedRuns             `json:"spawned_runs,omitempty"`
	RunExceptions           *RunExceptions           `json:"run_exceptions,omitempty"`
	ActiveDeadlineSeconds   *int64                   `json:"active_deadline_seconds,omitempty"`
	SparkExtension          *SparkExtension          `json:"spark_extension,omitempty"`
	MetricsUri              *string                  `json:"metrics_uri,omitempty"`
	Description             *string                  `json:"description,omitempty"`
	IdempotenceKey          *string                  `json:"idempotence_key,omitempty"`
	Arch                    *string                  `json:"arch,omitempty"`
	Labels                  Labels                   `json:"labels,omitempty"`
}

// UpdateWith updates this run with information from another
func (d *Run) UpdateWith(other Run) {
	if len(other.RunID) > 0 {
		d.RunID = other.RunID
	}
	if len(other.DefinitionID) > 0 {
		d.DefinitionID = other.DefinitionID
	}
	if len(other.Alias) > 0 {
		d.Alias = other.Alias
	}
	if len(other.Image) > 0 {
		d.Image = other.Image
	}
	if len(other.ClusterName) > 0 {
		d.ClusterName = other.ClusterName
	}
	if other.ExitCode != nil {
		d.ExitCode = other.ExitCode
	}
	if other.QueuedAt != nil {
		d.QueuedAt = other.QueuedAt
	}
	if other.StartedAt != nil {
		d.StartedAt = other.StartedAt
	}
	if other.FinishedAt != nil {
		d.FinishedAt = other.FinishedAt
	}
	if len(other.InstanceID) > 0 {
		d.InstanceID = other.InstanceID
	}
	if len(other.InstanceDNSName) > 0 {
		d.InstanceDNSName = other.InstanceDNSName
	}
	if len(other.GroupName) > 0 {
		d.GroupName = other.GroupName
	}
	if len(other.User) > 0 {
		d.User = other.User
	}
	if len(other.TaskType) > 0 {
		d.TaskType = other.TaskType
	}
	if other.Env != nil {
		d.Env = other.Env
	}

	if other.ExitReason != nil {
		d.ExitReason = other.ExitReason
	}

	if other.Command != nil && len(*other.Command) > 0 {
		d.Command = other.Command
	}

	if other.CommandHash != nil && len(*other.CommandHash) > 0 {
		d.CommandHash = other.CommandHash
	}

	if other.Memory != nil {
		d.Memory = other.Memory
	}

	if other.Cpu != nil {
		d.Cpu = other.Cpu
	}

	if other.Gpu != nil {
		d.Gpu = other.Gpu
	}

	if other.MaxMemoryUsed != nil {
		d.MaxMemoryUsed = other.MaxMemoryUsed
	}

	if other.MaxCpuUsed != nil {
		d.MaxCpuUsed = other.MaxCpuUsed
	}

	if other.Engine != nil {
		d.Engine = other.Engine
	}

	if other.EphemeralStorage != nil {
		d.EphemeralStorage = other.EphemeralStorage
	}

	if other.NodeLifecycle != nil {
		d.NodeLifecycle = other.NodeLifecycle
	}

	if other.PodName != nil {
		d.PodName = other.PodName
	}

	if other.Namespace != nil {
		d.Namespace = other.Namespace
	}

	if other.PodEvents != nil {
		d.PodEvents = other.PodEvents
	}

	if other.SpawnedRuns != nil {
		d.SpawnedRuns = other.SpawnedRuns
	}

	if other.RunExceptions != nil {
		d.RunExceptions = other.RunExceptions
	}

	if other.ExecutableID != nil {
		d.ExecutableID = other.ExecutableID
	}

	if other.ExecutableType != nil {
		d.ExecutableType = other.ExecutableType
	}

	if other.SparkExtension != nil {
		d.SparkExtension = other.SparkExtension
	}

	if other.CloudTrailNotifications != nil && len((*other.CloudTrailNotifications).Records) > 0 {
		d.CloudTrailNotifications = other.CloudTrailNotifications
	}

	if other.ExecutionRequestCustom != nil {
		d.ExecutionRequestCustom = other.ExecutionRequestCustom
	}

	if other.CpuLimit != nil {
		d.CpuLimit = other.CpuLimit
	}

	if other.MetricsUri != nil {
		d.MetricsUri = other.MetricsUri
	}

	if other.Description != nil {
		d.Description = other.Description
	}

	if other.IdempotenceKey != nil {
		d.IdempotenceKey = other.IdempotenceKey
	}

	if other.Arch != nil {
		d.Arch = other.Arch
	}

	if other.MemoryLimit != nil {
		d.MemoryLimit = other.MemoryLimit
	}

	if other.AttemptCount != nil {
		d.AttemptCount = other.AttemptCount
	}

	if other.Labels != nil {
		d.Labels = other.Labels
	}
	//
	// Runs have a deterministic lifecycle
	//
	// QUEUED --> PENDING --> RUNNING --> STOPPED
	// QUEUED --> PENDING --> NEEDS_RETRY --> QUEUED ...
	// QUEUED --> PENDING --> STOPPED ...
	//
	statusPrecedence := map[string]int{
		StatusNeedsRetry: -1,
		StatusQueued:     0,
		StatusPending:    1,
		StatusRunning:    2,
		StatusStopped:    3,
	}

	if other.Status == StatusNeedsRetry {
		d.Status = StatusNeedsRetry
	} else {
		if runStatus, ok := statusPrecedence[d.Status]; ok {
			if newStatus, ok := statusPrecedence[other.Status]; ok {
				if newStatus > runStatus {
					d.Status = other.Status
				}
			}
		}
	}
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

type byExecutorName []string

func (s byExecutorName) Len() int {
	return len(s)
}
func (s byExecutorName) Key(i int) int {
	r, _ := regexp.Compile("-exec-(\\d+)")
	matches := r.FindStringSubmatch(s[i])
	if matches == nil || len(matches) < 2 {
		return 0
	}
	key, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return key
}
func (s byExecutorName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byExecutorName) Less(i, j int) bool {
	return s.Key(i) < s.Key(j)
}

func (r Run) MarshalJSON() ([]byte, error) {
	type Alias Run
	instance := map[string]string{
		"instance_id": r.InstanceID,
		"dns_name":    r.InstanceDNSName,
	}
	podEvents := r.PodEvents
	if podEvents == nil {
		podEvents = &PodEvents{}
	}

	var executors []string
	for _, podEvent := range *podEvents {
		if strings.Contains(podEvent.SourceObject, "-exec-") {
			executors = append(executors, podEvent.SourceObject)
		}
	}

	if executors != nil && len(executors) > 0 && *r.Engine != EKSEngine {
		executors = removeDuplicateStr(executors)
		sort.Sort(byExecutorName(executors))
		r.SparkExtension.Executors = executors
	}

	cloudTrailNotifications := r.CloudTrailNotifications
	if cloudTrailNotifications == nil {
		cloudTrailNotifications = &CloudTrailNotifications{}
	}

	executionRequestCustom := r.ExecutionRequestCustom
	if executionRequestCustom == nil {
		executionRequestCustom = &ExecutionRequestCustom{}
	}

	if r.Description == nil {
		r.Description = aws.String(r.Alias)
	}

	sparkExtension := r.SparkExtension

	if sparkExtension == nil {
		sparkExtension = &SparkExtension{}
	} else {
		if sparkExtension.HiveConf != nil {
			for _, conf := range sparkExtension.HiveConf {
				if conf.Name != nil && strings.Contains(*conf.Name, "ConnectionPassword") {
					conf.Value = aws.String("****")
				}
			}
		}
		if r.Status != StatusStopped && r.SparkExtension.AppUri != nil {
			r.SparkExtension.HistoryUri = r.SparkExtension.AppUri
		}
	}

	return json.Marshal(&struct {
		Instance                map[string]string        `json:"instance"`
		PodEvents               *PodEvents               `json:"pod_events"`
		CloudTrailNotifications *CloudTrailNotifications `json:"cloudtrail_notifications"`
		SparkExtension          *SparkExtension          `json:"spark_extension"`
		Alias
	}{
		Instance:                instance,
		PodEvents:               podEvents,
		CloudTrailNotifications: cloudTrailNotifications,
		SparkExtension:          sparkExtension,
		Alias:                   (Alias)(r),
	})
}

// RunList wraps a list of Runs
type RunList struct {
	Total int   `json:"total"`
	Runs  []Run `json:"history"`
}

type PodEvents []PodEvent

type PodEventList struct {
	Total     int       `json:"total"`
	PodEvents PodEvents `json:"pod_events"`
}

type SpawnedRun struct {
	RunID string `json:"run_id"`
}

type SpawnedRuns []SpawnedRun

type RunExceptions []string

func (w *PodEvent) Equal(other PodEvent) bool {
	return w.Reason == other.Reason &&
		other.Timestamp != nil &&
		w.Timestamp.Equal(*other.Timestamp) &&
		w.SourceObject == other.SourceObject &&
		w.Message == other.Message &&
		w.EventType == other.EventType
}

type PodEvent struct {
	Timestamp    *time.Time `json:"timestamp,omitempty"`
	EventType    string     `json:"event_type"`
	Reason       string     `json:"reason"`
	SourceObject string     `json:"source_object"`
	Message      string     `json:"message"`
}

// GroupsList wraps a list of group names
type GroupsList struct {
	Groups []string
	Total  int
}

// TagsList wraps a list of tag names
type TagsList struct {
	Tags  []string
	Total int
}

// Worker represents a Flotilla Worker
type Worker struct {
	WorkerType       string `json:"worker_type"`
	CountPerInstance int    `json:"count_per_instance"`
	Engine           string `json:"engine"`
}

// UpdateWith updates this definition with information from another
func (w *Worker) UpdateWith(other Worker) {
	if other.CountPerInstance >= 0 {
		w.CountPerInstance = other.CountPerInstance
	}
}

// WorkersList wraps a list of Workers
type WorkersList struct {
	Total   int      `json:"total"`
	Workers []Worker `json:"workers"`
}

// User information making the API calls
type UserInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Internal object for tracking cpu / memory resources.
type TaskResources struct {
	Cpu    int64 `json:"cpu"`
	Memory int64 `json:"memory"`
}

// SQS notification object for CloudTrail S3 files.
type CloudTrailS3File struct {
	S3Bucket    string   `json:"s3Bucket"`
	S3ObjectKey []string `json:"s3ObjectKey"`
	Done        func() error
}

// Marshal method for CloudTrail SQS notifications.
func (e *CloudTrailNotifications) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// CloudTrail notification object that is persisted into the DB.
type CloudTrailNotifications struct {
	Records []Record `json:"Records"`
}

// CloudTrail notification record.
type Record struct {
	UserIdentity UserIdentity `json:"userIdentity"`
	EventSource  string       `json:"eventSource"`
	EventName    string       `json:"eventName"`
}

// User ARN who performed the AWS api action.
type UserIdentity struct {
	Arn string `json:"arn"`
}

// Equals helper method for Record.
func (w *Record) Equal(other Record) bool {
	return w.EventName == other.EventName && w.EventSource == other.EventSource
}

// String helper method for Record.
func (w *Record) String() string {
	return fmt.Sprintf("%s-%s", w.EventSource, w.EventName)
}

const TemplatePayloadKey = "template_payload"

type TemplatePayload map[string]interface{}

type TemplateExecutionRequest struct {
	*ExecutionRequestCommon
	TemplatePayload TemplatePayload `json:"template_payload"`
	DryRun          bool            `json:"dry_run,omitempty"`
}

// Returns ExecutionRequestCommon associated with a Template type.
func (t TemplateExecutionRequest) GetExecutionRequestCommon() *ExecutionRequestCommon {
	return t.ExecutionRequestCommon
}

// Returns ExecutionRequestCustom associated with a Template type.
func (t TemplateExecutionRequest) GetExecutionRequestCustom() *ExecutionRequestCustom {
	return &ExecutionRequestCustom{
		TemplatePayloadKey: t.TemplatePayload,
	}
}

// Templates uses JSON Schema types.
type TemplateJSONSchema map[string]interface{}

// Template Object Type. The CommandTemplate is a Go Template type.
type Template struct {
	TemplateID      string             `json:"template_id"`
	TemplateName    string             `json:"template_name"`
	Version         int64              `json:"version"`
	Schema          TemplateJSONSchema `json:"schema"`
	CommandTemplate string             `json:"command_template"`
	Defaults        TemplatePayload    `json:"defaults"`
	AvatarURI       string             `json:"avatar_uri"`
	ExecutableResources
}

type CreateTemplateRequest struct {
	TemplateName    string             `json:"template_name"`
	Schema          TemplateJSONSchema `json:"schema"`
	CommandTemplate string             `json:"command_template"`
	Defaults        TemplatePayload    `json:"defaults"`
	AvatarURI       string             `json:"avatar_uri"`
	ExecutableResources
}

type CreateTemplateResponse struct {
	DidCreate bool     `json:"did_create"`
	Template  Template `json:"template,omitempty"`
}

// Returns Template ID
func (t Template) GetExecutableID() *string {
	return &t.TemplateID
}

// Returns Template Type
func (t Template) GetExecutableType() *ExecutableType {
	et := ExecutableTypeTemplate
	return &et
}

// Returns default resources associated with that Template.
func (t Template) GetExecutableResources() *ExecutableResources {
	return &t.ExecutableResources
}

// Renders the command to be rendered for that Template.
func (t Template) GetExecutableCommand(req ExecutionRequest) (string, error) {
	var (
		err    error
		result bytes.Buffer
	)

	// Get the request's custom fields.
	customFields := *req.GetExecutionRequestCustom()
	executionPayload, ok := customFields[TemplatePayloadKey]
	if !ok || executionPayload == nil {
		return "", err
	}

	executionPayload, err = t.compositeUserAndDefaults(executionPayload)

	schemaLoader := gojsonschema.NewGoLoader(t.Schema)
	documentLoader := gojsonschema.NewGoLoader(executionPayload)

	// Perform JSON schema validation to ensure that the request's template
	// payload conforms to the template's JSON schema.
	validationResult, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return "", err
	}
	if validationResult != nil && validationResult.Valid() != true {
		var res []string
		for _, resultError := range validationResult.Errors() {
			res = append(res, resultError.String())
		}
		return "", errors.New(strings.Join(res, "\n"))
	}

	// Create a new template string based on the template.Template.
	textTemplate, err := template.New("command").Funcs(sprig.TxtFuncMap()).Parse(t.CommandTemplate)
	if err != nil {
		return "", err
	}

	// Dump payload into the template string.
	if err = textTemplate.Execute(&result, executionPayload); err != nil {
		return "", err
	}

	return result.String(), nil
}

// Returns the Template Id.
func (t Template) GetExecutableResourceName() string {
	return t.TemplateID
}

func (t Template) compositeUserAndDefaults(userPayload interface{}) (TemplatePayload, error) {
	var (
		final map[string]interface{}
		ok    bool
	)

	final, ok = userPayload.(TemplatePayload)
	if !ok {
		return final, errors.New("unable to cast request payload to TemplatePayload struct")
	}

	err := utils.MergeMaps(&final, t.Defaults)

	if err != nil {
		return final, err
	}

	return final, nil
}

// NewTemplateID returns a new uuid for a Template
func NewTemplateID(t Template) (string, error) {
	uuid4, err := newUUIDv4()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tpl-%s", uuid4[4:]), nil
}

// Checks validity of a template.
func (t *Template) IsValid() (bool, []string) {
	conditions := []validationCondition{
		{len(t.TemplateName) == 0, "string [template_name] must be specified"},
		{len(t.Schema) == 0, "schema must be specified"},
		{len(t.CommandTemplate) == 0, "string [command_template] must be specified"},
		{len(t.Image) == 0, "string [image] must be specified"},
		{t.Memory == nil, "int [memory] must be specified"},
	}

	valid := true
	var reasons []string
	for _, cond := range conditions {
		if cond.condition {
			valid = false
			reasons = append(reasons, cond.reason)
		}
	}
	return valid, reasons
}

// TemplateList wraps a list of Templates
type TemplateList struct {
	Total     int        `json:"total"`
	Templates []Template `json:"templates"`
}

// Template Marshal method.
func (tl *TemplateList) MarshalJSON() ([]byte, error) {
	type Alias TemplateList
	l := tl.Templates
	if l == nil {
		l = []Template{}
	}
	return json.Marshal(&struct {
		Templates []Template `json:"templates"`
		*Alias
	}{
		Templates: l,
		Alias:     (*Alias)(tl),
	})
}

func (r *KubernetesEvent) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type KubernetesEvent struct {
	Metadata           Metadata       `json:"metadata,omitempty"`
	Reason             string         `json:"reason,omitempty"`
	Message            string         `json:"message,omitempty"`
	Source             Source         `json:"source,omitempty"`
	FirstTimestamp     string         `json:"firstTimestamp,omitempty"`
	LastTimestamp      string         `json:"lastTimestamp,omitempty"`
	Count              int64          `json:"count,omitempty"`
	Type               string         `json:"type,omitempty"`
	EventTime          interface{}    `json:"eventTime,omitempty"`
	ReportingComponent string         `json:"reportingComponent,omitempty"`
	ReportingInstance  string         `json:"reportingInstance,omitempty"`
	InvolvedObject     InvolvedObject `json:"involvedObject,omitempty"`
	Done               func() error
}

type InvolvedObject struct {
	Kind            string      `json:"kind,omitempty"`
	Namespace       string      `json:"namespace,omitempty"`
	Name            string      `json:"name,omitempty"`
	Uid             string      `json:"uid,omitempty"`
	APIVersion      string      `json:"apiVersion,omitempty"`
	ResourceVersion string      `json:"resourceVersion,omitempty"`
	FieldPath       string      `json:"fieldPath,omitempty"`
	Labels          EventLabels `json:"labels,omitempty"`
}

type EventLabels struct {
	ControllerUid string `json:"controller-uid,omitempty"`
	JobName       string `json:"job-name,omitempty"`
}

type Metadata struct {
	Name              string `json:"name,omitempty"`
	Namespace         string `json:"namespace,omitempty"`
	SelfLink          string `json:"selfLink,omitempty"`
	Uid               string `json:"uid,omitempty"`
	ResourceVersion   string `json:"resourceVersion,omitempty"`
	CreationTimestamp string `json:"creationTimestamp,omitempty"`
}

type Source struct {
	Component string `json:"component,omitempty"`
	Host      string `json:"host,omitempty"`
}

func UnmarshalEmrEvents(data []byte) (EmrEvent, error) {
	var r EmrEvent
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *EmrEvent) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type EmrEvent struct {
	Version    *string       `json:"version,omitempty"`
	ID         *string       `json:"id,omitempty"`
	DetailType *string       `json:"detail-type,omitempty"`
	Source     *string       `json:"source,omitempty"`
	Account    *string       `json:"account,omitempty"`
	Time       *string       `json:"time,omitempty"`
	Region     *string       `json:"region,omitempty"`
	Resources  []interface{} `json:"resources,omitempty"`
	Detail     *Detail       `json:"detail,omitempty"`
	Done       func() error
}

type Detail struct {
	Severity         *string `json:"severity,omitempty"`
	Name             *string `json:"name,omitempty"`
	ID               *string `json:"id,omitempty"`
	Arn              *string `json:"arn,omitempty"`
	VirtualClusterID *string `json:"virtualClusterId,omitempty"`
	State            *string `json:"state,omitempty"`
	CreatedBy        *string `json:"createdBy,omitempty"`
	ReleaseLabel     *string `json:"releaseLabel,omitempty"`
	ExecutionRoleArn *string `json:"executionRoleArn,omitempty"`
	FailureReason    *string `json:"failureReason,omitempty"`
	StateDetails     *string `json:"stateDetails,omitempty"`
	Message          *string `json:"message,omitempty"`
}
