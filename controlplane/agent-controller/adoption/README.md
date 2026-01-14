# Adoption design 

Helm mgiht fail because:

1. Resources already exist
2. They lack Helm ownership metadata

So adoption must do this:

1. Render Helm manifests locally (no apply)
2. Parse rendered YAML into objects
3. For each object:
    - Check if it exists in the target cluster
    - If exists:
        - Patch Helm ownership metadata
    - If not exists:
        - Ignore (Helm will create it)

This guarantees:

- Helm install succeeds
- Adoption is deterministic
- No guessing resource names
- No hardcoded RBAC logic