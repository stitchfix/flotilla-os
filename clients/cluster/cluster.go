package cluster

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

//
// Client validates whether or not the given definition can be run
// on the specified cluster. This is to prevent infinite queue
// times - the case that the requested resources will -never- become
// available on the user's chosen cluster
//
type Client interface {
	Name() string
	Initialize(conf config.Config) error
	CanBeRun(clusterName string, definition state.Definition) (bool, error)
	ListClusters() ([]string, error)
}

//
// NewClusterClient returns a cluster client
//
func NewClusterClient(conf config.Config) (Client, error) {
	name := conf.GetString("cluster_client")
	if len(name) == 0 {
		name = "ecs"
	}

	switch name {
	case "ecs":
		ecsc := &ECSClusterClient{}
		return ecsc, ecsc.Initialize(conf)
	default:
		return nil, fmt.Errorf("No Client named [%s] was found", name)
	}
}
