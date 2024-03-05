package utils

import (
	"encoding/json"
	"fmt"
	"github.com/stitchfix/flotilla-os/state"
	"log"
	"regexp"
	"strings"
)

// SetSparkDatadogConfig sets the values needed for Spark Datadog integration
func SetSparkDatadogConfig(run state.Run) string {
	var customTags []string

	// This will always be present
	customTags = append(customTags, fmt.Sprintf("flotilla_run_id:%s", run.RunID))

	// These might not
	if team, exists := run.Labels["team"]; exists && team != "" {
		customTags = append(customTags, fmt.Sprintf("team:%s", team))
	}
	if kubeWorkflow, exists := run.Labels["kube_workflow"]; exists && kubeWorkflow != "" {
		customTags = append(customTags, fmt.Sprintf("kube_workflow:%s", kubeWorkflow))
	}
	if kubeTaskName, exists := run.Labels["kube_task_name"]; exists && kubeTaskName != "" {
		customTags = append(customTags, fmt.Sprintf("kube_task_name:%s", kubeTaskName))
	}

	existingConfig := map[string]interface{}{
		"spark_url":          "http://%host%:4040",
		"spark_cluster_mode": "spark_driver_mode",
		"cluster_name":       run.ClusterName,
		"tags":               customTags,
	}

	// Convert the existingConfig map into a JSON string
	existingConfigBytes, err := json.Marshal(existingConfig)

	// We should never reach here as this will always be a valid JSON
	if err != nil {
		log.Printf("Failed to marshal existingConfig: %v", err)
		return ""
	}
	return string(existingConfigBytes)
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
