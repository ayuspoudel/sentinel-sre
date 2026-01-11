package adoption

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func Adopt(ctx context.Context, client dynamic.Interface, obj runtime.Object, ownership Ownership) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	mapping := resourceMapping(gvk)
	if mapping == nil {
		return nil
	}

	res := client.Resource(*mapping).Namespace(accessor.GetNamespace())
	_, err = res.Get(ctx, accessor.GetName(), metav1.GetOptions{})
	if err != nil {
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

	_, err = res.Patch(ctx, accessor.GetName(), types.MergePatchType, patch, metav1.PatchOptions{})
	return err
}
