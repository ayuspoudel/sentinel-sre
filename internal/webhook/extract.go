package webhook

import "encoding/json"

type Meta struct {
	Metadata struct {
		Annotations map[string]string `json:"annotations"`
	} `json: "metadata"`
}

func ExtractGuardName(review *AdmissionReview) string {
	var meta Meta
	err := json.Unmarshal(review.Request.Object, &meta)
	if err != nil {
		return ""
	}
	return meta.Metadata.Annotations["sentinel.guard/name"]
}
