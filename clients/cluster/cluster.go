package cluster

import (
	"fmt"
	"github.com/pkg/errors"
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
	CanBeRun(clusterName string, executableResources state.ExecutableResources) (bool, error)
	ListClusters() ([]string, error)
}

//
// NewClusterClient returns a cluster client
//
func NewClusterClient(conf config.Config, name string) (Client, error) {
	switch name {
	case "ecs":
		ecsc := &ECSClusterClient{}
		if err := ecsc.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing ECSClusterClient")
		}
		return ecsc, nil
	case "eks":
		eksc := &EKSClusterClient{}
		if err := eksc.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing EKSClusterClient")
		}
		return eksc, nil
	default:
		return nil, fmt.Errorf("No Client named [%s] was found", name)
	}
}
