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
	customTags := []string{
		fmt.Sprintf("flotilla_run_id:%s", run.RunID),
		fmt.Sprintf("team:%s", run.Labels["team"]),
		fmt.Sprintf("kube_workflow:%s", run.Labels["kube_workflow"]),
		fmt.Sprintf("kube_task_name:%s", run.Labels["kube_task_name"]),
	}
	existingConfig := map[string]interface{}{
		"spark_url":          "http://%host%:4040",
		"spark_cluster_mode": "spark_driver_mode",
		"cluster_name":       run.ClusterName,
		"tags":               customTags,
	}

	// Convert the existingConfig map into a JSON string
	existingConfigBytes, err := json.Marshal(existingConfig)
	if err != nil {
		// Proper error handling should be in place
		log.Fatalf("Error marshaling config to JSON: %v", err)
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
