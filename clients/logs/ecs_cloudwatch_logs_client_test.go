package logs

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	"os"
	"testing"
)

type testLogsClient struct {
	t       *testing.T
	calls   []string
	nextTok string
}

func (tlc *testLogsClient) DescribeLogGroups(input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	tlc.calls = append(tlc.calls, "DescribeLogGroups")

	pref := input.LogGroupNamePrefix
	if pref == nil || len(*pref) == 0 {
		tlc.t.Errorf("Expected non-nil and non-empty LogGroupNamePrefix")
	}

	if *pref == "existing" {
		group1Name := "existing-cruft"
		group2Name := "existing"
		group1 := cloudwatchlogs.LogGroup{
			LogGroupName: &group1Name,
		}
		group2 := cloudwatchlogs.LogGroup{
			LogGroupName: &group2Name,
		}
		return &cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{
				&group1, &group2,
			},
		}, nil
	}
	return &cloudwatchlogs.DescribeLogGroupsOutput{
		LogGroups: []*cloudwatchlogs.LogGroup{},
	}, nil
}

func (tlc *testLogsClient) CreateLogGroup(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
	tlc.calls = append(tlc.calls, "CreateLogGroup")
	if input.LogGroupName == nil || len(*input.LogGroupName) == 0 {
		tlc.t.Errorf("Expected non-nil and non-empty LogGroupName")
	}
	return &cloudwatchlogs.CreateLogGroupOutput{}, nil
}

func (tlc *testLogsClient) PutRetentionPolicy(input *cloudwatchlogs.PutRetentionPolicyInput) (*cloudwatchlogs.PutRetentionPolicyOutput, error) {
	tlc.calls = append(tlc.calls, "PutRetentionPolicy")
	if input.LogGroupName == nil || len(*input.LogGroupName) == 0 {
		tlc.t.Errorf("Expected non-nil and non-empty LogGroupName")
	}

	if input.RetentionInDays == nil || *input.RetentionInDays <= 0 {
		tlc.t.Errorf("Expected non-nil RetentionInDays and a value > 0")
	}
	return &cloudwatchlogs.PutRetentionPolicyOutput{}, nil
}

func (tlc *testLogsClient) GetLogEvents(input *cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
	tlc.calls = append(tlc.calls, "GetLogEvents")
	if input.LogGroupName == nil || len(*input.LogGroupName) == 0 {
		tlc.t.Errorf("Expected non-nil and non-empty LogGroupName")
	}

	if input.LogStreamName == nil || len(*input.LogStreamName) == 0 {
		tlc.t.Errorf("Expected non-nil and non-empty LogStreamName")
	}

	m1 := "logs"
	t1 := int64(100)

	m2 := "are"
	t2 := int64(101)

	m3 := "loglike"
	t3 := int64(102)

	message1 := cloudwatchlogs.OutputLogEvent{
		Timestamp: &t1,
		Message:   &m1,
	}
	message2 := cloudwatchlogs.OutputLogEvent{
		Timestamp: &t2,
		Message:   &m2,
	}
	message3 := cloudwatchlogs.OutputLogEvent{
		Timestamp: &t3,
		Message:   &m3,
	}

	return &cloudwatchlogs.GetLogEventsOutput{
		Events: []*cloudwatchlogs.OutputLogEvent{
			&message2, &message1, &message3,
		},
		NextForwardToken: &tlc.nextTok,
	}, nil
}

func TestCloudWatchLogsClient_Logs(t *testing.T) {
	confDir := "../../conf"
	c, _ := config.NewConfig(&confDir)
	cwlc := ECSCloudWatchLogsClient{}

	tlc := testLogsClient{t: t}
	cwlc.logsClient = &tlc
	os.Setenv("ECS_LOG_NAMESPACE", "non-existing")
	err := cwlc.Initialize(c)
	if err != nil {
		t.Errorf("Failed to initialize logs client %v", err)
	}

	expectedInitializeCalls := map[string]bool{
		"DescribeLogGroups":  true,
		"CreateLogGroup":     true,
		"PutRetentionPolicy": true,
	}

	if len(tlc.calls) != len(expectedInitializeCalls) {
		t.Errorf(
			"Expected exactly %v initialization calls for non-existing log streams, but was %v",
			len(expectedInitializeCalls), len(tlc.calls))
	}

	for _, call := range tlc.calls {
		_, ok := expectedInitializeCalls[call]
		if !ok {
			t.Errorf("Unexpected initialization call for non-existing stream [%v]", call)
		}
	}

	tlc = testLogsClient{t: t}
	cwlc.logsClient = &tlc
	os.Setenv("ECS_LOG_NAMESPACE", "existing")
	cwlc.Initialize(c)
	expectedInitializeCalls = map[string]bool{
		"DescribeLogGroups": true,
	}

	if len(tlc.calls) != len(expectedInitializeCalls) {
		t.Errorf(
			"Expected exactly %v initialization calls for existing log streams, but was %v",
			len(expectedInitializeCalls), len(tlc.calls))
	}

	for _, call := range tlc.calls {
		_, ok := expectedInitializeCalls[call]
		if !ok {
			t.Errorf("Unexpected initialization call for existing stream [%v]", call)
		}
	}

	cwlc.logStreamPrefix = "cupcake"
	expectedMsg := "logs\nare\nloglike"
	expectedNextTok := "next!"
	tlc.nextTok = expectedNextTok

	d := state.Definition{
		ExecutableResources: state.ExecutableResources{ContainerName: "container"},
	}
	r := state.Run{TaskArn: "a/b/c"}

	// StreamName == cupcake/container/c
	msg, tok, _ := cwlc.Logs(d, r, nil)
	if msg != expectedMsg {
		t.Errorf("Expected log message [%v] but was [%v]", expectedMsg, msg)
	}

	if tok == nil {
		t.Errorf("Expected non-nil nextToken")
	} else if *tok != expectedNextTok {
		t.Errorf("Expected next token [%v] but was [%v]", expectedNextTok, *tok)
	}
}
