package cluster

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

type EKSClusterClient struct{}

func (EKSClusterClient) Name() string {
return ""
}

func (EKSClusterClient) Initialize(conf config.Config) error {
return nil
}

func (EKSClusterClient) CanBeRun(clusterName string, definition state.Definition) (bool, error) {
return true, nil
}

func (EKSClusterClient) ListClusters() ([]string, error) {
return []string{}, nil
}
