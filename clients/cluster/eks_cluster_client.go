package cluster

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

// EKSClusterClient is the cluster client for EKS
// [NOTE] This client assumes the EKS cluster is capable is running a mixed varieties of jobs.
type EKSClusterClient struct{}

func (EKSClusterClient) Name() string {
	return ""
}

func (EKSClusterClient) Initialize(conf config.Config) error {
	return nil
}

// CanBeRun for EKSCluster is always true
func (EKSClusterClient) CanBeRun(clusterName string, executableResources state.ExecutableResources) (bool, error) {
	return true, nil
}

// Since it is a single cluster environment for EKS, slice of clusters is empty.
func (EKSClusterClient) ListClusters() ([]state.ClusterMetadata, error) {
	return []state.ClusterMetadata{}, nil
}
