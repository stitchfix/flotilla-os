package state

import "time"

type EnvList []EnvVar
type PortsList []int

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

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

func (d *Definition) updateWith(other Definition) {
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

type DefinitionList struct {
	Total       int          `json:"total"`
	Definitions []Definition `json:"definitions"`
}

type Run struct {
	TaskArn         string     `json:"task_arn"`
	RunID           string     `json:"run_id"`
	DefinitionID    string     `json:"definition_id"`
	ClusterName     string     `json:"cluster"`
	ExitCode        *int       `json:"exit_code"`
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

func (d *Run) updateWith(other Run) {
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

type RunList struct {
	Total int   `json:"total"`
	Runs  []Run `json:"history"`
}
