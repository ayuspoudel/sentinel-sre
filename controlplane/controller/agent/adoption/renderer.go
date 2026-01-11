package adoption

import (
	"context"
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

/*
@ayusoudel
This renders helm templates without installing them
It is useful to identify what resources will this helm chart create and if those resources already exist
it will allow adpotion functions to check if any adoption is needed for those resources
Basically if resources that helm will install already exist, their ownership needs to be transferred to helm
This step renders manifest so that other functions can use it to get names of resources the helm chart will install
*/
func RenderManifests(ctx context.Context, kubeconfig []byte, contextName, namespace, chartRef string, values map[string]interface{}) (string, error) {
	settings := cli.New()

	tmp, err := os.CreateTemp("", "sentinel-render-kubeconfig-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(kubeconfig)
	if err != nil {
		return "", err
	}

	settings.KubeConfig = tmp.Name()
	settings.KubeContext = contextName

	actionCfg := new(action.Configuration)
	err = actionCfg.Init(settings.RESTClientGetter(), namespace, "", func(string, ...interface{}) {})
	if err != nil {
		return "", err
	}

	install := action.NewInstall(actionCfg)
	install.DryRun = true
	install.ReleaseName = "sentinel-agent"
	install.Namespace = namespace
	install.ClientOnly = true
	install.IncludeCRDs = true

	chartPath, err := install.LocateChart(chartRef, settings)
	if err != nil {
		return "", err
	}

	ch, err := loader.Load(chartPath)
	if err != nil {
		return "", err
	}

	rel, err := install.RunWithContext(ctx, ch, values)
	if err != nil {
		return "", err
	}

	if rel.Manifest == "" {
		return "", fmt.Errorf("empty rendered manifest")
	}

	return rel.Manifest, nil
}
