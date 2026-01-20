package config

import "github.com/ayuspoudel/sentinel-sre/pkg/env"

/*
Author: @ayuspoudel
Config holds runtime configuration for sentinel-api-server
This is the single source of truth for wiring internal services
*/
type Config struct {
	ListenAddr         string
	ClusterRegistryURL string
	PolicyRegistryURL  string
	AgentControllerURL string
}

func Load() Config {
	return Config{
		ListenAddr:         env.MustEnv("SENTINEL_API_ADDRESS", true),
		ClusterRegistryURL: "http://localhost:8080",
		PolicyRegistryURL:  "http://localhost:9001",
		AgentControllerURL: "http://localhost:9000",
	}
}
