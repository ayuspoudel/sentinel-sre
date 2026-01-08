# What does K8s API server send as a request

The kubernetes API always sends requests to your webhook server in a format (json) of kind:

```json
apiVersion:
kind:
request: {
    uid: string
    kind: { group: string, version: string, kind: string }
    resource: { group: string, version: string, resource: string }
    subResource: string
    requestKind: { group: string, version: string, kind: string }
    requestResource: { group: string, version: string, resource: string }
    requestSubResource: string
    name: string
    namespace: string
    operation: string
    userInfo: {
        username: string
        uid: string
        groups: [string]
        extra: { string: [string] }
    }
    object: {...}          // The new object being created/modified
    oldObject: {...}       // The existing object. Only populated for UPDATE and DELETE requests.
    dryRun: boolean
}
```

So a webhook server must be configured to recieve such HTTPS requests with POST http method. 

The webhook server must respond in the format:

```json
apiVersion:
kind:
response: {
    uid: string
    allowed: boolean
    status: {
        metadata: {...}
        status: string
        message: string
        reason: string
        code: int32
    }
    patchType: string
    patch: string
}
```
The `allowed` field is a boolean which tells kubernetes API server whether to allow or deny the request.

The `patch` field is a base64 encoded string of json patch which kubernetes API server will apply to the object being created/modified. The `patchType` field must be set to `JSONPatch` if a patch is provided.
For more information, refer to the official kubernetes documentation on [admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).