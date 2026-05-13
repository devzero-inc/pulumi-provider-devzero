# @devzero/pulumi-devzero

> The official **TypeScript / JavaScript** Pulumi provider for [DevZero](https://devzero.io/) — manage clusters, workload policies, and node policies as code.

[![npm version](https://img.shields.io/npm/v/@devzero/pulumi-devzero.svg?style=flat-square)](https://www.npmjs.com/package/@devzero/pulumi-devzero)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](https://github.com/devzero-inc/pulumi-provider-devzero/blob/main/LICENSE)
[![Pulumi Registry](https://img.shields.io/badge/Pulumi-Registry-blueviolet?style=flat-square&logo=pulumi)](https://www.pulumi.com/registry/)

---

## Installation

```bash
npm install @devzero/pulumi-devzero
# or
yarn add @devzero/pulumi-devzero
```

**Requires:** `@pulumi/pulumi` v3+

---

## Configuration

Before using the DevZero Pulumi provider, configure your credentials using Pulumi config.

### 1. Generate a Personal Access Token (PAT)

Go to the DevZero user settings page to generate your PAT token:

https://www.devzero.io/settings/user-settings/general

Create a **Personal Access Token** and copy it.

### 2. Find your Team ID

You can find your DevZero **Team ID** in the organization settings:

https://www.devzero.io/settings/organization-settings/account

Copy the **Team ID** value from this page.

### 3. Set Pulumi configuration

Run the following commands in your Pulumi project:

```bash
pulumi config set --secret devzero:token <YOUR_PAT_TOKEN>
pulumi config set devzero:teamId <TEAM_ID>
pulumi config set devzero:url https://dakr.devzero.io  # optional, this is the default
```

### Example

```bash
pulumi config set --secret devzero:token dz_pat_xxxxxxxxxxxxx
pulumi config set devzero:teamId team_123456789
```

The `--secret` flag ensures that your token is encrypted in the Pulumi state.

---

## Quick Start

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
const target = new resources.WorkloadPolicyTarget("prod-cluster-target", {
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

```bash
npm run build && pulumi up
```

---

## Resources

### `Cluster`

Provision and manage a DevZero cluster.

```typescript
const cluster = new resources.Cluster("my-cluster", {
    name: "my-cluster",
});

export const id    = cluster.id;
export const token = pulumi.secret(cluster.token);
```

---

### `WorkloadPolicy`

Configure vertical and horizontal scaling policies for workloads.

```typescript
const policy = new resources.WorkloadPolicy("my-policy", {
    name: "my-policy",
    description: "Vertical scaling for CPU and memory",
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
    memoryVerticalScaling: {
        enabled: true,
        targetPercentile: 0.9,
        minRequest: 128,
        maxRequest: 8192,
        maxScaleUpPercent: 50,
        maxScaleDownPercent: 20,
        overheadMultiplier: 1.2,
        limitsAdjustmentEnabled: true,
        limitMultiplier: 1.3,
    },
});
```

**`VerticalScalingArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `enabled` | `boolean` | Enable this scaling axis |
| `targetPercentile` | `number` | Percentile of observed usage to target (e.g. `0.95`) |
| `minRequest` | `number` | Minimum resource request (millicores / MiB) |
| `maxRequest` | `number` | Maximum resource request (millicores / MiB) |
| `maxScaleUpPercent` | `number` | Max % to scale up in one step. Default: `1000` |
| `maxScaleDownPercent` | `number` | Max % to scale down in one step. Default: `1.0` |
| `overheadMultiplier` | `number` | Multiplier added on top of the recommendation |
| `limitsAdjustmentEnabled` | `boolean` | Whether to also adjust resource limits |
| `limitMultiplier` | `number` | Limits = request × limitMultiplier |
| `minDataPoints` | `number` | Minimum data points before a recommendation is emitted. Default: `20` |
| `adjustReqEvenIfNotSet` | `boolean` | Recommend requests even when the workload has no existing requests set. Default: `false` |
| `limitsRemovalEnabled` | `boolean` | Actively remove limits from workloads (CPU axis only — memory limits removal is not supported). Takes precedence over `limitsAdjustmentEnabled`. Default: `false` |

**`WorkloadPolicy` pmax & VPA knob fields:**

| Field | Type | Description |
|---|---|---|
| `enablePmaxProtection` | `boolean` | Raise requests to cover peak usage when max/recommendation ratio exceeds `pmaxRatioThreshold`. Default: `false` |
| `pmaxRatioThreshold` | `number` | Max-to-recommendation ratio that triggers pmax protection. Default: `3.0` |
| `loopbackPeriodSeconds` | `number` | Look-back period in seconds for usage data. Default: `86400` (24 h) |
| `minDataPoints` | `number` | Global minimum data points for recommendations. Default: `15` |
| `minChangePercent` | `number` | Global minimum change threshold. Default: `0.2` (20%) |
| `minVpaWindowDataPoints` | `number` | Minimum data points in VPA analysis window. Default: `30` |
| `cooldownMinutes` | `number` | Minutes between applying recommendations. Default: `300` (5 h) |

---

### `WorkloadPolicyTarget`

Apply a workload policy to one or more clusters with optional filters.

```typescript
const target = new resources.WorkloadPolicyTarget("my-target", {
    name: "my-target",
    policyId: policy.id,
    clusterIds: [cluster.id],
    kindFilter: ["Deployment", "StatefulSet"],
    namespaceSelector: { matchLabels: { "env": "production" } },
    enabled: true,
});
```

**Fields:**

| Field | Type | Description |
|---|---|---|
| `name` | `string` | Unique target name |
| `policyId` | `string` | ID of the `WorkloadPolicy` to apply |
| `clusterIds` | `string[]` | Cluster IDs to target |
| `description` | `string` | Human-readable description (optional) |
| `priority` | `number` | Evaluation priority; higher value wins when targets overlap |
| `kindFilter` | `string[]` | `Pod` \| `Deployment` \| `StatefulSet` \| `DaemonSet` \| `Job` \| `CronJob` \| `ReplicaSet` \| `ReplicationController` \| `Rollout` |
| `workloadNames` | `string[]` | Explicit list of workload names to include |
| `nodeGroupNames` | `string[]` | Restrict matching to specific node groups by name |
| `namePattern` | `NamePatternArgs` | Regex pattern to match workload names |
| `namespaceSelector` | `LabelSelectorArgs` | Select namespaces by labels (matchLabels / matchExpressions) |
| `workloadSelector` | `LabelSelectorArgs` | Select workloads by labels |
| `enabled` | `boolean` | Activate the target |

---

### `NodePolicy`

Configure node provisioning and pooling (AWS / Azure) using Karpenter under the hood.

```typescript
const nodePolicy = new resources.NodePolicy("my-node-policy", {
    name: "my-node-policy",
    description: "AWS node policy with on-demand and spot capacity",
    weight: 10,
    capacityTypes: {
        matchExpressions: [{ key: "<capacity-type-label>", operator: "In", values: ["<value>"] }],
    },
    instanceCategories: {
        matchLabels: { "<label-key>": "<label-value>" },
    },
    labels: { "<key>": "<value>" },
    taints: [{ key: "<taint-key>", value: "<taint-value>", effect: "NoSchedule" }],
    disruption: {
        consolidationPolicy: "WhenEmptyOrUnderutilized",
        consolidateAfter: "30s",
        expireAfter: "720h",
    },
    limits: { cpu: "1000", memory: "1000Gi" },
    aws: {
        amiFamily: "AL2",
        role: "<iam-role-name>",
        subnetSelectorTerms: [{ tags: { "<tag-key>": "<tag-value>" } }],
        securityGroupSelectorTerms: [{ tags: { "<tag-key>": "<tag-value>" } }],
        blockDeviceMappings: [{
            deviceName: "/dev/xvda",
            rootVolume: true,
            ebs: { volumeSize: "100Gi", volumeType: "gp3", encrypted: true },
        }],
    },
});
```

**`NodePolicyArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `name` | `string` | Unique policy name |
| `description` | `string` | Human-readable description |
| `weight` | `number` | Priority weight when multiple policies match (higher = preferred) |
| `instanceCategories` | `LabelSelectorArgs` | Filter by instance category letter: e.g. `m`, `c`, `r` (AWS) or `D`, `E` (Azure) |
| `instanceFamilies` | `LabelSelectorArgs` | Filter instance families (e.g. `c5`, `m5`) |
| `instanceCpus` | `LabelSelectorArgs` | Filter by vCPU count |
| `instanceSizes` | `LabelSelectorArgs` | Filter instance sizes (e.g. `large`, `xlarge`) |
| `instanceTypes` | `LabelSelectorArgs` | Explicit instance types (e.g. `m5.xlarge`) |
| `instanceGenerations` | `LabelSelectorArgs` | Filter by instance generation |
| `instanceHypervisors` | `LabelSelectorArgs` | Filter by hypervisor type |
| `zones` | `LabelSelectorArgs` | Availability zones to provision into |
| `architectures` | `LabelSelectorArgs` | CPU architectures (e.g. `amd64`, `arm64`) |
| `capacityTypes` | `LabelSelectorArgs` | Capacity types: `on-demand` \| `spot` \| `reserved` |
| `operatingSystems` | `LabelSelectorArgs` | OS filter (e.g. `linux`, `windows`) |
| `labels` | `Record<string, string>` | Labels applied to provisioned nodes |
| `taints` | `TaintArgs[]` | Taints applied to provisioned nodes |
| `disruption` | `DisruptionPolicyArgs` | Node disruption / consolidation settings |
| `limits` | `ResourceLimitsArgs` | Max total CPU/memory this policy may provision |
| `nodePoolName` | `string` | Override the Karpenter NodePool name |
| `nodeClassName` | `string` | Override the Karpenter NodeClass name |
| `aws` | `AWSNodeClassSpecArgs` | AWS-specific node class configuration |
| `azure` | `AzureNodeClassSpecArgs` | Azure-specific node class configuration |
| `raw` | `RawKarpenterSpecArgs[]` | Raw Karpenter YAML (escape hatch) |

**`RawKarpenterSpecArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `nodepoolYaml` | `string` | Raw YAML for a complete Karpenter NodePool resource |
| `nodeclassYaml` | `string` | Raw YAML for a complete Karpenter NodeClass resource |

**`DisruptionPolicyArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `consolidationPolicy` | `string` | `WhenEmpty` \| `WhenEmptyOrUnderutilized` |
| `consolidateAfter` | `string` | Wait time after node is empty before consolidating (e.g. `30s`) |
| `expireAfter` | `string` | Force-replace nodes after this duration (e.g. `720h`) |
| `ttlSecondsAfterEmpty` | `number` | Seconds before an empty node is terminated (deprecated; prefer `consolidateAfter`) |
| `terminationGracePeriodSeconds` | `number` | Grace period before forcefully terminating a draining node |
| `budgets` | `DisruptionBudgetArgs[]` | Limits on how many nodes may be disrupted at once |

**`AWSNodeClassSpecArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `amiFamily` | `string` | AMI family: `AL2`, `AL2023`, `Bottlerocket`, `Windows2019`, `Windows2022` |
| `role` | `string` | IAM role name for nodes (Karpenter creates the instance profile) |
| `instanceProfile` | `string` | IAM instance profile name (alternative to `role`) |
| `subnetSelectorTerms` | `SubnetSelectorTermArgs[]` | Subnet selectors (by tag or ID) |
| `securityGroupSelectorTerms` | `SecurityGroupSelectorTermArgs[]` | Security group selectors |
| `capacityReservationSelectorTerms` | `CapacityReservationSelectorTermArgs[]` | EC2 capacity reservation selectors |
| `amiSelectorTerms` | `AMISelectorTermArgs[]` | AMI selectors (by alias, tag, or ID) |
| `blockDeviceMappings` | `BlockDeviceMappingArgs[]` | EBS volume configuration |
| `instanceStorePolicy` | `string` | NVMe instance store policy. Value: `INSTANCE_STORE_POLICY_RAID0` |
| `tags` | `Record<string, string>` | AWS tags on all provisioned resources |
| `associatePublicIpAddress` | `boolean` | Assign a public IP to nodes |
| `detailedMonitoring` | `boolean` | Enable CloudWatch detailed monitoring |
| `metadataOptions` | `MetadataOptionsArgs` | EC2 IMDS options (IMDSv2, hop limit, etc.) |
| `kubelet` | `KubeletConfigurationArgs` | Kubelet overrides (maxPods, eviction thresholds, etc.) |
| `userData` | `string` | Custom launch template user data |
| `context` | `string` | Additional EC2 launch template context ARN for advanced customization |

**`AzureNodeClassSpecArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `vnetSubnetId` | `string` | Azure VNet subnet resource ID |
| `imageFamily` | `string` | Image family: `AzureLinux`, `Ubuntu2204`, etc. |
| `osDiskSizeGb` | `number` | OS disk size in GB |
| `fipsMode` | `string` | `Enabled` \| `Disabled` |
| `maxPods` | `number` | Max pods per node |
| `tags` | `Record<string, string>` | Azure tags on provisioned resources |
| `kubelet` | `AzureKubeletConfigurationArgs` | Kubelet overrides for Azure nodes |

---

### `WorkloadRule`

Pin explicit resource rules directly to a single workload. Unlike `WorkloadPolicy`, which applies a shared policy to many workloads, a `WorkloadRule` targets one specific `kind/namespace/name` on a cluster.

```typescript
const rule = new resources.WorkloadRule("my-app-rule", {
    clusterId: "cluster-abc123",
    namespace:  "production",
    kind:       "Deployment",
    name:       "my-api",

    cpuRule: {
        enabled:                 true,
        minRequest:              100,   // 100m CPU
        maxRequest:              4000,  // 4 cores
        targetPercentile:        0.95,
        limitsAdjustmentEnabled: true,
        limitMultiplier:         1.5,
    },
    memoryRule: {
        enabled:    true,
        minRequest: 134217728,   // 128 MiB
        maxRequest: 1073741824,  // 1 GiB
    },
    actionTriggers:    ["on_detection"],
    detectionTriggers: ["pod_creation", "pod_reschedule"],
    cooldownMinutes:   60,
});

export const ruleId = rule.id;
```

> **Auto-generate:** Set `autoGenerate: true` to let the engine fill in all fields from observed usage:
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

**Fields:**

| Field | Type | Description |
|---|---|---|
| `clusterId` | `string` | ID of the cluster the workload lives in |
| `namespace` | `string` | Kubernetes namespace of the workload |
| `kind` | `string` | `Deployment` \| `StatefulSet` \| `DaemonSet` \| `CronJob` \| `Job` |
| `name` | `string` | Name of the Kubernetes workload |
| `autoGenerate` | `boolean` | When `true`, the engine fills all rule fields automatically |
| `cpuRule` | `ResourceRuleConfigArgs` | CPU vertical scaling rule |
| `memoryRule` | `ResourceRuleConfigArgs` | Memory vertical scaling rule |
| `gpuRule` | `ResourceRuleConfigArgs` | GPU vertical scaling rule |
| `hpaRule` | `HPARuleConfigArgs` | Horizontal (replica) scaling rule |
| `emergencyResponse` | `EmergencyResponseConfigArgs` | OOM and CPU-throttle emergency reactions |
| `actionTriggers` | `string[]` | `on_detection` \| `on_schedule` |
| `cronSchedule` | `string` | Cron expression (5-field UTC). Required when `actionTriggers` includes `on_schedule` |
| `detectionTriggers` | `string[]` | `pod_creation` \| `pod_update` \| `pod_reschedule` |
| `startupPeriodSeconds` | `number` | Seconds after workload start to exclude from usage data |
| `cooldownMinutes` | `number` | Minimum minutes between recommendation applications |
| `schedulerPlugins` | `string[]` | Kubernetes scheduler plugins to activate |
| `defragmentationSchedule` | `string` | Cron expression for node defragmentation |
| `liveMigrationEnabled` | `boolean` | Allow live pod migration when applying recommendations |
| `useInPlaceVerticalScaling` | `boolean` | Use in-place pod vertical scaling instead of pod restarts |
| `containers` | `ContainerResourceRuleConfigArgs[]` | Per-container resource overrides |

**`ResourceRuleConfigArgs` fields** (used for `cpuRule`, `memoryRule`, `gpuRule`):

> **Note:** `maxScaleUpPercent` and `maxScaleDownPercent` are **not** supported on per-container rules.

| Field | Type | Description |
|---|---|---|
| `enabled` | `boolean` | Enable this resource axis rule |
| `minRequest` | `number` | Minimum resource request (millicores for CPU, bytes for memory/GPU) |
| `maxRequest` | `number` | Maximum resource request |
| `targetPercentile` | `number` | Percentile of observed usage to target (0–1) |
| `maxScaleUpPercent` | `number` | Max % to scale up in one step *(workload-level only)* |
| `maxScaleDownPercent` | `number` | Max % to scale down in one step *(workload-level only)* |
| `limitsAdjustmentEnabled` | `boolean` | Also adjust resource limits |
| `limitMultiplier` | `number` | Limits = request × limitMultiplier |
| `limitsRemovalEnabled` | `boolean` | Actively remove limits (CPU only) |

**`HPARuleConfigArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `enabled` | `boolean` | Enable horizontal scaling |
| `minReplicas` | `number` | Minimum replicas |
| `maxReplicas` | `number` | Maximum replicas |
| `targetUtilization` | `number` | Target utilization ratio (0–1) |
| `primaryMetric` | `string` | `cpu` \| `memory` \| `gpu` \| `network_ingress` \| `network_egress` |
| `maxReplicaChangePercent` | `number` | Max % change in replica count per cycle |

**`EmergencyResponseConfigArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `oomEnabled` | `boolean` | React to OOM kills by increasing memory |
| `oomMemoryMultiplier` | `number` | Multiplier applied to memory on OOM. Example: `2.0` |
| `oomMaxReactions` | `number` | Max OOM reactions before giving up |
| `oomCooldownSeconds` | `number` | Seconds between OOM reactions |
| `cpuThrottlingEnabled` | `boolean` | React to CPU throttling |
| `cpuThrottlingThreshold` | `number` | Throttle ratio (0–1) that triggers a reaction |
| `cpuThrottlingMultiplier` | `number` | Multiplier applied to CPU request on throttle reaction |

---

### `NodePolicyTarget`

Apply a node policy to one or more clusters.

```typescript
const nodePolicyTarget = new resources.NodePolicyTarget("my-node-policy-target", {
    name: "my-node-policy-target",
    policyId: nodePolicy.id,
    clusterIds: [cluster.id],
    enabled: true,
});
```

**Fields:**

| Field | Type | Description |
|---|---|---|
| `name` | `string` | Unique target name |
| `policyId` | `string` | ID of the `NodePolicy` to apply |
| `clusterIds` | `string[]` | Cluster IDs to target. **At most 1 entry** — the backend rejects more than one. |
| `description` | `string` | Human-readable description (optional) |
| `enabled` | `boolean` | Activate the target |

> **Note:** `pulumi destroy` removes this resource from Pulumi state but does **not** delete it on the DevZero backend — no delete RPC exists for NodePolicyTarget.

---

## Data Sources

### `getClusterIdByName`

Look up an existing cluster by name and return its ID. Use this when a cluster was registered manually (not created by Pulumi) and you need its ID to attach policies or inject into `values.yaml` / a Kubernetes secret.

```typescript
import { resources } from "@devzero/pulumi-devzero";

const existing = await resources.getClusterIdByName({
    name: "my-existing-cluster",
    // teamId is optional — defaults to devzero:teamId from provider config
    // region: "us-east-1",        // optional: filter by region
    // cloudProvider: "AWS",       // optional: AWS | GCP | AKS | OCI
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

**Inputs:**

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` | yes | Cluster name to look up |
| `teamId` | `string` | no | Defaults to `devzero:teamId` from provider config |
| `region` | `string` | no | Filter by region name (e.g. `us-east-1`) |
| `cloudProvider` | `string` | no | Filter by cloud provider: `AWS` \| `GCP` \| `AKS` \| `OCI` |
| `liveness` | `string` | no | `IGNORE` (default) \| `PREFER_LIVE` \| `REQUIRE_LIVE` |

**Outputs:**

| Field | Type | Description |
|---|---|---|
| `clusterId` | `string` | UUID of the matching cluster |

> **Note:** If multiple clusters share the same name, the newest one is returned by default. Use `liveness: "PREFER_LIVE"` or `"REQUIRE_LIVE"` to filter by heartbeat freshness.

---

## Destroying Resources

```bash
# Tear down all resources in the stack
pulumi destroy

# Remove the stack itself
pulumi stack rm <stack-name>
```

---

## Links

- [DevZero](https://devzero.io/)
- [Pulumi Registry](https://www.pulumi.com/registry/)
- [GitHub — devzero-inc/pulumi-provider-devzero](https://github.com/devzero-inc/pulumi-provider-devzero)
- [Report an issue](https://github.com/devzero-inc/pulumi-provider-devzero/issues)

---

## License

[MIT](https://github.com/devzero-inc/pulumi-provider-devzero/blob/main/LICENSE) — Copyright (c) 2026 DevZero Inc.