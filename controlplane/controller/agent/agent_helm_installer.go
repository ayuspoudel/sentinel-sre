package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

type HelmInstaller struct {
	chart   string
	repoUrl string
}

func NewHelmInstaller(chart, repoUrl string) *HelmInstaller {
	return &HelmInstaller{chart: chart, repoUrl: repoUrl}
}

func (h *HelmInstaller) Install(ctx context.Context, cfg *InstallConfig) error {
	settings := cli.New()

	tmp, err := os.CreateTemp("", "sentinel-kubeconfig-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write([]byte(cfg.KubeConfig))
	if err != nil {
		return err
	}
	settings.KubeConfig = tmp.Name()
	settings.KubeContext = cfg.ContextName
	actionCfg := new(action.Configuration)
	err = actionCfg.Init(settings.RESTClientGetter(), AgentNamespace, DefaultHelmDriver, func(format string, v ...interface{}) {})
	if err != nil {
		return err
	}
	install := action.NewInstall(actionCfg)
	install.ReleaseName = AgentReleaseName
	install.Namespace = AgentNamespace
	install.CreateNamespace = true

	chartPath, err := install.LocateChart(h.chart, cli.New())
	if err != nil {
		return err
	}
	ch, err := loader.Load(chartPath)
	if err != nil {
		return err
	}
	_, err = install.RunWithContext(ctx, ch, cfg.Values)
	if err != nil {
		if strings.Contains(err.Error(), "cannot re-use a name that is still in use") {
			return nil
		}
		return fmt.Errorf("helm install failed: %w", err)
	}
	return nil
}
