package state

import "time"

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

func IsValidStatus(status string) bool {
	return status == StatusRunning ||
		status == StatusQueued ||
		status == StatusNeedsRetry ||
		status == StatusPending ||
		status == StatusStopped
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
// Definition represents a definition of a job
// - roughly 1-1 with an AWS ECS task definition
//
type Definition struct {
	Arn           string     `json:"arn"`
	DefinitionID  string     `json:"definition_id"`
	Image         string     `json:"image"`
	GroupName     string     `json:"group_name"`
	ContainerName string     `json:"container_name"`
	User          string     `json:"user,omitempty"`
	Alias         string     `json:"alias"`
	Memory        *int       `json:"memory"`
	Command       string     `json:"command,omitempty"`
	TaskType      string     `json:"-"`
	Env           *EnvList   `json:"env"`
	Ports         *PortsList `json:"ports,omitempty"`
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
	if len(other.TaskType) > 0 {
		d.TaskType = other.TaskType
	}
	if other.Env != nil {
		d.Env = other.Env
	}
	if other.Ports != nil {
		d.Ports = other.Ports
	}
}

//
// DefinitionList wraps a list of Definitions
//
type DefinitionList struct {
	Total       int          `json:"total"`
	Definitions []Definition `json:"definitions"`
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
	TaskArn         string     `json:"task_arn"`
	RunID           string     `json:"run_id"`
	DefinitionID    string     `json:"definition_id"`
	ClusterName     string     `json:"cluster"`
	ExitCode        *int64     `json:"exit_code"`
	Status          string     `json:"status"`
	StartedAt       *time.Time `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	InstanceID      string     `json:"-"`
	InstanceDNSName string     `json:"-"`
	GroupName       string     `json:"group_name"`
	User            string     `json:"user,omitempty"`
	TaskType        string     `json:"-"`
	Env             *EnvList   `json:"env"`
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
	if len(other.ClusterName) > 0 {
		d.ClusterName = other.ClusterName
	}
	if other.ExitCode != nil {
		d.ExitCode = other.ExitCode
	}
	if len(other.Status) > 0 {
		d.Status = other.Status
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
}

//
// RunList wraps a list of Runs
//
type RunList struct {
	Total int   `json:"total"`
	Runs  []Run `json:"history"`
}
