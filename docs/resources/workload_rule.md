# devzero:resources:WorkloadRule

Pins explicit resource rules directly to a single Kubernetes workload (a specific `kind`/`namespace`/`name` on a cluster). Unlike `WorkloadPolicy`, which applies a shared policy to many workloads via a `WorkloadPolicyTarget`, a `WorkloadRule` targets one workload and lets you override CPU, memory, GPU, and HPA settings with precise values — or set `autoGenerate: true` to let the engine compute them from observed usage.

## Example Usage

### Minimal example (auto-generated)

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const rule = new devzero.WorkloadRule("my-app-rule", {
    clusterId:    "cluster-abc123",
    namespace:    "production",
    kind:         "Deployment",
    name:         "my-api",
    autoGenerate: true,
});
```

### Manual CPU, memory, and emergency-response rules

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const rule = new devzero.WorkloadRule("my-app-rule", {
    clusterId: "cluster-abc123",
    namespace: "production",
    kind:      "Deployment",
    name:      "my-api",

    actionTriggers:    ["on_schedule", "on_detection"],
    cronSchedule:      "0 2 * * *",
    detectionTriggers: ["pod_creation", "pod_update"],

    cpuRule: {
        enabled:                 true,
        minRequest:              10,      // millicores
        maxRequest:              32000,   // millicores (32 cores)
        targetPercentile:        0.95,    // P95 of observed CPU usage
        limitsAdjustmentEnabled: true,
        limitMultiplier:         1.0,
    },
    memoryRule: {
        enabled:                 true,
        minRequest:              67108864,     // 64 MiB
        maxRequest:              68719476736,  // 64 GiB
        targetPercentile:        0.95,
        limitsAdjustmentEnabled: true,
    },
    emergencyResponse: {
        oomEnabled:              true,
        oomMemoryMultiplier:     1.5,
        cpuThrottlingEnabled:    true,
        cpuThrottlingThreshold:  0.20,
        cpuThrottlingMultiplier: 1.25,
    },
});

export const ruleId = rule.id;
```

```python
import pulumi_devzero as devzero

rule = devzero.WorkloadRule(
    "my-app-rule",
    cluster_id="cluster-abc123",
    namespace="production",
    kind="Deployment",
    name="my-api",
    cpu_rule=devzero.ResourceRuleConfigArgs(
        enabled=True,
        min_request=10,
        max_request=32000,
        target_percentile=0.95,
        limits_adjustment_enabled=True,
    ),
)
```

## Schema

### Required

| Name        | Type   | Description                                                                          |
|-------------|--------|---------------------------------------------------------------------------------------|
| `clusterId` | string | ID of the cluster the workload lives in.                                              |
| `namespace` | string | Kubernetes namespace of the workload.                                                 |
| `kind`      | string | Workload kind. Values: `"Deployment"`, `"StatefulSet"`, `"DaemonSet"`, `"CronJob"`, `"Job"`. |
| `name`      | string | Name of the Kubernetes workload.                                                       |

### Optional — Rules

| Name                | Type                       | Description                                                                     |
|---------------------|----------------------------|----------------------------------------------------------------------------------|
| `autoGenerate`       | `boolean` | When `true`, the engine fills all rule fields from observed usage; manual field overrides below are ignored. |
| `cpuRule`            | `ResourceRuleConfigArgs`   | CPU vertical scaling rule.                                                       |
| `memoryRule`         | `ResourceRuleConfigArgs`   | Memory vertical scaling rule.                                                    |
| `gpuRule`            | `ResourceRuleConfigArgs`   | GPU vertical scaling rule (units: GPU millicores).                               |
| `hpaRule`            | `HPARuleConfigArgs`        | Horizontal (replica) scaling rule.                                               |
| `emergencyResponse`  | `EmergencyResponseConfigArgs` | OOM and CPU-throttle emergency reactions.                                    |
| `containers`         | `ContainerResourceRuleConfigArgs[]` | Per-container resource overrides. When empty, workload-level rules apply to all containers. |
| `disabled`           | `boolean` | Whether the rule is currently disabled.                                                     |

### Optional — Triggers & Timing

| Name                      | Type       | Description                                                                     |
|---------------------------|------------|-----------------------------------------------------------------------------------|
| `actionTriggers`          | `string[]` | When to apply recommendations. Values: `"on_detection"`, `"on_schedule"`.         |
| `cronSchedule`            | `string`   | Cron expression for scheduled application (5-field UTC). Required when `actionTriggers` includes `"on_schedule"`. |
| `detectionTriggers`       | `string[]` | Events that trigger a recommendation. Values: `"pod_creation"`, `"pod_update"`, `"pod_reschedule"`. |
| `startupPeriodSeconds`    | `number`   | Seconds after workload start to exclude from usage data.                          |
| `cooldownMinutes`         | `number`   | Minimum minutes between consecutive recommendation applications.                  |
| `lookbackPeriodSeconds`   | `number`   | Seconds to look back for resource usage data.                                     |

### Optional — Scaling Behaviour

| Name                        | Type      | Description                                                            |
|-----------------------------|-----------|--------------------------------------------------------------------------|
| `schedulerPlugins`          | `string[]`| Kubernetes scheduler plugins to activate. Example: `["binpacking"]`.     |
| `defragmentationSchedule`   | `string`  | Cron expression for background node defragmentation.                    |
| `liveMigrationEnabled`      | `boolean` | Allow live pod migration when applying recommendations without a restart.|
| `useInPlaceVerticalScaling` | `boolean` | Use in-place pod vertical scaling instead of pod restarts.               |

### `ResourceRuleConfigArgs`

Used for `cpuRule`, `memoryRule`, and `gpuRule` at both the workload and per-container level. `maxScaleUpPercent`/`maxScaleDownPercent` are **not** supported on per-container rules.

| Field                     | Type      | Description                                                       |
|---------------------------|-----------|---------------------------------------------------------------------|
| `enabled`                 | `boolean` | Enable this resource axis rule.                                     |
| `minRequest`              | `number`  | Minimum resource request (millicores for CPU, bytes for memory/GPU).|
| `maxRequest`              | `number`  | Maximum resource request.                                            |
| `targetPercentile`        | `number`  | Percentile of observed usage to target (0–1). Example: `0.95`.       |
| `maxScaleUpPercent`       | `number`  | Maximum % to scale up in one step *(workload-level only)*.           |
| `maxScaleDownPercent`     | `number`  | Maximum % to scale down in one step *(workload-level only)*.         |
| `limitsAdjustmentEnabled` | `boolean` | Whether to also adjust resource limits.                              |
| `limitMultiplier`         | `number`  | Limits = request × `limitMultiplier`.                                |
| `limitsRemovalEnabled`    | `boolean` | Actively remove limits from workloads (CPU only).                    |

### `HPARuleConfigArgs`

| Field                      | Type      | Description                                                                                  |
|----------------------------|-----------|------------------------------------------------------------------------------------------------|
| `enabled`                  | `boolean` | Enable horizontal (replica) scaling.                                                            |
| `minReplicas`              | `number`  | Minimum number of replicas.                                                                     |
| `maxReplicas`              | `number`  | Maximum number of replicas.                                                                     |
| `maxReplicaChangePercent`  | `number`  | Maximum percentage change in replica count per cycle.                                           |
| `scaleDownCooldownSeconds` | `number`  | Seconds to wait between scale-down events.                                                      |
| `metrics`                  | `HPAMetricTriggerArgs[]` | External metric triggers only (Prometheus, queue depth, etc). CPU/Memory/Network are engine-generated. |
| `compositeFormula`         | `string`  | Expression combining multiple metric ratios into one scaling signal. Example: `"0.6*cpu + 0.4*memory"`. |
| `behavior`                 | `HPABehaviorArgs` | Fine-grained scale-up and scale-down behavior policies.                                 |
| `fallback`                 | `HPAFallbackArgs` | Replica fallback when metrics become unavailable.                                       |

### `EmergencyResponseConfigArgs`

| Field                     | Type      | Description                                                       |
|---------------------------|-----------|---------------------------------------------------------------------|
| `oomEnabled`              | `boolean` | React to OOM kills by increasing the memory request.                |
| `oomMemoryMultiplier`     | `number`  | Multiplier applied to memory on each OOM event.                     |
| `oomMaxReactions`         | `number`  | Maximum OOM reactions before giving up.                             |
| `oomCooldownSeconds`      | `number`  | Seconds to wait between OOM reactions.                              |
| `cpuThrottlingEnabled`    | `boolean` | React to CPU throttling by increasing the CPU request.               |
| `cpuThrottlingThreshold`  | `number`  | Throttle ratio threshold that triggers a reaction (0–1).            |
| `cpuThrottlingMultiplier` | `number`  | Multiplier applied to the CPU request on a throttle reaction.       |

### `ContainerResourceRuleConfigArgs`

| Field           | Type                     | Description                                            |
|-----------------|--------------------------|-----------------------------------------------------------|
| `containerName` | `string`                 | Name of the container this config applies to.              |
| `cpuRule`       | `ResourceRuleConfigArgs` | CPU resource rule for this container.                      |
| `memoryRule`    | `ResourceRuleConfigArgs` | Memory resource rule for this container.                   |
| `gpuRule`       | `ResourceRuleConfigArgs` | GPU resource rule for this container.                      |

### Read-Only

| Name | Type   | Description                                             |
|------|--------|------------------------------------------------------------|
| `id` | string | Unique identifier of the rule. Managed by the provider.     |

## Import

An existing workload rule can be imported using its rule ID:

```shell
pulumi import devzero:resources:WorkloadRule my-app-rule <rule-id>

# Example
pulumi import devzero:resources:WorkloadRule production-api "b7f3c1a2-9d4e-4a1b-8c3f-2e5d6f7a8b90"
```

> **Note:** `autoGenerate` is not stored as its own field on the API side — it is inferred on every `Read` (including at import time) from the rule's `current_source`: `true` when the rule is engine-managed (`auto_optimization`), unset otherwise. This keeps an imported rule's mode consistent with how it actually behaves on the platform.
