package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"text/template"
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

var ECSEngine = "ecs"

var EKSEngine = "eks"

var DefaultEngine = ECSEngine

var MinCPU = int64(125)

var MinMem = int64(125)

var TTLSecondsAfterFinished = int32(3600)

var ActiveDeadlineSeconds = int64(86400)

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

//
// Tags wraps a list of strings
// - abstraction to make it easier to read
//   and write to db
//
type Tags []string

// Definition represents a definition of a job
// - roughly 1-1 with an AWS ECS task definition
//
type Definition struct {
	Arn              string     `json:"arn"`
	DefinitionID     string     `json:"definition_id"`
	Image            string     `json:"image"`
	GroupName        string     `json:"group_name"`
	ContainerName    string     `json:"container_name"`
	User             string     `json:"user,omitempty"`
	Alias            string     `json:"alias"`
	Memory           *int64     `json:"memory"`
	Gpu              *int64     `json:"gpu,omitempty"`
	Cpu              *int64     `json:"cpu,omitempty"`
	Command          string     `json:"command,omitempty"`
	TaskType         string     `json:"-"`
	Env              *EnvList   `json:"env"`
	Ports            *PortsList `json:"ports,omitempty"`
	Tags             *Tags      `json:"tags,omitempty"`
	Privileged       *bool      `json:"privileged,omitempty"`
	SharedMemorySize *int64     `json:"sharedMemorySize,omitempty"`
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
	TaskArn          string     `json:"task_arn"`
	RunID            string     `json:"run_id"`
	DefinitionID     string     `json:"definition_id"`
	Alias            string     `json:"alias"`
	Image            string     `json:"image"`
	ClusterName      string     `json:"cluster"`
	ExitCode         *int64     `json:"exit_code,omitempty"`
	Status           string     `json:"status"`
	QueuedAt         *time.Time `json:"queued_at,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
	InstanceID       string     `json:"-"`
	InstanceDNSName  string     `json:"-"`
	GroupName        string     `json:"group_name"`
	User             string     `json:"user,omitempty"`
	TaskType         string     `json:"-"`
	Env              *EnvList   `json:"env,omitempty"`
	Command          *string    `json:"command,omitempty"`
	Memory           *int64     `json:"memory,omitempty"`
	Cpu              *int64     `json:"cpu,omitempty"`
	Gpu              *int64     `json:"gpu,omitempty"`
	ExitReason       *string    `json:"exit_reason,omitempty"`
	Engine           *string    `json:"engine,omitempty"`
	NodeLifecycle    *string    `json:"node_lifecycle,omitempty"`
	EphemeralStorage *int64     `json:"ephemeral_storage,omitempty"`
	PodName          *string    `json:"pod_name,omitempty"`
	Namespace        *string    `json:"namespace,omitempty"`
	ContainerName    *string    `json:"container_name,omitempty"`
	MaxMemoryUsed    *int64     `json:"max_memory_used,omitempty"`
	MaxCpuUsed       *int64     `json:"max_cpu_used,omitempty"`
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
	return json.Marshal(&struct {
		Instance map[string]string `json:"instance"`
		Alias
	}{
		Instance: instance,
		Alias:    (Alias)(r),
	})
}

//
// RunList wraps a list of Runs
//
type RunList struct {
	Total int   `json:"total"`
	Runs  []Run `json:"history"`
}

type RunEventList struct {
	Total     int        `json:"total"`
	RunEvents []RunEvent `json:"run_events"`
}

type RunEvent struct {
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
