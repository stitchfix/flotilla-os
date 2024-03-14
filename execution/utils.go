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

	// Always present tag
	customTags = append(customTags, fmt.Sprintf("flotilla_run_id:%s", run.RunID))

	// Labels that may or may not exist
	customTags = append(customTags, getTagOrDefault(run.Labels, "team", "unknown"))
	customTags = append(customTags, getTagOrDefault(run.Labels, "kube_workflow", "unknown"))
	customTags = append(customTags, getTagOrDefault(run.Labels, "kube_task_name", "unknown"))

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

func getTagOrDefault(labels map[string]string, labelName, defaultValue string) string {
	if value, exists := labels[labelName]; exists && value != "" {
		return fmt.Sprintf("%s:%s", labelName, value)
	}
	return fmt.Sprintf("%s:%s", labelName, defaultValue)
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
