package registryClient

/*
This is not a DB model, but it is what registry will expose over the wire
*/
type Cluster struct {
	Name       string
	Credential string
	Labels     map[string]string
}
