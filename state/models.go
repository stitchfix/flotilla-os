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
	Memory        *int       `json:"memory"`
	Command       string    `json:"command,omitempty"`
	TaskType      string    `json:"-"`
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
		// 0 is invalid anyway; we could use a pointer to be -even- clearer
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
	Total int                `json:"total"`
	Definitions []Definition `json:"definitions"`
}
