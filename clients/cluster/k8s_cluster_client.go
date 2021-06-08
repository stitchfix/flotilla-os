package cluster

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

//
// K8SClusterClient is the cluster client for K8S
// [NOTE] This client assumes the K8S cluster is capable is running a mixed varieties of jobs.
//
type K8SClusterClient struct{}

func (K8SClusterClient) Name() string {
	return ""
}

func (K8SClusterClient) Initialize(conf config.Config) error {
	return nil
}

// CanBeRun for K8SCluster is always true
func (K8SClusterClient) CanBeRun(clusterName string, executableResources state.ExecutableResources) (bool, error) {
	return true, nil
}

// Since it is a single cluster environment for K8S, slice of clusters is empty.
func (K8SClusterClient) ListClusters() ([]string, error) {
	return []string{}, nil
}
