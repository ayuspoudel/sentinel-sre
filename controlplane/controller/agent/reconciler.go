package agent

import (
	"context"
	"log"

	"github.com/ayuspoudel/sentinel-sre/controlplane/registryClient"
)

func (c *Controller) reconcileCluster(ctx context.Context, cl *registryClient.Cluster) {
	log.Printf("[reconcile] cluster=%s credential_ref=%s labels=%v", cl.Name, cl.CredentialRef, cl.Labels)

}
