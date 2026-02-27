# devzero:index/cluster:Cluster

Manages a DevZero cluster. A cluster represents a Kubernetes environment registered with the DevZero platform. Creating a cluster provisions it and returns a unique authentication token used by the DevZero agent running inside the cluster.

## Example Usage

### Minimal example

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const cluster = new devzero.Cluster("my-cluster", {
    name: "production",
});

export const clusterId = cluster.id;
export const clusterToken = cluster.token; // secret — stored encrypted in state
```

```python
import pulumi_devzero as devzero

cluster = devzero.Cluster("my-cluster",
    name="production",
)

pulumi.export("cluster_id", cluster.id)
pulumi.export("cluster_token", cluster.token)  # secret
```

```go
package main

import (
    devzero "github.com/devzero-inc/pulumi-provider-devzero/sdk/go/devzero"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        cluster, err := devzero.NewCluster(ctx, "my-cluster", &devzero.ClusterArgs{
            Name: pulumi.String("production"),
        })
        if err != nil {
            return err
        }
        ctx.Export("clusterId", cluster.ID())
        ctx.Export("clusterToken", cluster.Token)
        return nil
    })
}
```

### Multiple clusters

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const usEast = new devzero.Cluster("us-east", { name: "production-us-east-1" });
const usWest = new devzero.Cluster("us-west", { name: "production-us-west-2" });
const euWest = new devzero.Cluster("eu-west", { name: "production-eu-west-1" });
```

## Schema

### Required

| Name | Type   | Description                       |
|------|--------|-----------------------------------|
| `name` | string | The name of the cluster. Must be unique within your team. |

### Read-Only (Computed)

| Name    | Type   | Description                                                                               |
|---------|--------|-------------------------------------------------------------------------------------------|
| `id`    | string | Unique identifier of the cluster. Managed by the provider.                                |
| `token` | string | **(Secret)** Authentication token for the cluster agent. Stored encrypted in Pulumi state. Automatically rotated when the resource is imported and then updated. |

## Import

An existing cluster can be imported using its cluster ID:

```shell
pulumi import devzero:index/cluster:Cluster my-cluster <cluster-id>

# Example
pulumi import devzero:index/cluster:Cluster production "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

> **Note:** After importing, the `token` field will be empty in state. The next `pulumi up` will automatically call `ResetClusterToken` to obtain a fresh token.

## Notes

- The cluster `token` is a sensitive value and is encrypted at rest in Pulumi state.
- Deleting the resource via `pulumi destroy` permanently deletes the cluster from the DevZero platform.
