package agent

type InstallConfig struct {
	ClusterName string
	KubeConfig  []byte
	ContextName string
}
