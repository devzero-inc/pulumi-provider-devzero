# devzero:index/workloadPolicy:WorkloadPolicy

Manages a DevZero workload recommendation policy. A workload policy defines how the DevZero platform optimizes resource requests and replica counts for Kubernetes workloads — including vertical scaling (CPU, memory, GPU, VRAM) and horizontal scaling (HPA).

## Example Usage

### Minimal example

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const policy = new devzero.WorkloadPolicy("basic-policy", {
    name: "default-optimization",
});
```

### Full vertical and horizontal scaling

```typescript
import * as devzero from "@devzero/pulumi-provider-devzero";

const policy = new devzero.WorkloadPolicy("full-policy", {
    name: "production-optimization",
    description: "Optimizes CPU and memory with scheduled application",
    actionTriggers: ["on_schedule"],
    cronSchedule: "0 2 * * *",   // apply at 2 AM daily
    detectionTriggers: ["pod_creation", "pod_update"],
    loopbackPeriodSeconds: 604800, // 7 days
    startupPeriodSeconds: 300,
    liveMigrationEnabled: true,
    cooldownMinutes: 60,

    cpuVerticalScaling: {
        enabled: true,
        targetPercentile: 95,
        maxScaleUpPercent: 50,
        maxScaleDownPercent: 20,
        limitsAdjustmentEnabled: true,
        limitMultiplier: 1.5,
        minDataPoints: 10,
    },
    memoryVerticalScaling: {
        enabled: true,
        targetPercentile: 90,
        maxScaleUpPercent: 30,
        maxScaleDownPercent: 10,
        limitsAdjustmentEnabled: true,
    },
    horizontalScaling: {
        enabled: true,
        minReplicas: 2,
        maxReplicas: 20,
        targetUtilization: 70,
        primaryMetric: "cpu",
    },
});

export const policyId = policy.id;
```

```python
import pulumi_devzero as devzero

policy = devzero.WorkloadPolicy("full-policy",
    name="production-optimization",
    description="Optimizes CPU and memory with scheduled application",
    action_triggers=["on_schedule"],
    cron_schedule="0 2 * * *",
    detection_triggers=["pod_creation", "pod_update"],
    cpu_vertical_scaling=devzero.VerticalScalingArgs(
        enabled=True,
        target_percentile=95,
        max_scale_up_percent=50,
        limits_adjustment_enabled=True,
    ),
    horizontal_scaling=devzero.HorizontalScalingArgs(
        enabled=True,
        min_replicas=2,
        max_replicas=20,
        target_utilization=70,
        primary_metric="cpu",
    ),
)
```

## Schema

### Required

| Name   | Type   | Description                           |
|--------|--------|---------------------------------------|
| `name` | string | Human-friendly name for the policy.   |

### Optional — Triggers

| Name                  | Type       | Description                                                                 |
|-----------------------|------------|-----------------------------------------------------------------------------|
| `actionTriggers`      | `string[]` | When recommendations are applied. Values: `"on_schedule"`, `"on_detection"`. |
| `cronSchedule`        | `string`   | Cron expression for scheduled application (5-field format). Required when `"on_schedule"` is used. |
| `detectionTriggers`   | `string[]` | Events that trigger detection. Values: `"pod_creation"`, `"pod_update"`, `"pod_reschedule"`. |

### Optional — Timing

| Name                    | Type      | Description                                                     |
|-------------------------|-----------|-----------------------------------------------------------------|
| `description`           | `string`  | Free-form description of the policy.                            |
| `loopbackPeriodSeconds` | `number`  | Seconds to look back for resource usage data.                   |
| `startupPeriodSeconds`  | `number`  | Seconds to ignore usage data after workload starts.             |
| `cooldownMinutes`       | `number`  | Minutes to wait between applying successive recommendations.    |

### Optional — Scaling Behaviour

| Name                     | Type      | Description                                                                   |
|--------------------------|-----------|-------------------------------------------------------------------------------|
| `liveMigrationEnabled`   | `boolean` | Allow live migration when applying recommendations.                           |
| `schedulerPlugins`       | `string[]`| Kubernetes scheduler plugins to activate.                                     |
| `defragmentationSchedule`| `string`  | Cron expression for background defragmentation.                               |
| `minChangePercent`       | `number`  | Global minimum change threshold before recommendations are applied.           |
| `minDataPoints`          | `number`  | Global minimum data points required before recommendations are generated.     |
| `stabilityCvMax`         | `number`  | Max coefficient of variation for a workload to be considered stable.          |
| `hysteresisVsTarget`     | `number`  | Hysteresis threshold vs target for HPA coordination.                          |
| `driftDeltaPercent`      | `number`  | Percentage drift from baseline that triggers a VPA refresh.                   |
| `minVpaWindowDataPoints` | `number`  | Minimum data points in the VPA analysis window.                               |

### Optional — Vertical Scaling (per resource type)

Each of `cpuVerticalScaling`, `memoryVerticalScaling`, `gpuVerticalScaling`, and `gpuVramVerticalScaling` accepts:

| Field                    | Type      | Description                                              |
|--------------------------|-----------|----------------------------------------------------------|
| `enabled`                | `boolean` | Enable this scaler.                                      |
| `minRequest`             | `number`  | Minimum resource request (milli-units for CPU, bytes for memory). |
| `maxRequest`             | `number`  | Maximum resource request.                                |
| `overheadMultiplier`     | `number`  | Multiplier applied on top of observed usage.             |
| `targetPercentile`       | `number`  | Usage percentile to target (0–100).                      |
| `maxScaleUpPercent`      | `number`  | Maximum % to scale up in a single recommendation.        |
| `maxScaleDownPercent`    | `number`  | Maximum % to scale down in a single recommendation.      |
| `limitsAdjustmentEnabled`| `boolean` | Also adjust resource limits, not just requests.          |
| `limitMultiplier`        | `number`  | Multiplier applied to the recommended request to set limits. |
| `minDataPoints`          | `number`  | Minimum data points for this scaler.                     |

### Optional — Horizontal Scaling

`horizontalScaling` accepts:

| Field                    | Type      | Description                                                                     |
|--------------------------|-----------|---------------------------------------------------------------------------------|
| `enabled`                | `boolean` | Enable horizontal scaling.                                                      |
| `minReplicas`            | `number`  | Minimum replica count.                                                          |
| `maxReplicas`            | `number`  | Maximum replica count.                                                          |
| `targetUtilization`      | `number`  | Target utilization (0–100).                                                     |
| `primaryMetric`          | `string`  | HPA primary metric. Values: `"cpu"`, `"memory"`, `"gpu"`, `"network_ingress"`, `"network_egress"`. |
| `minDataPoints`          | `number`  | Minimum data points for this scaler.                                            |
| `maxReplicaChangePercent`| `number`  | Maximum % replica change in a single recommendation.                            |

### Read-Only

| Name | Type   | Description                              |
|------|--------|------------------------------------------|
| `id` | string | Unique identifier of the policy. Managed by the provider. |

## Import

```shell
pulumi import devzero:index/workloadPolicy:WorkloadPolicy my-policy <policy-id>
```
