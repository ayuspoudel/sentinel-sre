package admission

import "encoding/json"

type Meta struct {
	Metadata struct {
		Annotations map[string]string `json:"annotations"`
	} `json: "metadata"`
}

func extractGuardName(req *AdmissionRequest) (string, error) {
	if req == nil {
		return "", nil
	}

	var obj RawObject
	if err := json.Unmarshal(req.Object, &obj); err != nil {
		return "", err
	}

	return obj.Metadata.Labels["sentinel.guard"], nil
}
