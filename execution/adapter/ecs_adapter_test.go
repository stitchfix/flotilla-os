package adapter

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stitchfix/flotilla-os/config"
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
	confDir := "../../conf"
	conf, _ := config.NewConfig(&confDir)

	adapter := ecsAdapter{
		ec2Client: &client,
		ecsClient: &client,
		retriable: []string{
			"A", "B", "C",
		},
		conf: conf,
	}
	return adapter
}

func TestEcsAdapter_AdaptRun(t *testing.T) {
	adapter := setUp(t)

	definition := state.Definition{
		Arn:           "darn",
		GroupName:     "groupa",
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

	if rti.StartedBy == nil || *rti.StartedBy != definition.GroupName {
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

func TestEcsAdapter_AdaptDefinition(t *testing.T) {
	adapter := setUp(t)

	memory := int64(512)
	d := state.Definition{
		DefinitionID: "id:cupcake",
		GroupName:    "group:cupcake",
		Memory:       &memory,
		Alias:        "cupcake",
		Image:        "image:cupcake",
		Command:      "echo 'hi'",
		Env: &state.EnvList{
			{Name: "E1", Value: "V1"},
		},
		Ports: &state.PortsList{12345, 6789},
		Tags:  &state.Tags{"apple", "orange", "tiger"},
	}

	adapted := adapter.AdaptDefinition(d)
	if len(adapted.ContainerDefinitions) != 1 {
		t.Errorf("Expected exactly 1 container definition, was %v", len(adapted.ContainerDefinitions))
	}

	if adapted.Family == nil || len(*adapted.Family) == 0 {
		t.Errorf("Expected non-nil and non-empty Family")
	}

	if adapted.NetworkMode == nil || len(*adapted.NetworkMode) == 0 {
		t.Errorf("Expected non-nil and non-empty NetworkMode")
	}

	//
	// Just make sure that the data -from- the definition is in the right place; other
	// fields (defaults) subject to much change
	//
	container := adapted.ContainerDefinitions[0]
	if len(container.Command) == 0 {
		t.Errorf("Expected non-nil and non-empty command")
	}
	cmd := container.Command[len(container.Command)-1]
	wrapped, err := d.WrappedCommand()
	if err != nil {
		t.Errorf(err.Error())
	}

	if *cmd != wrapped {
		t.Errorf("Expected wrapped command [%s] but was [%s]", d.Command, *cmd)
	}

	alias, ok := container.DockerLabels["alias"]
	if !ok {
		t.Errorf("Expected non-empty DockerLabels with field [alias] set")
	}
	if *alias != d.Alias {
		t.Errorf("Expected alias %s but was %s", d.Alias, *alias)
	}

	groupName, ok := container.DockerLabels["group.name"]
	if !ok {
		t.Errorf("Expected non-empty DockerLabels with field [group.name] set")
	}
	if *groupName != d.GroupName {
		t.Errorf("Expected groupName %s but was %s", d.GroupName, groupName)
	}

	tagsList, ok := container.DockerLabels["tags"]
	if !ok {
		t.Errorf("Expected non-empty DockerLabels with field [tags] set")
	}
	if *tagsList != "apple,orange,tiger" {
		t.Errorf("Expected tags [apple,orange,tiger] but was %s", *tagsList)
	}

	env := container.Environment
	if len(env) != len(*d.Env) {
		t.Errorf("Expected %v environment variables but was %v", len(*d.Env), len(env))
	}
	for _, e := range env {
		if *e.Name == "E1" && *e.Value != "V1" {
			t.Errorf("Expected environment variable E1:V1 but was E1:%s", *e.Value)
		}
	}

	ports := container.PortMappings
	if len(ports) != len(*d.Ports) {
		t.Errorf("Expected %v ports but was %v", len(*d.Ports), len(ports))
	}

	allPresent := true
	for _, pm := range ports {
		present := false
		for _, p := range *d.Ports {
			if int64(p) == *pm.ContainerPort {
				present = true
				break
			}
		}
		allPresent = allPresent && present
	}
	if !allPresent {
		t.Errorf("Expected ports %v but was %v", *d.Ports, ports)
	}

	if container.Image == nil {
		t.Errorf("Expected non-nil image")
	}
	if *container.Image != d.Image {
		t.Errorf("Expected image %s but was %s", d.Image, *container.Image)
	}

	if container.Memory == nil {
		t.Errorf("Expected non-nil memory")
	}
	if *container.Memory != memory {
		t.Errorf("Expected memory %v but was %v", memory, *container.Memory)
	}
}

func TestEcsAdapter_AdaptTaskDef(t *testing.T) {
	adapter := setUp(t)

	arn := "arn:cupcake"
	family := "id:cupcake"
	group := "group:cupcake"
	memory := int64(512)
	port := int64(1234)
	image := "image:cupcake"
	alias := "alias:cupcake"
	tagsList := "apple,orange,tiger"
	k1 := "K1"
	v1 := "V1"
	env := []*ecs.KeyValuePair{
		{Name: &k1, Value: &v1},
	}
	ports := []*ecs.PortMapping{
		{ContainerPort: &port},
	}
	container := ecs.ContainerDefinition{
		Name:   &family,
		Memory: &memory,
		Image:  &image,
		DockerLabels: map[string]*string{
			"alias":      &alias,
			"group.name": &group,
			"tags":       &tagsList,
		},
		Environment:  env,
		PortMappings: ports,
	}
	taskDef := ecs.TaskDefinition{
		Family:               &family,
		TaskDefinitionArn:    &arn,
		ContainerDefinitions: []*ecs.ContainerDefinition{&container},
	}

	adapted := adapter.AdaptTaskDef(taskDef)
	if adapted.DefinitionID != family {
		t.Errorf("Expected DefinitionID %s but was %s", family, adapted.DefinitionID)
	}

	if adapted.Memory == nil {
		t.Errorf("Expected non-nil memory")
	}

	if *adapted.Memory != memory {
		t.Errorf("Expected memory %v but was %v", memory, *adapted.Memory)
	}

	if adapted.Image != image {
		t.Errorf("Expected image %s but was %s", image, adapted.Image)
	}

	if adapted.Alias != alias {
		t.Errorf("Expected alias %s but was %s", alias, adapted.Alias)
	}

	if adapted.GroupName != group {
		t.Errorf("Expected group %s but was %s", group, adapted.GroupName)
	}

	if adapted.Arn != arn {
		t.Errorf("Expected arn %s but was %s", arn, adapted.Arn)
	}

	if adapted.Ports == nil {
		t.Errorf("Expected non-nil ports")
	}

	if adapted.Env == nil {
		t.Errorf("Expected non-nil env")
	}

	if adapted.Tags == nil {
		t.Errorf("Expected non-nil tags")
	}

	if len(*adapted.Tags) != 3 {
		t.Errorf("Expected exactly 3 tags")
	}

	if len(*adapted.Ports) != 1 {
		t.Errorf("Expected exactly one port mapping")
	}

	if len(*adapted.Env) != 1 {
		t.Errorf("Expected exactly one env variable")
	}
}
