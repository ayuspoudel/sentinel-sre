package values

func BuildAgentValues(controlPlaneURL, clusterName, imageRepo, imageTag string) map[string]interface{} {
	return map[string]interface{}{
		"sentinelSRE": map[string]interface{}{
			"url": controlPlaneURL,
		},
		"cluster": map[string]interface{}{
			"name": clusterName,
		},
		"image": map[string]interface{}{
			"repository": imageRepo,
			"tag":        imageTag,
		},
	}
}
