package adapters

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"testing"
	"time"
)

type testClient struct {
	t               *testing.T
	instanceID      string
	instanceDNSName string
}

func (tc *testClient) DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
	if input.Cluster == nil || len(*input.Cluster) == 0 {
		tc.t.Errorf("Expected non-nil and non-empty cluster name")
	}

	if len(input.ContainerInstances) != 1 {
		tc.t.Errorf("Expected to describe exactly one container instance")
	}

	ci := ecs.ContainerInstance{
		Ec2InstanceId: &tc.instanceID,
	}
	res := ecs.DescribeContainerInstancesOutput{
		ContainerInstances: []*ecs.ContainerInstance{&ci},
	}
	return &res, nil
}

func (tc *testClient) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	if len(input.InstanceIds) != 1 {
		tc.t.Errorf("Expected exactly one instance id to describe")
	}

	instance := ec2.Instance{
		PrivateDnsName: &tc.instanceDNSName,
	}

	rsv := ec2.Reservation{
		Instances: []*ec2.Instance{&instance},
	}

	return &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{&rsv},
	}, nil
}

func TestEcsAdapter_AdaptTask(t *testing.T) {
	client := testClient{
		t:               t,
		instanceID:      "cupcake",
		instanceDNSName: "sprinkles",
	}
	adapter := ecsAdapter{
		ec2Client: &client,
		ecsClient: &client,
		retriable: []string{
			"A", "B", "C",
		},
	}

	arn := "ecs-task-arn"
	group := "ecs-task-group"
	clusterArn := "clusta"
	containerInstanceArn := "clusta-instance"
	containerName := "shoebox"

	k1 := "ENVVAR_A"
	k2 := "ENVVAR_B"
	v1 := "VALUEA"
	v2 := "VALUEB"

	kv1 := &ecs.KeyValuePair{
		Name:  &k1,
		Value: &v1,
	}
	kv2 := &ecs.KeyValuePair{
		Name:  &k2,
		Value: &v2,
	}
	pairs := []*ecs.KeyValuePair{kv1, kv2}

	envOverrides := ecs.ContainerOverride{
		Name:        &containerName,
		Environment: pairs,
	}

	overrides := ecs.TaskOverride{
		ContainerOverrides: []*ecs.ContainerOverride{&envOverrides},
	}

	exitCode := int64(0)
	reason := ""
	retriableReason := "CannotPullContainerError"

	container := ecs.Container{
		ExitCode: &exitCode,
		Reason:   &reason,
	}

	faultyContainer := ecs.Container{
		ExitCode: nil,
		Reason:   &retriableReason,
	}
	containers := []*ecs.Container{&container}
	faultyContainers := []*ecs.Container{&faultyContainer}

	lastStatus := "PENDING"
	desiredStatus := "STOPPED"

	t1, _ := time.Parse(time.RFC3339, "2017-07-04T00:01:00+00:00")
	t2, _ := time.Parse(time.RFC3339, "2017-07-04T00:02:00+00:00")

	// Normal
	task1 := ecs.Task{
		TaskArn:              &arn,
		Group:                &group,
		ClusterArn:           &clusterArn,
		ContainerInstanceArn: &containerInstanceArn,
		StartedAt:            &t1,
		StoppedAt:            &t2,
		Overrides:            &overrides,
		LastStatus:           &lastStatus,
		Containers:           containers,
	}

	// Killed
	task2 := ecs.Task{
		TaskArn:              &arn,
		Group:                &group,
		ClusterArn:           &clusterArn,
		ContainerInstanceArn: &containerInstanceArn,
		StartedAt:            &t1,
		StoppedAt:            &t2,
		Overrides:            &overrides,
		LastStatus:           &lastStatus,
		DesiredStatus:        &desiredStatus,
		Containers:           containers,
	}

	// To be retried
	task3 := ecs.Task{
		TaskArn:              &arn,
		Group:                &group,
		ClusterArn:           &clusterArn,
		ContainerInstanceArn: &containerInstanceArn,
		StartedAt:            &t1,
		StoppedAt:            &t2,
		Overrides:            &overrides,
		LastStatus:           &lastStatus,
		DesiredStatus:        &desiredStatus,
		Containers:           faultyContainers,
	}

	t.Log(adapter.AdaptTask(task1))
	t.Log(adapter.AdaptTask(task2))
	t.Log(adapter.AdaptTask(task3))
}
