package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/utils"
	"github.com/xeipuuv/gojsonschema"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var ECSEngine = "ecs"

var EKSEngine = "eks"

var DefaultEngine = ECSEngine

var DefaultARA = true

var MinCPU = int64(512)

var MaxCPU = int64(32000)

var MinMem = int64(256)

var MaxMem = int64(124000)

var TTLSecondsAfterFinished = int32(3600)

var SpotActiveDeadlineSeconds = int64(172800)

var OndemandActiveDeadlineSeconds = int64(604800)

var SpotLifecycle = "spot"

var OndemandLifecycle = "ondemand"

var DefaultLifecycle = SpotLifecycle

var NodeLifeCycles = []string{OndemandLifecycle, SpotLifecycle}

var Engines = []string{ECSEngine, EKSEngine}

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

var WorkerTypes = map[string]bool{
	"retry":  true,
	"submit": true,
	"status": true,
}

func IsValidWorkerType(workerType string) bool {
	return WorkerTypes[workerType]
}

//
// IsValidStatus checks that the given status
// string is one of the valid statuses
//
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
	return fmt.Sprintf("%s-%s", *engine, s[4:]), err
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

//
// EnvList wraps a list of EnvVar
// - abstraction to make it easier to read
//   and write to db
//
type EnvList []EnvVar

//
// PortsList wraps a list of int
// - abstraction to make it easier to read
//   and write to db
//
type PortsList []int

//
// EnvVar represents a single environment variable
// for either a definition or a run
//
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type NodeList []string

//
// Tags wraps a list of strings
// - abstraction to make it easier to read
//   and write to db
//
type Tags []string

// ExecutableResources define the resources and flags required to run an
// executable.
type ExecutableResources struct {
	Image                      string     `json:"image"`
	Memory                     *int64     `json:"memory"`
	Gpu                        *int64     `json:"gpu,omitempty"`
	Cpu                        *int64     `json:"cpu,omitempty"`
	Env                        *EnvList   `json:"env"`
	Privileged                 *bool      `json:"privileged,omitempty"`
	AdaptiveResourceAllocation *bool      `json:"adaptive_resource_allocation,omitempty"`
	ContainerName              string     `json:"container_name"`
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

// Common fields required to execute any Executable.
type ExecutionRequestCommon struct {
	ClusterName      string   `json:"cluster_name"`
	Env              *EnvList `json:"env"`
	OwnerID          string   `json:"owner_id"`
	Command          *string  `json:"command"`
	Memory           *int64   `json:"memory"`
	Cpu              *int64   `json:"cpu"`
	Gpu              *int64   `json:"gpu"`
	Engine           *string  `json:"engine"`
	EphemeralStorage *int64   `json:"ephemeral_storage"`
	NodeLifecycle    *string  `json:"node_lifecycle"`
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

// Definition represents a definition of a job - roughly 1-1 with an AWS ECS
// task definition. It implements the `Executable` interface.
type Definition struct {
	Arn              string `json:"arn"`
	DefinitionID     string `json:"definition_id"`
	GroupName        string `json:"group_name"`
	User             string `json:"user,omitempty"`
	Alias            string `json:"alias"`
	Command          string `json:"command,omitempty"`
	TaskType         string `json:"-"`
	SharedMemorySize *int64 `json:"shared_memory_size,omitempty"`
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
	return d.Arn
}

var commandWrapper = `
set -e
set -x

{{.Command}}
`
var CommandTemplate, _ = template.New("command").Parse(commandWrapper)

//
// WrappedCommand returns the wrapped command for the definition
// * wrapping ensures lines are logged and exit code is set
//
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

var validGroupName = regexp.MustCompile(`^[a-zA-Z0-9_\\-]+$`)

//
// IsValid returns true only if this is a valid definition with all
// required information
//
func (d *Definition) IsValid() (bool, []string) {
	conditions := []validationCondition{
		{len(d.Image) == 0, "string [image] must be specified"},
		{len(d.GroupName) == 0, "string [group_name] must be specified"},
		{!validGroupName.MatchString(d.GroupName), "Group name can only contain letters, numbers, hyphens, and underscores"},
		{len(d.GroupName) > 255, "Group name must be 255 characters or less"},
		{len(d.Alias) == 0, "string [alias] must be specified"},
		{d.Memory == nil, "int [memory] must be specified"},
		{len(d.Command) == 0, "string [command] must be specified"},
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

//
// UpdateWith updates this definition with information from another
//
func (d *Definition) UpdateWith(other Definition) {
	if len(other.Arn) > 0 {
		d.Arn = other.Arn
	}
	if len(other.DefinitionID) > 0 {
		d.DefinitionID = other.DefinitionID
	}
	if len(other.Image) > 0 {
		d.Image = other.Image
	}
	if len(other.GroupName) > 0 {
		d.GroupName = other.GroupName
	}
	if len(other.ContainerName) > 0 {
		d.ContainerName = other.ContainerName
	}
	if len(other.User) > 0 {
		d.User = other.User
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
	if other.Privileged != nil {
		d.Privileged = other.Privileged
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

//
// DefinitionList wraps a list of Definitions
//
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

//
// Run represents a single run of a Definition
//
// TODO:
//   Runs need to -copy- the run relevant information
//   from their associated definition when they are
//   created so they always have correct info. Currently
//   the definition can change during or after the run
//   is created and launched meaning the run is acting
//   on information that is no longer accessible.
//
type Run struct {
	TaskArn                 string                   `json:"task_arn"`
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
	TaskType                string                   `json:"-"`
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
	ContainerName           *string                  `json:"container_name,omitempty"`
	MaxMemoryUsed           *int64                   `json:"max_memory_used,omitempty"`
	MaxCpuUsed              *int64                   `json:"max_cpu_used,omitempty"`
	PodEvents               *PodEvents               `json:"pod_events,omitempty"`
	CloudTrailNotifications *CloudTrailNotifications `json:"cloudtrail_notifications,omitempty"`
	ExecutableID            *string                  `json:"executable_id,omitempty"`
	ExecutableType          *ExecutableType          `json:"executable_type,omitempty"`
	ExecutionRequestCustom  *ExecutionRequestCustom  `json:"execution_request_custom,omitempty"`
}

//
// UpdateWith updates this run with information from another
//
func (d *Run) UpdateWith(other Run) {
	if len(other.TaskArn) > 0 {
		d.TaskArn = other.TaskArn
	}
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

	if other.ContainerName != nil {
		d.ContainerName = other.ContainerName
	}

	if other.Namespace != nil {
		d.Namespace = other.Namespace
	}

	if other.PodEvents != nil {
		d.PodEvents = other.PodEvents
	}

	if other.ExecutableID != nil {
		d.ExecutableID = other.ExecutableID
	}

	if other.ExecutableType != nil {
		d.ExecutableType = other.ExecutableType
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

	if other.MemoryLimit != nil {
		d.MemoryLimit = other.MemoryLimit
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

	cloudTrailNotifications := r.CloudTrailNotifications

	if cloudTrailNotifications == nil {
		cloudTrailNotifications = &CloudTrailNotifications{}
	}

	executionRequestCustom := r.ExecutionRequestCustom

	if executionRequestCustom == nil {
		executionRequestCustom = &ExecutionRequestCustom{}
	}

	return json.Marshal(&struct {
		Instance                map[string]string        `json:"instance"`
		PodEvents               *PodEvents               `json:"pod_events"`
		CloudTrailNotifications *CloudTrailNotifications `json:"cloudtrail_notifications"`
		Alias
	}{
		Instance:                instance,
		PodEvents:               podEvents,
		CloudTrailNotifications: cloudTrailNotifications,
		Alias:                   (Alias)(r),
	})
}

//
// RunList wraps a list of Runs
//
type RunList struct {
	Total int   `json:"total"`
	Runs  []Run `json:"history"`
}

type PodEvents []PodEvent

type PodEventList struct {
	Total     int       `json:"total"`
	PodEvents PodEvents `json:"pod_events"`
}

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

//
// GroupsList wraps a list of group names
//
type GroupsList struct {
	Groups []string
	Total  int
}

//
// TagsList wraps a list of tag names
//
type TagsList struct {
	Tags  []string
	Total int
}

//
// Worker represents a Flotilla Worker
//
type Worker struct {
	WorkerType       string `json:"worker_type"`
	CountPerInstance int    `json:"count_per_instance"`
	Engine           string `json:"engine"`
}

//
// UpdateWith updates this definition with information from another
//
func (w *Worker) UpdateWith(other Worker) {
	if other.CountPerInstance >= 0 {
		w.CountPerInstance = other.CountPerInstance
	}
}

//
// WorkersList wraps a list of Workers
//
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

//
// TemplateList wraps a list of Templates
//
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
