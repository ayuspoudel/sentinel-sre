package adoption

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func resourceMapping(gvk schema.GroupVersionKind) *schema.GroupVersionResource {
	switch gvk.Kind {
	case "Deployment":
		return &schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case "Service":
		return &schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "ServiceAccount":
		return &schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccounts"}
	case "ClusterRole":
		return &schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}
	case "ClusterRoleBinding":
		return &schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}
	default:
		return nil
	}
}
