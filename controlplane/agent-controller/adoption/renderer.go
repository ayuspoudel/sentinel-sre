package adoption

import (
	"context"
	"fmt"
	"os"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/logging"
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
	log := logging.From(ctx)
	log.Info("rendering helm manifests", "namespace", namespace, "chart", chartRef)

	settings := cli.New()

	tmp, err := os.CreateTemp("", "sentinel-render-kubeconfig-*")
	if err != nil {
		log.Error("failed to create temp kubeconfig file for render", "error", err)
		return "", err
	}
	defer os.Remove(tmp.Name())

	_, err = tmp.Write(kubeconfig)
	if err != nil {
		log.Error("failed to write kubeconfig to temp file for render", "error", err)
		return "", err
	}

	settings.KubeConfig = tmp.Name()
	settings.KubeContext = contextName

	actionCfg := new(action.Configuration)
	err = actionCfg.Init(settings.RESTClientGetter(), namespace, "", func(string, ...interface{}) {})
	if err != nil {
		log.Error("failed to initialize helm action configuration for render", "error", err)
		return "", err
	}

	install := action.NewInstall(actionCfg)
	install.DryRun = true
	install.ReleaseName = "sentinel-agent"
	install.Namespace = namespace
	install.ClientOnly = true
	install.IncludeCRDs = true

	log.Info("locating helm chart for render")

	chartPath, err := install.LocateChart(chartRef, settings)
	if err != nil {
		log.Error("failed to locate helm chart for render", "error", err)
		return "", err
	}

	ch, err := loader.Load(chartPath)
	if err != nil {
		log.Error("failed to load helm chart for render", "error", err)
		return "", err
	}

	log.Info("executing helm dry-run render")

	rel, err := install.RunWithContext(ctx, ch, values)
	if err != nil {
		log.Error("helm dry-run render failed", "error", err)
		return "", err
	}

	if rel.Manifest == "" {
		log.Error("rendered manifest is empty")
		return "", fmt.Errorf("empty rendered manifest")
	}

	log.Info("helm manifests rendered successfully", "size_bytes", len(rel.Manifest))
	return rel.Manifest, nil
}
