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
| `maxScaleUpPercent` | `number` | Max % to scale up in one step |
| `maxScaleDownPercent` | `number` | Max % to scale down in one step |
| `overheadMultiplier` | `number` | Multiplier added on top of the recommendation |
| `limitsAdjustmentEnabled` | `boolean` | Whether to also adjust resource limits |
| `limitMultiplier` | `number` | Limits = request × limitMultiplier |

---

### `WorkloadPolicyTarget`

Apply a workload policy to one or more clusters with optional filters.

```typescript
const target = new resources.WorkloadPolicyTarget("my-target", {
    name: "my-target",
    policyId: policy.id,
    clusterIds: [cluster.id],
    kindFilter: ["Deployment", "StatefulSet"],
    namespaceFilter: ["production"],
    enabled: true,
});
```

**Fields:**

| Field | Type | Description |
|---|---|---|
| `name` | `string` | Unique target name |
| `policyId` | `string` | ID of the `WorkloadPolicy` to apply |
| `clusterIds` | `string[]` | Cluster IDs to target |
| `kindFilter` | `string[]` | `Pod` \| `Deployment` \| `StatefulSet` \| `DaemonSet` \| `Job` \| `CronJob` \| `ReplicaSet` \| `ReplicationController` \| `Rollout` |
| `namespaceFilter` | `string[]` | Restrict to specific namespaces (optional) |
| `enabled` | `boolean` | Activate the target |

---

### `NodePolicy`

Configure node provisioning and pooling (AWS / Azure).

```typescript
const nodePolicy = new resources.NodePolicy("my-node-policy", {
    name: "my-node-policy",
    // ...node policy configuration
});
```

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