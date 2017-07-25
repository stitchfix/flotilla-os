package cluster

import (
	"github.com/aws/aws-sdk-go/service/ecs"
	"testing"
)

type testResourceClient struct {
	t          *testing.T
	listCalled int
}

func (trc *testResourceClient) DescribeClusters(input *ecs.DescribeClustersInput) (*ecs.DescribeClustersOutput, error) {
	if len(input.Clusters) == 0 {
		trc.t.Errorf("Expected at least one cluster name search param, got 0")
	}

	first := input.Clusters[0]
	if first == nil || len(*first) == 0 {
		trc.t.Errorf("Expected non-nil and non-empty cluster name")
	}

	var res ecs.DescribeClustersOutput
	if *first == "clusta" {
		name := "clusta"
		clstr := ecs.Cluster{
			ClusterName: &name,
		}
		res = ecs.DescribeClustersOutput{
			Clusters: []*ecs.Cluster{
				&clstr,
			},
		}
	}
	return &res, nil
}

func (trc *testResourceClient) ListContainerInstances(input *ecs.ListContainerInstancesInput) (*ecs.ListContainerInstancesOutput, error) {
	tok := "next_token"

	if input.Cluster == nil || len(*input.Cluster) == 0 {
		trc.t.Errorf("Expected non-nil and non-empty cluster name")
	}

	if trc.listCalled > 0 && input.NextToken == nil && *input.NextToken != tok {
		trc.t.Errorf("Called ListContainerInstances already, yet NextToken provided was nil or incorrect")
	}

	var (
		res ecs.ListContainerInstancesOutput
		arn string
	)

	if trc.listCalled == 0 {
		res.NextToken = &tok
		arn = "arn1"
	} else {
		arn = "arn2"
	}

	res.ContainerInstanceArns = []*string{
		&arn,
	}
	trc.listCalled += 1
	return &res, nil
}

func (trc *testResourceClient) DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
	if input.Cluster == nil || len(*input.Cluster) == 0 {
		trc.t.Errorf("Expected non-nil and non-empty cluster name")
	}

	if len(input.ContainerInstances) == 0 {
		trc.t.Errorf("Shouldn't make this call with zero instances")
	}

	for _, arn := range input.ContainerInstances {
		if arn == nil || len(*arn) == 0 {
		}
		trc.t.Errorf("Expected non-nil and non-empty instance arns")
	}

	clusterName := *input.Cluster
	var res ecs.DescribeContainerInstancesOutput
	if clusterName == "failwhale" {
		arn := "arn"
		reason := "life is hard"
		failure := ecs.Failure{
			Arn:    &arn,
			Reason: &reason,
		}
		res.Failures = []*ecs.Failure{
			&failure,
		}
	} else {
		cpu := int64(10)
		cpuKey := "CPU"
		mem := int64(100)
		memKey := "MEMORY"
		cpuResource := ecs.Resource{
			Name:         &cpuKey,
			IntegerValue: &cpu,
		}
		memResource := ecs.Resource{
			Name:         &memKey,
			IntegerValue: &mem,
		}

		ci := ecs.ContainerInstance{
			RegisteredResources: []*ecs.Resource{
				&memResource,
				&cpuResource,
			},
		}
		res.ContainerInstances = []*ecs.ContainerInstance{
			&ci,
		}
	}

	return &res, nil
}

func TestECSClusterClient_CanBeRun(t *testing.T) {

}
