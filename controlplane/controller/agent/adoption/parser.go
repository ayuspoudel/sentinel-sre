package adoption

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func ParseManifests(manifest string) ([]runtime.Object, error) {
	var objects []runtime.Object
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 4096)

	for {
		var raw runtime.RawExtension
		if err := decoder.Decode(&raw); err != nil {
			break
		}
		if len(raw.Raw) == 0 {
			continue
		}
		obj, _, err := codecs.UniversalDeserializer().Decode(raw.Raw, nil, nil)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}
	return objects, nil
}
