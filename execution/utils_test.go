package utils

import (
	"encoding/json"
	"github.com/stitchfix/flotilla-os/state"
	"reflect"
	"strings"
	"testing"
)

func TestSanitizeLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "should truncate",
			input:    strings.Repeat("a", 64),
			expected: strings.Repeat("a", 63),
		},
		{
			name:     "leaves lowercase alone",
			input:    "lowercasealphanumeric11",
			expected: "lowercasealphanumeric11",
		},
		{
			name:     "lowercases stuff",
			input:    "UPPERCASEALPHANUMERIC11",
			expected: "uppercasealphanumeric11",
		},
		{
			name:     "replaces special chars",
			input:    "a*s",
			expected: "a_s",
		},
		{
			name:     "trims spaces",
			input:    " foo ",
			expected: "foo",
		},
		{
			name:     "removes leading _'s",
			input:    "_a",
			expected: "a",
		},
		{
			name:     "removes trailing _'s",
			input:    "a_",
			expected: "a",
		},
		{
			name:     "removes repeated trailing _'s",
			input:    "a_____",
			expected: "a",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SanitizeLabel(test.input)
			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestGetLabels(t *testing.T) {
	type args struct {
		run state.Run
	}
	var tests []struct {
		name string
		args args
		want map[string]string
	}

	tests = []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should return labels for run with definition",
			args: args{
				run: state.Run{
					DefinitionID: "A", ClusterName: "A", GroupName: "groupA", RunID: "runA", Labels: map[string]string{
						"kube_foo": "bar", "team": "awesomeness",
					},
				},
			},
			want: map[string]string{
				"cluster-name":    "A",
				"flotilla-run-id": "runa",
				"kube_foo":        "bar",
				"team":            "awesomeness",
			},
		},
		{
			name: "should return empty labels for run with no definition",
			args: args{
				run: state.Run{},
			},
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLabels(tt.args.run); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSetSparkDatadogConfig tests the SetSparkDatadogConfig function
func TestSetSparkDatadogConfig(t *testing.T) {
	// Define a test run object
	run := state.Run{
		RunID: "test-run-id",
		Labels: map[string]string{
			"team":           "test",
			"kube_workflow":  "test-workflow",
			"kube_task_name": "test-task",
		},
		ClusterName: "test-cluster",
	}

	// Expected tags in the JSON output
	expectedTags := []string{
		"flotilla_run_id:test-run-id",
		"team:test",
		"kube_workflow:test-workflow",
		"kube_task_name:test-task",
	}

	result := SetSparkDatadogConfig(run)

	if result == nil {
		t.Fatalf("Expected a non-nil result")
	}

	var resultMap map[string]interface{}
	err := json.Unmarshal([]byte(*result), &resultMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON result: %v", err)
	}

	checks, ok := resultMap["checks"].(map[string]interface{})
	if !ok {
		t.Fatalf("Checks are missing or not in the expected format")
	}

	spark, ok := checks["spark"].(map[string]interface{})
	if !ok {
		t.Fatalf("Spark configuration missing or not in the expected format")
	}

	instances, ok := spark["instances"].([]interface{})
	if !ok || len(instances) == 0 {
		t.Fatalf("Instances are missing, empty, or not in the expected format")
	}

	instance, ok := instances[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Instance is not in the expected format")
	}

	tags, ok := instance["tags"].([]interface{})
	if !ok {
		t.Fatalf("Tags are not in the expected format or missing")
	}

	for _, expectedTag := range expectedTags {
		found := false
		for _, tag := range tags {
			if strTag, ok := tag.(string); ok && strTag == expectedTag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tag %s not found in result", expectedTag)
		}
	}
}
