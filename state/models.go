package state

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

//
// Custom type for environment variables
//
type JsonEnvironment []EnvVar

func (e *JsonEnvironment) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}
func (e JsonEnvironment) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

//
// Custom type for ports
//
type JsonPorts []int

func (e *JsonPorts) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}
func (e JsonPorts) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Definition struct {
	Arn           sql.NullString  `db:"arn" json:"arn"`
	DefinitionID  string          `db:"definition_id" json:"definition_id"`
	Image         string          `db:"image" json:"image"`
	GroupName     string          `db:"group_name" json:"group_name"`
	ContainerName string          `db:"container_name" json:"container_name"`
	User          sql.NullString  `db:"user" json:"user,omitempty"`
	Alias         string          `db:"alias" json:"alias"`
	Memory        int             `db:"memory" json:"memory"`
	Command       sql.NullString  `db:"command" json:"command,omitempty"`
	TaskType      sql.NullString  `db:"task_type" json:"-"`
	Env           JsonEnvironment `db:"env" json:"env"`
	Ports         JsonPorts       `db:"ports" json:"ports,omitempty"`
}

type DefinitionList struct {
	Total int                `json:"total"`
	Definitions []Definition `json:"definitions"`
}