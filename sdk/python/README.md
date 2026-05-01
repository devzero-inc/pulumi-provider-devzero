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
    enable_pmax_protection=True,
    pmax_ratio_threshold=3.0,
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
        adjust_req_even_if_not_set=True,
        limits_removal_enabled=False,
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
        adjust_req_even_if_not_set=True,
        limits_removal_enabled=False,
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
| `max_scale_up_percent` | `float` | Max % to scale up in one step. Default: `1000` |
| `max_scale_down_percent` | `float` | Max % to scale down in one step. Default: `1.0` |
| `overhead_multiplier` | `float` | Multiplier added on top of the recommendation |
| `limits_adjustment_enabled` | `bool` | Whether to also adjust resource limits |
| `limit_multiplier` | `float` | Limits = request × limit_multiplier |
| `min_data_points` | `int` | Minimum data points before a recommendation is emitted. Default: `20` |
| `adjust_req_even_if_not_set` | `bool` | Recommend requests even when the workload has no existing requests set. Default: `true` |
| `limits_removal_enabled` | `bool` | Actively remove limits from workloads (CPU only). Takes precedence over `limits_adjustment_enabled`. Default: `true` for CPU, `false` for memory |

**`WorkloadPolicy` pmax & VPA knob fields:**

| Field | Type | Description |
|---|---|---|
| `enable_pmax_protection` | `bool` | Raise requests to cover peak usage when max/recommendation ratio exceeds `pmax_ratio_threshold`. Default: `true` |
| `pmax_ratio_threshold` | `float` | Max-to-recommendation ratio that triggers pmax protection. Default: `3.0` |
| `loopback_period_seconds` | `int` | Look-back period in seconds for usage data. Default: `86400` (24 h) |
| `min_data_points` | `int` | Global minimum data points for recommendations. Default: `15` |
| `min_change_percent` | `float` | Global minimum change threshold. Default: `0.2` (20%) |
| `min_vpa_window_data_points` | `int` | Minimum data points in VPA analysis window. Default: `30` |
| `cooldown_minutes` | `int` | Minutes between applying recommendations. Default: `300` (5 h) |

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

Configure node provisioning and pooling (AWS / Azure) using Karpenter under the hood.

```python
import pulumi_devzero as devzero

node_policy = devzero.resources.NodePolicy("my-node-policy",
    name="my-node-policy",
    description="AWS node policy with on-demand and spot capacity",
    weight=10,
    capacity_types=devzero.resources.LabelSelectorArgsArgs(match_expressions=[devzero.resources.LabelSelectorRequirementArgsArgs(key="<label-key>", operator="In", values=["<value>"])]),
    instance_categories=devzero.resources.LabelSelectorArgsArgs(match_labels={"<label-key>": "<label-value>"}),
    labels={"environment": "production"},
    taints=[devzero.resources.TaintArgsArgs(key="dedicated", value="gpu", effect="NoSchedule")],
    disruption=devzero.resources.DisruptionPolicyArgsArgs(
        consolidation_policy="WhenEmptyOrUnderutilized",
        consolidate_after="30s",
        expire_after="720h",
    ),
    limits=devzero.resources.ResourceLimitsArgsArgs(cpu="1000", memory="1000Gi"),
    aws=devzero.resources.AWSNodeClassSpecArgsArgs(
        ami_family="AL2",
        role="KarpenterNodeRole",
        subnet_selector_terms=[devzero.resources.SubnetSelectorTermArgsArgs(tags={"karpenter.sh/discovery": "my-cluster"})],
        security_group_selector_terms=[devzero.resources.SecurityGroupSelectorTermArgsArgs(tags={"karpenter.sh/discovery": "my-cluster"})],
        block_device_mappings=[devzero.resources.BlockDeviceMappingArgsArgs(
            device_name="/dev/xvda",
            root_volume=True,
            ebs=devzero.resources.BlockDeviceArgsArgs(volume_size="100Gi", volume_type="gp3", encrypted=True),
        )],
    ),
)
```

**`NodePolicyArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `name` | `str` | Unique policy name |
| `description` | `str` | Human-readable description |
| `weight` | `int` | Priority weight when multiple policies match (higher = preferred) |
| `instance_categories` | `LabelSelectorArgsArgs` | Filter by instance category letter: e.g. `m`, `c`, `r` (AWS) or `D`, `E` (Azure) |
| `instance_families` | `LabelSelectorArgsArgs` | Filter instance families (e.g. `c5`, `m5`) |
| `instance_cpus` | `LabelSelectorArgsArgs` | Filter by vCPU count |
| `instance_sizes` | `LabelSelectorArgsArgs` | Filter instance sizes (e.g. `large`, `xlarge`) |
| `instance_types` | `LabelSelectorArgsArgs` | Explicit instance types (e.g. `m5.xlarge`) |
| `instance_generations` | `LabelSelectorArgsArgs` | Filter by instance generation |
| `instance_hypervisors` | `LabelSelectorArgsArgs` | Filter by hypervisor type |
| `zones` | `LabelSelectorArgsArgs` | Availability zones to provision into |
| `architectures` | `LabelSelectorArgsArgs` | CPU architectures (e.g. `amd64`, `arm64`) |
| `capacity_types` | `LabelSelectorArgsArgs` | Capacity types: `on-demand` \| `spot` \| `reserved` |
| `operating_systems` | `LabelSelectorArgsArgs` | OS filter (e.g. `linux`, `windows`) |
| `labels` | `dict[str, str]` | Labels applied to provisioned nodes |
| `taints` | `list[TaintArgsArgs]` | Taints applied to provisioned nodes |
| `disruption` | `DisruptionPolicyArgsArgs` | Node disruption / consolidation settings |
| `limits` | `ResourceLimitsArgsArgs` | Max total CPU/memory this policy may provision |
| `node_pool_name` | `str` | Override the Karpenter NodePool name |
| `node_class_name` | `str` | Override the Karpenter NodeClass name |
| `aws` | `AWSNodeClassSpecArgsArgs` | AWS-specific node class configuration |
| `azure` | `AzureNodeClassSpecArgsArgs` | Azure-specific node class configuration |
| `raw` | `list[RawKarpenterSpecArgsArgs]` | Raw Karpenter YAML (escape hatch) |

**`DisruptionPolicyArgsArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `consolidation_policy` | `str` | `WhenEmpty` \| `WhenEmptyOrUnderutilized` |
| `consolidate_after` | `str` | Wait time after node is empty before consolidating (e.g. `30s`) |
| `expire_after` | `str` | Force-replace nodes after this duration (e.g. `720h`) |
| `termination_grace_period_seconds` | `int` | Grace period before forcefully terminating a draining node |
| `budgets` | `list[DisruptionBudgetArgsArgs]` | Limits on how many nodes may be disrupted at once |

**`AWSNodeClassSpecArgsArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `ami_family` | `str` | AMI family: `AL2`, `Bottlerocket`, `Windows2022`, etc. |
| `role` | `str` | IAM role name for nodes |
| `subnet_selector_terms` | `list[SubnetSelectorTermArgsArgs]` | Subnet selectors (by tag or ID) |
| `security_group_selector_terms` | `list[SecurityGroupSelectorTermArgsArgs]` | Security group selectors |
| `ami_selector_terms` | `list[AMISelectorTermArgsArgs]` | AMI selectors (by alias, tag, or ID) |
| `block_device_mappings` | `list[BlockDeviceMappingArgsArgs]` | EBS volume configuration |
| `tags` | `dict[str, str]` | AWS tags on all provisioned resources |
| `associate_public_ip_address` | `bool` | Assign a public IP to nodes |
| `detailed_monitoring` | `bool` | Enable CloudWatch detailed monitoring |
| `metadata_options` | `MetadataOptionsArgsArgs` | EC2 IMDS options (IMDSv2, hop limit, etc.) |
| `kubelet` | `KubeletConfigurationArgsArgs` | Kubelet overrides (max_pods, eviction thresholds, etc.) |
| `user_data` | `str` | Custom launch template user data |

**`AzureNodeClassSpecArgsArgs` fields:**

| Field | Type | Description |
|---|---|---|
| `vnet_subnet_id` | `str` | Azure VNet subnet resource ID |
| `image_family` | `str` | Image family: `AzureLinux`, `Ubuntu2204`, etc. |
| `os_disk_size_gb` | `int` | OS disk size in GB |
| `fips_mode` | `str` | `Enabled` \| `Disabled` |
| `max_pods` | `int` | Max pods per node |
| `tags` | `dict[str, str]` | Azure tags on provisioned resources |
| `kubelet` | `AzureKubeletConfigurationArgsArgs` | Kubelet overrides for Azure nodes |

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
    # region="us-east-1",        # optional: filter by region
    # cloud_provider="AWS",      # optional: AWS | GCP | AKS | OCI
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

**Inputs:**

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `str` | yes | Cluster name to look up |
| `team_id` | `str` | no | Defaults to `devzero:teamId` from provider config |
| `region` | `str` | no | Filter by region name (e.g. `us-east-1`) |
| `cloud_provider` | `str` | no | Filter by cloud provider: `AWS` \| `GCP` \| `AKS` \| `OCI` |
| `liveness` | `str` | no | `IGNORE` (default) \| `PREFER_LIVE` \| `REQUIRE_LIVE` |

**Outputs:**

| Field | Type | Description |
|---|---|---|
| `cluster_id` | `str` | UUID of the matching cluster |

> **Note:** If multiple clusters share the same name, the newest one is returned by default. Use `liveness="PREFER_LIVE"` or `"REQUIRE_LIVE"` to filter by heartbeat freshness.

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