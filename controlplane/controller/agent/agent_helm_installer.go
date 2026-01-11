package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	// Creates a new helm runtime environment
	// We need to set cluster, context for this
	settings := cli.New()
	// Creating a local kubeconfig file because helm does not accept in memory kubeconfig at runtime
	tmp, err := os.CreateTemp("", "sentinel-kubeconfig-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write([]byte(cfg.KubeConfig))
	if err != nil {
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
		repoFileObj.Add(repoEntry)
		if err := repoFileObj.WriteFile(repoFile, 0644); err != nil {
			return err
		}
	}
	chartRepo, err := repo.NewChartRepository(repoEntry, getter.All(settings))
	if err != nil {
		return err
	}
	chartRepo.CachePath = repoCache
	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("failed to download repo index: %w", err)
	}

	// At this point we are preparing command for helm install
	install := action.NewInstall(actionCfg)
	install.ReleaseName = AgentReleaseName
	install.Namespace = AgentNamespace
	install.CreateNamespace = true
	chartRef := fmt.Sprintf("%s/%s", h.repoName, h.chart)
	chartPath, err := install.LocateChart(chartRef, settings)
	if err != nil {
		return err
	}
	// Now helm repo has been added this can locate the chart and prepare values for chart, identify templates and validate against schema
	ch, err := loader.Load(chartPath)
	if err != nil {
		return err
	}
	// This step actually installs the chart
	_, err = install.RunWithContext(ctx, ch, cfg.Values)
	if err != nil {
		if strings.Contains(err.Error(), "cannot re-use a name that is still in use") {
			return nil
		}
		return fmt.Errorf("helm install failed: %w", err)
	}
	return nil
}

// This function checks if any adoption of existing rbac roles are needed
func needsAdoption(ctx context.Context, client *kubernetes.Clientset) (bool, error) {
	cr, err := client.RbacV1().ClusterRoles().Get(ctx, AgentDeploymentName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	managedBy, ok := cr.Labels["app.kubernetes.io/managed-by"]
	if ok && managedBy == "Helm" {
		return true, nil
	}
	releaseName, ok := cr.Labels["meta.helm.sh/release-name"]
	if ok && releaseName != AgentReleaseName {
		return true, nil
	}
	releaseNamespace, ok := cr.Labels["meta.helm.sh/release-namespace"]
	if ok && releaseNamespace != AgentNamespace {
		return true, nil
	}
	return false, nil

}

/*
This function checks if RBAC is already present, if so it will update to transfer ownership to helm
*/
func adoptClusterRole(ctx context.Context, client *kubernetes.Clientset) error {
	cr, err := client.RbacV1().ClusterRoles().Get(ctx, "sentinel-agent", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if cr.Labels == nil {
		cr.Labels = map[string]string{}
	}
	if cr.Annotations == nil {
		cr.Annotations = map[string]string{}
	}

	cr.Labels["app.kubernetes.io/managed-by"] = "Helm"
	cr.Annotations["meta.helm.sh/release-name"] = AgentReleaseName
	cr.Annotations["meta.helm.sh/release-namespace"] = AgentNamespace

	_, err = client.RbacV1().ClusterRoles().Update(ctx, cr, metav1.UpdateOptions{})
	return err
}
