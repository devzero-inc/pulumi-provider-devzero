# devzero:index/nodePolicyTarget:NodePolicyTarget

Attaches a node policy to specific clusters. A node policy target determines which clusters a node policy applies to, and whether the attachment is currently active.

> **Note:** There is no delete API for node policy targets. Destroying this resource removes it from Pulumi state only; the target continues to exist on the DevZero platform.

## Example Usage

### Minimal example

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const target = new devzero.NodePolicyTarget("production-target", {
    name: "production-clusters",
    policyId: nodePolicy.id,
    clusterIds: [cluster.id],
});
```

### Multiple clusters

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const usEast = new devzero.Cluster("us-east", { name: "production-us-east-1" });
const usWest = new devzero.Cluster("us-west", { name: "production-us-west-2" });
const euWest = new devzero.Cluster("eu-west", { name: "production-eu-west-1" });

const policy = new devzero.NodePolicy("general", {
    name: "general-purpose",
    nodePoolName:  "general-pool",
    nodeClassName: "general-class",
});

const target = new devzero.NodePolicyTarget("all-prod", {
    name:        "all-production-clusters",
    description: "Apply cost optimisation policy to all production clusters",
    policyId:    policy.id,
    enabled:     true,
    clusterIds:  [usEast.id, usWest.id, euWest.id],
});

export const targetId = target.id;
```

### Temporarily disabled target

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const stagingTarget = new devzero.NodePolicyTarget("staging-disabled", {
    name:        "staging-clusters",
    description: "Temporarily disabled while testing a new policy",
    policyId:    policy.id,
    enabled:     false,
    clusterIds:  [stagingCluster.id],
});
```

```python
import pulumi_devzero as devzero

target = devzero.NodePolicyTarget("prod-target",
    name="production-clusters",
    description="Applies general-purpose node policy",
    policy_id=policy.id,
    enabled=True,
    cluster_ids=[us_east.id, us_west.id],
)

pulumi.export("target_id", target.id)
```

```go
package main

import (
    devzero "github.com/devzero-inc/pulumi-provider-devzero/sdk/go/devzero"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        desc := "production clusters"
        target, err := devzero.NewNodePolicyTarget(ctx, "prod-target", &devzero.NodePolicyTargetArgs{
            Name:        pulumi.String("production-clusters"),
            Description: pulumi.StringRef(desc),
            PolicyId:    policy.ID(),
            Enabled:     pulumi.Bool(true),
            ClusterIds:  pulumi.StringArray{cluster.ID()},
        })
        if err != nil {
            return err
        }
        ctx.Export("targetId", target.ID())
        return nil
    })
}
```

## Schema

### Required

| Name         | Type       | Description                                                                       |
|--------------|------------|-----------------------------------------------------------------------------------|
| `name`       | `string`   | Human-friendly name for the target.                                               |
| `policyId`   | `string`   | ID of the node policy to attach. Must reference an existing `devzero.NodePolicy`. |
| `clusterIds` | `string[]` | List of cluster IDs to apply the node policy to.                                  |

### Optional

| Name          | Type      | Description                                                                                    |
|---------------|-----------|------------------------------------------------------------------------------------------------|
| `description` | `string`  | Free-form description of the target.                                                           |
| `enabled`     | `boolean` | Whether this target is active. When `false`, the policy is not applied to the clusters. Set explicitly to `true` to activate. |

### Read-Only

| Name | Type   | Description                                               |
|------|--------|-----------------------------------------------------------|
| `id` | string | Unique identifier of the target. Managed by the provider. |

## Import

An existing node policy target can be imported using its target ID:

```shell
pulumi import devzero:index/nodePolicyTarget:NodePolicyTarget my-target <target-id>

# Example
pulumi import devzero:index/nodePolicyTarget:NodePolicyTarget production "c84ccd96-d3f6-439d-9976-360577123fe0"
```

## Notes

- Because there is no delete API, `pulumi destroy` removes the resource from Pulumi state only. The target continues to exist on the DevZero platform.
- To effectively disable a target without destroying it, set `enabled: false` and run `pulumi up`.
- Multiple targets can reference the same policy, each covering a different set of clusters.
