package values

func BuildAgentValues(controlPlaneURL, clusterName string) map[string]interface{} {
	return map[string]interface{}{
		"sentinelSRE": map[string]interface{}{
			"url": controlPlaneURL,
		},
		"cluster": map[string]interface{}{
			"name": clusterName,
		},
	}
}
