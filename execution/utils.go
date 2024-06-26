package utils

import (
	"github.com/stitchfix/flotilla-os/state"
	"os"
	"regexp"
	"strings"
)

func GetLabels(run state.Run) map[string]string {
	var labels = make(map[string]string)

	if run.ClusterName != "" {
		labels["cluster-name"] = run.ClusterName
	}

	if run.RunID != "" {
		labels["flotilla-run-id"] = SanitizeLabel(run.RunID)
		labels["flotilla-run-mode"] = SanitizeLabel(os.Getenv("FLOTILLA_MODE"))
	}

	if run.User != "" {
		labels["owner"] = SanitizeLabel(run.User)
	}

	if _, workflowExists := run.Labels["kube_workflow"]; !workflowExists {
		if _, taskNameExists := run.Labels["kube_task_name"]; taskNameExists {
			labels["kube_workflow"] = SanitizeLabel(run.Labels["kube_task_name"])
		}
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
