package values

func BuildAgentValues(
	sentinelSREURL string,
	agentControllerURL string,
	clusterName string,
	agentID string,
	agentVersion string,
	imageRepo string,
	imageTag string,
) map[string]interface{} {
	return map[string]interface{}{
		"env": map[string]interface{}{
			"SENTINEL_SRE_URL":       sentinelSREURL,
			"AGENT_CONTROLLER_URL":   agentControllerURL,
			"SENTINEL_CLUSTER_NAME":  clusterName,
			"SENTINEL_AGENT_ID":      agentID,
			"SENTINEL_AGENT_VERSION": agentVersion,
		},
		"image": map[string]interface{}{
			"repository": imageRepo,
			"tag":        imageTag,
		},
	}
}
