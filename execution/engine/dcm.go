package engine

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	kubernetestrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	redisK8sConfigPrefix     = "flotilla:k8s_config:"
	redisMetricsConfigPrefix = "flotilla:metrics_config:"
	redisTokenPrefix         = "flotilla:eks_token:"
	defaultTokenTTL          = 15 * time.Minute
	defaultConfigTTL         = 60 * time.Minute
)

// ClusterConfig stores the serializable configuration for a k8s client
type ClusterConfig struct {
	Host      string `json:"host"`
	CAData    []byte `json:"ca_data"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
}

// DynamicClusterManager handles dynamic loading and caching of K8s clients
type DynamicClusterManager struct {
	mutex       sync.RWMutex
	log         flotillaLog.Logger
	eksClient   *eks.EKS
	awsRegion   string
	manager     state.Manager
	awsSession  *session.Session
	redisClient *redis.Client
	tokenTTL    time.Duration
	configTTL   time.Duration
}

// NewDynamicClusterManager creates a cluster manager that loads clusters from the state manager
func NewDynamicClusterManager(awsRegion string, log flotillaLog.Logger, manager state.Manager, redisClient *redis.Client) (*DynamicClusterManager, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))
	eksClient := eks.New(sess)

	return &DynamicClusterManager{
		log:         log,
		eksClient:   eksClient,
		awsRegion:   awsRegion,
		manager:     manager,
		awsSession:  sess,
		redisClient: redisClient,
		tokenTTL:    defaultTokenTTL,
		configTTL:   defaultConfigTTL,
	}, nil
}

// GetKubernetesClient returns a k8s client for the requested cluster, creating one if needed
func (dcm *DynamicClusterManager) GetKubernetesClient(clusterName string) (kubernetes.Clientset, error) {
	configKey := fmt.Sprintf("%s%s", redisK8sConfigPrefix, clusterName)
	cachedConfigStr, err := dcm.redisClient.Get(configKey).Result()

	var config *rest.Config

	if err == nil && cachedConfigStr != "" {
		var clusterConfig ClusterConfig
		if err := json.Unmarshal([]byte(cachedConfigStr), &clusterConfig); err == nil {
			if time.Now().Unix()-clusterConfig.Timestamp < int64(dcm.tokenTTL.Seconds()) {
				dcm.log.Log("message", "using cached config for cluster", "cluster", clusterName)
				config = &rest.Config{
					Host: clusterConfig.Host,
					TLSClientConfig: rest.TLSClientConfig{
						CAData: clusterConfig.CAData,
					},
					BearerToken: clusterConfig.Token,
				}
				config.WrapTransport = kubernetestrace.WrapRoundTripper
			}
		}
	}

	if config == nil {
		var err error
		clusters, _ := dcm.manager.ListClusterStates()
		for _, cluster := range clusters {
			if cluster.Name == clusterName {
				config, err = dcm.buildKubeConfig(cluster)
				break
			}
		}
		if err != nil {
			return kubernetes.Clientset{}, err
		}

		host := config.Host
		caData := config.TLSClientConfig.CAData
		token := config.BearerToken

		clusterConfig := ClusterConfig{
			Host:      host,
			CAData:    caData,
			Token:     token,
			Timestamp: time.Now().Unix(),
		}

		configBytes, err := json.Marshal(clusterConfig)
		if err == nil {
			err = dcm.redisClient.Set(configKey, string(configBytes), dcm.configTTL).Err()
			if err != nil {
				dcm.log.Log("message", "failed to cache config in Redis", "error", err.Error())
			}
		}
	}

	kClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return kubernetes.Clientset{}, errors.Wrap(err, "failed to create kubernetes client")
	}

	return *kClient, nil
}

// GetMetricsClient returns a metrics client for the requested cluster, creating one if needed
func (dcm *DynamicClusterManager) GetMetricsClient(clusterName string) (metricsv.Clientset, error) {
	configKey := fmt.Sprintf("%s%s", redisMetricsConfigPrefix, clusterName)
	cachedConfigStr, err := dcm.redisClient.Get(configKey).Result()

	var config *rest.Config

	if err == nil && cachedConfigStr != "" {
		var clusterConfig ClusterConfig
		if err := json.Unmarshal([]byte(cachedConfigStr), &clusterConfig); err == nil {
			if time.Now().Unix()-clusterConfig.Timestamp < int64(dcm.tokenTTL.Seconds()) {
				dcm.log.Log("message", "using cached metrics config for cluster", "cluster", clusterName)
				config = &rest.Config{
					Host: clusterConfig.Host,
					TLSClientConfig: rest.TLSClientConfig{
						CAData: clusterConfig.CAData,
					},
					BearerToken: clusterConfig.Token,
				}
				config.WrapTransport = kubernetestrace.WrapRoundTripper
			}
		}
	}

	if config == nil {
		var err error
		clusters, _ := dcm.manager.ListClusterStates()
		for _, cluster := range clusters {
			if cluster.Name == clusterName {
				config, err = dcm.buildKubeConfig(cluster)
				break
			}
		}
		if err != nil {
			return metricsv.Clientset{}, err
		}

		host := config.Host
		caData := config.TLSClientConfig.CAData
		token := config.BearerToken

		clusterConfig := ClusterConfig{
			Host:      host,
			CAData:    caData,
			Token:     token,
			Timestamp: time.Now().Unix(),
		}

		configBytes, err := json.Marshal(clusterConfig)
		if err == nil {
			err = dcm.redisClient.Set(configKey, string(configBytes), dcm.configTTL).Err()
			if err != nil {
				dcm.log.Log("message", "failed to cache metrics config in Redis", "error", err.Error())
			}
		}
	}

	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		return metricsv.Clientset{}, err
	}

	return *metricsClient, nil
}

// PreloadClusterClients preloads configs for all active clusters
func (dcm *DynamicClusterManager) PreloadClusterClients() error {
	clusters, err := dcm.manager.ListClusterStates()
	if err != nil {
		return errors.Wrap(err, "failed to list clusters for preloading")
	}

	for _, cluster := range clusters {
		if cluster.Status == state.StatusActive {
			config, err := dcm.buildKubeConfig(cluster)
			if err != nil {
				dcm.log.Log("message", "failed to preload config", "cluster", cluster.Name, "error", err.Error())
				continue
			}

			host := config.Host
			caData := config.TLSClientConfig.CAData
			token := config.BearerToken

			clusterConfig := ClusterConfig{
				Host:      host,
				CAData:    caData,
				Token:     token,
				Timestamp: time.Now().Unix(),
			}

			configBytes, err := json.Marshal(clusterConfig)
			if err == nil {
				k8sConfigKey := fmt.Sprintf("%s%s", redisK8sConfigPrefix, cluster.Name)
				err = dcm.redisClient.Set(k8sConfigKey, string(configBytes), dcm.configTTL).Err()
				if err != nil {
					dcm.log.Log("message", "failed to cache k8s config in Redis", "cluster", cluster.Name, "error", err.Error())
				}

				metricsConfigKey := fmt.Sprintf("%s%s", redisMetricsConfigPrefix, cluster.Name)
				err = dcm.redisClient.Set(metricsConfigKey, string(configBytes), dcm.configTTL).Err()
				if err != nil {
					dcm.log.Log("message", "failed to cache metrics config in Redis", "cluster", cluster.Name, "error", err.Error())
				}
			}
		}
	}

	return nil
}

// InitializeWithStaticClusters initializes the manager with a set of static clusters
func (dcm *DynamicClusterManager) InitializeWithStaticClusters(clusters []string, kubeConfigBasePath string) error {
	for _, clusterName := range clusters {
		filename := fmt.Sprintf("%s/%s", kubeConfigBasePath, clusterName)
		clientConf, err := clientcmd.BuildConfigFromFlags("", filename)
		if err != nil {
			return err
		}

		clusterConfig := ClusterConfig{
			Host:      clientConf.Host,
			CAData:    clientConf.TLSClientConfig.CAData,
			Token:     clientConf.BearerToken,
			Timestamp: time.Now().Unix(),
		}

		configBytes, err := json.Marshal(clusterConfig)
		if err == nil {
			k8sConfigKey := fmt.Sprintf("%s%s", redisK8sConfigPrefix, clusterName)
			err = dcm.redisClient.Set(k8sConfigKey, string(configBytes), dcm.configTTL).Err()
			if err != nil {
				dcm.log.Log("message", "failed to cache k8s config in Redis", "cluster", clusterName, "error", err.Error())
			}

			metricsConfigKey := fmt.Sprintf("%s%s", redisMetricsConfigPrefix, clusterName)
			err = dcm.redisClient.Set(metricsConfigKey, string(configBytes), dcm.configTTL).Err()
			if err != nil {
				dcm.log.Log("message", "failed to cache metrics config in Redis", "cluster", clusterName, "error", err.Error())
			}
		}
	}

	return nil
}

// buildKubeConfig creates a rest.Config for the given cluster
func (dcm *DynamicClusterManager) buildKubeConfig(clusterMetadata state.ClusterMetadata) (*rest.Config, error) {
	_, err := dcm.manager.GetClusterByID(clusterMetadata.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "cluster %s not found in system", clusterMetadata.Name)
	}

	result, err := dcm.eksClient.DescribeCluster(&eks.DescribeClusterInput{
		Name: aws.String(clusterMetadata.Name),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to describe EKS cluster %s", clusterMetadata.Name)
	}

	cluster := result.Cluster
	if cluster == nil {
		return nil, fmt.Errorf("cluster %s not found in AWS", clusterMetadata.Name)
	}

	token, err := dcm.getClusterToken(clusterMetadata.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get token for cluster %s", clusterMetadata.Name)
	}

	certData, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	dcm.log.Log("message", "certificate data", "data_length", len(certData), "data_prefix", certData[:20])
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode certificate data")
	}

	config := &rest.Config{
		Host: *cluster.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: certData,
		},
		BearerToken: token,
	}

	config.WrapTransport = kubernetestrace.WrapRoundTripper
	return config, nil
}

// getClusterToken gets an authentication token for the EKS cluster, using Redis cache if available
func (dcm *DynamicClusterManager) getClusterToken(clusterName string) (string, error) {
	tokenKey := fmt.Sprintf("%s%s", redisTokenPrefix, clusterName)

	token, err := dcm.redisClient.Get(tokenKey).Result()
	if err == nil && token != "" {
		dcm.log.Log("message", "using cached token for cluster", "cluster", clusterName)
		return token, nil
	}

	dcm.log.Log("message", "generating new token for cluster", "cluster", clusterName)

	stsClient := sts.New(dcm.awsSession)

	request, _ := stsClient.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	presignedURL, err := request.Presign(15 * time.Minute)
	if err != nil {
		return "", errors.Wrap(err, "failed to presign request")
	}

	token = fmt.Sprintf("k8s-aws-v1.%s", presignedURL)

	err = dcm.redisClient.Set(tokenKey, token, dcm.tokenTTL).Err()
	if err != nil {
		dcm.log.Log("message", "failed to cache token in Redis", "error", err.Error())
	}

	return token, nil
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
	result, err := dcm.eksClient.DescribeCluster(&eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		return nil, err
	}

	cluster := result.Cluster
	if cluster == nil {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}

	kubeConfig := api.NewConfig()
	kubeConfig.Clusters[clusterName] = &api.Cluster{
		Server:                   *cluster.Endpoint,
		CertificateAuthorityData: []byte(*cluster.CertificateAuthority.Data),
	}

	kubeConfig.AuthInfos[clusterName] = &api.AuthInfo{
		Exec: &api.ExecConfig{
			APIVersion: "client.authentication.k8s.io/v1beta1",
			Command:    "aws",
			Args: []string{
				"eks",
				"get-token",
				"--cluster-name",
				clusterName,
				"--region",
				dcm.awsRegion,
			},
		},
	}

	kubeConfig.Contexts[clusterName] = &api.Context{
		Cluster:  clusterName,
		AuthInfo: clusterName,
	}

	kubeConfig.CurrentContext = clusterName

	return clientcmd.NewDefaultClientConfig(*kubeConfig, &clientcmd.ConfigOverrides{}), nil
}

// InvalidateClusterClient removes a client config from cache to force recreation
func (dcm *DynamicClusterManager) InvalidateClusterClient(clusterName string) {
	k8sConfigKey := fmt.Sprintf("%s%s", redisK8sConfigPrefix, clusterName)
	metricsConfigKey := fmt.Sprintf("%s%s", redisMetricsConfigPrefix, clusterName)
	tokenKey := fmt.Sprintf("%s%s", redisTokenPrefix, clusterName)

	dcm.redisClient.Del(k8sConfigKey)
	dcm.redisClient.Del(metricsConfigKey)
	dcm.redisClient.Del(tokenKey)

	dcm.log.Log("message", "invalidated cluster client cache", "cluster", clusterName)
}
