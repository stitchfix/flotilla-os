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
	Initialize(conf config.Config) error
	CanBeRun(clusterName string, definition state.Definition) (bool, error)
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
		ecsm := &ECSClusterClient{}
		return ecsm, ecsm.Initialize(conf)
	default:
		return nil, fmt.Errorf("No ClusterClient named [%s] was found", name)
	}
}
