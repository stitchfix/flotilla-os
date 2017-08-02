package exceptions

import (
	"errors"
)

//
// MalformedInput describes malformed or otherwise incorrect input
//
type MalformedInput struct {
	ErrorString string
}

func (e MalformedInput) Error() string {
	return e.ErrorString
}

// ErrorDefinitionNotFound indicates that the definition does not exist in the source of state
var ErrorDefinitionNotFound = errors.New("Referenced definition not found")

// ErrorDefinitionExists indicates that the definition already exists
var ErrorDefinitionExists = errors.New("Definition with that alias or id already exists")

// ErrorImageNotFound indicates that the image ref does not exist in the configured docker image repos
var ErrorImageNotFound = errors.New("Referenced image does not exist in any of the configured repositories")

// ErrorClusterConfigurationIssue indicates that the cluster is not configured to run the run given
var ErrorClusterConfigurationIssue = errors.New("Defintion cannot be run on the cluster specified")

// ErrorReservedEnvironmentVariable indicates that one of the environment variables specified is reserved
var ErrorReservedEnvironmentVariable = errors.New("Using reserved environment variable")
