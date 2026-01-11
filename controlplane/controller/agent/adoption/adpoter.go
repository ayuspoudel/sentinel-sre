package adoption

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/logging"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func Adopt(ctx context.Context, client dynamic.Interface, obj runtime.Object, ownership Ownership) error {
	log := logging.From(ctx)

	accessor, err := meta.Accessor(obj)
	if err != nil {
		log.Error("failed to get object accessor", "error", err)
		return err
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	mapping := resourceMapping(gvk)
	if mapping == nil {
		log.Info("resource kind not supported for adoption", "kind", gvk.Kind)
		return nil
	}

	log.Info("checking resource for adoption", "kind", gvk.Kind, "name", accessor.GetName(), "namespace", accessor.GetNamespace())

	res := client.Resource(*mapping).Namespace(accessor.GetNamespace())
	_, err = res.Get(ctx, accessor.GetName(), metav1.GetOptions{})
	if err != nil {
		log.Info("resource not present, skipping adoption", "kind", gvk.Kind, "name", accessor.GetName())
		return nil
	}

	patch := []byte(`{
		"metadata": {
			"labels": {
				"app.kubernetes.io/managed-by": "Helm"
			},
			"annotations": {
				"meta.helm.sh/release-name": "` + ownership.ReleaseName + `",
				"meta.helm.sh/release-namespace": "` + ownership.ReleaseNamespace + `"
			}
		}
	}`)

	log.Info("adopting existing resource", "kind", gvk.Kind, "name", accessor.GetName(), "release", ownership.ReleaseName)

	_, err = res.Patch(ctx, accessor.GetName(), types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		log.Error("failed to adopt resource", "kind", gvk.Kind, "name", accessor.GetName(), "error", err)
		return err
	}

	log.Info("resource adopted successfully", "kind", gvk.Kind, "name", accessor.GetName())
	return nil
}
