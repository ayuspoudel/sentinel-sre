package api

type RegisterRequest struct {
	Name          string            `json:"name"`
	CredentialRef string            `json:"credential_ref"`
	Labels        map[string]string `json:"labels"`
}

type ClusterResponse struct {
	ClusterName   string            `json:"cluster_name"`
	CredentialRef string            `json:"credential_ref,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	RegisteredAt  string            `json:"registered_at"`
	Source        string            `json:"source"`
}
