package registryClient

/*
This is not a DB model, but it is what registry will expose over the wire
*/
type Cluster struct {
	Name          string            `json:"name"`
	CredentialRef string            `json:"credential_ref"`
	Labels        map[string]string `json:"labels"`
}
