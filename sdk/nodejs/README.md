# pulumi-provider-devzero

The official [Pulumi](https://www.pulumi.com/) provider for [DevZero](https://devzero.io/), enabling you to manage DevZero infrastructure — Clusters, Workload Policies, and Node Policies — using your preferred programming language.

## Resources

| Resource | Description |
|---|---|
| `Cluster` | Provision and manage a DevZero cluster |
| `WorkloadPolicy` | Configure vertical/horizontal scaling policies for workloads |
| `WorkloadPolicyTarget` | Apply a workload policy to one or more clusters with filters |
| `WorkloadRule` | Pin explicit resource rules to a specific workload (MPA v3) |
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

---

## Quick Start

Pick your language below. Each example creates a **Cluster**, a **WorkloadPolicy** with CPU/memory vertical scaling, a **WorkloadPolicyTarget**, a **NodePolicy**, and a **NodePolicyTarget**.

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

// 2. Create a workload policy with CPU and memory vertical scaling
const policy = new resources.WorkloadPolicy("cpu-scaling-policy", {
    name: "cpu-scaling-policy",
    description: "Policy with CPU and memory vertical scaling enabled",
    actionTriggers: ["on_detection", "on_schedule"],  // apply on pod events AND on schedule
    cronSchedule: "0 2 * * *",                        // daily at 2 am UTC (required for on_schedule)
    detectionTriggers: ["pod_creation", "pod_reschedule"],
    cpuVerticalScaling: {
                enabled: true,
                targetPercentile: 0.75,        // P75 of observed usage
                minRequest: 25,                // millicores; hard floor
                maxScaleUpPercent: 1000,       // % per step
                maxScaleDownPercent: 1,        // % per step
                minDataPoints: 20,             // min CPU samples
                adjustReqEvenIfNotSet: true,   // set requests even if workload has none
                limitsRemovalEnabled: true,    // strip CPU limits (cycles compress safely)
            },
    memoryVerticalScaling: {
                enabled: true,
                targetPercentile: 1,           // P100 — guard against OOMKills
                minRequest: 134217728,         // 128 MiB in bytes; hard floor
                maxScaleUpPercent: 1000,       // % per step
                maxScaleDownPercent: 1,        // % per step
                overheadMultiplier: 0.3,       // extra headroom over the recommendation
                limitsAdjustmentEnabled: true, // adjust limits alongside requests
                limitMultiplier: 1,            // limits = request × this
                minDataPoints: 20,             // min memory samples
                adjustReqEvenIfNotSet: true,   // set requests even if workload has none
            },
            enablePmaxProtection: true,                        // guard against spike-induced OOMKills
            pmaxRatioThreshold: 3,                             // raise requests 3× on an OOM event
            minChangePercent: 0.2,                             // apply only if change > 20%
        });

// 3. Apply the workload policy to the cluster for all Deployments in prod-* namespaces
const workloadTarget = new resources.WorkloadPolicyTarget("prod-cluster-deployments-target", {
    name: "prod-cluster-deployments-target",
    description: "Apply cpu-scaling-policy to all Deployments in prod-cluster",
    policyId: policy.id,
    clusterIds: [cluster.id],
    kindFilter: ["Deployment"],
    enabled: true,
    // Target namespaces by name pattern instead of (or in addition to) labels.
    // Matches any namespace whose name starts with "prod-" (case-insensitive).
    namespacePattern: {
        pattern: "^prod-",
        flags: "i",
    },
});

// 4. Create a node policy for dzkarp-based node provisioning
const nodePolicy = new resources.NodePolicy("prod-node-policy", {
    name: "prod-node-policy",
    description: "Cost-efficient node provisioning for production workloads",

    // Higher weight wins when multiple policies match the same node request.
    weight: 10,

    // Instance categories: c (compute), m (general), r (memory), t (burstable).
    // Kept broad to maximise the instance pool and minimise cost.
    instanceCategories: {
        matchExpressions: [{
            operator: "In",
            values: ["c", "m", "r", "t"],
        }],
    },

    // Instance generation: prefer modern hardware (gen 3+) for better performance/cost ratio.
    instanceGenerations: {
        matchExpressions: [{
            operator: "In",
            values: ["3", "4", "5", "6"],
        }],
    },

    // CPU architecture: amd64 (x86_64) — derived from active nodes in the cluster.
    architectures: {
        matchExpressions: [{ operator: "In", values: ["amd64"] }],
    },

    // Capacity types: prefer spot for savings, fall back to on-demand for availability.
    capacityTypes: {
        matchExpressions: [{ operator: "In", values: ["spot", "on-demand"] }],
    },

    // Operating system: linux only.
    operatingSystems: {
        matchExpressions: [{ operator: "In", values: ["linux"] }],
    },

    // Disruption: how dzkarp consolidates and rotates nodes.
    disruption: {
        consolidationPolicy: "WhenEmptyOrUnderutilized", // reclaim empty and underused nodes
        consolidateAfter: "2h0m0s",                      // wait 2 h before consolidating
        expireAfter: "168h",                             // rotate nodes after 7 days
        budgets: [
            {
                // Disrupt up to 10% of nodes at once for these reasons.
                reasons: ["Empty", "Drifted", "Underutilized"],
                nodes: "10%",
            },
            {
                // Always protect at least 1 node from disruption at any time.
                nodes: "1",
            },
        ],
    },

    // Override the generated dzkarp CRD names (helps avoid collisions in shared clusters).
    nodePoolName: "prod-nodepool",   // name of the dzkarp NodePool CR
    nodeClassName: "prod-nodeclass", // name of the dzkarp NodeClass CR

    // AWS-specific EC2 configuration.
    aws: {
        // Subnets where nodes will launch — discovered via the cluster tag.
        subnetSelectorTerms: [{
            tags: { "karpenter.sh/discovery": "my-prod-cluster" },
        }],
        // Security groups for node instances — same discovery tag pattern.
        securityGroupSelectorTerms: [{
            tags: { "karpenter.sh/discovery": "my-prod-cluster" },
        }],
        // AMI: latest Amazon Linux 2023 managed alias (dzkarp keeps it up to date).
        amiSelectorTerms: [{ alias: "al2023@latest" }],
        // IAM role dzkarp uses to launch and manage nodes (must already exist in AWS).
        role: "KarpenterNodeRole-my-prod-cluster",
    },
});

// 5. Attach the node policy to the cluster
const nodePolicyTarget = new resources.NodePolicyTarget("prod-node-policy-target", {
    name: "prod-node-policy-target",
    description: "Apply prod-node-policy to prod-cluster",
    policyId: nodePolicy.id,
    clusterIds: [cluster.id],
    enabled: true,
});

export const clusterId    = cluster.id;
export const clusterToken = pulumi.secret(cluster.token);
export const policyId     = policy.id;
export const targetId     = workloadTarget.id;
export const nodePolicyId = nodePolicy.id;
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
    NodePolicy, NodePolicyArgs,
    NodePolicyTarget, NodePolicyTargetArgs,
    NamePatternArgs,
)
from pulumi_devzero.resources.types import (
    VerticalScalingArgs,
    LabelSelectorArgs,
    MatchExpressionArgs,
    DisruptionPolicyArgs,
    DisruptionBudgetArgs,
    AWSNodeClassSpecArgs,
    AMISelectorTermArgs,
    SubnetSelectorTermArgs,
    SecurityGroupSelectorTermArgs,
)

# 1. Create a cluster
cluster = Cluster(
    "prod-cluster",
    args=ClusterArgs(name="prod-cluster"),
)

# 2. Create a workload policy with CPU and memory vertical scaling
policy = WorkloadPolicy(
    "cpu-scaling-policy",
    args=WorkloadPolicyArgs(
        name="cpu-scaling-policy",
        description="Policy with CPU and memory vertical scaling enabled",
        action_triggers=["on_detection", "on_schedule"],  # apply on pod events AND on schedule
        cron_schedule="0 2 * * *",                        # daily at 2 am UTC (required for on_schedule)
        detection_triggers=["pod_creation", "pod_reschedule"],
        cpu_vertical_scaling=VerticalScalingArgs(
            enabled=True,
            target_percentile=0.75,          # P75 of observed usage
            min_request=25,                  # millicores; hard floor
            max_scale_up_percent=1000,       # % per step
            max_scale_down_percent=1,        # % per step
            min_data_points=20,              # min CPU samples
            adjust_req_even_if_not_set=True, # set requests even if workload has none
            limits_removal_enabled=True,     # strip CPU limits (cycles compress safely)
        ),
        memory_vertical_scaling=VerticalScalingArgs(
            enabled=True,
            target_percentile=1,             # P100 — guard against OOMKills
            min_request=134217728,           # 128 MiB in bytes; hard floor
            max_scale_up_percent=1000,       # % per step
            max_scale_down_percent=1,        # % per step
            overhead_multiplier=0.3,         # extra headroom over the recommendation
            limits_adjustment_enabled=True,  # adjust limits alongside requests
            limit_multiplier=1,              # limits = request × this
            min_data_points=20,              # min memory samples
            adjust_req_even_if_not_set=True, # set requests even if workload has none
        ),
        enable_pmax_protection=True,         # guard against spike-induced OOMKills
        pmax_ratio_threshold=3,              # raise requests 3× on an OOM event
        min_change_percent=0.2,              # apply only if change > 20%
    ),
)

# 3. Apply the workload policy to the cluster for all Deployments in prod-* namespaces
workload_target = WorkloadPolicyTarget(
    "prod-cluster-deployments-target",
    args=WorkloadPolicyTargetArgs(
        name="prod-cluster-deployments-target",
        policy_id=policy.id,
        cluster_ids=[cluster.id],
        kind_filter=["Deployment"],
        enabled=True,
        # Target namespaces by name pattern instead of (or in addition to) labels.
        # Matches any namespace whose name starts with "prod-" (case-insensitive).
        namespace_pattern=NamePatternArgs(pattern="^prod-", flags="i"),
    ),
)

# 4. Create a node policy for dzkarp-based node provisioning
node_policy = NodePolicy("prod-node-policy", args=NodePolicyArgs(
    name="prod-node-policy",
    description="Cost-efficient node provisioning for production workloads",

    # Higher weight wins when multiple policies match the same node request.
    weight=10,

    # Instance categories: c (compute), m (general), r (memory), t (burstable).
    # Kept broad to maximise the instance pool and minimise cost.
    instance_categories=LabelSelectorArgs(
        match_expressions=[MatchExpressionArgs(
            operator="In",
            values=["c", "m", "r", "t"],
        )],
    ),

    # Instance generation: prefer modern hardware (gen 3+) for better performance/cost ratio.
    instance_generations=LabelSelectorArgs(
        match_expressions=[MatchExpressionArgs(
            operator="In",
            values=["3", "4", "5", "6"],
        )],
    ),

    # CPU architecture: amd64 (x86_64) — derived from active nodes in the cluster.
    architectures=LabelSelectorArgs(
        match_expressions=[MatchExpressionArgs(operator="In", values=["amd64"])],
    ),

    # Capacity types: prefer spot for savings, fall back to on-demand for availability.
    capacity_types=LabelSelectorArgs(
        match_expressions=[MatchExpressionArgs(operator="In", values=["spot", "on-demand"])],
    ),

    # Operating system: linux only.
    operating_systems=LabelSelectorArgs(
        match_expressions=[MatchExpressionArgs(operator="In", values=["linux"])],
    ),

    # Disruption: how dzkarp consolidates and rotates nodes.
    disruption=DisruptionPolicyArgs(
        consolidation_policy="WhenEmptyOrUnderutilized", # reclaim empty and underused nodes
        consolidate_after="2h0m0s",                      # wait 2 h before consolidating
        expire_after="168h",                             # rotate nodes after 7 days
        budgets=[
            DisruptionBudgetArgs(
                # Disrupt up to 10% of nodes at once for these reasons.
                reasons=["Empty", "Drifted", "Underutilized"],
                nodes="10%",
            ),
            DisruptionBudgetArgs(nodes="1"),  # always protect at least 1 node
        ],
    ),

    # Override the generated dzkarp CRD names (helps avoid collisions in shared clusters).
    node_pool_name="prod-nodepool",   # name of the dzkarp NodePool CR
    node_class_name="prod-nodeclass", # name of the dzkarp NodeClass CR

    # AWS-specific EC2 configuration.
    aws=AWSNodeClassSpecArgs(
        # Subnets where nodes will launch — discovered via the cluster tag.
        subnet_selector_terms=[SubnetSelectorTermArgs(
            tags={"karpenter.sh/discovery": "my-prod-cluster"},
        )],
        # Security groups for node instances — same discovery tag pattern.
        security_group_selector_terms=[SecurityGroupSelectorTermArgs(
            tags={"karpenter.sh/discovery": "my-prod-cluster"},
        )],
        # AMI: latest Amazon Linux 2023 managed alias (dzkarp keeps it up to date).
        ami_selector_terms=[AMISelectorTermArgs(alias="al2023@latest")],
        # IAM role dzkarp uses to launch and manage nodes (must already exist in AWS).
        role="KarpenterNodeRole-my-prod-cluster",
    ),
))

# 5. Attach the node policy to the cluster
node_policy_target = NodePolicyTarget("prod-node-policy-target", args=NodePolicyTargetArgs(
    name="prod-node-policy-target",
    description="Apply prod-node-policy to prod-cluster",
    policy_id=node_policy.id,
    cluster_ids=[cluster.id],
    enabled=True,
))

pulumi.export("cluster_id",     cluster.id)
pulumi.export("cluster_token",  pulumi.Output.secret(cluster.token))
pulumi.export("policy_id",      policy.id)
pulumi.export("target_id",      workload_target.id)
pulumi.export("node_policy_id", node_policy.id)
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

        // 2. Create a workload policy with CPU and memory vertical scaling
        policy, err := resources.NewWorkloadPolicy(ctx, "cpu-scaling-policy", &resources.WorkloadPolicyArgs{
            Name:              pulumi.String("cpu-scaling-policy"),
            Description:       pulumi.StringPtr("Policy with CPU and memory vertical scaling enabled"),
            ActionTriggers:    pulumi.StringArray{pulumi.String("on_detection"), pulumi.String("on_schedule")},
            CronSchedule:      pulumi.StringPtr("0 2 * * *"), // daily at 2 am UTC (required for on_schedule)
            DetectionTriggers: pulumi.StringArray{pulumi.String("pod_creation"), pulumi.String("pod_reschedule")},
            CpuVerticalScaling: resources.VerticalScalingArgsArgs{
                Enabled:               pulumi.BoolPtr(true),
                TargetPercentile:      pulumi.Float64Ptr(0.75),  // P75 of observed usage
                MinRequest:            pulumi.IntPtr(25),         // millicores; hard floor
                MaxScaleUpPercent:     pulumi.Float64Ptr(1000),  // % per step
                MaxScaleDownPercent:   pulumi.Float64Ptr(1),     // % per step
                MinDataPoints:         pulumi.IntPtr(20),         // min CPU samples
                AdjustReqEvenIfNotSet: pulumi.BoolPtr(true),     // set requests even if workload has none
                LimitsRemovalEnabled:  pulumi.BoolPtr(true),     // strip CPU limits (cycles compress safely)
            }.ToVerticalScalingArgsPtrOutput(),
            MemoryVerticalScaling: resources.VerticalScalingArgsArgs{
                Enabled:                 pulumi.BoolPtr(true),
                TargetPercentile:        pulumi.Float64Ptr(1.0),      // P100 — guard against OOMKills
                MinRequest:              pulumi.IntPtr(134217728),     // 128 MiB in bytes; hard floor
                MaxScaleUpPercent:       pulumi.Float64Ptr(1000),     // % per step
                MaxScaleDownPercent:     pulumi.Float64Ptr(1),        // % per step
                OverheadMultiplier:      pulumi.Float64Ptr(0.3),      // extra headroom over the recommendation
                LimitsAdjustmentEnabled: pulumi.BoolPtr(true),        // adjust limits alongside requests
                LimitMultiplier:         pulumi.Float64Ptr(1),        // limits = request × this
                MinDataPoints:           pulumi.IntPtr(20),            // min memory samples
                AdjustReqEvenIfNotSet:   pulumi.BoolPtr(true),        // set requests even if workload has none
            }.ToVerticalScalingArgsPtrOutput(),
            EnablePmaxProtection: pulumi.BoolPtr(true),       // guard against spike-induced OOMKills
            PmaxRatioThreshold:   pulumi.Float64Ptr(3),       // raise requests 3× on an OOM event
            MinChangePercent:     pulumi.Float64Ptr(0.2),     // apply only if change > 20%
        })
        if err != nil {
            return err
        }

        // 3. Apply the workload policy to the cluster for all Deployments in prod-* namespaces
        workloadTarget, err := resources.NewWorkloadPolicyTarget(ctx, "prod-cluster-deployments-target", &resources.WorkloadPolicyTargetArgs{
            Name:       pulumi.String("prod-cluster-deployments-target"),
            PolicyId:   policy.ID(),
            ClusterIds: pulumi.StringArray{cluster.ID()},
            KindFilter: pulumi.StringArray{pulumi.String("Deployment")},
            Enabled:    pulumi.BoolPtr(true),
            // Target namespaces by name pattern instead of (or in addition to) labels.
            // Matches any namespace whose name starts with "prod-" (case-insensitive).
            NamespacePattern: resources.NamePatternArgsArgs{
                Pattern: pulumi.StringPtr("^prod-"),
                Flags:   pulumi.StringPtr("i"),
            },
        })
        if err != nil {
            return err
        }

        // 4. Create a node policy for dzkarp-based node provisioning
        nodePolicy, err := resources.NewNodePolicy(ctx, "prod-node-policy", &resources.NodePolicyArgs{
            Name:        pulumi.String("prod-node-policy"),
            Description: pulumi.StringPtr("Cost-efficient node provisioning for production workloads"),
            // Higher weight wins when multiple policies match the same node request.
            Weight: pulumi.IntPtr(10),
            // Instance categories: c (compute), m (general), r (memory), t (burstable).
            // Kept broad to maximise the instance pool and minimise cost.
            InstanceCategories: &resources.LabelSelectorArgs{
                MatchExpressions: resources.MatchExpressionArray{
                    {Operator: pulumi.String("In"),
                        Values: pulumi.StringArray{pulumi.String("c"), pulumi.String("m"), pulumi.String("r"), pulumi.String("t")}},
                },
            },
            // Instance generation: prefer modern hardware (gen 3+) for better performance/cost ratio.
            InstanceGenerations: &resources.LabelSelectorArgs{
                MatchExpressions: resources.MatchExpressionArray{
                    {Operator: pulumi.String("In"),
                        Values: pulumi.StringArray{pulumi.String("3"), pulumi.String("4"), pulumi.String("5"), pulumi.String("6")}},
                },
            },
            // CPU architecture: amd64 (x86_64) — derived from active nodes in the cluster.
            Architectures: &resources.LabelSelectorArgs{
                MatchExpressions: resources.MatchExpressionArray{
                    {Operator: pulumi.String("In"), Values: pulumi.StringArray{pulumi.String("amd64")}},
                },
            },
            // Capacity types: prefer spot for savings, fall back to on-demand for availability.
            CapacityTypes: &resources.LabelSelectorArgs{
                MatchExpressions: resources.MatchExpressionArray{
                    {Operator: pulumi.String("In"), Values: pulumi.StringArray{pulumi.String("spot"), pulumi.String("on-demand")}},
                },
            },
            // Operating system: linux only.
            OperatingSystems: &resources.LabelSelectorArgs{
                MatchExpressions: resources.MatchExpressionArray{
                    {Operator: pulumi.String("In"), Values: pulumi.StringArray{pulumi.String("linux")}},
                },
            },
            // Disruption: how dzkarp consolidates and rotates nodes.
            Disruption: &resources.DisruptionPolicyArgs{
                ConsolidationPolicy: pulumi.StringPtr("WhenEmptyOrUnderutilized"), // reclaim empty and underused nodes
                ConsolidateAfter:    pulumi.StringPtr("2h0m0s"),                   // wait 2 h before consolidating
                ExpireAfter:         pulumi.StringPtr("168h"),                     // rotate nodes after 7 days
                Budgets: resources.DisruptionBudgetArray{
                    {
                        // Disrupt up to 10% of nodes at once for these reasons.
                        Reasons: pulumi.StringArray{pulumi.String("Empty"), pulumi.String("Drifted"), pulumi.String("Underutilized")},
                        Nodes:   pulumi.StringPtr("10%"),
                    },
                    {
                        Nodes: pulumi.StringPtr("1"), // always protect at least 1 node
                    },
                },
            },
            // Override the generated dzkarp CRD names (helps avoid collisions in shared clusters).
            NodePoolName:  pulumi.StringPtr("prod-nodepool"),   // name of the dzkarp NodePool CR
            NodeClassName: pulumi.StringPtr("prod-nodeclass"),  // name of the dzkarp NodeClass CR
            // AWS-specific EC2 configuration.
            Aws: &resources.AWSNodeClassSpecArgs{
                // Subnets where nodes will launch — discovered via the cluster tag.
                SubnetSelectorTerms: resources.SubnetSelectorTermArray{
                    {Tags: pulumi.StringMap{"karpenter.sh/discovery": pulumi.String("my-prod-cluster")}},
                },
                // Security groups for node instances — same discovery tag pattern.
                SecurityGroupSelectorTerms: resources.SecurityGroupSelectorTermArray{
                    {Tags: pulumi.StringMap{"karpenter.sh/discovery": pulumi.String("my-prod-cluster")}},
                },
                // AMI: latest Amazon Linux 2023 managed alias (dzkarp keeps it up to date).
                AmiSelectorTerms: resources.AMISelectorTermArray{
                    {Alias: pulumi.StringPtr("al2023@latest")},
                },
                // IAM role dzkarp uses to launch and manage nodes (must already exist in AWS).
                Role: pulumi.StringPtr("KarpenterNodeRole-my-prod-cluster"),
            },
        })
        if err != nil {
            return err
        }

        // 5. Attach the node policy to the cluster
        _, err = resources.NewNodePolicyTarget(ctx, "prod-node-policy-target", &resources.NodePolicyTargetArgs{
            Name:        pulumi.String("prod-node-policy-target"),
            Description: pulumi.StringPtr("Apply prod-node-policy to prod-cluster"),
            PolicyId:    nodePolicy.ID(),
            ClusterIds:  pulumi.StringArray{cluster.ID()},
            Enabled:     pulumi.BoolPtr(true),
        })
        if err != nil {
            return err
        }

        ctx.Export("clusterId",    cluster.ID())
        ctx.Export("clusterToken", cluster.Token)
        ctx.Export("policyId",     policy.ID())
        ctx.Export("targetId",     workloadTarget.ID())
        ctx.Export("nodePolicyId", nodePolicy.ID())

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

const existing = await resources.getClusterIdByName({
    name: "my-existing-cluster",
    // teamId: "my-team-id",      // optional — defaults to devzero:teamId from provider config
    // region: "us-east-1",       // optional: filter by region
    // cloudProvider: "AWS",      // optional: filter by cloud provider (AWS | GCP | AKS | OCI)
    // liveness: "PREFER_LIVE",   // optional: IGNORE | PREFER_LIVE | REQUIRE_LIVE
});

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

existing = devzero.resources.get_cluster_id_by_name(
    name="my-existing-cluster",
    # team_id="my-team-id",      # optional — defaults to devzero:teamId from provider config
    # region="us-east-1",        # optional: filter by region
    # cloud_provider="AWS",      # optional: filter by cloud provider (AWS | GCP | AKS | OCI)
    # liveness="PREFER_LIVE",    # optional: IGNORE | PREFER_LIVE | REQUIRE_LIVE
)

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
existing, err := resources.GetClusterIdByName(ctx, &resources.GetClusterIdByNameArgs{
    Name: "my-existing-cluster",
    // TeamId:        pulumi.StringRef("my-team-id"),    // optional — defaults to devzero:teamId from provider config
    // Region:        pulumi.StringRef("us-east-1"),     // optional: filter by region
    // CloudProvider: pulumi.StringRef("AWS"),           // optional: filter by cloud provider (AWS | GCP | AKS | OCI)
    // Liveness:      pulumi.StringRef("PREFER_LIVE"),   // optional: IGNORE | PREFER_LIVE | REQUIRE_LIVE
})
if err != nil {
    return err
}

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

ctx.Export("existingClusterId", pulumi.String(existing.ClusterId))
```

**Inputs:**

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Cluster name to look up |
| `teamId` | string | no | Team to search within. Defaults to `devzero:teamId` from provider config |
| `region` | string | no | Filter by region name (e.g. `us-east-1`) |
| `cloudProvider` | string | no | Filter by cloud provider: `AWS`, `GCP`, `AKS`, `OCI` |
| `liveness` | string | no | Heartbeat filter: `IGNORE` (default), `PREFER_LIVE`, `REQUIRE_LIVE` |

**Outputs:**

| Field | Type | Description |
|---|---|---|
| `clusterId` | string | UUID of the matching cluster |

---

## WorkloadPolicy — Key Fields

| Field | Type | Description |
|---|---|---|
| `name` | string | Unique policy name (per team) |
| `description` | string | Human-readable description |
| `cpuVerticalScaling` | `VerticalScalingArgs` | CPU vertical scaling configuration |
| `memoryVerticalScaling` | `VerticalScalingArgs` | Memory vertical scaling configuration |
| `gpuVerticalScaling` | `VerticalScalingArgs` | GPU core vertical scaling configuration (units: GPU millicores) |
| `gpuVramVerticalScaling` | `VerticalScalingArgs` | GPU VRAM vertical scaling configuration (units: bytes) |
| `horizontalScaling` | `HorizontalScalingArgs` | Horizontal (replica) scaling configuration |
| `actionTriggers` | string[] | When to apply recommendations: `on_detection` \| `on_schedule`. Both can be used together. |
| `cronSchedule` | string | 5-field UTC cron expression for scheduled application. Required when `actionTriggers` includes `on_schedule`. Example: `0 2 * * *` |
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

Python uses snake_case for all fields (e.g. `cpu_vertical_scaling`, `action_triggers`, `cron_schedule`, `detection_triggers`, `enable_pmax_protection`, `loopback_period_seconds`, `min_data_points`, `min_change_percent`, `cooldown_minutes`). Go uses PascalCase equivalents.

### VerticalScalingArgs

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Enable this scaling axis |
| `targetPercentile` | float | Percentile of observed usage to target (e.g. `0.95`) |
| `minRequest` | int | Minimum resource request (millicores for CPU, bytes for memory) |
| `maxRequest` | int | Maximum resource request |
| `maxScaleUpPercent` | float | Maximum percentage to scale up in one step. Default: `1000` |
| `maxScaleDownPercent` | float | Maximum percentage to scale down in one step. Default: `1.0` |
| `overheadMultiplier` | float | Safety margin multiplier applied on top of the recommendation |
| `limitsAdjustmentEnabled` | bool | Whether to also adjust resource limits alongside requests |
| `limitMultiplier` | float | Limits = request × limitMultiplier |
| `minDataPoints` | int | Minimum data points required before a recommendation is emitted. Default: `20` |
| `adjustReqEvenIfNotSet` | bool | Recommend requests even when the workload has no existing requests set. Default: `false` |
| `limitsRemovalEnabled` | bool | Actively remove limits from workloads (CPU axis only — memory limits removal is not supported). Takes precedence over `limitsAdjustmentEnabled`. Default: `false` |

Python: `target_percentile`, `min_request`, `max_request`, `max_scale_up_percent`, `max_scale_down_percent`, `overhead_multiplier`, `limits_adjustment_enabled`, `limit_multiplier`, `min_data_points`, `adjust_req_even_if_not_set`, `limits_removal_enabled`. Go uses PascalCase equivalents.

### HorizontalScalingArgs

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Enable horizontal (replica) scaling |
| `minReplicas` | int | Minimum number of replicas to maintain |
| `maxReplicas` | int | Maximum number of replicas to scale to |
| `targetUtilization` | float | Target utilization ratio (0–1) for the primary metric. Example: `0.8` |
| `primaryMetric` | string | Metric driving HPA: `cpu` \| `memory` \| `gpu` \| `network_ingress` \| `network_egress`. Example: `"memory"` |
| `minDataPoints` | int | Minimum data points before a recommendation is emitted |
| `maxReplicaChangePercent` | float | Maximum fraction of current replicas that can change per cycle (0–1). `0.25` = at most 25% added/removed at once. Example: `0.25` |

Python: `min_replicas`, `max_replicas`, `target_utilization`, `primary_metric`, `min_data_points`, `max_replica_change_percent`. Go uses PascalCase equivalents.

---

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
| `namespacePattern` | `NamePatternArgs` | Regex pattern to match namespace names. Useful when namespaces follow a naming convention but lack consistent labels (e.g. `^prod-`). Combined with other criteria using AND logic. |
| `namespaceSelector` | `LabelSelectorArgs` | Select namespaces by labels (`matchLabels` / `matchExpressions`) |
| `workloadSelector` | `LabelSelectorArgs` | Select workloads by labels |
| `enabled` | bool | Activate the target. Default: `true` |

### NamePatternArgs

| Field | Type | Description |
|---|---|---|
| `pattern` | string | Regular expression to match against the name |
| `flags` | string | Optional regex flags. Use `"i"` for case-insensitive matching. Can also be embedded inline in the pattern with `(?i)`. |

Python: `policy_id`, `cluster_ids`, `kind_filter`, `workload_names`, `node_group_names`, `name_pattern`, `namespace_pattern`, `namespace_selector`, `workload_selector`. Go uses PascalCase equivalents.

---

## NodePolicy — Key Fields

`NodePolicy` configures dzkarp-based node provisioning rules. Ensure dzkarp is installed on your target clusters before attaching node policies.

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
| `disruption` | `DisruptionPolicyArgs` | Node disruption and consolidation settings |
| `limits` | `ResourceLimitsArgs` | Max total CPU/memory this policy may provision |
| `nodePoolName` | string | Override name for the generated dzkarp NodePool CR |
| `nodeClassName` | string | Override name for the generated dzkarp NodeClass CR |
| `aws` | `AWSNodeClassSpecArgs` | AWS-specific configuration (AMI, subnets, IAM role, EBS, etc.) |
| `azure` | `AzureNodeClassSpecArgs` | Azure-specific configuration (subnet, image family, disk, etc.) |
| `raw` | `RawKarpenterSpecArgs[]` | Raw Karpenter NodePool/NodeClass YAML (escape hatch) |

Python uses snake_case (e.g. `capacity_types`, `instance_categories`, `instance_families`, `instance_cpus`, `instance_sizes`, `instance_types`, `operating_systems`, `node_pool_name`, `node_class_name`). Go uses PascalCase equivalents.

### DisruptionPolicyArgs

| Field | Type | Description |
|---|---|---|
| `consolidationPolicy` | string | `WhenEmpty` \| `WhenEmptyOrUnderutilized` |
| `consolidateAfter` | string | Wait time after a node is empty before consolidating (e.g. `30s`) |
| `expireAfter` | string | Force-replace nodes after this duration (e.g. `720h`) |
| `ttlSecondsAfterEmpty` | int | Seconds before an empty node is terminated. Deprecated — prefer `consolidateAfter` |
| `terminationGracePeriodSeconds` | int | Grace period before forcefully terminating a draining node |
| `budgets` | `DisruptionBudgetArgs[]` | Limits on how many nodes may be disrupted at once |

### DisruptionBudgetArgs

| Field | Type | Description |
|---|---|---|
| `nodes` | string | Max nodes that can be disrupted at once. Absolute (e.g. `"1"`) or percentage (e.g. `"10%"`) |
| `reasons` | string[] | Disruption reasons this budget applies to: `Empty`, `Drifted`, `Underutilized`. Omit to apply to all. |
| `schedule` | string | Cron expression restricting when this budget is active |
| `duration` | string | Duration the budget is active per schedule cycle (e.g. `"1h"`) |

### AWSNodeClassSpecArgs

| Field | Type | Description |
|---|---|---|
| `amiFamily` | string | AMI family: `AL2`, `AL2023`, `Bottlerocket`, `Windows2019`, `Windows2022` |
| `role` | string | IAM role name for nodes (dzkarp creates the instance profile) |
| `instanceProfile` | string | IAM instance profile name (alternative to `role`) |
| `subnetSelectorTerms` | `SubnetSelectorTermArgs[]` | Subnet selectors (by tag or ID) |
| `securityGroupSelectorTerms` | `SecurityGroupSelectorTermArgs[]` | Security group selectors |
| `capacityReservationSelectorTerms` | `CapacityReservationSelectorTermArgs[]` | EC2 capacity reservation selectors |
| `amiSelectorTerms` | `AMISelectorTermArgs[]` | AMI selectors (by alias, tag, or ID) |
| `blockDeviceMappings` | `BlockDeviceMappingArgs[]` | EBS volume configuration |
| `instanceStorePolicy` | string | NVMe instance store policy. Value: `INSTANCE_STORE_POLICY_RAID0` |
| `tags` | map[string]string | AWS tags applied to all provisioned resources |
| `associatePublicIpAddress` | bool | Assign a public IP to nodes |
| `detailedMonitoring` | bool | Enable CloudWatch detailed monitoring |
| `metadataOptions` | `MetadataOptionsArgs` | EC2 IMDS options (IMDSv2, hop limit, etc.) |
| `kubelet` | `KubeletConfigurationArgs` | Kubelet overrides (maxPods, eviction thresholds, etc.) |
| `userData` | string | Custom launch template user data |
| `context` | string | Additional EC2 launch template context ARN for advanced customization |

Python: `ami_family`, `instance_profile`, `subnet_selector_terms`, `security_group_selector_terms`, `capacity_reservation_selector_terms`, `ami_selector_terms`, `block_device_mappings`, `instance_store_policy`, `associate_public_ip_address`, `detailed_monitoring`, `metadata_options`. Go uses PascalCase equivalents.

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

Python: `vnet_subnet_id`, `image_family`, `os_disk_size_gb`, `fips_mode`, `max_pods`. Go uses PascalCase equivalents.

### RawKarpenterSpecArgs

Use this as an escape hatch when the structured fields don't cover your use case and you need full control over the Karpenter NodePool/NodeClass resources.

| Field | Type | Description |
|---|---|---|
| `nodepoolYaml` | string | Raw YAML for a complete dzkarp NodePool resource |
| `nodeclassYaml` | string | Raw YAML for a complete dzkarp NodeClass resource |

Python: `nodepool_yaml`, `nodeclass_yaml`. Go: `NodepoolYaml`, `NodeclassYaml`.

---

## WorkloadRule

A `WorkloadRule` pins explicit resource rules directly to a single workload (a specific `kind/namespace/name` on a cluster). Unlike `WorkloadPolicy`, which applies a shared policy to many workloads via a `WorkloadPolicyTarget`, a `WorkloadRule` targets one workload and lets you override CPU, memory, GPU, and HPA settings with precise values.

Set `autoGenerate: true` to have the engine automatically compute all rule fields from observed usage. Omit it (or set it to `false`) to provide your own values via `cpuRule`, `memoryRule`, `hpaRule`, etc.

### TypeScript

```typescript
import * as pulumi from "@pulumi/pulumi";
import { resources } from "@devzero/pulumi-devzero";

const rule = new resources.WorkloadRule("my-app-rule", {
    clusterId: "cluster-abc123",
    namespace:  "production",
    kind:       "Deployment",
    name:       "my-api",

    cpuRule: {
        enabled:                 true,    // activate CPU vertical scaling
        minRequest:              10,      // millicores; hard floor for CPU requests
        maxRequest:              4000,    // millicores; hard ceiling for CPU requests (4 cores)
        targetPercentile:        0.95,    // P95 of observed CPU usage to target
        limitsAdjustmentEnabled: true,    // adjust CPU limits alongside requests
        limitMultiplier:         1.5,     // limits = request × 1.5
    },
    memoryRule: {
        enabled:    true,                 // activate memory vertical scaling for this workload
        minRequest: 67108864,             // bytes; hard floor for memory requests (64 MiB)
        maxRequest: 536870912,            // bytes; hard ceiling for memory requests (512 MiB)
    },
    emergencyResponse: {
        oomEnabled:              true,    // react to OOMKills by increasing memory requests
        oomMemoryMultiplier:     1.5,     // multiply memory request by 1.5× on each OOM event
        cpuThrottlingEnabled:    true,    // react to CPU throttling by increasing CPU requests
        cpuThrottlingThreshold:  0.1,     // trigger when throttle ratio exceeds 10%
        cpuThrottlingMultiplier: 1.25,    // multiply CPU request by 1.25× on throttle reaction
    },
    actionTriggers:    ["on_detection"],                   // apply recommendations immediately on pod events
    detectionTriggers: ["pod_creation", "pod_reschedule"], // pod events that trigger a recommendation
});

export const ruleId = rule.id;
```

> **Auto-generate:** Replace the rule body with `autoGenerate: true` to let the engine fill in all fields from observed usage data.
>
> ```typescript
> const rule = new resources.WorkloadRule("my-app-rule", {
>     clusterId:    "cluster-abc123",
>     namespace:    "production",
>     kind:         "Deployment",
>     name:         "my-api",
>     autoGenerate: true,
> });
> ```

---

### Python

```python
import pulumi
from pulumi_devzero.resources import (
    WorkloadRule, WorkloadRuleArgs,
    ResourceRuleConfigArgsArgs,
    EmergencyResponseConfigArgsArgs,
)

rule = WorkloadRule(
    "my-app-rule",
    args=WorkloadRuleArgs(
        cluster_id="cluster-abc123",
        namespace="production",
        kind="Deployment",
        name="my-api",
        cpu_rule=ResourceRuleConfigArgsArgs(
            enabled=True,                    # activate CPU vertical scaling
            min_request=10,                  # millicores; hard floor for CPU requests
            max_request=4000,                # millicores; hard ceiling for CPU requests (4 cores)
            target_percentile=0.95,          # P95 of observed CPU usage to target
            limits_adjustment_enabled=True,  # adjust CPU limits alongside requests
            limit_multiplier=1.5,            # limits = request × 1.5
        ),
        memory_rule=ResourceRuleConfigArgsArgs(
            enabled=True,           # activate memory vertical scaling for this workload
            min_request=67108864,   # bytes; hard floor for memory requests (64 MiB)
            max_request=536870912,  # bytes; hard ceiling for memory requests (512 MiB)
        ),
        emergency_response=EmergencyResponseConfigArgsArgs(
            oom_enabled=True,                  # react to OOMKills by increasing memory requests
            oom_memory_multiplier=1.5,         # multiply memory request by 1.5× on each OOM event
            cpu_throttling_enabled=True,       # react to CPU throttling by increasing CPU requests
            cpu_throttling_threshold=0.1,      # trigger when throttle ratio exceeds 10%
            cpu_throttling_multiplier=1.25,    # multiply CPU request by 1.25× on throttle reaction
        ),
        action_triggers=["on_detection"],                    # apply recommendations immediately on pod events
        detection_triggers=["pod_creation", "pod_reschedule"],  # pod events that trigger a recommendation
    ),
)

pulumi.export("rule_id", rule.id)
```

> **Auto-generate:**
>
> ```python
> rule = WorkloadRule("my-app-rule", args=WorkloadRuleArgs(
>     cluster_id="cluster-abc123", namespace="production",
>     kind="Deployment", name="my-api", auto_generate=True,
> ))
> ```

---

### Go

```go
rule, err := resources.NewWorkloadRule(ctx, "my-app-rule", &resources.WorkloadRuleArgs{
    ClusterId: pulumi.String("cluster-abc123"),
    Namespace: pulumi.String("production"),
    Kind:      pulumi.String("Deployment"),
    Name:      pulumi.String("my-api"),

    CpuRule: resources.ResourceRuleConfigArgsArgs{
        Enabled:                 pulumi.BoolPtr(true),          // activate CPU vertical scaling
        MinRequest:              pulumi.IntPtr(10),             // millicores; hard floor for CPU requests
        MaxRequest:              pulumi.IntPtr(4000),           // millicores; hard ceiling for CPU requests (4 cores)
        TargetPercentile:        pulumi.Float64Ptr(0.95),       // P95 of observed CPU usage to target
        LimitsAdjustmentEnabled: pulumi.BoolPtr(true),          // adjust CPU limits alongside requests
        LimitMultiplier:         pulumi.Float64Ptr(1.5),        // limits = request × 1.5
    }.ToResourceRuleConfigArgsPtrOutput(),
    MemoryRule: resources.ResourceRuleConfigArgsArgs{
        Enabled:    pulumi.BoolPtr(true),           // activate memory vertical scaling for this workload
        MinRequest: pulumi.IntPtr(67108864),        // bytes; hard floor for memory requests (64 MiB)
        MaxRequest: pulumi.IntPtr(536870912),       // bytes; hard ceiling for memory requests (512 MiB)
    }.ToResourceRuleConfigArgsPtrOutput(),
    EmergencyResponse: resources.EmergencyResponseConfigArgsArgs{
        OomEnabled:              pulumi.BoolPtr(true),          // react to OOMKills by increasing memory requests
        OomMemoryMultiplier:     pulumi.Float64Ptr(1.5),        // multiply memory request by 1.5× on each OOM event
        CpuThrottlingEnabled:    pulumi.BoolPtr(true),          // react to CPU throttling by increasing CPU requests
        CpuThrottlingThreshold:  pulumi.Float64Ptr(0.1),        // trigger when throttle ratio exceeds 10%
        CpuThrottlingMultiplier: pulumi.Float64Ptr(1.25),       // multiply CPU request by 1.25× on throttle reaction
    }.ToEmergencyResponseConfigArgsPtrOutput(),
    ActionTriggers:    pulumi.StringArray{pulumi.String("on_detection")},                                        // apply recommendations immediately on pod events
    DetectionTriggers: pulumi.StringArray{pulumi.String("pod_creation"), pulumi.String("pod_reschedule")},       // pod events that trigger a recommendation
})
if err != nil {
    return err
}

ctx.Export("ruleId", rule.ID())
```

> **Auto-generate:**
>
> ```go
> rule, err := resources.NewWorkloadRule(ctx, "my-app-rule", &resources.WorkloadRuleArgs{
>     ClusterId:    pulumi.String("cluster-abc123"),
>     Namespace:    pulumi.String("production"),
>     Kind:         pulumi.String("Deployment"),
>     Name:         pulumi.String("my-api"),
>     AutoGenerate: pulumi.BoolPtr(true),
> })
> ```

---

## WorkloadRule — Key Fields

| Field | Type | Description |
|---|---|---|
| `clusterId` | string | ID of the cluster the workload lives in |
| `namespace` | string | Kubernetes namespace of the workload |
| `kind` | string | Workload kind: `Deployment` \| `StatefulSet` \| `DaemonSet` \| `CronJob` \| `Job` |
| `name` | string | Name of the Kubernetes workload |
| `autoGenerate` | bool | When `true`, the engine fills all rule fields from observed usage; manual fields are ignored |
| `cpuRule` | `ResourceRuleConfigArgs` | CPU vertical scaling rule |
| `memoryRule` | `ResourceRuleConfigArgs` | Memory vertical scaling rule |
| `gpuRule` | `ResourceRuleConfigArgs` | GPU vertical scaling rule (units: GPU millicores) |
| `hpaRule` | `HPARuleConfigArgs` | Horizontal (replica) scaling rule |
| `emergencyResponse` | `EmergencyResponseConfigArgs` | OOM and CPU-throttle emergency reactions |
| `actionTriggers` | string[] | When to apply: `on_detection` \| `on_schedule` |
| `cronSchedule` | string | Cron expression for scheduled application (5-field UTC). Required when `actionTriggers` includes `on_schedule` |
| `detectionTriggers` | string[] | Events that trigger a recommendation: `pod_creation` \| `pod_update` \| `pod_reschedule` |
| `startupPeriodSeconds` | int | Seconds after workload start to exclude from usage data |
| `schedulerPlugins` | string[] | Kubernetes scheduler plugins to activate. Example: `["binpacking"]` |
| `defragmentationSchedule` | string | Cron expression for node defragmentation |
| `liveMigrationEnabled` | bool | Allow live pod migration when applying recommendations without restart |
| `useInPlaceVerticalScaling` | bool | Use in-place pod vertical scaling instead of pod restarts |
| `containers` | `ContainerResourceRuleConfigArgs[]` | Per-container resource overrides. When empty, workload-level rules apply to all containers |

### ResourceRuleConfigArgs

Used for `cpuRule`, `memoryRule`, and `gpuRule` at both the workload and per-container level.

> **Note:** `maxScaleUpPercent` and `maxScaleDownPercent` are **not** supported on per-container rules — set them on the workload-level fields instead.

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Enable this resource axis rule |
| `minRequest` | int | Minimum resource request (millicores for CPU, bytes for memory/GPU) |
| `maxRequest` | int | Maximum resource request |
| `targetPercentile` | float | Percentile of observed usage to target (0–1). Example: `0.95` |
| `maxScaleUpPercent` | float | Maximum percentage to scale up in one step *(workload-level only)* |
| `maxScaleDownPercent` | float | Maximum percentage to scale down in one step *(workload-level only)* |
| `limitsAdjustmentEnabled` | bool | Whether to also adjust resource limits |
| `limitMultiplier` | float | Limits = request × limitMultiplier |
| `limitsRemovalEnabled` | bool | Actively remove limits from workloads (CPU only) |

### HPARuleConfigArgs

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Enable horizontal (replica) scaling |
| `minReplicas` | int | Minimum number of replicas |
| `maxReplicas` | int | Maximum number of replicas |
| `targetUtilization` | float | Target CPU utilization ratio (0–1). Example: `0.8` |
| `targetMemoryUtilization` | float | Target memory utilization ratio (0–1), tuned independently of CPU. Example: `0.65` |
| `primaryMetric` | string | Primary metric driving HPA (used when `metrics` is empty): `HPA_METRIC_TYPE_CPU` \| `HPA_METRIC_TYPE_MEMORY` \| `HPA_METRIC_TYPE_GPU` \| `HPA_METRIC_TYPE_NETWORK_INGRESS` \| `HPA_METRIC_TYPE_NETWORK_EGRESS`. Example: `"HPA_METRIC_TYPE_MEMORY"` |
| `maxReplicaChangePercent` | float | Maximum fraction of current replicas that can change in one scale event (0–1). `0.25` means at most 25% added or removed at once. Example: `0.25` |
| `scaleDownCooldownSeconds` | int | Seconds to wait between scale-down events. Example: `300` |
| `metrics` | `HPAMetricTriggerArgs[]` | External metric triggers only (e.g. Prometheus, queue depth). CPU/Memory/Network are auto-generated by the engine from `primaryMetric` + `targetUtilization` — redeclaring them here has no effect; the engine silently drops them and regenerates its own triggers |
| `compositeFormula` | string | Expression combining multiple metric ratios into one scaling signal. Example: `"0.6*cpu + 0.4*memory"` |
| `behavior` | `HPABehaviorArgs` | Fine-grained scale-up and scale-down behavior policies |
| `fallback` | `HPAFallbackArgs` | Replica fallback when metrics become unavailable |

Python uses snake_case (e.g. `target_memory_utilization`, `scale_down_cooldown_seconds`, `composite_formula`). Go uses PascalCase equivalents.

### HPAMetricTriggerArgs

> **When to use `metrics[]`:** Only add entries here when you need to scale on **external metrics** (e.g. a Prometheus query, request queue depth, or custom business metric). CPU, Memory, and Network triggers are auto-generated by the engine from `primaryMetric` + `targetUtilization` — redeclaring them in `metrics[]` has no effect; the engine silently drops them and regenerates its own triggers.

| Field | Type | Description |
|---|---|---|
| `type` | string | Metric source type. Built-in: `CPU`, `Memory`, `NetworkIngress`, `NetworkEgress`. External: `prometheus` |
| `targetUtilization` | string | Target utilization as a decimal string (resource metrics). Example: `"0.70"` |
| `targetValue` | string | Absolute target value as a string (external/object metrics). Example: `"50000000"` |
| `weight` | string | Weight for composite formula scaling (decimal string). Example: `"0.5"` |
| `metadata` | map[string]string | Free-form key-value pairs passed to the external scaler |
| `serverAddress` | string | Prometheus server URL — packed into `metadata` by the service layer. Example: `"http://prometheus:9090"` |
| `query` | string | PromQL query string — packed into `metadata` by the service layer. Example: `"sum(rate(http_requests_total[2m]))"` |

Python: `target_utilization`, `target_value`, `server_address`. Go uses PascalCase equivalents.

**Prometheus-driven HPA example**

```typescript
// TypeScript
const rule = new resources.WorkloadRule("my-app-rule", {
    clusterId: "cluster-abc123",
    namespace: "production",
    kind: "Deployment",
    name: "my-api",
    hpaRule: {
        enabled: true,                                         // activate horizontal (replica) scaling
        minReplicas: 1,
        maxReplicas: 8,
        primaryMetric: "HPA_METRIC_TYPE_MEMORY",               // primary metric driving HPA decisions
        targetUtilization: 0.8,                                // target 80% utilization for the primary metric
        targetMemoryUtilization: 0.65,                         // target 65% memory utilization, tuned independently
        maxReplicaChangePercent: 0.25,                         // cap scale events at ±25% of current replicas per cycle
        metrics: [
            {
                type: "prometheus",                            // external Prometheus metric
                targetValue: "50000000",                       // absolute target value (e.g. 50 req/s)
                serverAddress: "http://prometheus:9090",       // Prometheus server URL
                query: "sum(rate(http_requests_total[2m]))",   // PromQL query
            },
        ],
        fallback: {
            replicas: 1,                                       // hold at 1 replica when metrics are unavailable
            behavior: "currentReplicas",                       // use the current live replica count as the fallback value
            failureThreshold: 3,                               // activate fallback after 3 consecutive metric failures
        },
        behavior: {
            scaleDown: {
                selectPolicy: "Min",   // apply the most conservative (smallest) scale-down step
                policies: [
                    { type: "Percent", value: 10 }, // remove at most 10% of replicas per cycle
                ],
            },
            scaleUp: {
                selectPolicy: "Max",   // apply the most aggressive (largest) scale-up step
                policies: [
                    { type: "Percent", value: 100 }, // allow up to 100% more replicas per cycle
                ],
            },
        },
    },
});
```

```python
# Python
from pulumi_devzero.resources import (
    WorkloadRule, WorkloadRuleArgs,
    HPARuleConfigArgsArgs,
    HPAMetricTriggerArgsArgs,
    HPAFallbackArgsArgs,
    HPABehaviorArgsArgs,
    HPAScalingRulesArgsArgs,
    HPAScalingPolicyArgsArgs,
)

rule = WorkloadRule("my-app-rule", args=WorkloadRuleArgs(
    cluster_id="cluster-abc123",
    namespace="production",
    kind="Deployment",
    name="my-api",
    hpa_rule=HPARuleConfigArgsArgs(
        enabled=True,                                          # activate horizontal (replica) scaling
        min_replicas=1,
        max_replicas=8,
        primary_metric="HPA_METRIC_TYPE_MEMORY",               # primary metric driving HPA decisions
        target_utilization=0.8,                                # target 80% utilization for the primary metric
        target_memory_utilization=0.65,                        # target 65% memory utilization, tuned independently
        max_replica_change_percent=0.25,                       # cap scale events at ±25% of current replicas per cycle
        metrics=[HPAMetricTriggerArgsArgs(
            type="prometheus",                                 # external Prometheus metric
            target_value="50000000",                           # absolute target value (e.g. 50 req/s)
            server_address="http://prometheus:9090",           # Prometheus server URL
            query="sum(rate(http_requests_total[2m]))",        # PromQL query
        )],
        fallback=HPAFallbackArgsArgs(
            replicas=1,                                        # hold at 1 replica when metrics are unavailable
            behavior="currentReplicas",                        # use the current live replica count as the fallback value
            failure_threshold=3,                               # activate fallback after 3 consecutive metric failures
        ),
        behavior=HPABehaviorArgsArgs(
            scale_down=HPAScalingRulesArgsArgs(
                select_policy="Min",                           # apply the most conservative (smallest) scale-down step
                policies=[HPAScalingPolicyArgsArgs(type="Percent", value=10)],  # remove at most 10% of replicas per cycle
            ),
            scale_up=HPAScalingRulesArgsArgs(
                select_policy="Max",                           # apply the most aggressive (largest) scale-up step
                policies=[
                    HPAScalingPolicyArgsArgs(type="Percent", value=100),  # allow up to 100% more replicas per cycle
                ],
            ),
        ),
    ),
))
```

```go
// Go
rule, err := resources.NewWorkloadRule(ctx, "my-app-rule", &resources.WorkloadRuleArgs{
    ClusterId: pulumi.String("cluster-abc123"),
    Namespace: pulumi.String("production"),
    Kind:      pulumi.String("Deployment"),
    Name:      pulumi.String("my-api"),
    HpaRule: resources.HPARuleConfigArgsArgs{
        Enabled:                 pulumi.BoolPtr(true),                       // activate horizontal (replica) scaling
        MinReplicas:             pulumi.IntPtr(1),
        MaxReplicas:             pulumi.IntPtr(8),
        PrimaryMetric:           pulumi.StringPtr("HPA_METRIC_TYPE_MEMORY"), // primary metric driving HPA decisions
        TargetUtilization:       pulumi.Float64Ptr(0.8),                     // target 80% utilization for the primary metric
        TargetMemoryUtilization: pulumi.Float64Ptr(0.65),                    // target 65% memory utilization, tuned independently
        MaxReplicaChangePercent: pulumi.Float64Ptr(0.25),                    // cap scale events at ±25% of current replicas per cycle
        Metrics: resources.HPAMetricTriggerArgsArray{
            resources.HPAMetricTriggerArgsArgs{
                Type:          pulumi.String("prometheus"),                         // external Prometheus metric
                TargetValue:   pulumi.StringPtr("50000000"),                        // absolute target value (e.g. 50 req/s)
                ServerAddress: pulumi.StringPtr("http://prometheus:9090"),          // Prometheus server URL
                Query:         pulumi.StringPtr("sum(rate(http_requests_total[2m]))"), // PromQL query
            },
        },
        Fallback: resources.HPAFallbackArgsArgs{
            Replicas:         pulumi.Int(1),                            // hold at 1 replica when metrics are unavailable
            Behavior:         pulumi.String("currentReplicas"),         // use the current live replica count as the fallback value
            FailureThreshold: pulumi.Int(3),                            // activate fallback after 3 consecutive metric failures
        }.ToHPAFallbackArgsPtrOutput(),
        Behavior: resources.HPABehaviorArgsArgs{
            ScaleDown: resources.HPAScalingRulesArgsArgs{
                SelectPolicy: pulumi.String("Min"), // apply the most conservative (smallest) scale-down step
                Policies: resources.HPAScalingPolicyArgsArray{
                    resources.HPAScalingPolicyArgsArgs{Type: pulumi.String("Percent"), Value: pulumi.Int(10)}, // remove at most 10% of replicas per cycle
                },
            }.ToHPAScalingRulesArgsPtrOutput(),
            ScaleUp: resources.HPAScalingRulesArgsArgs{
                SelectPolicy: pulumi.String("Max"), // apply the most aggressive (largest) scale-up step
                Policies: resources.HPAScalingPolicyArgsArray{
                    resources.HPAScalingPolicyArgsArgs{Type: pulumi.String("Percent"), Value: pulumi.Int(100)}, // allow up to 100% more replicas per cycle
                },
            }.ToHPAScalingRulesArgsPtrOutput(),
        }.ToHPABehaviorArgsPtrOutput(),
    }.ToHPARuleConfigArgsPtrOutput(),
})
```

### HPAFallbackArgs

| Field | Type | Description |
|---|---|---|
| `replicas` | int | Replica count to use when metrics are unavailable. Example: `1` |
| `behavior` | string | How to apply fallback replicas. One of: `static` (always use `replicas`), `currentReplicas` (keep whatever is running), `currentReplicasIfHigher` (use current only if higher), `currentReplicasIfLower` (use current only if lower). Example: `"currentReplicas"` |
| `failureThreshold` | int | Consecutive metric failures before fallback activates. Example: `3` |

### HPABehaviorArgs

| Field | Type | Description |
|---|---|---|
| `scaleUp` | `HPAScalingRulesArgs` | Scale-up rate limiting and stabilization |
| `scaleDown` | `HPAScalingRulesArgs` | Scale-down rate limiting and stabilization |

### HPAScalingRulesArgs

| Field | Type | Description |
|---|---|---|
| `stabilizationWindowSeconds` | int | Seconds to look back when selecting replica count to avoid flapping. Default: `0` for scale-up, `300` for scale-down |
| `selectPolicy` | string | Which policy wins when multiple match: `Max` \| `Min` \| `Disabled`. Example: `"Max"` |
| `policies` | `HPAScalingPolicyArgs[]` | List of rate-limiting step policies |

### HPAScalingPolicyArgs

| Field | Type | Description |
|---|---|---|
| `type` | string | Policy type: `Pods` (absolute count) \| `Percent` (% of current replicas). Example: `"Percent"` |
| `value` | int | Maximum change allowed per period. Example: `100` |
| `periodSeconds` | int | Time window for this policy in seconds. Example: `60` |

Python: `stabilization_window_seconds`, `select_policy`, `period_seconds`. Go uses PascalCase equivalents.

### EmergencyResponseConfigArgs

| Field | Type | Description |
|---|---|---|
| `oomEnabled` | bool | React to OOM kills by increasing memory requests |
| `oomMemoryMultiplier` | float | Multiplier applied to memory on OOM. Example: `1.5` |
| `cpuThrottlingEnabled` | bool | React to CPU throttling by increasing CPU requests |
| `cpuThrottlingThreshold` | float | Throttle ratio (0–1) that triggers a reaction. Example: `0.1` |
| `cpuThrottlingMultiplier` | float | Multiplier applied to CPU request on throttle reaction. Example: `1.25` |

### ContainerResourceRuleConfigArgs

| Field | Type | Description |
|---|---|---|
| `containerName` | string | Name of the container this config applies to |
| `cpuRule` | `ResourceRuleConfigArgs` | CPU rule for this container |
| `memoryRule` | `ResourceRuleConfigArgs` | Memory rule for this container |
| `gpuRule` | `ResourceRuleConfigArgs` | GPU rule for this container |

## NodePolicyTarget — Key Fields

| Field | Type | Description |
|---|---|---|
| `name` | string | Unique target name |
| `policyId` | string | ID of the `NodePolicy` to apply |
| `clusterIds` | string[] | Cluster IDs to target. **At most 1 entry** — the backend rejects more than one. |
| `description` | string | Human-readable description (optional) |
| `enabled` | bool | Activate the target. Default: `true` |

Python: `policy_id`, `cluster_ids`. Go: `PolicyId`, `ClusterIds`.

> **Note:** `pulumi destroy` removes this resource from Pulumi state but does **not** delete it on the DevZero backend. You must remove it manually via the dashboard or API if needed.

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

---

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

---

## Examples

Ready-to-run examples live in [`examples/`](examples/):

| Language | Path |
|---|---|
| TypeScript | [`examples/typescript/`](examples/typescript/) |
| Python | [`examples/python/`](examples/python/) |
| Go | [`examples/go/`](examples/go/) |

---

## License

[MIT](LICENSE) — Copyright (c) 2026 DevZero Inc.