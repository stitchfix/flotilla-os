package utils

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/stitchfix/flotilla-os/state"
	"k8s.io/client-go/rest"
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

func GetClusterConfig(clusterName string, region string) (*rest.Config, error) {
	sess := session.Must(session.NewSession())
	eksSvc := eks.New(sess, aws.NewConfig().WithRegion(region))

	input := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	result, err := eksSvc.DescribeCluster(input)
	if err != nil {
		return nil, err
	}

	cluster := result.Cluster
	caData, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return nil, err
	}

	kubeConfig := &rest.Config{
		Host: aws.StringValue(cluster.Endpoint),
		TLSClientConfig: rest.TLSClientConfig{
			CAData: caData,
		},
		BearerTokenFile: "",
	}

	return kubeConfig, nil
}
