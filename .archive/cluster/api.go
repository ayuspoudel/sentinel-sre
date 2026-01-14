package cluster

type RegisterRequest struct {
	Name          string            `json:"name"`
	CredentialRef string            `json:"credential_ref"`
	Labels        map[string]string `json:"labels"`
}
