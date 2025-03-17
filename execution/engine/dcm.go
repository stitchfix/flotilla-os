package engine

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/pkg/errors"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	kubernetestrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes"
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

func getKubeconfigBaseDir() string {
	dir := os.Getenv("KUBECONFIG_BASE_DIR")
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

// GetKubernetesClient returns a k8s client for the requested cluster
func (dcm *DynamicClusterManager) GetKubernetesClient(clusterName string) (kubernetes.Clientset, error) {
	kubeconfigBaseDir := getKubeconfigBaseDir()
	kubeconfigPath := filepath.Join(kubeconfigBaseDir, clusterName)

	needsGeneration := false
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		needsGeneration = true
	} else {
		// Test if the existing kubeconfig works by trying to load it
		_, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			needsGeneration = true
		}
	}

	if needsGeneration {
		if err := os.MkdirAll(kubeconfigBaseDir, 0755); err != nil {
			return kubernetes.Clientset{}, errors.Wrap(err, "failed to create directory for kubeconfigs")
		}

		cmd := exec.Command("aws", "eks", "update-kubeconfig",
			"--name", clusterName,
			"--region", dcm.awsRegion,
			"--kubeconfig", kubeconfigPath)

		if output, err := cmd.CombinedOutput(); err != nil {
			dcm.log.Log("message", "Failed to generate kubeconfig",
				"cluster", clusterName,
				"error", err.Error(),
				"output", string(output))
			return kubernetes.Clientset{}, errors.Wrapf(err, "failed to generate kubeconfig: %s", string(output))
		}

		dcm.log.Log("message", "Generated new kubeconfig", "cluster", clusterName, "path", kubeconfigPath)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return kubernetes.Clientset{}, errors.Wrap(err, "failed to load kubeconfig")
	}

	config.WrapTransport = kubernetestrace.WrapRoundTripper

	kClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return kubernetes.Clientset{}, errors.Wrap(err, "failed to create kubernetes client")
	}

	return *kClient, nil
}

// GetMetricsClient returns a metrics client for the requested cluster
func (dcm *DynamicClusterManager) GetMetricsClient(clusterName string) (metricsv.Clientset, error) {
	kubeconfigBaseDir := getKubeconfigBaseDir()
	kubeconfigPath := filepath.Join(kubeconfigBaseDir, clusterName)

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
		if err := os.MkdirAll(kubeconfigBaseDir, 0755); err != nil {
			return metricsv.Clientset{}, errors.Wrap(err, "failed to create directory for kubeconfigs")
		}

		cmd := exec.Command("aws", "eks", "update-kubeconfig",
			"--name", clusterName,
			"--region", dcm.awsRegion,
			"--kubeconfig", kubeconfigPath)

		if output, err := cmd.CombinedOutput(); err != nil {
			dcm.log.Log("message", "Failed to generate kubeconfig",
				"cluster", clusterName,
				"error", err.Error(),
				"output", string(output))
			return metricsv.Clientset{}, errors.Wrapf(err, "failed to generate kubeconfig: %s", string(output))
		}

		dcm.log.Log("message", "Generated new kubeconfig", "cluster", clusterName, "path", kubeconfigPath)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return metricsv.Clientset{}, errors.Wrap(err, "failed to load kubeconfig")
	}

	config.WrapTransport = kubernetestrace.WrapRoundTripper

	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		return metricsv.Clientset{}, errors.Wrap(err, "failed to create metrics client")
	}

	return *metricsClient, nil
}

// InitializeClusters handles both static and dynamic cluster configurations
func (dcm *DynamicClusterManager) InitializeClusters(staticClusters []string) error {
	kubeconfigBaseDir := getKubeconfigBaseDir()
	if err := os.MkdirAll(kubeconfigBaseDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create directory for kubeconfigs")
	}

	generateKubeconfig := func(clusterName string) error {
		kubeconfigPath := filepath.Join(kubeconfigBaseDir, clusterName)
		cmd := exec.Command("aws", "eks", "update-kubeconfig",
			"--name", clusterName,
			"--region", dcm.awsRegion,
			"--kubeconfig", kubeconfigPath)

		if output, err := cmd.CombinedOutput(); err != nil {
			dcm.log.Log("message", "Failed to generate kubeconfig",
				"cluster", clusterName,
				"error", err.Error(),
				"output", string(output))
			return err
		}

		dcm.log.Log("message", "Successfully initialized kubeconfig",
			"cluster", clusterName,
			"path", kubeconfigPath)
		return nil
	}

	for _, clusterName := range staticClusters {
		if err := generateKubeconfig(clusterName); err != nil {
			continue
		}
	}

	clusters, err := dcm.manager.ListClusterStates()
	if err != nil {
		return errors.Wrap(err, "failed to list clusters")
	}

	for _, cluster := range clusters {
		if cluster.Status == state.StatusActive {
			if err := generateKubeconfig(cluster.Name); err != nil {
				continue
			}
		}
	}

	return nil
}

// GetClusters returns a list of all active cluster names
func (dcm *DynamicClusterManager) GetClusters() ([]string, error) {
	clusters, err := dcm.manager.ListClusterStates()
	if err != nil {
		return nil, err
	}

	var clusterNames []string
	for _, cluster := range clusters {
		if cluster.Status == state.StatusActive {
			clusterNames = append(clusterNames, cluster.Name)
		}
	}

	return clusterNames, nil
}

// PrepareKubeConfigFromCluster creates a clientcmd.ClientConfig from cluster details
func (dcm *DynamicClusterManager) PrepareKubeConfigFromCluster(clusterName string) (clientcmd.ClientConfig, error) {
	kubeconfigBaseDir := getKubeconfigBaseDir()
	kubeconfigPath := filepath.Join(kubeconfigBaseDir, clusterName)
	cmd := exec.Command("aws", "eks", "update-kubeconfig",
		"--name", clusterName,
		"--region", dcm.awsRegion,
		"--kubeconfig", kubeconfigPath)

	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, errors.Wrapf(err, "failed to generate kubeconfig: %s", string(output))
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{},
	), nil
}
