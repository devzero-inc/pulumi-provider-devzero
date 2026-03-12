# pulumi-provider-devzero

The official [Pulumi](https://www.pulumi.com/) provider for [DevZero](https://devzero.io/), enabling you to manage DevZero infrastructure — Clusters, Workload Policies, and Node Policies — using your preferred programming language.

## Resources

| Resource | Description |
|---|---|
| `Cluster` | Provision and manage a DevZero cluster |
| `WorkloadPolicy` | Configure vertical/horizontal scaling policies for workloads |
| `WorkloadPolicyTarget` | Apply a workload policy to one or more clusters with filters |
| `NodePolicy` | Configure node provisioning and pooling (AWS / Azure) |
| `NodePolicyTarget` | Apply a node policy to one or more clusters |

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/) v3+
- A DevZero account and API token
- The provider binary in `bin/` (see [Building from source](#building-from-source))

## Installation

### TypeScript / JavaScript

```bash
npm install @devzero/pulumi-devzero
```

### Python

```bash
pip install pulumi-devzero
```

### Go

```bash
go get github.com/devzero-inc/pulumi-provider-devzero/sdk/go/devzero
```

## Configuration

Set your DevZero API endpoint and credentials via Pulumi config or environment variables:

```bash
pulumi config set --secret devzero:token <YOUR_PAT_TOKEN>
pulumi config set devzero:teamId <TEAM_ID>
pulumi config set devzero:url https://dakr.devzero.io  # optional, this is the default
```

## Quick Start

Pick your language below. Each example creates a **Cluster**, a **WorkloadPolicy** with CPU vertical scaling, and a **WorkloadPolicyTarget** that applies the policy to all `Deployment` workloads in that cluster.

---

### TypeScript

**Setup**

```bash
mkdir my-devzero-ts && cd my-devzero-ts
pulumi new typescript
npm install @devzero/pulumi-devzero
```

**`index.ts`**

```typescript
import * as pulumi from "@pulumi/pulumi";
import { resources } from "@devzero/pulumi-devzero";

// 1. Create a cluster
const cluster = new resources.Cluster("prod-cluster", {
    name: "prod-cluster",
});

// 2. Create a workload policy with CPU vertical scaling
const policy = new resources.WorkloadPolicy("cpu-scaling-policy", {
    name: "cpu-scaling-policy",
    description: "Policy with CPU vertical scaling enabled",
    cpuVerticalScaling: {
        enabled: true,
        targetPercentile: 0.95,
        minRequest: 50,
        maxRequest: 4000,
        maxScaleUpPercent: 100,
        maxScaleDownPercent: 25,
        overheadMultiplier: 1.1,
        limitsAdjustmentEnabled: true,
        limitMultiplier: 1.5,
    },
});

// 3. Apply the policy to the cluster for all Deployments
const target = new resources.WorkloadPolicyTarget("prod-cluster-deployments-target", {
    name: "prod-cluster-deployments-target",
    description: "Apply cpu-scaling-policy to all Deployments in prod-cluster",
    policyId: policy.id,
    clusterIds: [cluster.id],
    kindFilter: ["Deployment"],
    enabled: true,
});

export const clusterId    = cluster.id;
export const clusterToken = pulumi.secret(cluster.token);
export const policyId     = policy.id;
export const targetId     = target.id;
```

**Deploy**

```bash
npm run build
pulumi up
```

---

### Python

**Setup**

```bash
mkdir my-devzero-py && cd my-devzero-py
pulumi new python
pip install pulumi-devzero
```

**`__main__.py`**

```python
import pulumi
from pulumi_devzero.resources import (
    Cluster, ClusterArgs,
    WorkloadPolicy, WorkloadPolicyArgs,
    WorkloadPolicyTarget, WorkloadPolicyTargetArgs,
    VerticalScalingArgsArgs,
)

# 1. Create a cluster
cluster = Cluster(
    "prod-cluster",
    args=ClusterArgs(name="prod-cluster"),
)

# 2. Create a workload policy with CPU vertical scaling
policy = WorkloadPolicy(
    "cpu-scaling-policy",
    args=WorkloadPolicyArgs(
        name="cpu-scaling-policy",
        description="Workload policy with CPU vertical scaling for production cluster",
        cpu_vertical_scaling=VerticalScalingArgsArgs(
            enabled=True,
            target_percentile=0.95,
            min_request=50,
            max_request=4000,
            max_scale_up_percent=100,
            max_scale_down_percent=25,
            overhead_multiplier=1.1,
            limits_adjustment_enabled=True,
            limit_multiplier=1.5,
        ),
    ),
)

# 3. Apply the policy to the cluster for all Deployments
target = WorkloadPolicyTarget(
    "prod-cluster-deployments-target",
    args=WorkloadPolicyTargetArgs(
        name="prod-cluster-deployments-target",
        policy_id=policy.id,
        cluster_ids=[cluster.id],
        kind_filter=["Deployment"],
        enabled=True,
    ),
)

pulumi.export("cluster_id",    cluster.id)
pulumi.export("cluster_token", pulumi.Output.secret(cluster.token))
pulumi.export("policy_id",     policy.id)
pulumi.export("target_id",     target.id)
```

**Deploy**

```bash
pulumi up
```

---

### Go

**Setup**

```bash
mkdir my-devzero-go && cd my-devzero-go
pulumi new go
go get github.com/devzero-inc/pulumi-provider-devzero/sdk/go/devzero
```

**`main.go`**

```go
package main

import (
    "github.com/devzero-inc/pulumi-provider-devzero/sdk/go/devzero/resources"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        // 1. Create a cluster
        cluster, err := resources.NewCluster(ctx, "prod-cluster", &resources.ClusterArgs{
            Name: pulumi.String("prod-cluster"),
        })
        if err != nil {
            return err
        }

        // 2. Create a workload policy with CPU vertical scaling
        policy, err := resources.NewWorkloadPolicy(ctx, "cpu-scaling-policy", &resources.WorkloadPolicyArgs{
            Name:        pulumi.String("cpu-scaling-policy"),
            Description: pulumi.StringPtr("Policy with CPU vertical scaling enabled"),
            CpuVerticalScaling: resources.VerticalScalingArgsArgs{
                Enabled:                 pulumi.BoolPtr(true),
                TargetPercentile:        pulumi.Float64Ptr(0.95),
                MinRequest:              pulumi.IntPtr(50),
                MaxRequest:              pulumi.IntPtr(4000),
                MaxScaleUpPercent:       pulumi.Float64Ptr(100),
                MaxScaleDownPercent:     pulumi.Float64Ptr(25),
                OverheadMultiplier:      pulumi.Float64Ptr(1.1),
                LimitsAdjustmentEnabled: pulumi.BoolPtr(true),
                LimitMultiplier:         pulumi.Float64Ptr(1.5),
            }.ToVerticalScalingArgsPtrOutput(),
        })
        if err != nil {
            return err
        }

        // 3. Apply the policy to the cluster for all Deployments
        _, err = resources.NewWorkloadPolicyTarget(ctx, "prod-cluster-deployments-target", &resources.WorkloadPolicyTargetArgs{
            Name:       pulumi.String("prod-cluster-deployments-target"),
            PolicyId:   policy.ID(),
            ClusterIds: pulumi.StringArray{cluster.ID()},
            KindFilter: pulumi.StringArray{pulumi.String("Deployment")},
            Enabled:    pulumi.BoolPtr(true),
        })
        if err != nil {
            return err
        }

        ctx.Export("clusterId",    cluster.ID())
        ctx.Export("clusterToken", cluster.Token)
        ctx.Export("policyId",     policy.ID())

        return nil
    })
}
```

**Deploy**

```bash
go build -o devzero-example .
pulumi up
```

---

## Destroying Resources

To tear down all resources managed by your stack:

```bash
pulumi destroy
```

To also remove the stack itself:

```bash
pulumi stack rm <stack-name>
```

## WorkloadPolicy — Key Fields

| Field | Type | Description |
|---|---|---|
| `name` | string | Unique policy name (per team) |
| `description` | string | Human-readable description |
| `cpuVerticalScaling` | `VerticalScalingArgs` | CPU vertical scaling configuration |
| `memoryVerticalScaling` | `VerticalScalingArgs` | Memory vertical scaling configuration |
| `horizontalScaling` | `HorizontalScalingArgs` | Horizontal (replica) scaling configuration |
| `actionTriggers` | string[] | `on_detection` \| `on_schedule` |
| `detectionTriggers` | string[] | `pod_creation` \| `pod_update` \| `pod_reschedule` |

### VerticalScalingArgs

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Enable this scaling axis |
| `targetPercentile` | float | Percentile of observed usage to target (e.g. `0.95`) |
| `minRequest` | int | Minimum resource request (millicores / MiB) |
| `maxRequest` | int | Maximum resource request (millicores / MiB) |
| `maxScaleUpPercent` | float | Maximum percentage to scale up in one step |
| `maxScaleDownPercent` | float | Maximum percentage to scale down in one step |
| `overheadMultiplier` | float | Multiplier added on top of the recommendation |
| `limitsAdjustmentEnabled` | bool | Whether to also adjust resource limits |
| `limitMultiplier` | float | Limits = request × limitMultiplier |

## WorkloadPolicyTarget — Key Fields

| Field | Type | Description |
|---|---|---|
| `name` | string | Unique target name |
| `policyId` | string | ID of the `WorkloadPolicy` to apply |
| `clusterIds` | string[] | IDs of clusters to target |
| `kindFilter` | string[] | Workload kinds: `Pod`, `Deployment`, `StatefulSet`, `DaemonSet`, `Job`, `CronJob`, `ReplicaSet`, `ReplicationController`, `Rollout` |
| `namespaceFilter` | string[] | Restrict to specific namespaces |
| `enabled` | bool | Activate the target |

## Building from Source

```bash
# Build the provider binary
make build

# Run tests
make test

# Regenerate schema and all SDKs (requires Pulumi CLI)
make sdk

# Install binary to $GOPATH/bin
make install
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full development instructions.

## Examples

Ready-to-run examples live in [`examples/`](examples/):

| Language | Path |
|---|---|
| TypeScript | [`examples/typescript/`](examples/typescript/) |
| Python | [`examples/python/`](examples/python/) |
| Go | [`examples/go/`](examples/go/) |

## License

[MIT](LICENSE) — Copyright (c) 2026 DevZero Inc.