package exceptions

import (
	"errors"
)

type MalformedInput struct {
	ErrorString string
}

func (e MalformedInput) Error() string {
	return e.ErrorString
}

var DefinitionNotFound = errors.New("Referenced definition not found")
var DefinitionExists = errors.New("Definition with that alias or id already exists")
var ImageNotFound = errors.New("Referenced image does not exist in any of the configured repositories")
var ClusterConfigurationIssue = errors.New("Defintion cannot be run on the cluster specified")
