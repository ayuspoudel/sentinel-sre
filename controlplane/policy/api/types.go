package api

import "github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"

type ApplyPolicyRequest struct {
	Metadata spec.Metadata `json:"metadata"`
	Target   spec.Target   `json:"target"`
	Signals  spec.Signals  `json:"signals"`
	Policy   spec.Policy   `json:"policy"`
}

type ApplyPolicyResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
