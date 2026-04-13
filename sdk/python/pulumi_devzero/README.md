# pulumi-devzero

> The official **Python** Pulumi provider for [DevZero](https://devzero.io/) — manage clusters, workload policies, and node policies as code.

[![PyPI version](https://img.shields.io/pypi/v/pulumi-devzero.svg?style=flat-square)](https://pypi.org/project/pulumi-devzero/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](https://github.com/devzero-inc/pulumi-provider-devzero/blob/main/LICENSE)
[![Pulumi Registry](https://img.shields.io/badge/Pulumi-Registry-blueviolet?style=flat-square&logo=pulumi)](https://www.pulumi.com/registry/)

---

## Installation

```bash
pip install pulumi-devzero
```

**Requires:** `pulumi>=3.0.0,<4.0.0`, Python 3.9+

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

```python
import pulumi
import pulumi_devzero as devzero

# 1. Create a cluster
cluster = devzero.resources.Cluster("prod-cluster",
    name="prod-cluster",
)

# 2. Create a workload policy with CPU vertical scaling
policy = devzero.resources.WorkloadPolicy("cpu-scaling-policy",
    name="cpu-scaling-policy",
    description="Policy with CPU vertical scaling enabled",
    cpu_vertical_scaling=devzero.resources.VerticalScalingArgsArgs(
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
)

# 3. Apply the policy to the cluster for all Deployments
target = devzero.resources.WorkloadPolicyTarget("prod-cluster-target",
    name="prod-cluster-deployments-target",
    description="Apply cpu-scaling-policy to all Deployments in prod-cluster",
    policy_id=policy.id,
    cluster_ids=[cluster.id],
    kind_filter=["Deployment"],
    enabled=True,
)

pulumi.export("cluster_id", cluster.id)
pulumi.export("cluster_token", pulumi.Output.secret(cluster.token))
pulumi.export("policy_id", policy.id)
pulumi.export("target_id", target.id)
```

```bash
pulumi up
```

---

## Resources

### `Cluster`

Provision and manage a DevZero cluster.

```python
import pulumi
import pulumi_devzero as devzero

cluster = devzero.resources.Cluster("my-cluster",
    name="my-cluster",
)

pulumi.export("id", cluster.id)
pulumi.export("token", pulumi.Output.secret(cluster.token))
```

---

### `WorkloadPolicy`

Configure vertical and horizontal scaling policies for workloads.

```python
import pulumi_devzero as devzero

policy = devzero.resources.WorkloadPolicy("my-policy",
    name="my-policy",
    description="Vertical scaling for CPU and memory",
    cpu_vertical_scaling=devzero.resources.VerticalScalingArgsArgs(
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
    memory_vertical_scaling=devzero.resources.VerticalScalingArgsArgs(
        enabled=True,
        target_percentile=0.9,
        min_request=128,
        max_request=8192,
        max_scale_up_percent=50,
        max_scale_down_percent=20,
        overhead_multiplier=1.2,
        limits_adjustment_enabled=True,
        limit_multiplier=1.3,
    ),
)
```

**`VerticalScalingArgsArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `enabled` | `bool` | Enable this scaling axis |
| `target_percentile` | `float` | Percentile of observed usage to target (e.g. `0.95`) |
| `min_request` | `int` | Minimum resource request (millicores / MiB) |
| `max_request` | `int` | Maximum resource request (millicores / MiB) |
| `max_scale_up_percent` | `int` | Max % to scale up in one step |
| `max_scale_down_percent` | `int` | Max % to scale down in one step |
| `overhead_multiplier` | `float` | Multiplier added on top of the recommendation |
| `limits_adjustment_enabled` | `bool` | Whether to also adjust resource limits |
| `limit_multiplier` | `float` | Limits = request × limit_multiplier |

---

### `WorkloadPolicyTarget`

Apply a workload policy to one or more clusters with optional filters.

```python
import pulumi_devzero as devzero

target = devzero.resources.WorkloadPolicyTarget("my-target",
    name="my-target",
    policy_id=policy.id,
    cluster_ids=[cluster.id],
    kind_filter=["Deployment", "StatefulSet"],
    namespace_filter=["production"],
    enabled=True,
)
```

**Fields:**

| Field | Type | Description |
|---|---|---|
| `name` | `str` | Unique target name |
| `policy_id` | `str` | ID of the `WorkloadPolicy` to apply |
| `cluster_ids` | `list[str]` | Cluster IDs to target |
| `kind_filter` | `list[str]` | `Pod` \| `Deployment` \| `StatefulSet` \| `DaemonSet` \| `Job` \| `CronJob` \| `ReplicaSet` \| `ReplicationController` \| `Rollout` |
| `namespace_filter` | `list[str]` | Restrict to specific namespaces (optional) |
| `enabled` | `bool` | Activate the target |

---

### `NodePolicy`

Configure node provisioning and pooling (AWS / Azure).

```python
import pulumi_devzero as devzero

node_policy = devzero.resources.NodePolicy("my-node-policy",
    name="my-node-policy",
)
```

---

### `NodePolicyTarget`

Apply a node policy to one or more clusters.

```python
import pulumi_devzero as devzero

node_policy_target = devzero.resources.NodePolicyTarget("my-node-policy-target",
    name="my-node-policy-target",
    policy_id=node_policy.id,
    cluster_ids=[cluster.id],
    enabled=True,
)
```

---

## Data Sources

### `get_cluster_id_by_name`

Look up an existing cluster by name and return its ID. Use this when a cluster was registered manually (not created by Pulumi) and you need its ID to attach policies or inject into `values.yaml` / a Kubernetes secret.

```python
import pulumi
import pulumi_devzero as devzero

existing = devzero.resources.get_cluster_id_by_name(
    name="my-existing-cluster",
    # team_id is optional — defaults to devzero:teamId from provider config
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

**Inputs:**

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `str` | yes | Cluster name to look up |
| `team_id` | `str` | no | Defaults to `devzero:teamId` from provider config |

**Outputs:**

| Field | Type | Description |
|---|---|---|
| `cluster_id` | `str` | UUID of the matching cluster |

> **Warning:** If multiple clusters share the same name, only the first active (non-deleted) one is returned. Ensure cluster names are unique within your team to avoid unexpected results.

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