package engine

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/pkg/errors"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	kubernetestrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// DynamicClusterManager handles dynamic loading of K8s clients
type DynamicClusterManager struct {
	mutex      sync.RWMutex
	log        flotillaLog.Logger
	eksClient  *eks.EKS
	awsRegion  string
	manager    state.Manager
	awsSession *session.Session
}

// getKubeconfigBaseDir returns the base directory for kubeconfig files
func getKubeconfigBaseDir() string {
	dir := os.Getenv("EKS_KUBECONFIG_BASEPATH")
	if dir != "" {
		dir, _ = os.Getwd()
	}
	return dir
}

// NewDynamicClusterManager creates a cluster manager that loads clusters from the state manager
func NewDynamicClusterManager(awsRegion string, log flotillaLog.Logger, manager state.Manager) (*DynamicClusterManager, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))
	eksClient := eks.New(sess)

	return &DynamicClusterManager{
		log:        log,
		eksClient:  eksClient,
		awsRegion:  awsRegion,
		manager:    manager,
		awsSession: sess,
	}, nil
}

// getOrCreateKubeconfig ensures a valid kubeconfig exists for the given cluster
func (dcm *DynamicClusterManager) getOrCreateKubeconfig(clusterName string) (string, error) {
	kubeconfigBaseDir := getKubeconfigBaseDir()
	kubeconfigPath := filepath.Join(kubeconfigBaseDir, clusterName)

	if _, err := os.Stat(kubeconfigBaseDir); os.IsNotExist(err) {
		if err := os.MkdirAll(kubeconfigBaseDir, 0755); err != nil {
			return "", errors.Wrap(err, "failed to create directory for kubeconfigs")
		}
	}

	needsGeneration := false
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		needsGeneration = true
	} else {
		_, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			needsGeneration = true
		}
	}

	if needsGeneration {
		if err := dcm.generateKubeconfig(clusterName, kubeconfigPath); err != nil {
			return "", err
		}
	}

	return kubeconfigPath, nil
}

// generateKubeconfig creates a kubeconfig file for the specified cluster
func (dcm *DynamicClusterManager) generateKubeconfig(clusterName, kubeconfigPath string) error {
	cmd := exec.Command("aws", "eks", "update-kubeconfig",
		"--name", clusterName,
		"--region", dcm.awsRegion,
		"--kubeconfig", kubeconfigPath)

	if output, err := cmd.CombinedOutput(); err != nil {
		dcm.log.Log("level", "error", "message", "Failed to generate kubeconfig",
			"cluster", clusterName,
			"error", err.Error(),
			"output", string(output))
		return errors.Wrapf(err, "failed to generate kubeconfig: %s", string(output))
	}

	dcm.log.Log("level", "info", "message", "Successfully generated kubeconfig",
		"cluster", clusterName,
		"path", kubeconfigPath)
	return nil
}

// createRestConfig builds a rest.Config from a kubeconfig path
func (dcm *DynamicClusterManager) createRestConfig(kubeconfigPath string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load kubeconfig")
	}

	config.WrapTransport = kubernetestrace.WrapRoundTripper
	return config, nil
}

// GetKubernetesClient returns a k8s client for the requested cluster
func (dcm *DynamicClusterManager) GetKubernetesClient(clusterName string) (kubernetes.Clientset, error) {
	kubeconfigPath, err := dcm.getOrCreateKubeconfig(clusterName)
	if err != nil {
		return kubernetes.Clientset{}, err
	}

	config, err := dcm.createRestConfig(kubeconfigPath)
	if err != nil {
		return kubernetes.Clientset{}, err
	}

	kClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return kubernetes.Clientset{}, errors.Wrap(err, "failed to create kubernetes client")
	}

	return *kClient, nil
}

// GetMetricsClient returns a metrics client for the requested cluster
func (dcm *DynamicClusterManager) GetMetricsClient(clusterName string) (metricsv.Clientset, error) {
	kubeconfigPath, err := dcm.getOrCreateKubeconfig(clusterName)
	if err != nil {
		return metricsv.Clientset{}, err
	}

	config, err := dcm.createRestConfig(kubeconfigPath)
	if err != nil {
		return metricsv.Clientset{}, err
	}

	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		return metricsv.Clientset{}, errors.Wrap(err, "failed to create metrics client")
	}

	return *metricsClient, nil
}

// InitializeClusters handles both static and dynamic cluster configurations
func (dcm *DynamicClusterManager) InitializeClusters(ctx context.Context, staticClusters []string) error {
	kubeconfigBaseDir := getKubeconfigBaseDir()
	if err := os.MkdirAll(kubeconfigBaseDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create directory for kubeconfigs")
	}

	// Initialize static clusters
	for _, clusterName := range staticClusters {
		kubeconfigPath := filepath.Join(kubeconfigBaseDir, clusterName)
		if err := dcm.generateKubeconfig(clusterName, kubeconfigPath); err != nil {
			dcm.log.Log("level", "error", "message", "Failed to initialize static cluster",
				"cluster", clusterName,
				"error", err.Error())
		}
	}

	// Initialize dynamic clusters from state manager
	clusters, err := dcm.manager.ListClusterStates(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list clusters")
	}

	for _, cluster := range clusters {
		if cluster.Status == state.StatusActive {
			kubeconfigPath := filepath.Join(kubeconfigBaseDir, cluster.Name)
			if err := dcm.generateKubeconfig(cluster.Name, kubeconfigPath); err != nil {
				dcm.log.Log("level", "error", "message", "Failed to initialize dynamic cluster",
					"cluster", cluster.Name,
					"error", err.Error())
			}
		}
	}

	return nil
}
