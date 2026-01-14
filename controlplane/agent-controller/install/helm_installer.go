package install

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/logging"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

type HelmInstaller struct {
	chart    string
	repoUrl  string
	repoName string
}

func NewHelmInstaller(chart, repoUrl, repoName string) *HelmInstaller {
	return &HelmInstaller{chart: chart, repoUrl: repoUrl, repoName: repoName}
}

/*
	@ayuspoudel
	This is a simple agent installer using Helm SDK
	It represents the following helm commands
		helm repo add sentinel-sre <link>
		helm install sentinel-agent sentinel-sre/sentinel-agent --namespace sentinel-agent --create-namespace --kubeconfig <kubeconfig> --kube-context <context> --values <values>
	The function recieves a cfg *InstallConfig which has ClusterName, KubeConfig, ContextName, Values map[string]interface
	It first creates a helm runtime environment supported by cli.New() provided by "helm.sh/helm/v3/pkg/cli" package
	Then it writes kubeconfig from cfg.Kubeconfig to temporary file called sentinel-kubeconfig-*
	In settings we have to update two things
		1. settings.Kubeconfig ... This is set to be the path of temporary kubeconfig file so Helm knows which cluster
		2. settings.Kubecontext ... This is set from cfg.ContextName so helm knows which context
	Then it initializes action.Configuration using the kubeconfig file and context name from cfg
	action.Configruation is a core object that wires helm actions (install, upgrade, uninstall())
	We also add helm repo using repo.NewChartRepository
	The actioncfg object represents helm's internation execution context, it has k8s client, release storage backend (secrets
	configmaps) and namespace info
	Now we build helm command using install.<field>
	Then we locate chart using action.NewInstall()'s method called LocateChart
		This parses Chart.yaml, templates, values schema and dependencies into a chart object helm can execute
	Then we finally do install.RunWithContext() and provide ctx, chart and values (values file is coming from agent_values_builder.go)
*/

func (h *HelmInstaller) Install(ctx context.Context, cfg *InstallConfig) error {
	log := logging.From(ctx)
	log.Info("starting helm install", "release", AgentReleaseName, "namespace", AgentNamespace, "chart", h.repoName+"/"+h.chart)

	// Creates a new helm runtime environment
	// We need to set cluster, context for this
	settings := cli.New()

	// Creating a local kubeconfig file because helm does not accept in memory kubeconfig at runtime
	tmp, err := os.CreateTemp("", "sentinel-kubeconfig-*")
	if err != nil {
		log.Error("failed to create temp kubeconfig file", "error", err)
		return err
	}
	defer os.Remove(tmp.Name())

	_, err = tmp.Write([]byte(cfg.KubeConfig))
	if err != nil {
		log.Error("failed to write kubeconfig to temp file", "error", err)
		return err
	}

	// Setting cluster kubeconfig path and context name
	settings.KubeConfig = tmp.Name()
	settings.KubeContext = cfg.ContextName

	// This creates a new core helm object that wires helm actions
	actionCfg := new(action.Configuration)

	// Initializing : Equiavalent to writing helm install ...
	err = actionCfg.Init(settings.RESTClientGetter(), AgentNamespace, DefaultHelmDriver, func(format string, v ...interface{}) {})
	if err != nil {
		log.Error("failed to initialize helm action configuration", "error", err)
		return err
	}

	// We need to do helm repo add sentinel-sre <link> before we perform the actual installation
	repoFile := settings.RepositoryConfig
	repoCache := settings.RepositoryCache
	repoEntry := &repo.Entry{
		Name: h.repoName,
		URL:  h.repoUrl,
	}

	repoFileObj, err := repo.LoadFile(repoFile)
	if err != nil {
		repoFileObj = repo.NewFile()
	}

	if !repoFileObj.Has(h.repoName) {
		log.Info("adding helm repository", "repo", h.repoName, "url", h.repoUrl)
		repoFileObj.Add(repoEntry)
		if err := repoFileObj.WriteFile(repoFile, 0644); err != nil {
			log.Error("failed to write helm repo file", "error", err)
			return err
		}
	}

	chartRepo, err := repo.NewChartRepository(repoEntry, getter.All(settings))
	if err != nil {
		log.Error("failed to create chart repository", "error", err)
		return err
	}

	chartRepo.CachePath = repoCache
	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		log.Error("failed to download helm repo index", "error", err)
		return fmt.Errorf("failed to download repo index: %w", err)
	}

	// At this point we are preparing command for helm install
	install := action.NewInstall(actionCfg)
	install.ReleaseName = AgentReleaseName
	install.Namespace = AgentNamespace
	install.CreateNamespace = true
	install.Wait = false
	install.Timeout = 30 * time.Second

	chartRef := fmt.Sprintf("%s/%s", h.repoName, h.chart)
	log.Info("locating helm chart", "chart_ref", chartRef)

	chartPath, err := install.LocateChart(chartRef, settings)
	if err != nil {
		log.Error("failed to locate helm chart", "error", err)
		return err
	}

	// Now helm repo has been added this can locate the chart and prepare values for chart, identify templates and validate against schema
	ch, err := loader.Load(chartPath)
	if err != nil {
		log.Error("failed to load helm chart", "error", err)
		return err
	}

	// This step actually installs the chart
	log.Info("running helm install")

	_, err = install.RunWithContext(ctx, ch, cfg.Values)
	if err != nil {
		if strings.Contains(err.Error(), "cannot re-use a name that is still in use") {
			log.Warn("helm release already exists, skipping install")
			upgrade := action.NewUpgrade(actionCfg)
			upgrade.Namespace = AgentNamespace
			upgrade.Install = true
			upgrade.Wait = false
			upgrade.Timeout = 30 * time.Second

			_, uerr := upgrade.RunWithContext(ctx, AgentReleaseName, ch, cfg.Values)
			if uerr != nil {
				log.Error("helm upgrade failed", "error", uerr)
				return fmt.Errorf("helm upgrade failed: %w", uerr)
			}

			log.Info("helm upgrade completed successfully")
			return nil
		}
		log.Error("helm install failed", "error", err)
		return fmt.Errorf("helm install failed: %w", err)
	}

	log.Info("helm install completed successfully")
	return nil
}
