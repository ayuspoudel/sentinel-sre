package install

type InstallConfig struct {
	ClusterName string
	KubeConfig  []byte
	ContextName string
	Values      map[string]interface{}
}
