package cluster

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

type EKSClusterClient struct {
	clusterName string
}

func (ecc *EKSClusterClient) Name() string {
	return "eks"
}

func (ecc *EKSClusterClient) Initialize(conf config.Config) error {
	name := "default"
	if conf.IsSet("experimental__cluster_name") {
		name = conf.GetString("experimental__cluster_name")
	}
	ecc.clusterName = name
	return nil
}

func (ecc *EKSClusterClient) CanBeRun(clusterName string, definition state.Definition) (bool, error) {
	return true, nil
}

func (ecc *EKSClusterClient) ListClusters() ([]string, error) {
	return []string{ecc.clusterName}, nil
}
