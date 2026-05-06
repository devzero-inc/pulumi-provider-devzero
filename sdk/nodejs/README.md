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
    actionTriggers: ["on_detection", "on_schedule"],  // apply on pod events AND on schedule
    cronSchedule: "0 2 * * *",                        // daily at 2 am UTC (required for on_schedule)
    detectionTriggers: ["pod_creation", "pod_reschedule"],
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