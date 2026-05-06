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

## Data Sources

### `getClusterIdByName`

Look up an existing cluster by name and return its ID. Use this when a cluster was **registered manually** (not created by Pulumi) and you need its ID to attach policies, inject into `values.yaml`, or pass to a Kubernetes secret.

> **Note:** If multiple clusters share the same name, the newest one (by `created_at`) is returned by default. Use the `liveness` field to prefer or require a cluster whose zxporter agent has reported a heartbeat within the last 60 minutes.

#### TypeScript

```typescript
import { resources } from "@devzero/pulumi-devzero";

// Look up a manually registered cluster by name
const existing = await resources.getClusterIdByName({
    name: "my-existing-cluster",
    // teamId is optional — defaults to devzero:teamId from provider config
    // region: "us-east-1",        // optional: filter by region
    // cloudProvider: "AWS",       // optional: filter by cloud provider (AWS | GCP | AKS | OCI)
    // liveness: "PREFER_LIVE",    // optional: IGNORE | PREFER_LIVE | REQUIRE_LIVE
});

// Attach a policy to the existing cluster
const target = new resources.WorkloadPolicyTarget("my-target", {
    name: "my-target",
    policyId: policy.id,
    clusterIds: [existing.clusterId],
    kindFilter: ["Deployment"],
    enabled: true,
});

export const existingClusterId = existing.clusterId;
```

#### Python

```python
import pulumi
import pulumi_devzero as devzero

# Look up a manually registered cluster by name
existing = devzero.resources.get_cluster_id_by_name(
    name="my-existing-cluster",
    # team_id is optional — defaults to devzero:teamId from provider config
    # region="us-east-1",        # optional: filter by region
    # cloud_provider="AWS",      # optional: filter by cloud provider (AWS | GCP | AKS | OCI)
    # liveness="PREFER_LIVE",    # optional: IGNORE | PREFER_LIVE | REQUIRE_LIVE
)

# Attach a policy to the existing cluster
target = devzero.resources.WorkloadPolicyTarget("my-target",
    name="my-target",
    policy_id=policy.id,
    cluster_ids=[existing.cluster_id],
    kind_filter=["Deployment"],
    enabled=True,
)

pulumi.export("existing_cluster_id", existing.cluster_id)
```

#### Go

```go
// Look up a manually registered cluster by name
existing, err := resources.GetClusterIdByName(ctx, &resources.GetClusterIdByNameArgs{
    Name: "my-existing-cluster",
    // TeamId is optional — defaults to devzero:teamId from provider config
    // Region:        pulumi.StringRef("us-east-1"),     // optional: filter by region
    // CloudProvider: pulumi.StringRef("AWS"),           // optional: filter by cloud provider (AWS | GCP | AKS | OCI)
    // Liveness:      pulumi.StringRef("PREFER_LIVE"),   // optional: IGNORE | PREFER_LIVE | REQUIRE_LIVE
})
if err != nil {
    return err
}

// Attach a policy to the existing cluster using its looked-up ID
_, err = resources.NewWorkloadPolicyTarget(ctx, "my-target", &resources.WorkloadPolicyTargetArgs{
    Name:       pulumi.String("my-target"),
    PolicyId:   policy.ID(),
    ClusterIds: pulumi.StringArray{pulumi.String(existing.ClusterId)},
    KindFilter: pulumi.StringArray{pulumi.String("Deployment")},
    Enabled:    pulumi.BoolPtr(true),
})
if err != nil {
    return err
}

// Or inject the ID into a Kubernetes secret / values.yaml
ctx.Export("existingClusterId", pulumi.String(existing.ClusterId))
```

**Inputs:**

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Cluster name to look up |
| `teamId` | string | no | Team to search within. Defaults to `devzero:teamId` from provider config |
| `region` | string | no | Filter by region name (e.g. `us-east-1`) |
| `cloudProvider` | string | no | Filter by cloud provider (e.g. `AWS`, `GCP`,`AKS`,`OCI`) |
| `liveness` | string | no | Heartbeat filter. One of: `IGNORE` (default — newest by `created_at`), `PREFER_LIVE` (live clusters first, fallback to newest), `REQUIRE_LIVE` (404 if no heartbeat within 60 min) |

**Outputs:**

| Field | Type | Description |
|---|---|---|
| `clusterId` | string | UUID of the matching cluster |

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
| `gpuVerticalScaling` | `VerticalScalingArgs` | GPU core vertical scaling configuration (same fields as `cpuVerticalScaling`; units: GPU millicores) |
| `gpuVramVerticalScaling` | `VerticalScalingArgs` | GPU VRAM vertical scaling configuration (same fields as `cpuVerticalScaling`; units: bytes) |
| `horizontalScaling` | `HorizontalScalingArgs` | Horizontal (replica) scaling configuration |
| `actionTriggers` | string[] | When to apply recommendations: `on_detection` \| `on_schedule` |
| `cronSchedule` | string | Cron expression for scheduled application (5-field UTC). Required when `actionTriggers` includes `on_schedule`. Example: `0 2 * * *` |
| `detectionTriggers` | string[] | Events that trigger a recommendation: `pod_creation` \| `pod_update` \| `pod_reschedule` |
| `loopbackPeriodSeconds` | int | Seconds of historical usage data to consider. Default: `86400` (24 h) |
| `startupPeriodSeconds` | int | Seconds after workload start to exclude from usage data (avoids cold-start spikes). Example: `300` |
| `liveMigrationEnabled` | bool | Allow live pod migration when applying recommendations without restart. Default: `false` |
| `schedulerPlugins` | string[] | Kubernetes scheduler plugins to activate. Example: `["binpacking"]` |
| `defragmentationSchedule` | string | Cron expression for background node defragmentation. Example: `0 3 * * 0` |
| `enablePmaxProtection` | bool | Raise requests to cover peak usage when max/recommendation ratio exceeds `pmaxRatioThreshold`. Default: `false` |
| `pmaxRatioThreshold` | float | Peak-to-recommendation ratio that triggers pmax protection. Default: `3.0` |
| `minDataPoints` | int | Global minimum data points required before a recommendation is emitted. Default: `15` |
| `minChangePercent` | float | Global minimum relative change (0–1) required before applying a recommendation. Default: `0.2` (20%) |
| `stabilityCvMax` | float | Maximum coefficient of variation (stddev/mean) for a workload to be considered stable enough for VPA. Example: `0.3` |
| `hysteresisVsTarget` | float | Dead-band ratio around the HPA target to suppress VPA/HPA oscillation. Example: `0.1` |
| `driftDeltaPercent` | float | Percentage change from baseline recommendation that triggers a VPA refresh. Example: `20.0` |
| `minVpaWindowDataPoints` | int | Minimum data points in VPA analysis window. Default: `30` |
| `cooldownMinutes` | int | Minutes to wait between applying recommendations. Default: `300` (5 h) |

### VerticalScalingArgs

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Enable this scaling axis |
| `targetPercentile` | float | Percentile of observed usage to target (e.g. `0.95`) |
| `minRequest` | int | Minimum resource request (millicores / MiB) |
| `maxRequest` | int | Maximum resource request (millicores / MiB) |
| `maxScaleUpPercent` | float | Maximum percentage to scale up in one step. Default: `1000` |
| `maxScaleDownPercent` | float | Maximum percentage to scale down in one step. Default: `1.0` |
| `overheadMultiplier` | float | Multiplier added on top of the recommendation |
| `limitsAdjustmentEnabled` | bool | Whether to also adjust resource limits |
| `limitMultiplier` | float | Limits = request × limitMultiplier |
| `minDataPoints` | int | Minimum data points required before a recommendation is emitted. Default: `20` |
| `adjustReqEvenIfNotSet` | bool | Recommend requests even when the workload has no existing requests set. Default: `false` |
| `limitsRemovalEnabled` | bool | Actively remove limits from workloads (CPU axis only — memory limits removal is not supported). Takes precedence over `limitsAdjustmentEnabled`. Default: `false` |

### HorizontalScalingArgs

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Enable horizontal (replica) scaling |
| `minReplicas` | int | Minimum number of replicas to maintain |
| `maxReplicas` | int | Maximum number of replicas to scale to |
| `targetUtilization` | float | Target utilization ratio (0–1) for the primary metric. Example: `0.7` |
| `primaryMetric` | string | Metric driving HPA: `cpu` \| `memory` \| `gpu` \| `network_ingress` \| `network_egress` |
| `minDataPoints` | int | Minimum data points before a recommendation is emitted |
| `maxReplicaChangePercent` | float | Maximum % change in replica count per cycle. Example: `50.0` |

## WorkloadPolicyTarget — Key Fields

| Field | Type | Description |
|---|---|---|
| `name` | string | Unique target name |
| `policyId` | string | ID of the `WorkloadPolicy` to apply |
| `clusterIds` | string[] | IDs of clusters to target |
| `description` | string | Human-readable description (optional) |
| `priority` | int | Evaluation priority; higher value wins when targets overlap |
| `kindFilter` | string[] | Workload kinds: `Pod` \| `Deployment` \| `StatefulSet` \| `DaemonSet` \| `Job` \| `CronJob` \| `ReplicaSet` \| `ReplicationController` \| `Rollout` |
| `workloadNames` | string[] | Explicit list of workload names to include |
| `nodeGroupNames` | string[] | Restrict matching to specific node groups by name |
| `namePattern` | `NamePatternArgs` | Regex pattern to match workload names |
| `namespaceSelector` | `LabelSelectorArgs` | Select namespaces by labels (matchLabels / matchExpressions) |
| `workloadSelector` | `LabelSelectorArgs` | Select workloads by labels |
| `enabled` | bool | Activate the target |

## NodePolicy — Key Fields

| Field | Type | Description |
|---|---|---|
| `name` | string | Unique policy name |
| `description` | string | Human-readable description |
| `weight` | int | Priority when multiple policies match (higher = preferred) |
| `capacityTypes` | `LabelSelectorArgs` | Capacity types: `on-demand` \| `spot` \| `reserved` |
| `instanceCategories` | `LabelSelectorArgs` | Filter by instance category letter: e.g. `m`, `c`, `r` (AWS) or `D`, `E` (Azure) |
| `instanceFamilies` | `LabelSelectorArgs` | Filter instance families (e.g. `c5`, `m5`) |
| `instanceCpus` | `LabelSelectorArgs` | Filter by vCPU count |
| `instanceSizes` | `LabelSelectorArgs` | Filter instance sizes (e.g. `large`, `xlarge`) |
| `instanceTypes` | `LabelSelectorArgs` | Explicit instance types (e.g. `m5.xlarge`) |
| `instanceGenerations` | `LabelSelectorArgs` | Filter by instance generation number (e.g. `2`, `3`) |
| `instanceHypervisors` | `LabelSelectorArgs` | Filter by hypervisor type (e.g. `nitro`) |
| `zones` | `LabelSelectorArgs` | Availability zones to provision into |
| `architectures` | `LabelSelectorArgs` | CPU architectures (e.g. `amd64`, `arm64`) |
| `operatingSystems` | `LabelSelectorArgs` | OS filter (e.g. `linux`, `windows`) |
| `labels` | map[string]string | Labels applied to provisioned nodes |
| `taints` | `TaintArgs[]` | Taints applied to provisioned nodes |
| `disruption` | `DisruptionPolicyArgs` | Node disruption / consolidation settings |
| `limits` | `ResourceLimitsArgs` | Max total CPU/memory this policy may provision |
| `nodePoolName` | string | Override name for the generated Karpenter NodePool resource |
| `nodeClassName` | string | Override name for the generated Karpenter NodeClass resource |
| `aws` | `AWSNodeClassSpecArgs` | AWS-specific configuration (AMI, subnets, IAM role, EBS, etc.) |
| `azure` | `AzureNodeClassSpecArgs` | Azure-specific configuration (subnet, image family, disk, etc.) |
| `raw` | `RawKarpenterSpecArgs[]` | Raw Karpenter NodePool/NodeClass YAML (escape hatch) |

### DisruptionPolicyArgs

| Field | Type | Description |
|---|---|---|
| `consolidationPolicy` | string | `WhenEmpty` \| `WhenEmptyOrUnderutilized` |
| `consolidateAfter` | string | Wait time after node is empty before consolidating (e.g. `30s`) |
| `expireAfter` | string | Force-replace nodes after this duration (e.g. `720h`) |
| `ttlSecondsAfterEmpty` | int | Seconds before an empty node is terminated (deprecated; prefer `consolidateAfter`) |
| `terminationGracePeriodSeconds` | int | Grace period before forcefully terminating a draining node |
| `budgets` | `DisruptionBudgetArgs[]` | Limits on how many nodes may be disrupted at once |

### AWSNodeClassSpecArgs

| Field | Type | Description |
|---|---|---|
| `amiFamily` | string | AMI family: `AL2`, `AL2023`, `Bottlerocket`, `Windows2019`, `Windows2022` |
| `role` | string | IAM role name for nodes (Karpenter creates the instance profile) |
| `instanceProfile` | string | IAM instance profile name (alternative to `role`) |
| `subnetSelectorTerms` | `SubnetSelectorTermArgs[]` | Subnet selectors (by tag or ID) |
| `securityGroupSelectorTerms` | `SecurityGroupSelectorTermArgs[]` | Security group selectors |
| `capacityReservationSelectorTerms` | `CapacityReservationSelectorTermArgs[]` | EC2 capacity reservation selectors |
| `amiSelectorTerms` | `AMISelectorTermArgs[]` | AMI selectors (by alias, tag, or ID) |
| `blockDeviceMappings` | `BlockDeviceMappingArgs[]` | EBS volume configuration |
| `instanceStorePolicy` | string | NVMe instance store policy. Value: `INSTANCE_STORE_POLICY_RAID0` |
| `tags` | map[string]string | AWS tags on all provisioned resources |
| `associatePublicIpAddress` | bool | Assign a public IP to nodes |
| `detailedMonitoring` | bool | Enable CloudWatch detailed monitoring |
| `metadataOptions` | `MetadataOptionsArgs` | EC2 IMDS options (IMDSv2, hop limit, etc.) |
| `kubelet` | `KubeletConfigurationArgs` | Kubelet overrides (maxPods, eviction thresholds, etc.) |
| `userData` | string | Custom launch template user data |
| `context` | string | Additional EC2 launch template context ARN for advanced customization |

### AzureNodeClassSpecArgs

| Field | Type | Description |
|---|---|---|
| `vnetSubnetId` | string | Azure VNet subnet resource ID |
| `imageFamily` | string | Image family: `AzureLinux`, `Ubuntu2204`, etc. |
| `osDiskSizeGb` | int | OS disk size in GB |
| `fipsMode` | string | `Enabled` \| `Disabled` |
| `maxPods` | int | Max pods per node |
| `tags` | map[string]string | Azure tags on provisioned resources |
| `kubelet` | `AzureKubeletConfigurationArgs` | Kubelet overrides for Azure nodes |

### RawKarpenterSpecArgs

| Field | Type | Description |
|---|---|---|
| `nodepoolYaml` | string | Raw YAML for a complete Karpenter NodePool resource |
| `nodeclassYaml` | string | Raw YAML for a complete Karpenter NodeClass resource |

## NodePolicyTarget — Key Fields

| Field | Type | Description |
|---|---|---|
| `name` | string | Unique target name |
| `policyId` | string | ID of the `NodePolicy` to apply |
| `clusterIds` | string[] | Cluster IDs to target. **At most 1 entry** — the backend rejects more than one. |
| `description` | string | Human-readable description (optional) |
| `enabled` | bool | Activate the target |

> **Note:** `pulumi destroy` removes this resource from Pulumi state but does **not** delete it on the DevZero backend — no delete RPC exists for NodePolicyTarget.

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