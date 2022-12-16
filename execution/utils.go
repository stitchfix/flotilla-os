package utils

import (
	"regexp"
	"strings"

	"github.com/stitchfix/flotilla-os/state"
)

func GetLabels(run state.Run) map[string]string {
	var labels = make(map[string]string)

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
	return key
}
