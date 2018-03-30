package cluster

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	"log"
	"strings"
	"time"
)

var LIST_CLUSTER_RESULTS int64 = 10

//
// ECSClusterClient is the default cluster client and maintains a
// cached map[string]instanceResources which is used to check that
// the resources requested by a definition  -at some point- could
// become available on the specified cluster.
//
// [NOTE] This client assumes homogenous clusters
//
type ECSClusterClient struct {
	ecsClient    resourceClient
	clusters     resourceCache
	clusterNames clusterNamesCache
}

type resourceClient interface {
	ListClusters(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error)
	DescribeClusters(input *ecs.DescribeClustersInput) (*ecs.DescribeClustersOutput, error)
	ListContainerInstances(input *ecs.ListContainerInstancesInput) (*ecs.ListContainerInstancesOutput, error)
	DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
}

type instanceResources struct {
	memory int64
	cpu    int64
}

//
// Name is the name of the client
//
func (ecc *ECSClusterClient) Name() string {
	return "ecs"
}

//
// Initialize the ecs cluster client with config
//
func (ecc *ECSClusterClient) Initialize(conf config.Config) error {
	if !conf.IsSet("aws_default_region") {
		return errors.Errorf("ecsClusterClient needs [aws_default_region] set in config")
	}

	flotillaMode := conf.GetString("flotilla_mode")
	if flotillaMode != "test" {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(conf.GetString("aws_default_region"))}))

		ecc.ecsClient = ecs.New(sess)
	}
	ecc.clusters = resourceCache{
		duration:      15 * time.Minute,
		internalCache: cache.New(15*time.Minute, 5*time.Minute),
	}
	ecc.clusterNames = clusterNamesCache{
		duration:      15 * time.Minute,
		internalCache: cache.New(15*time.Minute, 5*time.Minute),
	}
	return nil
}

//
// CanBeRun determines whether a task formed from the specified definition
// can be run on clusterName
//
func (ecc *ECSClusterClient) CanBeRun(clusterName string, definition state.Definition) (bool, error) {
	var (
		resources *instanceResources
		err       error
	)
	resources, found := ecc.clusters.getInstanceResources(clusterName)
	if !found {
		resources, err = ecc.fetchResources(clusterName)
		if err != nil {
			return false, errors.Wrapf(err, "problem getting available resources for cluster [%s]", clusterName)
		}
		if resources == nil {
			return false, nil
		}
		ecc.clusters.setInstanceResources(clusterName, resources)
	}

	return ecc.validate(resources, definition), nil
}

//
// ListClusters gets a list of cluster names
//
func (ecc *ECSClusterClient) ListClusters() ([]string, error) {
	clusterNames, found := ecc.clusterNames.getClusterNames()
	if found {
		return clusterNames, nil
	}
	clusterNames, err := ecc.getClusterNamesFromApi()
	if err != nil {
		return clusterNames, errors.Wrap(err, "problem listing cluster names")
	}
	ecc.clusterNames.setInstanceResources(clusterNames)
	return clusterNames, nil
}

func (ecc *ECSClusterClient) getClusterNamesFromApi() ([]string, error) {
	var nextToken *string
	clusterNames := make([]string, 0, 100)
	for {
		clusterArns, err := ecc.ecsClient.ListClusters(
			&ecs.ListClustersInput{
				MaxResults: &LIST_CLUSTER_RESULTS,
				NextToken:  nextToken})
		if err != nil {
			return nil, errors.Wrap(err, "problem listing ecs clusters")
		}
		clusters, err := ecc.ecsClient.DescribeClusters(
			&ecs.DescribeClustersInput{Clusters: clusterArns.ClusterArns})
		if err != nil {
			return nil, errors.Wrap(err, "problem describing ecs clusters")
		}
		for _, cluster := range clusters.Clusters {
			if cluster.ClusterName != nil {
				clusterNames = append(clusterNames, *cluster.ClusterName)
			} else {
				log.Printf("Nil cluster name in cluster %v", cluster)
			}
		}
		if clusterArns.NextToken == nil {
			break
		} else {
			nextToken = clusterArns.NextToken
		}
	}
	return clusterNames, nil
}

func (ecc *ECSClusterClient) validate(resources *instanceResources, definition state.Definition) bool {
	if resources != nil && definition.Memory != nil && int64(*definition.Memory) < resources.memory {
		// TODO - check cpu when available on the definition
		return true
	}
	return false
}

func (ecc *ECSClusterClient) fetchResources(clusterName string) (*instanceResources, error) {
	exists, err := ecc.clusterExists(clusterName)
	if err != nil {
		return nil, errors.Wrapf(err, "problem checking for cluster existence of cluster [%s]", clusterName)
	}
	if exists {
		rsrc, err := ecc.clusterInstanceResources(clusterName)
		if err != nil {
			return nil, errors.Wrap(err, "problem fetching cluster resources")
		}
		return rsrc, nil
	}
	return nil, nil
}

func (ecc *ECSClusterClient) clusterExists(clusterName string) (bool, error) {
	result, err := ecc.ecsClient.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			&clusterName,
		},
	})

	if err != nil {
		return false, errors.Wrapf(err, "problem describing ecs cluster with name: [%s]", clusterName)
	}
	if len(result.Failures) != 0 {
		msg := make([]string, len(result.Failures))
		for i, failure := range result.Failures {
			msg[i] = *failure.Reason
		}
		return false, errors.Errorf("ERRORS: %s", strings.Join(msg, "\n"))
	}
	if len(result.Clusters) == 0 {
		return false, nil
	}
	return true, nil
}

func (ecc *ECSClusterClient) clusterInstanceResources(clusterName string) (*instanceResources, error) {
	var result instanceResources

	instances, err := ecc.listInstances(&ecs.ListContainerInstancesInput{
		Cluster: &clusterName,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "problem listing ecs container instances for cluster [%s]", clusterName)
	}

	if len(instances) == 0 {
		return nil, nil // short-circuit to avoid additional spurious api call with zero instances
	}

	resources, err := ecc.describeInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            &clusterName,
		ContainerInstances: instances,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "problem describing container instances for cluster [%s]", clusterName)
	}

	//
	// Assumes -all- instances in the cluster have identical registered resources
	// While this is certainly the most manageable way to configure ecs clusters
	// it is -not- the only way in practice
	//
	if resources != nil && len(resources) > 0 {
		// An alternative, and potentially better handling of heterogeneous clusters,
		// is to return the resources with the -highest- memory and cpu
		result = resources[0]
	}
	return &result, nil
}

func (ecc *ECSClusterClient) listInstances(input *ecs.ListContainerInstancesInput) ([]*string, error) {
	result, err := ecc.ecsClient.ListContainerInstances(input)
	if err != nil {
		return nil, errors.Wrap(err, "problem listing container instances")
	}
	var subset []*string
	for _, arn := range result.ContainerInstanceArns {
		subset = append(subset, arn)
	}

	if result.NextToken != nil {
		input.NextToken = result.NextToken
		more, err := ecc.listInstances(input)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		subset = append(subset, more...)
	}
	return subset, nil
}

func (ecc *ECSClusterClient) describeInstances(input *ecs.DescribeContainerInstancesInput) ([]instanceResources, error) {
	result, err := ecc.ecsClient.DescribeContainerInstances(input)
	if err != nil {
		return nil, errors.Wrap(err, "problem describing ecs container instances")
	}

	if len(result.Failures) != 0 {
		msg := make([]string, len(result.Failures))
		for i, failure := range result.Failures {
			msg[i] = *failure.Reason
		}
		return nil, errors.Errorf("ERRORS: %s", strings.Join(msg, "\n"))
	}

	res := make([]instanceResources, len(result.ContainerInstances))
	for i, ci := range result.ContainerInstances {
		irs := instanceResources{}
		for _, rsrc := range ci.RegisteredResources {
			if *rsrc.Name == "CPU" {
				irs.cpu = *rsrc.IntegerValue
			} else if *rsrc.Name == "MEMORY" {
				irs.memory = *rsrc.IntegerValue
			}
		}
		res[i] = irs
	}
	return res, nil
}

type resourceCache struct {
	duration      time.Duration
	internalCache *cache.Cache
}

func (rc *resourceCache) getInstanceResources(clusterName string) (*instanceResources, bool) {
	resources, found := rc.internalCache.Get(clusterName)
	if found {
		ir := resources.(instanceResources)
		return &ir, true
	}
	return &instanceResources{}, false
}

func (rc *resourceCache) setInstanceResources(clusterName string, resources *instanceResources) {
	rc.internalCache.Set(clusterName, *resources, rc.duration)
}

type clusterNamesCache struct {
	duration      time.Duration
	internalCache *cache.Cache
}

func (cnc *clusterNamesCache) getClusterNames() ([]string, bool) {
	rawClusterNames, found := cnc.internalCache.Get("clusterNames")
	if found {
		clusterNames, ok := rawClusterNames.(*[]string)
		return *clusterNames, ok
	}
	return make([]string, 0, 0), false
}

func (cnc *clusterNamesCache) setInstanceResources(clusterNames []string) {
	cnc.internalCache.Set("clusterNames", &clusterNames, cnc.duration)
}
