package resources

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// VerticalScalingArgs configures vertical scaling for a single resource type (CPU/Memory/GPU/VRAM).
type VerticalScalingArgs struct {
	Enabled                 bool     `pulumi:"enabled,optional"`
	MinRequest              *int     `pulumi:"minRequest,optional"`
	MaxRequest              *int     `pulumi:"maxRequest,optional"`
	OverheadMultiplier      *float64 `pulumi:"overheadMultiplier,optional"`
	TargetPercentile        *float64 `pulumi:"targetPercentile,optional"`
	MaxScaleUpPercent       *float64 `pulumi:"maxScaleUpPercent,optional"`
	MaxScaleDownPercent     *float64 `pulumi:"maxScaleDownPercent,optional"`
	LimitsAdjustmentEnabled *bool    `pulumi:"limitsAdjustmentEnabled,optional"`
	LimitMultiplier         *float64 `pulumi:"limitMultiplier,optional"`
	MinDataPoints           *int     `pulumi:"minDataPoints,optional"`
	AdjustReqEvenIfNotSet   *bool    `pulumi:"adjustReqEvenIfNotSet,optional"`
	LimitsRemovalEnabled    *bool    `pulumi:"limitsRemovalEnabled,optional"`
}

// Annotate provides SDK documentation for VerticalScalingArgs fields.
func (v *VerticalScalingArgs) Annotate(a infer.Annotator) {
	a.Describe(&v.Enabled, "Enable vertical scaling for this resource type.")
	a.Describe(&v.MinRequest, "Minimum resource request in millicores (CPU) or bytes (memory/GPU).")
	a.Describe(&v.MaxRequest, "Maximum resource request in millicores (CPU) or bytes (memory/GPU).")
	a.Describe(&v.OverheadMultiplier, "Multiplier applied on top of the recommendation to add headroom.")
	a.Describe(&v.TargetPercentile, "Percentile of usage data used as the recommendation target (e.g. 0.95).")
	a.Describe(&v.MaxScaleUpPercent, "Maximum percentage increase allowed in a single recommendation cycle.")
	a.Describe(&v.MaxScaleDownPercent, "Maximum percentage decrease allowed in a single recommendation cycle.")
	a.Describe(&v.LimitsAdjustmentEnabled, "Whether to also adjust resource limits alongside requests.")
	a.Describe(&v.LimitMultiplier, "Multiplier applied to the request to derive the resource limit.")
	a.Describe(&v.MinDataPoints, "Minimum number of data points required before a recommendation is emitted.")
	a.Describe(&v.AdjustReqEvenIfNotSet, "Recommend requests even when the workload has no existing requests set. Server/web default: true.")
	a.Describe(&v.LimitsRemovalEnabled, "Actively remove limits from workloads (CPU only). Takes precedence over limitsAdjustmentEnabled. Web default: true for CPU, false for Memory.")

	a.SetDefault(&v.MinDataPoints, 20)
	a.SetDefault(&v.MaxScaleUpPercent, 1000.0)
	a.SetDefault(&v.MaxScaleDownPercent, 1.0)
}

// HorizontalScalingArgs configures horizontal (replica) scaling.
type HorizontalScalingArgs struct {
	Enabled                 bool     `pulumi:"enabled,optional"`
	MinReplicas             *int     `pulumi:"minReplicas,optional"`
	MaxReplicas             *int     `pulumi:"maxReplicas,optional"`
	TargetUtilization       *float64 `pulumi:"targetUtilization,optional"`
	PrimaryMetric           *string  `pulumi:"primaryMetric,optional"`
	MinDataPoints           *int     `pulumi:"minDataPoints,optional"`
	MaxReplicaChangePercent *float64 `pulumi:"maxReplicaChangePercent,optional"`
}

// Annotate provides SDK documentation for HorizontalScalingArgs fields.
func (h *HorizontalScalingArgs) Annotate(a infer.Annotator) {
	a.Describe(&h.Enabled, "Enable horizontal (replica) scaling.")
	a.Describe(&h.MinReplicas, "Minimum number of replicas to maintain.")
	a.Describe(&h.MaxReplicas, "Maximum number of replicas to scale up to.")
	a.Describe(&h.TargetUtilization, "Target utilization ratio (0-1) for the primary metric.")
	a.Describe(&h.PrimaryMetric, "Primary metric for HPA decisions. One of: 'cpu', 'memory', 'gpu', 'network_ingress', 'network_egress'.")
	a.Describe(&h.MinDataPoints, "Minimum data points required before a recommendation is emitted.")
	a.Describe(&h.MaxReplicaChangePercent, "Maximum percentage change in replica count per recommendation cycle.")
}

// WorkloadPolicyArgs are the user-configurable inputs for a WorkloadPolicy resource.
type WorkloadPolicyArgs struct {
	Name                    string                 `pulumi:"name"`
	Description             *string                `pulumi:"description,optional"`
	ActionTriggers          []string               `pulumi:"actionTriggers,optional"`
	CronSchedule            *string                `pulumi:"cronSchedule,optional"`
	DetectionTriggers       []string               `pulumi:"detectionTriggers,optional"`
	LoopbackPeriodSeconds   *int                   `pulumi:"loopbackPeriodSeconds,optional"`
	StartupPeriodSeconds    *int                   `pulumi:"startupPeriodSeconds,optional"`
	LiveMigrationEnabled    bool                   `pulumi:"liveMigrationEnabled,optional"`
	SchedulerPlugins        []string               `pulumi:"schedulerPlugins,optional"`
	DefragmentationSchedule *string                `pulumi:"defragmentationSchedule,optional"`
	CpuVerticalScaling      *VerticalScalingArgs   `pulumi:"cpuVerticalScaling,optional"`
	MemoryVerticalScaling   *VerticalScalingArgs   `pulumi:"memoryVerticalScaling,optional"`
	GpuVerticalScaling      *VerticalScalingArgs   `pulumi:"gpuVerticalScaling,optional"`
	GpuVramVerticalScaling  *VerticalScalingArgs   `pulumi:"gpuVramVerticalScaling,optional"`
	HorizontalScaling       *HorizontalScalingArgs `pulumi:"horizontalScaling,optional"`
	MinChangePercent        *float64               `pulumi:"minChangePercent,optional"`
	MinDataPoints           *int                   `pulumi:"minDataPoints,optional"`
	StabilityCvMax          *float64               `pulumi:"stabilityCvMax,optional"`
	HysteresisVsTarget      *float64               `pulumi:"hysteresisVsTarget,optional"`
	DriftDeltaPercent       *float64               `pulumi:"driftDeltaPercent,optional"`
	MinVpaWindowDataPoints  *int                   `pulumi:"minVpaWindowDataPoints,optional"`
	CooldownMinutes         *int                   `pulumi:"cooldownMinutes,optional"`
	EnablePmaxProtection    *bool                  `pulumi:"enablePmaxProtection,optional"`
	PmaxRatioThreshold      *float64               `pulumi:"pmaxRatioThreshold,optional"`
}

// Annotate provides SDK documentation and default values for WorkloadPolicyArgs fields.
// Placed on Args (not State) so that inputProperties in the schema pick up defaults.
func (a *WorkloadPolicyArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Name, "Human-friendly name for the policy.")
	ann.Describe(&a.Description, "Free-form description of the policy.")
	ann.Describe(&a.ActionTriggers, "Action triggers: 'on_detection' or 'on_schedule'.")
	ann.Describe(&a.CronSchedule, "Cron expression for scheduled application (5-field format).")
	ann.Describe(&a.DetectionTriggers, "Detection triggers: 'pod_creation', 'pod_update', or 'pod_reschedule'.")
	ann.Describe(&a.LoopbackPeriodSeconds, "Period in seconds to look back for resource usage data. Default: 86400 (24 h).")
	ann.Describe(&a.StartupPeriodSeconds, "Period in seconds to ignore usage data after workload starts.")
	ann.Describe(&a.LiveMigrationEnabled, "Allow live migration when applying recommendations.")
	ann.Describe(&a.SchedulerPlugins, "Kubernetes scheduler plugins to activate.")
	ann.Describe(&a.DefragmentationSchedule, "Cron expression for background defragmentation.")
	ann.Describe(&a.MinChangePercent, "Global minimum change threshold for applying recommendations. Default: 0.2 (20%).")
	ann.Describe(&a.MinDataPoints, "Global minimum data points required for recommendations. Default: 15.")
	ann.Describe(&a.StabilityCvMax, "Maximum coefficient of variation for workload to be considered stable.")
	ann.Describe(&a.HysteresisVsTarget, "Hysteresis threshold vs target for HPA coordination.")
	ann.Describe(&a.DriftDeltaPercent, "Percentage drift from baseline that triggers VPA refresh.")
	ann.Describe(&a.MinVpaWindowDataPoints, "Minimum data points in VPA analysis window. Default: 30.")
	ann.Describe(&a.CooldownMinutes, "Minutes to wait between applying recommendations. Default: 300 (5 h).")
	ann.Describe(&a.EnablePmaxProtection, "Raise requests to cover peak usage when max/recommendation ratio exceeds pmaxRatioThreshold. Server/web default: true.")
	ann.Describe(&a.PmaxRatioThreshold, "Max-to-recommendation ratio that triggers pmax protection. Default: 3.0.")
	ann.SetDefault(&a.PmaxRatioThreshold, 3.0)
	ann.SetDefault(&a.LoopbackPeriodSeconds, 86400)
	ann.SetDefault(&a.MinDataPoints, 15)
	ann.SetDefault(&a.MinChangePercent, 0.2)
	ann.SetDefault(&a.MinVpaWindowDataPoints, 30)
	ann.SetDefault(&a.CooldownMinutes, 300)
}

// WorkloadPolicyState is the full persisted state (identical to args — no additional computed fields).
type WorkloadPolicyState struct {
	WorkloadPolicyArgs
}

// WorkloadPolicy is the resource implementation.
type WorkloadPolicy struct{}

// ---------- CRUD ----------

func (w *WorkloadPolicy) Create(ctx context.Context, req infer.CreateRequest[WorkloadPolicyArgs]) (infer.CreateResponse[WorkloadPolicyState], error) {
	if req.DryRun {
		return infer.CreateResponse[WorkloadPolicyState]{Output: WorkloadPolicyState{WorkloadPolicyArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.CreateResponse[WorkloadPolicyState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.CreateWorkloadRecommendationPolicy(ctx, connect.NewRequest(&apiv1.CreateWorkloadRecommendationPolicyRequest{
		TeamId: cs.TeamID,
		Policy: argsToProto(cs.TeamID, "", req.Inputs),
	}))
	if err != nil {
		return infer.CreateResponse[WorkloadPolicyState]{}, fmt.Errorf("CreateWorkloadRecommendationPolicy: %w", err)
	}
	if resp.Msg.Policy == nil {
		return infer.CreateResponse[WorkloadPolicyState]{}, fmt.Errorf("CreateWorkloadRecommendationPolicy: empty response from server")
	}

	return infer.CreateResponse[WorkloadPolicyState]{
		ID:     resp.Msg.Policy.PolicyId,
		Output: WorkloadPolicyState{WorkloadPolicyArgs: protoToArgs(resp.Msg.Policy)},
	}, nil
}

func (w *WorkloadPolicy) Read(ctx context.Context, req infer.ReadRequest[WorkloadPolicyArgs, WorkloadPolicyState]) (infer.ReadResponse[WorkloadPolicyArgs, WorkloadPolicyState], error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.ReadResponse[WorkloadPolicyArgs, WorkloadPolicyState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.GetWorkloadRecommendationPolicy(ctx, connect.NewRequest(&apiv1.GetWorkloadRecommendationPolicyRequest{
		TeamId:   cs.TeamID,
		PolicyId: req.ID,
	}))
	if err != nil {
		return infer.ReadResponse[WorkloadPolicyArgs, WorkloadPolicyState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("GetWorkloadRecommendationPolicy: %w", err)
	}
	if resp.Msg.Policy == nil {
		return infer.ReadResponse[WorkloadPolicyArgs, WorkloadPolicyState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("GetWorkloadRecommendationPolicy: policy not found")
	}

	updatedArgs := protoToArgs(resp.Msg.Policy)
	return infer.ReadResponse[WorkloadPolicyArgs, WorkloadPolicyState]{
		ID:     req.ID,
		Inputs: updatedArgs,
		State:  WorkloadPolicyState{WorkloadPolicyArgs: updatedArgs},
	}, nil
}

func (w *WorkloadPolicy) Update(ctx context.Context, req infer.UpdateRequest[WorkloadPolicyArgs, WorkloadPolicyState]) (infer.UpdateResponse[WorkloadPolicyState], error) {
	if req.DryRun {
		return infer.UpdateResponse[WorkloadPolicyState]{Output: WorkloadPolicyState{WorkloadPolicyArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.UpdateResponse[WorkloadPolicyState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.UpdateWorkloadRecommendationPolicy(ctx, connect.NewRequest(&apiv1.UpdateWorkloadRecommendationPolicyRequest{
		TeamId: cs.TeamID,
		Policy: argsToProto(cs.TeamID, req.ID, req.Inputs),
	}))
	if err != nil {
		return infer.UpdateResponse[WorkloadPolicyState]{}, fmt.Errorf("UpdateWorkloadRecommendationPolicy: %w", err)
	}
	if resp.Msg.Policy == nil {
		return infer.UpdateResponse[WorkloadPolicyState]{}, fmt.Errorf("UpdateWorkloadRecommendationPolicy: empty response from server")
	}

	return infer.UpdateResponse[WorkloadPolicyState]{
		Output: WorkloadPolicyState{WorkloadPolicyArgs: protoToArgs(resp.Msg.Policy)},
	}, nil
}

func (w *WorkloadPolicy) Delete(ctx context.Context, req infer.DeleteRequest[WorkloadPolicyState]) (infer.DeleteResponse, error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.DeleteResponse{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	_, err := cs.RecommendationClient.DeleteWorkloadRecommendationPolicy(ctx, connect.NewRequest(&apiv1.DeleteWorkloadRecommendationPolicyRequest{
		TeamId:   cs.TeamID,
		PolicyId: req.ID,
	}))
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("DeleteWorkloadRecommendationPolicy: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

// ---------- proto conversion helpers ----------

func argsToProto(teamID, policyID string, a WorkloadPolicyArgs) *apiv1.WorkloadRecommendationPolicy {
	p := &apiv1.WorkloadRecommendationPolicy{
		PolicyId:               policyID,
		TeamId:                 teamID,
		Name:                   a.Name,
		ActionTriggers:         actionTriggersToProto(a.ActionTriggers),
		DetectionTriggers:      detectionTriggersToProto(a.DetectionTriggers),
		LiveMigrationEnabled:   a.LiveMigrationEnabled,
		SchedulerPlugins:       a.SchedulerPlugins,
		CpuVerticalScaling:     verticalScalingToProto(a.CpuVerticalScaling),
		MemoryVerticalScaling:  verticalScalingToProto(a.MemoryVerticalScaling),
		GpuVerticalScaling:     verticalScalingToProto(a.GpuVerticalScaling),
		GpuVramVerticalScaling: verticalScalingToProto(a.GpuVramVerticalScaling),
		HorizontalScaling:      horizontalScalingToProto(a.HorizontalScaling),
	}
	if a.Description != nil {
		p.Description = *a.Description
	}
	if a.CronSchedule != nil {
		p.CronSchedule = a.CronSchedule
	}
	if a.DefragmentationSchedule != nil {
		p.DefragmentationSchedule = a.DefragmentationSchedule
	}
	if a.LoopbackPeriodSeconds != nil {
		v := int32(*a.LoopbackPeriodSeconds)
		p.LoopbackPeriodSeconds = &v
	}
	if a.StartupPeriodSeconds != nil {
		v := int64(*a.StartupPeriodSeconds)
		p.StartupPeriodSeconds = &v
	}
	if a.MinChangePercent != nil {
		v := float32(*a.MinChangePercent)
		p.MinChangePercent = &v
	}
	if a.MinDataPoints != nil {
		v := int32(*a.MinDataPoints)
		p.MinDataPoints = &v
	}
	if a.StabilityCvMax != nil {
		v := float32(*a.StabilityCvMax)
		p.StabilityCvMax = &v
	}
	if a.HysteresisVsTarget != nil {
		v := float32(*a.HysteresisVsTarget)
		p.HysteresisVsTarget = &v
	}
	if a.DriftDeltaPercent != nil {
		v := float32(*a.DriftDeltaPercent)
		p.DriftDeltaPercent = &v
	}
	if a.MinVpaWindowDataPoints != nil {
		v := int32(*a.MinVpaWindowDataPoints)
		p.MinVpaWindowDataPoints = &v
	}
	if a.CooldownMinutes != nil {
		v := int32(*a.CooldownMinutes)
		p.CooldownMinutes = &v
	}
	if a.EnablePmaxProtection != nil {
		p.EnablePmaxProtection = *a.EnablePmaxProtection
	}
	if a.PmaxRatioThreshold != nil {
		v := float32(*a.PmaxRatioThreshold)
		p.PmaxRatioThreshold = &v
	}
	return p
}

func protoToArgs(p *apiv1.WorkloadRecommendationPolicy) WorkloadPolicyArgs {
	a := WorkloadPolicyArgs{
		Name:                   p.Name,
		ActionTriggers:         actionTriggersFromProto(p.ActionTriggers),
		DetectionTriggers:      detectionTriggersFromProto(p.DetectionTriggers),
		LiveMigrationEnabled:   p.LiveMigrationEnabled,
		SchedulerPlugins:       p.SchedulerPlugins,
		CpuVerticalScaling:     verticalScalingFromProto(p.CpuVerticalScaling),
		MemoryVerticalScaling:  verticalScalingFromProto(p.MemoryVerticalScaling),
		GpuVerticalScaling:     verticalScalingFromProto(p.GpuVerticalScaling),
		GpuVramVerticalScaling: verticalScalingFromProto(p.GpuVramVerticalScaling),
		HorizontalScaling:      horizontalScalingFromProto(p.HorizontalScaling),
	}
	if p.Description != "" {
		a.Description = &p.Description
	}
	if p.CronSchedule != nil {
		a.CronSchedule = p.CronSchedule
	}
	if p.DefragmentationSchedule != nil {
		a.DefragmentationSchedule = p.DefragmentationSchedule
	}
	if p.LoopbackPeriodSeconds != nil {
		v := int(*p.LoopbackPeriodSeconds)
		a.LoopbackPeriodSeconds = &v
	}
	if p.StartupPeriodSeconds != nil {
		v := int(*p.StartupPeriodSeconds)
		a.StartupPeriodSeconds = &v
	}
	if p.MinChangePercent != nil {
		v := float64(*p.MinChangePercent)
		a.MinChangePercent = &v
	}
	if p.MinDataPoints != nil {
		v := int(*p.MinDataPoints)
		a.MinDataPoints = &v
	}
	if p.StabilityCvMax != nil {
		v := float64(*p.StabilityCvMax)
		a.StabilityCvMax = &v
	}
	if p.HysteresisVsTarget != nil {
		v := float64(*p.HysteresisVsTarget)
		a.HysteresisVsTarget = &v
	}
	if p.DriftDeltaPercent != nil {
		v := float64(*p.DriftDeltaPercent)
		a.DriftDeltaPercent = &v
	}
	if p.MinVpaWindowDataPoints != nil {
		v := int(*p.MinVpaWindowDataPoints)
		a.MinVpaWindowDataPoints = &v
	}
	if p.CooldownMinutes != nil {
		v := int(*p.CooldownMinutes)
		a.CooldownMinutes = &v
	}
	v := p.EnablePmaxProtection
	a.EnablePmaxProtection = &v
	if p.PmaxRatioThreshold != nil {
		v := float64(*p.PmaxRatioThreshold)
		a.PmaxRatioThreshold = &v
	}
	return a
}

func actionTriggersToProto(triggers []string) []apiv1.ActionTrigger {
	result := make([]apiv1.ActionTrigger, 0, len(triggers))
	for _, t := range triggers {
		switch t {
		case "on_schedule":
			result = append(result, apiv1.ActionTrigger_ACTION_TRIGGER_ON_SCHEDULE)
		case "on_detection":
			result = append(result, apiv1.ActionTrigger_ACTION_TRIGGER_ON_DETECTION)
		}
	}
	return result
}

func actionTriggersFromProto(triggers []apiv1.ActionTrigger) []string {
	if len(triggers) == 0 {
		return nil
	}
	result := make([]string, 0, len(triggers))
	for _, t := range triggers {
		switch t {
		case apiv1.ActionTrigger_ACTION_TRIGGER_ON_SCHEDULE:
			result = append(result, "on_schedule")
		case apiv1.ActionTrigger_ACTION_TRIGGER_ON_DETECTION:
			result = append(result, "on_detection")
		}
	}
	return result
}

func detectionTriggersToProto(triggers []string) []apiv1.WorkloadDetectionTrigger {
	result := make([]apiv1.WorkloadDetectionTrigger, 0, len(triggers))
	for _, t := range triggers {
		switch t {
		case "pod_creation":
			result = append(result, apiv1.WorkloadDetectionTrigger_DETECTION_TRIGGER_POD_CREATION)
		case "pod_update":
			result = append(result, apiv1.WorkloadDetectionTrigger_DETECTION_TRIGGER_POD_UPDATE)
		case "pod_reschedule":
			result = append(result, apiv1.WorkloadDetectionTrigger_DETECTION_TRIGGER_POD_EVICT)
		}
	}
	return result
}

func detectionTriggersFromProto(triggers []apiv1.WorkloadDetectionTrigger) []string {
	if len(triggers) == 0 {
		return nil
	}
	result := make([]string, 0, len(triggers))
	for _, t := range triggers {
		switch t {
		case apiv1.WorkloadDetectionTrigger_DETECTION_TRIGGER_POD_CREATION:
			result = append(result, "pod_creation")
		case apiv1.WorkloadDetectionTrigger_DETECTION_TRIGGER_POD_UPDATE:
			result = append(result, "pod_update")
		case apiv1.WorkloadDetectionTrigger_DETECTION_TRIGGER_POD_EVICT:
			result = append(result, "pod_reschedule")
		}
	}
	return result
}

func verticalScalingToProto(v *VerticalScalingArgs) *apiv1.VerticalScalingOptimizationTarget {
	if v == nil {
		return nil
	}
	t := &apiv1.VerticalScalingOptimizationTarget{
		Enabled: v.Enabled,
	}
	if v.MinRequest != nil {
		x := int64(*v.MinRequest)
		t.MinRequest = &x
	}
	if v.MaxRequest != nil {
		x := int64(*v.MaxRequest)
		t.MaxRequest = &x
	}
	if v.OverheadMultiplier != nil {
		x := float32(*v.OverheadMultiplier)
		t.OverheadMultiplier = &x
	}
	if v.TargetPercentile != nil {
		x := float32(*v.TargetPercentile)
		t.TargetPercentile = &x
	}
	if v.MaxScaleUpPercent != nil {
		x := float32(*v.MaxScaleUpPercent)
		t.MaxScaleUpPercent = &x
	}
	if v.MaxScaleDownPercent != nil {
		x := float32(*v.MaxScaleDownPercent)
		t.MaxScaleDownPercent = &x
	}
	if v.LimitsAdjustmentEnabled != nil {
		t.LimitsAdjustmentEnabled = v.LimitsAdjustmentEnabled
	}
	if v.LimitMultiplier != nil {
		x := float32(*v.LimitMultiplier)
		t.LimitMultiplier = &x
	}
	if v.MinDataPoints != nil {
		x := int32(*v.MinDataPoints)
		t.MinDataPoints = &x
	}
	if v.AdjustReqEvenIfNotSet != nil {
		t.AdjustReqEvenIfNotSet = *v.AdjustReqEvenIfNotSet
	}
	if v.LimitsRemovalEnabled != nil {
		t.LimitsRemovalEnabled = *v.LimitsRemovalEnabled
	}
	return t
}

func verticalScalingFromProto(t *apiv1.VerticalScalingOptimizationTarget) *VerticalScalingArgs {
	if t == nil {
		return nil
	}
	v := &VerticalScalingArgs{
		Enabled: t.Enabled,
	}
	if t.MinRequest != nil {
		x := int(*t.MinRequest)
		v.MinRequest = &x
	}
	if t.MaxRequest != nil {
		x := int(*t.MaxRequest)
		v.MaxRequest = &x
	}
	if t.OverheadMultiplier != nil {
		x := float64(*t.OverheadMultiplier)
		v.OverheadMultiplier = &x
	}
	if t.TargetPercentile != nil {
		x := float64(*t.TargetPercentile)
		v.TargetPercentile = &x
	}
	if t.MaxScaleUpPercent != nil {
		x := float64(*t.MaxScaleUpPercent)
		v.MaxScaleUpPercent = &x
	}
	if t.MaxScaleDownPercent != nil {
		x := float64(*t.MaxScaleDownPercent)
		v.MaxScaleDownPercent = &x
	}
	if t.LimitsAdjustmentEnabled != nil {
		v.LimitsAdjustmentEnabled = t.LimitsAdjustmentEnabled
	}
	if t.LimitMultiplier != nil {
		x := float64(*t.LimitMultiplier)
		v.LimitMultiplier = &x
	}
	if t.MinDataPoints != nil {
		x := int(*t.MinDataPoints)
		v.MinDataPoints = &x
	}
	adj := t.AdjustReqEvenIfNotSet
	v.AdjustReqEvenIfNotSet = &adj
	lre := t.LimitsRemovalEnabled
	v.LimitsRemovalEnabled = &lre
	return v
}

func horizontalScalingToProto(h *HorizontalScalingArgs) *apiv1.HorizontalScalingOptimizationTarget {
	if h == nil {
		return nil
	}
	t := &apiv1.HorizontalScalingOptimizationTarget{
		Enabled: h.Enabled,
	}
	if h.MinReplicas != nil {
		x := int32(*h.MinReplicas)
		t.MinReplicas = &x
	}
	if h.MaxReplicas != nil {
		x := int32(*h.MaxReplicas)
		t.MaxReplicas = &x
	}
	if h.TargetUtilization != nil {
		x := float32(*h.TargetUtilization)
		t.TargetUtilization = &x
	}
	if h.PrimaryMetric != nil {
		t.PrimaryMetric = hpaMetricToProto(h.PrimaryMetric)
	}
	if h.MinDataPoints != nil {
		x := int32(*h.MinDataPoints)
		t.MinDataPoints = &x
	}
	if h.MaxReplicaChangePercent != nil {
		x := float32(*h.MaxReplicaChangePercent)
		t.MaxReplicaChangePercent = &x
	}
	return t
}

func horizontalScalingFromProto(t *apiv1.HorizontalScalingOptimizationTarget) *HorizontalScalingArgs {
	if t == nil {
		return nil
	}
	h := &HorizontalScalingArgs{
		Enabled: t.Enabled,
	}
	if t.MinReplicas != nil {
		x := int(*t.MinReplicas)
		h.MinReplicas = &x
	}
	if t.MaxReplicas != nil {
		x := int(*t.MaxReplicas)
		h.MaxReplicas = &x
	}
	if t.TargetUtilization != nil {
		x := float64(*t.TargetUtilization)
		h.TargetUtilization = &x
	}
	if t.PrimaryMetric != nil {
		h.PrimaryMetric = hpaMetricFromProto(t.PrimaryMetric)
	}
	if t.MinDataPoints != nil {
		x := int(*t.MinDataPoints)
		h.MinDataPoints = &x
	}
	if t.MaxReplicaChangePercent != nil {
		x := float64(*t.MaxReplicaChangePercent)
		h.MaxReplicaChangePercent = &x
	}
	return h
}

func hpaMetricToProto(s *string) *apiv1.HPAMetricType {
	if s == nil {
		return nil
	}
	var v apiv1.HPAMetricType
	switch *s {
	case "cpu":
		v = apiv1.HPAMetricType_HPA_METRIC_TYPE_CPU
	case "memory":
		v = apiv1.HPAMetricType_HPA_METRIC_TYPE_MEMORY
	case "gpu":
		v = apiv1.HPAMetricType_HPA_METRIC_TYPE_GPU
	case "network_ingress":
		v = apiv1.HPAMetricType_HPA_METRIC_TYPE_NETWORK_INGRESS
	case "network_egress":
		v = apiv1.HPAMetricType_HPA_METRIC_TYPE_NETWORK_EGRESS
	default:
		return nil
	}
	return &v
}

func hpaMetricFromProto(m *apiv1.HPAMetricType) *string {
	if m == nil {
		return nil
	}
	var s string
	switch *m {
	case apiv1.HPAMetricType_HPA_METRIC_TYPE_CPU:
		s = "cpu"
	case apiv1.HPAMetricType_HPA_METRIC_TYPE_MEMORY:
		s = "memory"
	case apiv1.HPAMetricType_HPA_METRIC_TYPE_GPU:
		s = "gpu"
	case apiv1.HPAMetricType_HPA_METRIC_TYPE_NETWORK_INGRESS:
		s = "network_ingress"
	case apiv1.HPAMetricType_HPA_METRIC_TYPE_NETWORK_EGRESS:
		s = "network_egress"
	default:
		return nil
	}
	return &s
}
