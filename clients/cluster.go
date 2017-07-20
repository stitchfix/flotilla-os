package clients

import "github.com/stitchfix/flotilla-os/state"

//
// ClusterClient validates whether or not the given definition
// can be run on the specified cluster. This is to prevent
// infinite queue times - the case that the requested resources
// will -never- become available on the user's chosen cluster
//
type ClusterClient interface {
	CanBeRun(clusterName string, definition state.Definition) (bool, error)
}

//
// Default cluster client maintains a cached map[string]instanceResources
// which is used to check that the resources requested by a definition
// -at some point- could become available on the specified cluster.
//
type clusterClient struct {
	clusters map[string]instanceResources
}

type instanceResources struct {
	memory int
	cpu    int
	ports  []string
}

func (cc *clusterClient) CanBeRun(clusterName string, definition state.Definition) (bool, error) {
	return false, nil
}
