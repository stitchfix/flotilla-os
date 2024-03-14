package utils

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stitchfix/flotilla-os/state"
	"log"
	"regexp"
	"strings"
)

type DatadogConfig struct {
	Checks map[string]IntegrationConfig `json:"checks"`
}

type IntegrationConfig struct {
	InitConfig map[string]interface{} `json:"init_config"`
	Instances  []InstanceConfig       `json:"instances"`
}

type InstanceConfig struct {
	SparkURL         string   `json:"spark_url"`
	SparkClusterMode string   `json:"spark_cluster_mode"`
	ClusterName      string   `json:"cluster_name"`
	Tags             []string `json:"tags"`
}

// SetSparkDatadogConfig sets the values needed for Spark Datadog integration
func SetSparkDatadogConfig(run state.Run) *string {
	var customTags []string

	customTags = append(customTags, fmt.Sprintf("flotilla_run_id:%s", run.RunID))

	if team, exists := run.Labels["team"]; exists && team != "" {
		customTags = append(customTags, fmt.Sprintf("team:%s", team))
	} else {
		customTags = append(customTags, "team:unknown")
	}

	if kubeWorkflow, exists := run.Labels["kube_workflow"]; exists && kubeWorkflow != "" {
		customTags = append(customTags, fmt.Sprintf("kube_workflow:%s", kubeWorkflow))
	} else {
		customTags = append(customTags, "kube_workflow:unknown")
	}

	if kubeTaskName, exists := run.Labels["kube_task_name"]; exists && kubeTaskName != "" {
		customTags = append(customTags, fmt.Sprintf("kube_task_name:%s", kubeTaskName))
	} else {
		customTags = append(customTags, "kube_task_name:unknown")
	}

	datadogConfig := DatadogConfig{
		Checks: map[string]IntegrationConfig{
			"spark": {
				InitConfig: map[string]interface{}{},
				Instances: []InstanceConfig{
					{
						SparkURL:         "http://%host%:4040",
						SparkClusterMode: "spark_driver_mode",
						ClusterName:      run.ClusterName,
						Tags:             customTags,
					},
				},
			},
		},
	}

	datadogConfigBytes, err := json.Marshal(datadogConfig)

	// We should never reach here as this will always be a valid JSON
	if err != nil {
		log.Printf("Failed to marshal existingConfig: %v", err)
		return nil
	}
	return aws.String(string(datadogConfigBytes))
}

func GetLabels(run state.Run) map[string]string {
	var labels = make(map[string]string)

	if run.ClusterName != "" {
		labels["cluster-name"] = run.ClusterName
	}

	if run.RunID != "" {
		labels["flotilla-run-id"] = SanitizeLabel(run.RunID)
	}

	if run.User != "" {
		labels["owner"] = SanitizeLabel(run.User)
	}

	for k, v := range run.Labels {
		labels[k] = SanitizeLabel(v)
	}

	return labels
}

func SanitizeLabel(key string) string {
	key = strings.TrimSpace(key)
	key = regexp.MustCompile(`[^-a-z0-9A-Z_.]+`).ReplaceAllString(key, "_")
	key = strings.TrimPrefix(key, "_")
	key = strings.ToLower(key)
	if len(key) > 63 {
		key = key[:63]
	}
	for {
		tempKey := strings.TrimSuffix(key, "_")
		if tempKey == key {
			break
		}
		key = tempKey
	}

	return key
}
