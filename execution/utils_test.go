package utils

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stitchfix/flotilla-os/state"
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
			input:    "a*",
			expected: "a_",
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
