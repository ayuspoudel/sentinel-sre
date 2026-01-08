package admission

import "encoding/json"

/*
@ayuspoudel
As we have already known that k8s api server will send POST request to our webhook server in format of json with apiVersion, kind, request : {with ....}
We will write a struct so that we can decode json request into structs with struct tags with validations, to extract only pieces we need from the request
*/

// We will extract request in here (we have ignored apiVersion and Kind)
type AdmissionReview struct {
	Request *AdmissionRequest `json:"request,omitempty"`
}

// From request we are only getting UID, Kind (also a nested Json), Namespace, Name, Object (object again has a long json object)
type AdmissionRequest struct {
	UID       string          `json:"uid"`
	Kind      GroupKind       `json:"kind"`
	Namespace string          `json:"namespace"`
	Name      string          `json:"name"`
	Object    json.RawMessage `json:"object"`
}

// From kind we only get kind as top level kind has a child level kind
type GroupKind struct {
	Kind string `json:"kind"`
}

// From Object that is raw we will get the metadata
type RawObject struct {
	Metadata ObjectMeta `json:"metadata"`
}

// From object meta we will get Name and Labels
type ObjectMeta struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}
