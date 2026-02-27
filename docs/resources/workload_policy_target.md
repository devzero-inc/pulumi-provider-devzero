# devzero:index/workloadPolicyTarget:WorkloadPolicyTarget

Attaches a workload policy to specific workloads within clusters. A workload policy target scopes a policy to matching namespaces, workloads, or Kubernetes object kinds, and determines which clusters are affected.

## Example Usage

### Minimal example

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const target = new devzero.WorkloadPolicyTarget("basic-target", {
    name: "all-deployments",
    policyId: policy.id,
    clusterIds: [cluster.id],
});
```

### Full example with selectors

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const target = new devzero.WorkloadPolicyTarget("production-target", {
    name: "production-deployments",
    description: "Apply optimization policy to all production deployments",
    policyId: policy.id,
    clusterIds: [usEast.id, usWest.id, euWest.id],
    priority: 10,
    enabled: true,
    kindFilter: ["Deployment", "StatefulSet"],
    namespaceSelector: {
        matchLabels: { environment: "production" },
    },
    workloadSelector: {
        matchExpressions: [
            { key: "app.kubernetes.io/part-of", operator: "In", values: ["backend"] },
        ],
    },
    namePattern: {
        pattern: "^api-.*",
    },
});
```

```python
import pulumi_devzero as devzero

target = devzero.WorkloadPolicyTarget("production-target",
    name="production-deployments",
    description="Apply optimization policy to all production deployments",
    policy_id=policy.id,
    cluster_ids=[us_east.id, us_west.id],
    priority=10,
    enabled=True,
    kind_filter=["Deployment", "StatefulSet"],
    namespace_selector=devzero.LabelSelectorArgs(
        match_labels={"environment": "production"},
    ),
    workload_selector=devzero.LabelSelectorArgs(
        match_expressions=[
            devzero.LabelSelectorRequirementArgs(
                key="app.kubernetes.io/part-of",
                operator="In",
                values=["backend"],
            ),
        ],
    ),
)
```

## Schema

### Required

| Name         | Type       | Description                                                      |
|--------------|------------|------------------------------------------------------------------|
| `name`       | `string`   | Human-friendly name for this target.                             |
| `policyId`   | `string`   | Workload policy ID this target is attached to.                   |
| `clusterIds` | `string[]` | Cluster IDs where this target applies.                           |

### Optional — Filtering

| Name                | Type       | Description                                                                           |
|---------------------|------------|---------------------------------------------------------------------------------------|
| `description`       | `string`   | Free-form description of the target.                                                  |
| `priority`          | `number`   | Evaluation priority. Higher values take precedence when targets overlap. Default `0`. |
| `enabled`           | `boolean`  | Enable or disable this target. Default `true`.                                        |
| `kindFilter`        | `string[]` | Restrict matching to specific Kubernetes kinds. Supported values: `"Pod"`, `"Deployment"`, `"StatefulSet"`, `"DaemonSet"`, `"ReplicaSet"`, `"Job"`, `"CronJob"`, `"ReplicationController"`, `"Rollout"`. |
| `workloadNames`     | `string[]` | Explicit list of workload names to include.                                           |
| `nodeGroupNames`    | `string[]` | Restrict matching to specific node groups by name.                                    |
| `namePattern`       | object     | Regex to match workload names. See **NamePattern** below.                             |
| `namespaceSelector` | object     | Select namespaces by labels. See **LabelSelector** below.                             |
| `workloadSelector`  | object     | Select workloads by labels. See **LabelSelector** below.                              |

### NamePattern

| Field     | Type     | Description                              |
|-----------|----------|------------------------------------------|
| `pattern` | `string` | Regular expression pattern.              |
| `flags`   | `string` | Optional regex flags (e.g. `"i"` for case-insensitive). |

### LabelSelector

| Field              | Type                          | Description                                      |
|--------------------|-------------------------------|--------------------------------------------------|
| `matchLabels`      | `map[string]string`           | Key-value label pairs that must all match.       |
| `matchExpressions` | `LabelSelectorRequirement[]`  | Set-based requirements. See table below.         |

**LabelSelectorRequirement**

| Field      | Type       | Description                                                                                     |
|------------|------------|-------------------------------------------------------------------------------------------------|
| `key`      | `string`   | Label key.                                                                                      |
| `operator` | `string`   | Operator. Values: `"In"`, `"NotIn"`, `"Exists"`, `"DoesNotExist"`.                             |
| `values`   | `string[]` | Label values. Required for `In` and `NotIn`. Must be empty for `Exists` and `DoesNotExist`.     |

### Read-Only

| Name | Type   | Description                                            |
|------|--------|--------------------------------------------------------|
| `id` | string | Unique identifier of the target. Managed by the provider. |

## Import

```shell
pulumi import devzero:index/workloadPolicyTarget:WorkloadPolicyTarget my-target <target-id>
```
