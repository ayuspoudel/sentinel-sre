# What is this agent

Agent is a small go binary which will run in target clusters. It will have three responsibilites:
1. Expose an admission webhook
    - This intercepts CREATE/UPDATE of workloads (Deployments)

2. Ask Sentinel Control plane
    - Is deployment X allowed right now?

3. Enforce the answer
    - Allow -> Let kubernetes continue
    - Block -> Reject Admission