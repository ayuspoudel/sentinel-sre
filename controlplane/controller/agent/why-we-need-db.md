# Why Agent Controller needs a DB

The agent controller only answers one question:
> Given the declared intent, what is the observed operational state of sentinel on that cluster

First let's clarify what intent is:

1. Declarative Intent
Cluster Registry already stores this and exposes via registryClient, so we do not need to store this again. 
It has cluster name, credential ref and labels. 
Credential is also stored as secret in the cluster sentinel control plane is running.

2. Reconciliation Metadata
We need these answers out of the box.

- When was the last time we reconciled this cluster?
- How long did it take to reconcile the cluster?
- What was the result of last reconciliation?
- What was the error if reconciliation failed?

We need to store these.

3. Connectivity and Auth State
We need to know if we can reach the cluster and if the credentials are valid.

4. Agent Lifecycle State 
We need to know if the agent is installed, healthy, what version it is running, when was the last heartbeat etc.

### DB Model

```sql
`CREATE TABLE IF NOT EXISTS cluster_status (
		cluster_name TEXT PRIMARY KEY,
		last_reconcile_at TIMESTAMPTZ,
		last_reconcile_duration_ms INTEGER,
		last_reconcile_success BOOLEAN,
		last_error TEXT,
		reachable BOOLEAN,
		auth_valid BOOLEAN,
		api_server_version TEXT,
		last_successful_connection TIMESTAMPTZ,
		agent_installed BOOLEAN,
		agent_version TEXT,
		agent_namespace TEXT,
		agent_healthy BOOLEAN,
		agent_last_heartbeat TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`
```