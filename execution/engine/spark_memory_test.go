package engine

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stitchfix/flotilla-os/state"
)

func TestDriverMemoryMiB(t *testing.T) {
	tests := []struct {
		name     string
		run      state.Run
		expected int64
	}{
		{
			name:     "nil spark extension",
			run:      state.Run{},
			expected: 0,
		},
		{
			name: "no spark.driver.memory in conf",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						SparkSubmitConf: []state.Conf{
							{Name: aws.String("spark.executor.memory"), Value: aws.String("4g")},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "spark.driver.memory 40g",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						SparkSubmitConf: []state.Conf{
							{Name: aws.String("spark.driver.memory"), Value: aws.String("40g")},
						},
					},
				},
			},
			expected: 40 * 1024,
		},
		{
			name: "spark.driver.memory 512m",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						SparkSubmitConf: []state.Conf{
							{Name: aws.String("spark.driver.memory"), Value: aws.String("512m")},
						},
					},
				},
			},
			expected: 512,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := driverMemoryMiB(tt.run)
			if got != tt.expected {
				t.Errorf("driverMemoryMiB() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestExecutorMemoryMiB(t *testing.T) {
	tests := []struct {
		name     string
		run      state.Run
		expected int64
	}{
		{
			name:     "nil spark extension",
			run:      state.Run{},
			expected: 0,
		},
		{
			name: "ExecutorMemory field set",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						ExecutorMemory: aws.Int64(8192),
					},
				},
			},
			expected: 8192,
		},
		{
			name: "spark.executor.memory in conf",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						SparkSubmitConf: []state.Conf{
							{Name: aws.String("spark.executor.memory"), Value: aws.String("16g")},
						},
					},
				},
			},
			expected: 16 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executorMemoryMiB(tt.run)
			if got != tt.expected {
				t.Errorf("executorMemoryMiB() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestDriverCPUMillis(t *testing.T) {
	tests := []struct {
		name     string
		run      state.Run
		expected int64
	}{
		{
			name:     "nil spark extension",
			run:      state.Run{},
			expected: 0,
		},
		{
			name: "no spark.driver.cores in conf",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						SparkSubmitConf: []state.Conf{
							{Name: aws.String("spark.executor.cores"), Value: aws.String("4")},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "spark.driver.cores 4",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						SparkSubmitConf: []state.Conf{
							{Name: aws.String("spark.driver.cores"), Value: aws.String("4")},
						},
					},
				},
			},
			expected: 4000,
		},
		{
			name: "spark.driver.cores 1",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						SparkSubmitConf: []state.Conf{
							{Name: aws.String("spark.driver.cores"), Value: aws.String("1")},
						},
					},
				},
			},
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := driverCPUMillis(tt.run)
			if got != tt.expected {
				t.Errorf("driverCPUMillis() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestExecutorCPUMillis(t *testing.T) {
	tests := []struct {
		name     string
		run      state.Run
		expected int64
	}{
		{
			name:     "nil spark extension",
			run:      state.Run{},
			expected: 0,
		},
		{
			name: "no spark.executor.cores in conf",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						SparkSubmitConf: []state.Conf{
							{Name: aws.String("spark.driver.cores"), Value: aws.String("2")},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "spark.executor.cores 3",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						SparkSubmitConf: []state.Conf{
							{Name: aws.String("spark.executor.cores"), Value: aws.String("3")},
						},
					},
				},
			},
			expected: 3000,
		},
		{
			name: "spark.executor.cores 8",
			run: state.Run{
				SparkExtension: &state.SparkExtension{
					SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
						SparkSubmitConf: []state.Conf{
							{Name: aws.String("spark.executor.cores"), Value: aws.String("8")},
						},
					},
				},
			},
			expected: 8000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executorCPUMillis(tt.run)
			if got != tt.expected {
				t.Errorf("executorCPUMillis() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestParseSparkMemoryMiB(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"40g", 40 * 1024},
		{"1t", 1024 * 1024},
		{"512m", 512},
		{"2048k", 2},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSparkMemoryMiB(tt.input)
			if got != tt.expected {
				t.Errorf("parseSparkMemoryMiB(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}
