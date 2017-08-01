package adapters

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stitchfix/flotilla-os/state"
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

func setUp(t *testing.T) ecsAdapter {
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
	return adapter
}

func TestEcsAdapter_AdaptRun(t *testing.T) {
	adapter := setUp(t)

	definition := state.Definition{
		Arn:           "darn",
		ContainerName: "mynameiswhat",
	}

	k1 := "ENVVAR_A"
	k2 := "ENVVAR_B"
	v1 := "VALUEA"
	v2 := "VALUEB"
	env := state.EnvList([]state.EnvVar{
		{Name: k1, Value: v1},
		{Name: k2, Value: v2},
	})

	run := state.Run{
		ClusterName: "clusta",
		GroupName:   "groupa",
		Env:         &env,
	}
	rti := adapter.AdaptRun(definition, run)

	if rti.StartedBy == nil || *rti.StartedBy != run.GroupName {
		t.Errorf("Expected startedBy name groupa")
	}

	if rti.Cluster == nil || *rti.Cluster != run.ClusterName {
		t.Errorf("Expected cluster name clusta")
	}

	if rti.Overrides != nil && len(rti.Overrides.ContainerOverrides) > 0 {
		envOverrides := rti.Overrides.ContainerOverrides[0].Environment
		if len(envOverrides) != len(env) {
			t.Errorf("Expected %v env vars, got %v", len(env), len(envOverrides))
		}

		for _, e := range envOverrides {
			if *e.Name != k1 && *e.Name != k2 {
				t.Errorf("Unexpected env var %s", *e.Name)
			}
			if *e.Name == k1 && *e.Value != v1 {
				t.Errorf("Expected %s value %v but was %v", k1, v1, *e.Value)
			}
			if *e.Name == k2 && *e.Value != v2 {
				t.Errorf("Expected %s value %v but was %v", k2, v2, *e.Value)
			}
		}
	} else {
		t.Errorf("Expected non-nil and non empty container overrides")
	}
}

func TestEcsAdapter_AdaptTask(t *testing.T) {
	adapter := setUp(t)

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

	lastStatus := state.StatusPending
	desiredStatus := state.StatusStopped

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
	adapted := adapter.AdaptTask(task1)

	if adapted.TaskArn != arn {
		t.Errorf("Expected arn: %s, was %s", arn, adapted.TaskArn)
	}

	if adapted.GroupName != group {
		t.Errorf("Expected group: %s, was %s", group, adapted.GroupName)
	}

	if adapted.StartedAt.UTC().String() != t1.UTC().String() {
		t.Errorf("Expected startedAt: %v, was %v", t1.UTC().String(), adapted.StartedAt.UTC())
	}

	if adapted.FinishedAt.UTC().String() != t2.UTC().String() {
		t.Errorf("Expected finishedAt: %v, was %v", t2.UTC().String(), adapted.FinishedAt.UTC())
	}

	if adapted.InstanceDNSName != "sprinkles" {
		t.Errorf("Expected instanceDNSName 'sprinkles', was %s", adapted.InstanceDNSName)
	}

	if adapted.InstanceID != "cupcake" {
		t.Errorf("Expected instanceID 'cupcake', was %s", adapted.InstanceID)
	}

	if adapted.ExitCode == nil || *adapted.ExitCode != exitCode {
		t.Errorf("Expected exit code 0")
	}

	if adapted.Status != lastStatus {
		t.Errorf("Expected status %s, was %s", lastStatus, adapted.Status)
	}

	if adapted.Env != nil && len(*adapted.Env) > 0 {
		if len(*adapted.Env) != 2 {
			t.Errorf("Expected %v env vars, got %v", 2, len(*adapted.Env))
		}

		for _, e := range *adapted.Env {
			if e.Name != k1 && e.Name != k2 {
				t.Errorf("Unexpected env var %s", e.Name)
			}
			if e.Name == k1 && e.Value != v1 {
				t.Errorf("Expected %s value %v but was %v", k1, v1, e.Value)
			}
			if e.Name == k2 && e.Value != v2 {
				t.Errorf("Expected %s value %v but was %v", k2, v2, e.Value)
			}
		}
	} else {
		t.Errorf("Expected non-nil and non-empty env")
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
	adapted = adapter.AdaptTask(task2)
	if adapted.Status != desiredStatus {
		t.Errorf("Expected status %s, was %s", desiredStatus, adapted.Status)
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
	adapted = adapter.AdaptTask(task3)
	if adapted.Status != state.StatusNeedsRetry {
		t.Errorf("Expected status %s, was %s", state.StatusNeedsRetry, adapted.Status)
	}
}
