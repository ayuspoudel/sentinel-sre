package admission

import "encoding/json"

type AdmissionReview struct {
	Request *AdmissionRequest `json:"request,omitempty"`
}

type AdmissionRequest struct {
	UID       string          `json:"uid"`
	Kind      GroupKind       `json:"kind"`
	Namespace string          `json:"namespace"`
	Name      string          `json:"name"`
	Object    json.RawMessage `json:"object"`
}

type GroupKind struct {
	Kind string `json:"kind"`
}

type RawObject struct {
	Metadata ObjectMeta `json:"metadata"`
}

type ObjectMeta struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}
