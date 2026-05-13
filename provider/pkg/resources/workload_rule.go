package resources

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// ResourceRuleConfigArgs configures vertical scaling for a single resource axis on a WorkloadRule.
type ResourceRuleConfigArgs struct {
	Enabled                 bool     `pulumi:"enabled,optional"`
	MinRequest              *int     `pulumi:"minRequest,optional"`
	MaxRequest              *int     `pulumi:"maxRequest,optional"`
	LimitMultiplier         *float64 `pulumi:"limitMultiplier,optional"`
	LimitsAdjustmentEnabled bool     `pulumi:"limitsAdjustmentEnabled,optional"`
	TargetPercentile        *float64 `pulumi:"targetPercentile,optional"`
	MaxScaleUpPercent       *float64 `pulumi:"maxScaleUpPercent,optional"`
	MaxScaleDownPercent     *float64 `pulumi:"maxScaleDownPercent,optional"`
	LimitsRemovalEnabled    bool     `pulumi:"limitsRemovalEnabled,optional"`
}

// Annotate provides SDK documentation for ResourceRuleConfigArgs fields.
func (r *ResourceRuleConfigArgs) Annotate(a infer.Annotator) {
	a.Describe(&r.Enabled, "Enable this resource axis rule. Example: true.")
	a.Describe(&r.MinRequest, "Minimum resource request (millicores for CPU, bytes for memory/GPU). Example: 100.")
	a.Describe(&r.MaxRequest, "Maximum resource request (millicores for CPU, bytes for memory/GPU). Example: 4000.")
	a.Describe(&r.LimitMultiplier, "Multiplier applied to the request to derive the resource limit. Example: 1.5.")
	a.Describe(&r.LimitsAdjustmentEnabled, "Whether to also adjust resource limits alongside requests. Example: true.")
	a.Describe(&r.TargetPercentile, "Percentile of usage data used as the recommendation target (0-1). Example: 0.95.")
	a.Describe(&r.MaxScaleUpPercent, "Maximum percentage increase allowed in a single cycle. Example: 50.0.")
	a.Describe(&r.MaxScaleDownPercent, "Maximum percentage decrease allowed in a single cycle. Example: 20.0.")
	a.Describe(&r.LimitsRemovalEnabled, "Actively remove limits from workloads. Example: false.")
}

// HPARuleConfigArgs configures horizontal (replica) scaling on a WorkloadRule.
type HPARuleConfigArgs struct {
	Enabled                 bool     `pulumi:"enabled,optional"`
	MinReplicas             *int     `pulumi:"minReplicas,optional"`
	MaxReplicas             *int     `pulumi:"maxReplicas,optional"`
	TargetUtilization       *float64 `pulumi:"targetUtilization,optional"`
	PrimaryMetric           *string  `pulumi:"primaryMetric,optional"`
	MaxReplicaChangePercent *float64 `pulumi:"maxReplicaChangePercent,optional"`
}

// Annotate provides SDK documentation for HPARuleConfigArgs fields.
func (h *HPARuleConfigArgs) Annotate(a infer.Annotator) {
	a.Describe(&h.Enabled, "Enable horizontal (replica) scaling. Example: true.")
	a.Describe(&h.MinReplicas, "Minimum number of replicas. Example: 2.")
	a.Describe(&h.MaxReplicas, "Maximum number of replicas. Example: 10.")
	a.Describe(&h.TargetUtilization, "Target utilization ratio (0-1) for the primary metric. Example: 0.7.")
	a.Describe(&h.PrimaryMetric, "Primary metric for HPA. One of: 'cpu', 'memory', 'gpu', 'network_ingress', 'network_egress'. Example: 'cpu'.")
	a.Describe(&h.MaxReplicaChangePercent, "Maximum percentage change in replica count per cycle. Example: 50.0.")
}

// EmergencyResponseConfigArgs configures OOM and CPU-throttle emergency reactions.
type EmergencyResponseConfigArgs struct {
	OomEnabled              bool    `pulumi:"oomEnabled,optional"`
	OomMemoryMultiplier     float64 `pulumi:"oomMemoryMultiplier,optional"`
	OomMaxReactions         int     `pulumi:"oomMaxReactions,optional"`
	OomCooldownSeconds      int     `pulumi:"oomCooldownSeconds,optional"`
	CpuThrottlingEnabled    bool    `pulumi:"cpuThrottlingEnabled,optional"`
	CpuThrottlingThreshold  float64 `pulumi:"cpuThrottlingThreshold,optional"`
	CpuThrottlingMultiplier float64 `pulumi:"cpuThrottlingMultiplier,optional"`
}

// Annotate provides SDK documentation for EmergencyResponseConfigArgs fields.
func (e *EmergencyResponseConfigArgs) Annotate(a infer.Annotator) {
	a.Describe(&e.OomEnabled, "React to OOM kills by increasing memory. Example: true.")
	a.Describe(&e.OomMemoryMultiplier, "Multiplier applied to memory on OOM. Example: 2.0.")
	a.Describe(&e.OomMaxReactions, "Maximum number of OOM reactions before giving up. Example: 3.")
	a.Describe(&e.OomCooldownSeconds, "Seconds to wait between OOM reactions. Example: 60.")
	a.Describe(&e.CpuThrottlingEnabled, "React to CPU throttling by increasing CPU request. Example: true.")
	a.Describe(&e.CpuThrottlingThreshold, "Throttle ratio threshold that triggers a reaction (0-1). Example: 0.8.")
	a.Describe(&e.CpuThrottlingMultiplier, "Multiplier applied to CPU request on throttle reaction. Example: 1.5.")
}

// ContainerResourceRuleConfigArgs holds per-container resource rules.
type ContainerResourceRuleConfigArgs struct {
	ContainerName string                  `pulumi:"containerName"`
	CpuRule       *ResourceRuleConfigArgs `pulumi:"cpuRule,optional"`
	MemoryRule    *ResourceRuleConfigArgs `pulumi:"memoryRule,optional"`
	GpuRule       *ResourceRuleConfigArgs `pulumi:"gpuRule,optional"`
}

// Annotate provides SDK documentation for ContainerResourceRuleConfigArgs fields.
func (c *ContainerResourceRuleConfigArgs) Annotate(a infer.Annotator) {
	a.Describe(&c.ContainerName, "Name of the container this config applies to. Example: 'main'.")
	a.Describe(&c.CpuRule, "CPU resource rule for this container.")
	a.Describe(&c.MemoryRule, "Memory resource rule for this container.")
	a.Describe(&c.GpuRule, "GPU resource rule for this container.")
}

// WorkloadRuleArgs are the user-configurable inputs for a WorkloadRule resource.
type WorkloadRuleArgs struct {
	ClusterID                string                            `pulumi:"clusterId"`
	Namespace                string                            `pulumi:"namespace"`
	Kind                     string                            `pulumi:"kind"`
	Name                     string                            `pulumi:"name"`
	AutoGenerate             *bool                             `pulumi:"autoGenerate,optional"`
	CpuRule                  *ResourceRuleConfigArgs           `pulumi:"cpuRule,optional"`
	MemoryRule               *ResourceRuleConfigArgs           `pulumi:"memoryRule,optional"`
	GpuRule                  *ResourceRuleConfigArgs           `pulumi:"gpuRule,optional"`
	HpaRule                  *HPARuleConfigArgs                `pulumi:"hpaRule,optional"`
	EmergencyResponse        *EmergencyResponseConfigArgs      `pulumi:"emergencyResponse,optional"`
	ActionTriggers           []string                          `pulumi:"actionTriggers,optional"`
	StartupPeriodSeconds     *int                              `pulumi:"startupPeriodSeconds,optional"`
	CronSchedule             *string                           `pulumi:"cronSchedule,optional"`
	CooldownMinutes          *int                              `pulumi:"cooldownMinutes,optional"`
	DetectionTriggers        []string                          `pulumi:"detectionTriggers,optional"`
	SchedulerPlugins         []string                          `pulumi:"schedulerPlugins,optional"`
	DefragmentationSchedule  *string                           `pulumi:"defragmentationSchedule,optional"`
	LiveMigrationEnabled     bool                              `pulumi:"liveMigrationEnabled,optional"`
	UseInPlaceVerticalScaling bool                             `pulumi:"useInPlaceVerticalScaling,optional"`
	Containers               []ContainerResourceRuleConfigArgs `pulumi:"containers,optional"`
}

// Annotate provides SDK documentation for WorkloadRuleArgs fields.
func (a *WorkloadRuleArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.ClusterID, "ID of the cluster this rule targets. Example: 'cluster-abc123'.")
	ann.Describe(&a.Namespace, "Kubernetes namespace of the workload. Example: 'production'.")
	ann.Describe(&a.Kind, "Kubernetes workload kind. One of: 'Deployment', 'StatefulSet', 'DaemonSet', 'CronJob', 'Job'. Example: 'Deployment'.")
	ann.Describe(&a.Name, "Name of the Kubernetes workload. Example: 'my-api'.")
	ann.Describe(&a.AutoGenerate, "When true the engine generates all rule fields automatically; manual field overrides are ignored. Example: false.")
	ann.Describe(&a.CpuRule, "CPU vertical scaling rule configuration.")
	ann.Describe(&a.MemoryRule, "Memory vertical scaling rule configuration.")
	ann.Describe(&a.GpuRule, "GPU vertical scaling rule configuration.")
	ann.Describe(&a.HpaRule, "Horizontal (replica) scaling rule configuration.")
	ann.Describe(&a.EmergencyResponse, "Emergency response configuration for OOM and CPU throttle events.")
	ann.Describe(&a.ActionTriggers, "When to apply recommendations. Valid values: 'on_detection', 'on_schedule'. Example: [\"on_detection\"].")
	ann.Describe(&a.StartupPeriodSeconds, "Seconds after workload start to exclude from usage data. Example: 300.")
	ann.Describe(&a.CronSchedule, "Cron expression for scheduled application (5-field UTC). Example: '0 2 * * *'.")
	ann.Describe(&a.CooldownMinutes, "Minimum minutes between consecutive recommendation applications. Example: 60.")
	ann.Describe(&a.DetectionTriggers, "Events that trigger a recommendation. Valid values: 'pod_creation', 'pod_update', 'pod_reschedule'. Example: [\"pod_creation\"].")
	ann.Describe(&a.SchedulerPlugins, "Kubernetes scheduler plugins to activate. Example: [\"binpacking\"].")
	ann.Describe(&a.DefragmentationSchedule, "Cron expression for node defragmentation. Example: '0 3 * * 0'.")
	ann.Describe(&a.LiveMigrationEnabled, "Allow live pod migration when applying recommendations. Example: false.")
	ann.Describe(&a.UseInPlaceVerticalScaling, "Use in-place pod vertical scaling instead of pod restarts. Example: false.")
	ann.Describe(&a.Containers, "Per-container resource rule configurations. When empty, workload-level rules apply to all containers.")
}

// WorkloadRuleState is the full persisted state.
type WorkloadRuleState struct {
	WorkloadRuleArgs
}

// WorkloadRule is the resource implementation.
type WorkloadRule struct{}

// ---------- CRUD ----------

func (w *WorkloadRule) Create(ctx context.Context, req infer.CreateRequest[WorkloadRuleArgs]) (infer.CreateResponse[WorkloadRuleState], error) {
	if err := validateContainerRules(req.Inputs.Containers); err != nil {
		return infer.CreateResponse[WorkloadRuleState]{}, err
	}
	if req.DryRun {
		return infer.CreateResponse[WorkloadRuleState]{Output: WorkloadRuleState{WorkloadRuleArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.CreateResponse[WorkloadRuleState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.UpsertManualWorkloadRule(ctx, connect.NewRequest(ruleArgsToUpsertRequest(cs.TeamID, req.Inputs)))
	if err != nil {
		return infer.CreateResponse[WorkloadRuleState]{}, fmt.Errorf("UpsertManualWorkloadRule: %w", err)
	}
	if resp.Msg.Rule == nil {
		return infer.CreateResponse[WorkloadRuleState]{}, fmt.Errorf("UpsertManualWorkloadRule: empty response from server")
	}

	out := ruleProtoToArgs(resp.Msg.Rule)
	out.AutoGenerate = req.Inputs.AutoGenerate
	return infer.CreateResponse[WorkloadRuleState]{
		ID:     resp.Msg.Rule.RuleId,
		Output: WorkloadRuleState{WorkloadRuleArgs: out},
	}, nil
}

func (w *WorkloadRule) Read(ctx context.Context, req infer.ReadRequest[WorkloadRuleArgs, WorkloadRuleState]) (infer.ReadResponse[WorkloadRuleArgs, WorkloadRuleState], error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.ReadResponse[WorkloadRuleArgs, WorkloadRuleState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.GetWorkloadRuleByID(ctx, connect.NewRequest(&apiv1.GetWorkloadRuleByIDRequest{
		TeamId: cs.TeamID,
		RuleId: req.ID,
	}))
	if err != nil {
		return infer.ReadResponse[WorkloadRuleArgs, WorkloadRuleState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("GetWorkloadRuleByID: %w", err)
	}
	if resp.Msg.Rule == nil {
		return infer.ReadResponse[WorkloadRuleArgs, WorkloadRuleState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("GetWorkloadRuleByID: rule not found")
	}

	updatedArgs := ruleProtoToArgs(resp.Msg.Rule)
	return infer.ReadResponse[WorkloadRuleArgs, WorkloadRuleState]{
		ID:     req.ID,
		Inputs: updatedArgs,
		State:  WorkloadRuleState{WorkloadRuleArgs: updatedArgs},
	}, nil
}

func (w *WorkloadRule) Update(ctx context.Context, req infer.UpdateRequest[WorkloadRuleArgs, WorkloadRuleState]) (infer.UpdateResponse[WorkloadRuleState], error) {
	if err := validateContainerRules(req.Inputs.Containers); err != nil {
		return infer.UpdateResponse[WorkloadRuleState]{}, err
	}
	if req.DryRun {
		return infer.UpdateResponse[WorkloadRuleState]{Output: WorkloadRuleState{WorkloadRuleArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.UpdateResponse[WorkloadRuleState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.UpsertManualWorkloadRule(ctx, connect.NewRequest(ruleArgsToUpsertRequest(cs.TeamID, req.Inputs)))
	if err != nil {
		return infer.UpdateResponse[WorkloadRuleState]{}, fmt.Errorf("UpsertManualWorkloadRule: %w", err)
	}
	if resp.Msg.Rule == nil {
		return infer.UpdateResponse[WorkloadRuleState]{}, fmt.Errorf("UpsertManualWorkloadRule: empty response from server")
	}

	out := ruleProtoToArgs(resp.Msg.Rule)
	out.AutoGenerate = req.Inputs.AutoGenerate
	return infer.UpdateResponse[WorkloadRuleState]{
		Output: WorkloadRuleState{WorkloadRuleArgs: out},
	}, nil
}

func (w *WorkloadRule) Delete(ctx context.Context, req infer.DeleteRequest[WorkloadRuleState]) (infer.DeleteResponse, error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.DeleteResponse{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	_, err := cs.RecommendationClient.DeleteWorkloadRule(ctx, connect.NewRequest(&apiv1.DeleteWorkloadRuleRequest{
		TeamId: cs.TeamID,
		RuleId: req.ID,
	}))
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("DeleteWorkloadRule: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

// ---------- proto conversion helpers ----------

func ruleArgsToUpsertRequest(teamID string, a WorkloadRuleArgs) *apiv1.UpsertManualWorkloadRuleRequest {
	req := &apiv1.UpsertManualWorkloadRuleRequest{
		TeamId:    teamID,
		ClusterId: a.ClusterID,
		Namespace: a.Namespace,
		Kind:      a.Kind,
		Name:      a.Name,
	}
	if a.AutoGenerate != nil && *a.AutoGenerate {
		req.AutoGenerate = true
		return req
	}
	req.Fields = &apiv1.ManualRuleFields{
		CpuRule:                   resourceRuleConfigToProto(a.CpuRule),
		MemoryRule:                resourceRuleConfigToProto(a.MemoryRule),
		GpuRule:                   resourceRuleConfigToProto(a.GpuRule),
		HpaRule:                   hpaRuleConfigToProto(a.HpaRule),
		EmergencyResponse:         emergencyResponseToProto(a.EmergencyResponse),
		ActionTriggers:            actionTriggersToProto(a.ActionTriggers),
		DetectionTriggers:         detectionTriggersToProto(a.DetectionTriggers),
		SchedulerPlugins:          a.SchedulerPlugins,
		LiveMigrationEnabled:      a.LiveMigrationEnabled,
		UseInPlaceVerticalScaling: a.UseInPlaceVerticalScaling,
		Containers:                containerRuleConfigsToProto(a.Containers),
	}
	if a.StartupPeriodSeconds != nil {
		v := int64(*a.StartupPeriodSeconds)
		req.Fields.StartupPeriodSeconds = &v
	}
	if a.CronSchedule != nil {
		req.Fields.CronSchedule = a.CronSchedule
	}
	if a.CooldownMinutes != nil {
		v := int32(*a.CooldownMinutes)
		req.Fields.CooldownMinutes = &v
	}
	if a.DefragmentationSchedule != nil {
		req.Fields.DefragmentationSchedule = a.DefragmentationSchedule
	}
	return req
}

func ruleProtoToArgs(r *apiv1.WorkloadRule) WorkloadRuleArgs {
	a := WorkloadRuleArgs{
		ClusterID:                 r.ClusterId,
		Namespace:                 r.Namespace,
		Kind:                      r.Kind,
		Name:                      r.Name,
		CpuRule:                   resourceRuleConfigFromProto(r.CpuRule),
		MemoryRule:                resourceRuleConfigFromProto(r.MemoryRule),
		GpuRule:                   resourceRuleConfigFromProto(r.GpuRule),
		HpaRule:                   hpaRuleConfigFromProto(r.HpaRule),
		EmergencyResponse:         emergencyResponseFromProto(r.EmergencyResponse),
		ActionTriggers:            actionTriggersFromProto(r.ActionTriggers),
		DetectionTriggers:         detectionTriggersFromProto(r.DetectionTriggers),
		SchedulerPlugins:          r.SchedulerPlugins,
		LiveMigrationEnabled:      r.LiveMigrationEnabled,
		UseInPlaceVerticalScaling: r.UseInPlaceVerticalScaling,
		Containers:                containerRuleConfigsFromProto(r.Containers),
	}
	if r.StartupPeriodSeconds != nil {
		v := int(*r.StartupPeriodSeconds)
		a.StartupPeriodSeconds = &v
	}
	if r.CronSchedule != nil {
		a.CronSchedule = r.CronSchedule
	}
	if r.CooldownMinutes != nil {
		v := int(*r.CooldownMinutes)
		a.CooldownMinutes = &v
	}
	if r.DefragmentationSchedule != nil {
		a.DefragmentationSchedule = r.DefragmentationSchedule
	}
	if r.CurrentSource == "auto_optimization" {
		autoGen := true
		a.AutoGenerate = &autoGen
	}
	return a
}

// validateContainerRules returns an error if any container rule uses fields that the
// ContainerResourceConfig proto does not support (MaxScaleUpPercent, MaxScaleDownPercent).
// Setting these would silently discard the values rather than applying them.
func validateContainerRules(containers []ContainerResourceRuleConfigArgs) error {
	for _, c := range containers {
		for _, r := range []*ResourceRuleConfigArgs{c.CpuRule, c.MemoryRule, c.GpuRule} {
			if r == nil {
				continue
			}
			if r.MaxScaleUpPercent != nil {
				return fmt.Errorf("devzero: maxScaleUpPercent is not supported on container-level rules (container %q) — set it on the workload-level cpuRule/memoryRule/gpuRule instead", c.ContainerName)
			}
			if r.MaxScaleDownPercent != nil {
				return fmt.Errorf("devzero: maxScaleDownPercent is not supported on container-level rules (container %q) — set it on the workload-level cpuRule/memoryRule/gpuRule instead", c.ContainerName)
			}
		}
	}
	return nil
}

func resourceRuleConfigToProto(r *ResourceRuleConfigArgs) *apiv1.ResourceRuleConfig {
	if r == nil {
		return nil
	}
	p := &apiv1.ResourceRuleConfig{
		Enabled:                 r.Enabled,
		LimitsAdjustmentEnabled: r.LimitsAdjustmentEnabled,
		LimitsRemovalEnabled:    r.LimitsRemovalEnabled,
	}
	if r.MinRequest != nil {
		v := int64(*r.MinRequest)
		p.MinRequest = &v
	}
	if r.MaxRequest != nil {
		v := int64(*r.MaxRequest)
		p.MaxRequest = &v
	}
	if r.LimitMultiplier != nil {
		v := float32(*r.LimitMultiplier)
		p.LimitMultiplier = &v
	}
	if r.TargetPercentile != nil {
		v := float32(*r.TargetPercentile)
		p.TargetPercentile = &v
	}
	if r.MaxScaleUpPercent != nil {
		v := float32(*r.MaxScaleUpPercent)
		p.MaxScaleUpPercent = &v
	}
	if r.MaxScaleDownPercent != nil {
		v := float32(*r.MaxScaleDownPercent)
		p.MaxScaleDownPercent = &v
	}
	return p
}

func resourceRuleConfigFromProto(p *apiv1.ResourceRuleConfig) *ResourceRuleConfigArgs {
	if p == nil {
		return nil
	}
	r := &ResourceRuleConfigArgs{
		Enabled:                 p.Enabled,
		LimitsAdjustmentEnabled: p.LimitsAdjustmentEnabled,
		LimitsRemovalEnabled:    p.LimitsRemovalEnabled,
	}
	if p.MinRequest != nil {
		v := int(*p.MinRequest)
		r.MinRequest = &v
	}
	if p.MaxRequest != nil {
		v := int(*p.MaxRequest)
		r.MaxRequest = &v
	}
	if p.LimitMultiplier != nil {
		v := float64(*p.LimitMultiplier)
		r.LimitMultiplier = &v
	}
	if p.TargetPercentile != nil {
		v := float64(*p.TargetPercentile)
		r.TargetPercentile = &v
	}
	if p.MaxScaleUpPercent != nil {
		v := float64(*p.MaxScaleUpPercent)
		r.MaxScaleUpPercent = &v
	}
	if p.MaxScaleDownPercent != nil {
		v := float64(*p.MaxScaleDownPercent)
		r.MaxScaleDownPercent = &v
	}
	return r
}

func hpaRuleConfigToProto(h *HPARuleConfigArgs) *apiv1.HPARuleConfig {
	if h == nil {
		return nil
	}
	p := &apiv1.HPARuleConfig{
		Enabled: h.Enabled,
	}
	if h.MinReplicas != nil {
		v := int32(*h.MinReplicas)
		p.MinReplicas = &v
	}
	if h.MaxReplicas != nil {
		v := int32(*h.MaxReplicas)
		p.MaxReplicas = &v
	}
	if h.TargetUtilization != nil {
		v := float32(*h.TargetUtilization)
		p.TargetUtilization = &v
	}
	if h.PrimaryMetric != nil {
		p.PrimaryMetric = hpaMetricToProto(h.PrimaryMetric)
	}
	if h.MaxReplicaChangePercent != nil {
		v := float32(*h.MaxReplicaChangePercent)
		p.MaxReplicaChangePercent = &v
	}
	return p
}

func hpaRuleConfigFromProto(p *apiv1.HPARuleConfig) *HPARuleConfigArgs {
	if p == nil {
		return nil
	}
	h := &HPARuleConfigArgs{
		Enabled: p.Enabled,
	}
	if p.MinReplicas != nil {
		v := int(*p.MinReplicas)
		h.MinReplicas = &v
	}
	if p.MaxReplicas != nil {
		v := int(*p.MaxReplicas)
		h.MaxReplicas = &v
	}
	if p.TargetUtilization != nil {
		v := float64(*p.TargetUtilization)
		h.TargetUtilization = &v
	}
	if p.PrimaryMetric != nil {
		h.PrimaryMetric = hpaMetricFromProto(p.PrimaryMetric)
	}
	if p.MaxReplicaChangePercent != nil {
		v := float64(*p.MaxReplicaChangePercent)
		h.MaxReplicaChangePercent = &v
	}
	return h
}

func emergencyResponseToProto(e *EmergencyResponseConfigArgs) *apiv1.EmergencyResponseConfig {
	if e == nil {
		return nil
	}
	return &apiv1.EmergencyResponseConfig{
		OomEnabled:              e.OomEnabled,
		OomMemoryMultiplier:     float32(e.OomMemoryMultiplier),
		OomMaxReactions:         int32(e.OomMaxReactions),
		OomCooldownSeconds:      int32(e.OomCooldownSeconds),
		CpuThrottlingEnabled:    e.CpuThrottlingEnabled,
		CpuThrottlingThreshold:  float32(e.CpuThrottlingThreshold),
		CpuThrottlingMultiplier: float32(e.CpuThrottlingMultiplier),
	}
}

func emergencyResponseFromProto(p *apiv1.EmergencyResponseConfig) *EmergencyResponseConfigArgs {
	if p == nil {
		return nil
	}
	return &EmergencyResponseConfigArgs{
		OomEnabled:              p.OomEnabled,
		OomMemoryMultiplier:     float64(p.OomMemoryMultiplier),
		OomMaxReactions:         int(p.OomMaxReactions),
		OomCooldownSeconds:      int(p.OomCooldownSeconds),
		CpuThrottlingEnabled:    p.CpuThrottlingEnabled,
		CpuThrottlingThreshold:  float64(p.CpuThrottlingThreshold),
		CpuThrottlingMultiplier: float64(p.CpuThrottlingMultiplier),
	}
}

func containerRuleConfigsToProto(cs []ContainerResourceRuleConfigArgs) []*apiv1.ContainerResourceRuleConfig {
	if len(cs) == 0 {
		return nil
	}
	result := make([]*apiv1.ContainerResourceRuleConfig, len(cs))
	for i, c := range cs {
		result[i] = &apiv1.ContainerResourceRuleConfig{
			ContainerName: c.ContainerName,
			CpuRule:       containerResourceConfigToProto(c.CpuRule),
			MemoryRule:    containerResourceConfigToProto(c.MemoryRule),
			GpuRule:       containerResourceConfigToProto(c.GpuRule),
		}
	}
	return result
}

func containerRuleConfigsFromProto(ps []*apiv1.ContainerResourceRuleConfig) []ContainerResourceRuleConfigArgs {
	if len(ps) == 0 {
		return nil
	}
	result := make([]ContainerResourceRuleConfigArgs, len(ps))
	for i, p := range ps {
		result[i] = ContainerResourceRuleConfigArgs{
			ContainerName: p.ContainerName,
			CpuRule:       containerResourceConfigFromProto(p.CpuRule),
			MemoryRule:    containerResourceConfigFromProto(p.MemoryRule),
			GpuRule:       containerResourceConfigFromProto(p.GpuRule),
		}
	}
	return result
}

// containerResourceConfigToProto maps ResourceRuleConfigArgs to the proto ContainerResourceConfig
// (a subset type used inside ContainerResourceRuleConfig).
func containerResourceConfigToProto(r *ResourceRuleConfigArgs) *apiv1.ContainerResourceConfig {
	if r == nil {
		return nil
	}
	p := &apiv1.ContainerResourceConfig{
		Enabled:                 r.Enabled,
		LimitsAdjustmentEnabled: r.LimitsAdjustmentEnabled,
		LimitsRemovalEnabled:    r.LimitsRemovalEnabled,
	}
	if r.MinRequest != nil {
		v := int64(*r.MinRequest)
		p.MinRequest = &v
	}
	if r.MaxRequest != nil {
		v := int64(*r.MaxRequest)
		p.MaxRequest = &v
	}
	if r.LimitMultiplier != nil {
		v := float32(*r.LimitMultiplier)
		p.LimitMultiplier = &v
	}
	if r.TargetPercentile != nil {
		v := float32(*r.TargetPercentile)
		p.TargetPercentile = &v
	}
	return p
}

func containerResourceConfigFromProto(p *apiv1.ContainerResourceConfig) *ResourceRuleConfigArgs {
	if p == nil {
		return nil
	}
	r := &ResourceRuleConfigArgs{
		Enabled:                 p.Enabled,
		LimitsAdjustmentEnabled: p.LimitsAdjustmentEnabled,
		LimitsRemovalEnabled:    p.LimitsRemovalEnabled,
	}
	if p.MinRequest != nil {
		v := int(*p.MinRequest)
		r.MinRequest = &v
	}
	if p.MaxRequest != nil {
		v := int(*p.MaxRequest)
		r.MaxRequest = &v
	}
	if p.LimitMultiplier != nil {
		v := float64(*p.LimitMultiplier)
		r.LimitMultiplier = &v
	}
	if p.TargetPercentile != nil {
		v := float64(*p.TargetPercentile)
		r.TargetPercentile = &v
	}
	return r
}
