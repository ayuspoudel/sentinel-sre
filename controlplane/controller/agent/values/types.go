package values

type AgentValues struct {
	SentinelSRE SentinelConfig `json:"sentinelSRE"`
	Cluster     ClusterConfig  `json:"cluster"`
}

type SentinelConfig struct {
	URL string `json:"url"`
}

type ClusterConfig struct {
	Name string `json:"name"`
}
