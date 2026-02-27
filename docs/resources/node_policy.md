# devzero:index/nodePolicy:NodePolicy

Manages a DevZero node policy. A node policy configures Karpenter-based node provisioning for a cluster — defining which instance types, availability zones, capacity types, and architectures are permitted, along with disruption behaviour, resource limits, and cloud-specific settings.

> **Note:** There is no delete API for node policies. Destroying this resource removes it from Pulumi state only; the policy continues to exist on the DevZero platform.

## Example Usage

### Minimal example

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const policy = new devzero.NodePolicy("general-purpose", {
    name: "general-purpose",
});
```

### Structured configuration — spot instances with disruption

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const policy = new devzero.NodePolicy("spot-policy", {
    name: "cost-optimized",
    description: "Spot instances for non-critical workloads",
    weight: 10,

    capacityTypes: {
        matchLabels: { "karpenter.sh/capacity-type": "spot" },
    },
    architectures: {
        matchLabels: { "kubernetes.io/arch": "amd64" },
    },
    instanceCategories: {
        matchLabels: { "karpenter.k8s.aws/instance-category": "c" },
    },

    disruption: {
        consolidationPolicy: "WhenEmptyOrUnderutilized",
        consolidateAfter: "30m",
        expireAfter: "720h",
        budgets: [
            { reasons: ["Underutilized", "Empty"], nodes: "10%", duration: "1h" },
        ],
    },

    limits: {
        cpu: "1000",
        memory: "4000Gi",
    },

    labels: { team: "platform", env: "production" },
    taints: [
        { key: "dedicated", value: "gpu", effect: "NoSchedule" },
    ],
});
```

### AWS-specific node class

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const policy = new devzero.NodePolicy("aws-policy", {
    name: "aws-optimized",
    nodePoolName:  "general-pool",
    nodeClassName: "general-class",

    aws: {
        amiFamily: "AL2",
        amiSelectorTerms: [{ alias: "al2@latest" }],
        subnetSelectorTerms: [
            { tags: { Name: "private-subnet" } },
        ],
        securityGroupSelectorTerms: [
            { tags: { "kubernetes.io/cluster/my-cluster": "owned" } },
        ],
        tags: { Environment: "production" },
        blockDeviceMappings: [
            {
                deviceName: "/dev/xvda",
                ebs: {
                    volumeSize: "100Gi",
                    volumeType: "gp3",
                    encrypted: true,
                    deleteOnTermination: true,
                },
            },
        ],
        metadataOptions: {
            httpTokens: "required",
            httpEndpoint: "enabled",
            httpPutResponseHopLimit: 2,
        },
        kubelet: {
            maxPods: 110,
            evictionHard: { "memory.available": "5%" },
        },
    },
});
```

### Raw Karpenter YAML (escape hatch)

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const policy = new devzero.NodePolicy("raw-policy", {
    name: "custom-karpenter",
    raw: [
        {
            nodepoolYaml: `
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: custom-pool
spec:
  template:
    spec:
      nodeClassRef:
        name: custom-class
      requirements:
        - key: karpenter.sh/capacity-type
          operator: In
          values: ["spot"]
`,
            nodeclassYaml: `
apiVersion: karpenter.k8s.aws/v1
kind: EC2NodeClass
metadata:
  name: custom-class
spec:
  amiFamily: AL2
  role: "KarpenterNodeRole"
`,
        },
    ],
});
```

```python
import pulumi_devzero as devzero

policy = devzero.NodePolicy("cost-optimized",
    name="cost-optimized",
    description="Spot instances for batch workloads",
    weight=5,
    capacity_types=devzero.LabelSelectorArgs(
        match_labels={"karpenter.sh/capacity-type": "spot"},
    ),
    disruption=devzero.DisruptionPolicyArgs(
        consolidation_policy="WhenEmptyOrUnderutilized",
        consolidate_after="30m",
    ),
    limits=devzero.ResourceLimitsArgs(cpu="500", memory="2000Gi"),
)
```

## Schema

### Required

| Name   | Type   | Description                        |
|--------|--------|------------------------------------|
| `name` | string | Human-friendly name for the policy.|

### Optional — General

| Name           | Type     | Description                                                                     |
|----------------|----------|---------------------------------------------------------------------------------|
| `description`  | `string` | Free-form description of the policy.                                            |
| `weight`       | `number` | Priority weight. Higher values take precedence when multiple policies match. Default `0`. |
| `nodePoolName` | `string` | Override name for the generated Karpenter NodePool.                             |
| `nodeClassName`| `string` | Override name for the generated Karpenter NodeClass.                            |
| `labels`       | `map`    | Labels applied to provisioned nodes.                                            |
| `taints`       | `Taint[]`| Taints applied to provisioned nodes. See **Taint** below.                      |

### Optional — Instance Selectors (LabelSelector)

All fields accept a **LabelSelector** object (see structure below).

| Name                  | Description                                                   |
|-----------------------|---------------------------------------------------------------|
| `instanceCategories`  | Filter by instance category (e.g. `general-purpose`, `compute-optimized`). |
| `instanceFamilies`    | Filter by instance family (e.g. `m5`, `c6i`).                |
| `instanceCpus`        | Filter by CPU count.                                          |
| `instanceHypervisors` | Filter by hypervisor type.                                    |
| `instanceGenerations` | Filter by generation.                                         |
| `instanceSizes`       | Filter by size (e.g. `large`, `xlarge`).                     |
| `instanceTypes`       | Explicitly select specific instance types.                   |
| `zones`               | Availability zones where nodes may be provisioned.            |
| `architectures`       | CPU architectures (e.g. `amd64`, `arm64`).                   |
| `capacityTypes`       | Capacity types (e.g. `on-demand`, `spot`).                   |
| `operatingSystems`    | Operating systems (e.g. `linux`, `windows`).                 |

### Optional — Resource Limits

`limits` accepts:

| Field    | Type     | Description                    |
|----------|----------|--------------------------------|
| `cpu`    | `string` | Max total CPU across all nodes (e.g. `"1000"`). |
| `memory` | `string` | Max total memory (e.g. `"4000Gi"`).             |

### Optional — Disruption

`disruption` accepts:

| Field                           | Type             | Description                                                              |
|---------------------------------|------------------|--------------------------------------------------------------------------|
| `consolidationPolicy`           | `string`         | `"WhenEmpty"` or `"WhenEmptyOrUnderutilized"`.                           |
| `consolidateAfter`              | `string`         | Duration to wait before consolidating (e.g. `"30m"`).                   |
| `expireAfter`                   | `string`         | Node lifetime before forced replacement (e.g. `"720h"`).                |
| `ttlSecondsAfterEmpty`          | `number`         | Seconds to wait after node is empty before terminating.                  |
| `terminationGracePeriodSeconds` | `number`         | Grace period when terminating nodes.                                     |
| `budgets`                       | `Budget[]`       | Disruption budgets. See **DisruptionBudget** below.                      |

**DisruptionBudget**

| Field      | Type       | Description                                                          |
|------------|------------|----------------------------------------------------------------------|
| `reasons`  | `string[]` | Disruption reasons this budget applies to (e.g. `"Underutilized"`, `"Empty"`). |
| `nodes`    | `string`   | Max nodes to disrupt concurrently — a count (`"2"`) or percentage (`"10%"`). |
| `schedule` | `string`   | Optional cron schedule for when this budget is active.               |
| `duration` | `string`   | Duration the budget window is active (e.g. `"1h30m"`).              |

### Optional — Taint

| Field    | Type     | Description                                                         |
|----------|----------|---------------------------------------------------------------------|
| `key`    | `string` | Taint key.                                                          |
| `value`  | `string` | Taint value.                                                        |
| `effect` | `string` | Taint effect: `"NoSchedule"`, `"PreferNoSchedule"`, `"NoExecute"`. |

### Optional — AWS Node Class (`aws`)

| Field                              | Type       | Description                                                       |
|------------------------------------|------------|-------------------------------------------------------------------|
| `amiFamily`                        | `string`   | AMI family (e.g. `"AL2"`, `"Bottlerocket"`, `"Custom"`).         |
| `amiSelectorTerms`                 | `AMISelectorTerm[]` | AMI selection rules. See **AMISelectorTerm** below.    |
| `subnetSelectorTerms`              | `SubnetSelectorTerm[]` | Subnet selection rules. See below.                   |
| `securityGroupSelectorTerms`       | `SecurityGroupSelectorTerm[]` | Security group selection rules. See below.    |
| `capacityReservationSelectorTerms` | `CapacityReservationSelectorTerm[]` | Capacity reservation selection rules.     |
| `associatePublicIpAddress`         | `boolean`  | Whether to associate a public IP to launched instances.           |
| `role`                             | `string`   | IAM role for nodes.                                               |
| `instanceProfile`                  | `string`   | IAM instance profile ARN.                                         |
| `tags`                             | `map`      | AWS tags applied to all resources created by Karpenter.           |
| `userData`                         | `string`   | Custom user data script injected into the launch template.        |
| `context`                          | `string`   | Arbitrary context string passed through to CloudFormation.        |
| `detailedMonitoring`               | `boolean`  | Enable detailed CloudWatch monitoring.                            |
| `instanceStorePolicy`              | `string`   | Instance store policy. Value: `"INSTANCE_STORE_POLICY_RAID0"`.   |
| `blockDeviceMappings`              | `BlockDeviceMapping[]` | EBS block device mappings. See **BlockDeviceMapping** below. |
| `metadataOptions`                  | object     | EC2 instance metadata options. See **MetadataOptions** below.    |
| `kubelet`                          | object     | Kubelet configuration overrides. See **KubeletConfiguration** below. |

**AMISelectorTerm** — `alias`, `id`, `name`, `owner`, `ssmParameter`, `tags`

**SubnetSelectorTerm** — `id`, `tags`

**SecurityGroupSelectorTerm** — `id`, `name`, `tags`

**CapacityReservationSelectorTerm** — `id`, `ownerId`, `tags`

**BlockDeviceMapping**

| Field        | Type      | Description                       |
|--------------|-----------|-----------------------------------|
| `deviceName` | `string`  | Device name (e.g. `"/dev/xvda"`). |
| `rootVolume` | `boolean` | Mark as root volume.              |
| `ebs`        | object    | EBS block device config. Fields: `deleteOnTermination`, `encrypted`, `iops`, `kmsKeyId`, `snapshotId`, `throughput`, `volumeInitializationRate`, `volumeSize`, `volumeType`. |

**MetadataOptions** — `httpEndpoint`, `httpProtocolIpv6`, `httpPutResponseHopLimit`, `httpTokens`

**KubeletConfiguration** — `clusterDns`, `maxPods`, `podsPerCore`, `systemReserved`, `kubeReserved`, `evictionHard`, `evictionSoft`, `evictionSoftGracePeriod`, `evictionMaxPodGracePeriod`, `imageGcHighThresholdPercent`, `imageGcLowThresholdPercent`, `cpuCfsQuota`

### Optional — Azure Node Class (`azure`)

| Field          | Type      | Description                                                   |
|----------------|-----------|---------------------------------------------------------------|
| `vnetSubnetId` | `string`  | Azure Virtual Network subnet ID.                             |
| `osDiskSizeGb` | `number`  | OS disk size in GB.                                          |
| `imageFamily`  | `string`  | Image family (e.g. `"Ubuntu2204"`, `"Ubuntu2404"`, `"AzureLinux"`). |
| `fipsMode`     | `string`  | FIPS mode: `"FIPS"` or `"Disabled"`.                        |
| `tags`         | `map`     | Azure tags applied to all resources.                         |
| `maxPods`      | `number`  | Maximum pods per node.                                       |
| `kubelet`      | object    | Azure kubelet configuration. Fields: `cpuManagerPolicy`, `cpuCfsQuota`, `cpuCfsQuotaPeriod`, `imageGcHighThresholdPercent`, `imageGcLowThresholdPercent`, `topologyManagerPolicy`, `allowedUnsafeSysctls`, `containerLogMaxSize`, `containerLogMaxFiles`, `podPidsLimit`. |

### Optional — Raw Karpenter YAML (`raw`)

| Field           | Type     | Description                                        |
|-----------------|----------|----------------------------------------------------|
| `nodepoolYaml`  | `string` | Full YAML for a custom Karpenter `NodePool` object. |
| `nodeclassYaml` | `string` | Full YAML for a custom Karpenter `NodeClass` object.|

### LabelSelector

Used by all instance/zone/architecture selectors:

| Field              | Type                          | Description                                  |
|--------------------|-------------------------------|----------------------------------------------|
| `matchLabels`      | `map[string]string`           | Key-value pairs that must all match.         |
| `matchExpressions` | `LabelSelectorRequirement[]`  | Set-based requirements.                      |

**LabelSelectorRequirement** — `key`, `operator` (`"In"`, `"NotIn"`, `"Exists"`, `"DoesNotExist"`), `values`

### Read-Only

| Name | Type   | Description                                         |
|------|--------|-----------------------------------------------------|
| `id` | string | Unique identifier of the policy. Managed by the provider. |

## Import

```shell
pulumi import devzero:index/nodePolicy:NodePolicy my-policy <policy-id>
```

> **Note:** Because there is no delete API, `pulumi destroy` only removes the resource from Pulumi state. The policy continues to exist on the DevZero platform.
