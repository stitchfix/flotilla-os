package state

type EnvList []EnvVar
type PortsList []int

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Definition struct {
	Arn           string    `json:"arn"`
	DefinitionID  string    `json:"definition_id"`
	Image         string    `json:"image"`
	GroupName     string    `json:"group_name"`
	ContainerName string    `json:"container_name"`
	User          string    `json:"user,omitempty"`
	Alias         string    `json:"alias"`
	Memory        int       `json:"memory"`
	Command       string    `json:"command,omitempty"`
	TaskType      string    `json:"-"`
	Env           EnvList   `json:"env"`
	Ports         PortsList `json:"ports,omitempty"`
}

type DefinitionList struct {
	Total int                `json:"total"`
	Definitions []Definition `json:"definitions"`
}