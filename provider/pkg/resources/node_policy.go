package resources

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// TaintArgs represents a Kubernetes node taint.
type TaintArgs struct {
	Key    string `pulumi:"key"`
	Value  string `pulumi:"value,optional"`
	Effect string `pulumi:"effect"`
}

// DisruptionBudgetArgs represents a single disruption budget entry.
type DisruptionBudgetArgs struct {
	Reasons  []string `pulumi:"reasons,optional"`
	Nodes    string   `pulumi:"nodes,optional"`
	Schedule string   `pulumi:"schedule,optional"`
	Duration string   `pulumi:"duration,optional"`
}

// DisruptionPolicyArgs configures how Karpenter disrupts nodes.
type DisruptionPolicyArgs struct {
	ConsolidateAfter              string                 `pulumi:"consolidateAfter,optional"`
	ConsolidationPolicy           string                 `pulumi:"consolidationPolicy,optional"`
	ExpireAfter                   string                 `pulumi:"expireAfter,optional"`
	TtlSecondsAfterEmpty          int                    `pulumi:"ttlSecondsAfterEmpty,optional"`
	TerminationGracePeriodSeconds int                    `pulumi:"terminationGracePeriodSeconds,optional"`
	Budgets                       []DisruptionBudgetArgs `pulumi:"budgets,optional"`
}

// ResourceLimitsArgs sets resource limits on the node pool.
type ResourceLimitsArgs struct {
	Cpu    string `pulumi:"cpu,optional"`
	Memory string `pulumi:"memory,optional"`
}

// SubnetSelectorTermArgs selects subnets by tag or ID.
type SubnetSelectorTermArgs struct {
	Tags map[string]string `pulumi:"tags,optional"`
	Id   string            `pulumi:"id,optional"`
}

// SecurityGroupSelectorTermArgs selects security groups.
type SecurityGroupSelectorTermArgs struct {
	Tags map[string]string `pulumi:"tags,optional"`
	Id   string            `pulumi:"id,optional"`
	Name string            `pulumi:"name,optional"`
}

// CapacityReservationSelectorTermArgs selects capacity reservations.
type CapacityReservationSelectorTermArgs struct {
	Tags    map[string]string `pulumi:"tags,optional"`
	Id      string            `pulumi:"id,optional"`
	OwnerId string            `pulumi:"ownerId,optional"`
}

// AMISelectorTermArgs selects AMIs for AWS nodes.
type AMISelectorTermArgs struct {
	Alias        string            `pulumi:"alias,optional"`
	Tags         map[string]string `pulumi:"tags,optional"`
	Id           string            `pulumi:"id,optional"`
	Name         string            `pulumi:"name,optional"`
	Owner        string            `pulumi:"owner,optional"`
	SsmParameter string            `pulumi:"ssmParameter,optional"`
}

// BlockDeviceArgs configures an EBS block device.
type BlockDeviceArgs struct {
	DeleteOnTermination      *bool   `pulumi:"deleteOnTermination,optional"`
	Encrypted                *bool   `pulumi:"encrypted,optional"`
	Iops                     *int    `pulumi:"iops,optional"`
	KmsKeyId                 *string `pulumi:"kmsKeyId,optional"`
	SnapshotId               *string `pulumi:"snapshotId,optional"`
	Throughput               *int    `pulumi:"throughput,optional"`
	VolumeInitializationRate *int    `pulumi:"volumeInitializationRate,optional"`
	VolumeSize               *string `pulumi:"volumeSize,optional"`
	VolumeType               *string `pulumi:"volumeType,optional"`
}

// BlockDeviceMappingArgs maps an EBS block device to a device name.
type BlockDeviceMappingArgs struct {
	DeviceName *string          `pulumi:"deviceName,optional"`
	Ebs        *BlockDeviceArgs `pulumi:"ebs,optional"`
	RootVolume *bool            `pulumi:"rootVolume,optional"`
}

// KubeletConfigurationArgs configures kubelet on AWS nodes.
type KubeletConfigurationArgs struct {
	ClusterDns                  []string          `pulumi:"clusterDns,optional"`
	MaxPods                     *int              `pulumi:"maxPods,optional"`
	PodsPerCore                 *int              `pulumi:"podsPerCore,optional"`
	SystemReserved              map[string]string `pulumi:"systemReserved,optional"`
	KubeReserved                map[string]string `pulumi:"kubeReserved,optional"`
	EvictionHard                map[string]string `pulumi:"evictionHard,optional"`
	EvictionSoft                map[string]string `pulumi:"evictionSoft,optional"`
	EvictionSoftGracePeriod     map[string]string `pulumi:"evictionSoftGracePeriod,optional"`
	EvictionMaxPodGracePeriod   *int              `pulumi:"evictionMaxPodGracePeriod,optional"`
	ImageGcHighThresholdPercent *int              `pulumi:"imageGcHighThresholdPercent,optional"`
	ImageGcLowThresholdPercent  *int              `pulumi:"imageGcLowThresholdPercent,optional"`
	CpuCfsQuota                 *bool             `pulumi:"cpuCfsQuota,optional"`
}

// MetadataOptionsArgs configures EC2 instance metadata options.
type MetadataOptionsArgs struct {
	HttpEndpoint            string `pulumi:"httpEndpoint,optional"`
	HttpProtocolIpv6        string `pulumi:"httpProtocolIpv6,optional"`
	HttpPutResponseHopLimit int    `pulumi:"httpPutResponseHopLimit,optional"`
	HttpTokens              string `pulumi:"httpTokens,optional"`
}

// AWSNodeClassSpecArgs holds AWS-specific node class configuration.
type AWSNodeClassSpecArgs struct {
	SubnetSelectorTerms              []SubnetSelectorTermArgs              `pulumi:"subnetSelectorTerms,optional"`
	SecurityGroupSelectorTerms       []SecurityGroupSelectorTermArgs       `pulumi:"securityGroupSelectorTerms,optional"`
	CapacityReservationSelectorTerms []CapacityReservationSelectorTermArgs `pulumi:"capacityReservationSelectorTerms,optional"`
	AssociatePublicIpAddress         *bool                                 `pulumi:"associatePublicIpAddress,optional"`
	AmiSelectorTerms                 []AMISelectorTermArgs                 `pulumi:"amiSelectorTerms,optional"`
	AmiFamily                        *string                               `pulumi:"amiFamily,optional"`
	UserData                         *string                               `pulumi:"userData,optional"`
	Role                             *string                               `pulumi:"role,optional"`
	InstanceProfile                  *string                               `pulumi:"instanceProfile,optional"`
	Tags                             map[string]string                     `pulumi:"tags,optional"`
	Kubelet                          *KubeletConfigurationArgs             `pulumi:"kubelet,optional"`
	BlockDeviceMappings              []BlockDeviceMappingArgs              `pulumi:"blockDeviceMappings,optional"`
	InstanceStorePolicy              *string                               `pulumi:"instanceStorePolicy,optional"`
	DetailedMonitoring               *bool                                 `pulumi:"detailedMonitoring,optional"`
	MetadataOptions                  *MetadataOptionsArgs                  `pulumi:"metadataOptions,optional"`
	Context                          *string                               `pulumi:"context,optional"`
}

// AzureKubeletConfigurationArgs configures kubelet on Azure nodes.
type AzureKubeletConfigurationArgs struct {
	CpuManagerPolicy            *string  `pulumi:"cpuManagerPolicy,optional"`
	CpuCfsQuota                 *bool    `pulumi:"cpuCfsQuota,optional"`
	CpuCfsQuotaPeriod           *string  `pulumi:"cpuCfsQuotaPeriod,optional"`
	ImageGcHighThresholdPercent *int     `pulumi:"imageGcHighThresholdPercent,optional"`
	ImageGcLowThresholdPercent  *int     `pulumi:"imageGcLowThresholdPercent,optional"`
	TopologyManagerPolicy       *string  `pulumi:"topologyManagerPolicy,optional"`
	AllowedUnsafeSysctls        []string `pulumi:"allowedUnsafeSysctls,optional"`
	ContainerLogMaxSize         *string  `pulumi:"containerLogMaxSize,optional"`
	ContainerLogMaxFiles        *int     `pulumi:"containerLogMaxFiles,optional"`
	PodPidsLimit                *int     `pulumi:"podPidsLimit,optional"`
}

// AzureNodeClassSpecArgs holds Azure-specific node class configuration.
type AzureNodeClassSpecArgs struct {
	VnetSubnetId string                         `pulumi:"vnetSubnetId,optional"`
	OsDiskSizeGb *int                           `pulumi:"osDiskSizeGb,optional"`
	ImageFamily  *string                        `pulumi:"imageFamily,optional"`
	FipsMode     *string                        `pulumi:"fipsMode,optional"`
	Tags         map[string]string              `pulumi:"tags,optional"`
	Kubelet      *AzureKubeletConfigurationArgs `pulumi:"kubelet,optional"`
	MaxPods      *int                           `pulumi:"maxPods,optional"`
}

// RawKarpenterSpecArgs provides raw YAML for a custom Karpenter node pool / node class.
type RawKarpenterSpecArgs struct {
	NodepoolYaml  string `pulumi:"nodepoolYaml,optional"`
	NodeclassYaml string `pulumi:"nodeclassYaml,optional"`
}

// NodePolicyArgs are the user-configurable inputs for a NodePolicy resource.
type NodePolicyArgs struct {
	Name        string  `pulumi:"name"`
	Description *string `pulumi:"description,optional"`
	Weight      int     `pulumi:"weight,optional"`

	// Instance selectors
	InstanceCategories  *LabelSelectorArgs `pulumi:"instanceCategories,optional"`
	InstanceFamilies    *LabelSelectorArgs `pulumi:"instanceFamilies,optional"`
	InstanceCpus        *LabelSelectorArgs `pulumi:"instanceCpus,optional"`
	InstanceHypervisors *LabelSelectorArgs `pulumi:"instanceHypervisors,optional"`
	InstanceGenerations *LabelSelectorArgs `pulumi:"instanceGenerations,optional"`
	InstanceSizes       *LabelSelectorArgs `pulumi:"instanceSizes,optional"`
	InstanceTypes       *LabelSelectorArgs `pulumi:"instanceTypes,optional"`

	// Node attribute selectors
	Zones            *LabelSelectorArgs `pulumi:"zones,optional"`
	Architectures    *LabelSelectorArgs `pulumi:"architectures,optional"`
	CapacityTypes    *LabelSelectorArgs `pulumi:"capacityTypes,optional"`
	OperatingSystems *LabelSelectorArgs `pulumi:"operatingSystems,optional"`

	// Node metadata
	Labels map[string]string `pulumi:"labels,optional"`
	Taints []TaintArgs       `pulumi:"taints,optional"`

	// Policy configuration
	Disruption *DisruptionPolicyArgs `pulumi:"disruption,optional"`
	Limits     *ResourceLimitsArgs   `pulumi:"limits,optional"`

	// Karpenter node pool / class override names
	NodePoolName  string `pulumi:"nodePoolName,optional"`
	NodeClassName string `pulumi:"nodeClassName,optional"`

	// Cloud-specific configuration
	Aws   *AWSNodeClassSpecArgs   `pulumi:"aws,optional"`
	Azure *AzureNodeClassSpecArgs `pulumi:"azure,optional"`

	// Raw Karpenter YAML (escape hatch for full customization)
	Raw []RawKarpenterSpecArgs `pulumi:"raw,optional"`
}

// NodePolicyState is the full persisted state (identical to args — no additional computed fields).
type NodePolicyState struct {
	NodePolicyArgs
}

// Annotate provides SDK documentation for NodePolicy fields.
func (s *NodePolicyState) Annotate(a infer.Annotator) {
	a.Describe(&s.Name, "Human-friendly name for the node policy.")
	a.Describe(&s.Description, "Free-form description of the node policy.")
	a.Describe(&s.Weight, "Priority weight for this policy; higher values take precedence.")
	a.Describe(&s.InstanceCategories, "Filter instances by category (e.g. general-purpose, compute-optimized).")
	a.Describe(&s.InstanceFamilies, "Filter instances by family (e.g. m5, c6i).")
	a.Describe(&s.InstanceCpus, "Filter instances by CPU count.")
	a.Describe(&s.InstanceHypervisors, "Filter instances by hypervisor type.")
	a.Describe(&s.InstanceGenerations, "Filter instances by generation.")
	a.Describe(&s.InstanceSizes, "Filter instances by size (e.g. large, xlarge).")
	a.Describe(&s.InstanceTypes, "Explicitly select specific instance types.")
	a.Describe(&s.Zones, "Availability zones where nodes may be provisioned.")
	a.Describe(&s.Architectures, "CPU architectures (e.g. amd64, arm64).")
	a.Describe(&s.CapacityTypes, "Capacity types (e.g. on-demand, spot).")
	a.Describe(&s.OperatingSystems, "Operating systems for nodes (e.g. linux, windows).")
	a.Describe(&s.Labels, "Labels applied to provisioned nodes.")
	a.Describe(&s.Taints, "Taints applied to provisioned nodes.")
	a.Describe(&s.Disruption, "Karpenter disruption policy for consolidation and expiry.")
	a.Describe(&s.Limits, "Resource limits on the total capacity managed by this policy.")
	a.Describe(&s.NodePoolName, "Override name for the generated Karpenter NodePool.")
	a.Describe(&s.NodeClassName, "Override name for the generated Karpenter NodeClass.")
	a.Describe(&s.Aws, "AWS-specific node class configuration.")
	a.Describe(&s.Azure, "Azure-specific node class configuration.")
	a.Describe(&s.Raw, "Raw Karpenter YAML for full NodePool/NodeClass customization.")
}

// NodePolicy is the resource implementation.
type NodePolicy struct{}

// ---------- CRUD ----------

func (n *NodePolicy) Create(ctx context.Context, req infer.CreateRequest[NodePolicyArgs]) (infer.CreateResponse[NodePolicyState], error) {
	if req.DryRun {
		return infer.CreateResponse[NodePolicyState]{Output: NodePolicyState{NodePolicyArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.CreateResponse[NodePolicyState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.CreateNodePolicies(ctx, connect.NewRequest(&apiv1.CreateNodePoliciesRequest{
		TeamId:   cs.TeamID,
		Policies: []*apiv1.NodePolicy{nodePolicyArgsToProto(cs.TeamID, "", req.Inputs)},
	}))
	if err != nil {
		return infer.CreateResponse[NodePolicyState]{}, fmt.Errorf("CreateNodePolicies: %w", err)
	}
	if len(resp.Msg.Policies) == 0 {
		return infer.CreateResponse[NodePolicyState]{}, fmt.Errorf("CreateNodePolicies: empty response from server")
	}

	created := resp.Msg.Policies[0]
	return infer.CreateResponse[NodePolicyState]{
		ID:     created.Id,
		Output: NodePolicyState{NodePolicyArgs: nodePolicyProtoToArgs(created)},
	}, nil
}

func (n *NodePolicy) Read(ctx context.Context, req infer.ReadRequest[NodePolicyArgs, NodePolicyState]) (infer.ReadResponse[NodePolicyArgs, NodePolicyState], error) {
	cs := clientset.Get()
	if cs == nil {
		return infer.ReadResponse[NodePolicyArgs, NodePolicyState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.ListNodePolicies(ctx, connect.NewRequest(&apiv1.ListNodePoliciesRequest{
		TeamId: cs.TeamID,
	}))
	if err != nil {
		return infer.ReadResponse[NodePolicyArgs, NodePolicyState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
			fmt.Errorf("ListNodePolicies: %w", err)
	}

	for _, p := range resp.Msg.Policies {
		if p.Id == req.ID {
			updatedArgs := nodePolicyProtoToArgs(p)
			return infer.ReadResponse[NodePolicyArgs, NodePolicyState]{
				ID:     req.ID,
				Inputs: updatedArgs,
				State:  NodePolicyState{NodePolicyArgs: updatedArgs},
			}, nil
		}
	}

	return infer.ReadResponse[NodePolicyArgs, NodePolicyState]{ID: req.ID, Inputs: req.Inputs, State: req.State},
		fmt.Errorf("ListNodePolicies: policy %q not found", req.ID)
}

func (n *NodePolicy) Update(ctx context.Context, req infer.UpdateRequest[NodePolicyArgs, NodePolicyState]) (infer.UpdateResponse[NodePolicyState], error) {
	if req.DryRun {
		return infer.UpdateResponse[NodePolicyState]{Output: NodePolicyState{NodePolicyArgs: req.Inputs}}, nil
	}

	cs := clientset.Get()
	if cs == nil {
		return infer.UpdateResponse[NodePolicyState]{}, fmt.Errorf("devzero: provider not configured (ClientSet is nil)")
	}

	resp, err := cs.RecommendationClient.UpdateNodePolicy(ctx, connect.NewRequest(&apiv1.UpdateNodePolicyRequest{
		TeamId: cs.TeamID,
		Policy: nodePolicyArgsToProto(cs.TeamID, req.ID, req.Inputs),
	}))
	if err != nil {
		return infer.UpdateResponse[NodePolicyState]{}, fmt.Errorf("UpdateNodePolicy: %w", err)
	}
	if resp.Msg.Policy == nil {
		return infer.UpdateResponse[NodePolicyState]{}, fmt.Errorf("UpdateNodePolicy: empty response from server")
	}

	return infer.UpdateResponse[NodePolicyState]{
		Output: NodePolicyState{NodePolicyArgs: nodePolicyProtoToArgs(resp.Msg.Policy)},
	}, nil
}

// Delete removes the resource from Pulumi state only — no delete endpoint exists for NodePolicy.
func (n *NodePolicy) Delete(_ context.Context, _ infer.DeleteRequest[NodePolicyState]) (infer.DeleteResponse, error) {
	return infer.DeleteResponse{}, nil
}

// ---------- proto conversion ----------

func nodePolicyArgsToProto(teamID, id string, a NodePolicyArgs) *apiv1.NodePolicy {
	p := &apiv1.NodePolicy{
		Id:                  id,
		TeamId:              teamID,
		Name:                a.Name,
		Weight:              int32(a.Weight),
		Labels:              a.Labels,
		NodePoolName:        a.NodePoolName,
		NodeClassName:       a.NodeClassName,
		InstanceCategories:  labelSelectorToProto(a.InstanceCategories),
		InstanceFamilies:    labelSelectorToProto(a.InstanceFamilies),
		InstanceCpus:        labelSelectorToProto(a.InstanceCpus),
		InstanceHypervisors: labelSelectorToProto(a.InstanceHypervisors),
		InstanceGenerations: labelSelectorToProto(a.InstanceGenerations),
		InstanceSizes:       labelSelectorToProto(a.InstanceSizes),
		InstanceTypes:       labelSelectorToProto(a.InstanceTypes),
		Zones:               labelSelectorToProto(a.Zones),
		Architectures:       labelSelectorToProto(a.Architectures),
		CapacityTypes:       labelSelectorToProto(a.CapacityTypes),
		OperatingSystems:    labelSelectorToProto(a.OperatingSystems),
		Taints:              taintsToProto(a.Taints),
		Disruption:          disruptionPolicyToProto(a.Disruption),
		Limits:              resourceLimitsToProto(a.Limits),
		Aws:                 awsNodeClassSpecToProto(a.Aws),
		Azure:               azureNodeClassSpecToProto(a.Azure),
		Raw:                 rawKarpenterSpecsToProto(a.Raw),
	}
	if a.Description != nil {
		p.Description = *a.Description
	}
	return p
}

func nodePolicyProtoToArgs(p *apiv1.NodePolicy) NodePolicyArgs {
	a := NodePolicyArgs{
		Name:                p.Name,
		Weight:              int(p.Weight),
		Labels:              p.Labels,
		NodePoolName:        p.NodePoolName,
		NodeClassName:       p.NodeClassName,
		InstanceCategories:  labelSelectorFromProto(p.InstanceCategories),
		InstanceFamilies:    labelSelectorFromProto(p.InstanceFamilies),
		InstanceCpus:        labelSelectorFromProto(p.InstanceCpus),
		InstanceHypervisors: labelSelectorFromProto(p.InstanceHypervisors),
		InstanceGenerations: labelSelectorFromProto(p.InstanceGenerations),
		InstanceSizes:       labelSelectorFromProto(p.InstanceSizes),
		InstanceTypes:       labelSelectorFromProto(p.InstanceTypes),
		Zones:               labelSelectorFromProto(p.Zones),
		Architectures:       labelSelectorFromProto(p.Architectures),
		CapacityTypes:       labelSelectorFromProto(p.CapacityTypes),
		OperatingSystems:    labelSelectorFromProto(p.OperatingSystems),
		Taints:              taintsFromProto(p.Taints),
		Disruption:          disruptionPolicyFromProto(p.Disruption),
		Limits:              resourceLimitsFromProto(p.Limits),
		Aws:                 awsNodeClassSpecFromProto(p.Aws),
		Azure:               azureNodeClassSpecFromProto(p.Azure),
		Raw:                 rawKarpenterSpecsFromProto(p.Raw),
	}
	if p.Description != "" {
		a.Description = &p.Description
	}
	return a
}

// ---------- taint helpers ----------

func taintsToProto(taints []TaintArgs) []*apiv1.Taint {
	if len(taints) == 0 {
		return nil
	}
	result := make([]*apiv1.Taint, len(taints))
	for i, t := range taints {
		result[i] = &apiv1.Taint{Key: t.Key, Value: t.Value, Effect: t.Effect}
	}
	return result
}

func taintsFromProto(taints []*apiv1.Taint) []TaintArgs {
	if len(taints) == 0 {
		return nil
	}
	result := make([]TaintArgs, len(taints))
	for i, t := range taints {
		result[i] = TaintArgs{Key: t.Key, Value: t.Value, Effect: t.Effect}
	}
	return result
}

// ---------- disruption helpers ----------

func disruptionPolicyToProto(d *DisruptionPolicyArgs) *apiv1.DisruptionPolicy {
	if d == nil {
		return nil
	}
	p := &apiv1.DisruptionPolicy{
		ConsolidateAfter:              d.ConsolidateAfter,
		ConsolidationPolicy:           d.ConsolidationPolicy,
		ExpireAfter:                   d.ExpireAfter,
		TtlSecondsAfterEmpty:          int32(d.TtlSecondsAfterEmpty),
		TerminationGracePeriodSeconds: int32(d.TerminationGracePeriodSeconds),
	}
	for _, b := range d.Budgets {
		p.Budgets = append(p.Budgets, &apiv1.DisruptionBudget{
			Reasons:  b.Reasons,
			Nodes:    b.Nodes,
			Schedule: b.Schedule,
			Duration: b.Duration,
		})
	}
	return p
}

func disruptionPolicyFromProto(p *apiv1.DisruptionPolicy) *DisruptionPolicyArgs {
	if p == nil {
		return nil
	}
	d := &DisruptionPolicyArgs{
		ConsolidateAfter:              p.ConsolidateAfter,
		ConsolidationPolicy:           p.ConsolidationPolicy,
		ExpireAfter:                   p.ExpireAfter,
		TtlSecondsAfterEmpty:          int(p.TtlSecondsAfterEmpty),
		TerminationGracePeriodSeconds: int(p.TerminationGracePeriodSeconds),
	}
	for _, b := range p.Budgets {
		d.Budgets = append(d.Budgets, DisruptionBudgetArgs{
			Reasons:  b.Reasons,
			Nodes:    b.Nodes,
			Schedule: b.Schedule,
			Duration: b.Duration,
		})
	}
	return d
}

// ---------- resource limits helpers ----------

func resourceLimitsToProto(l *ResourceLimitsArgs) *apiv1.ResourceLimits {
	if l == nil {
		return nil
	}
	return &apiv1.ResourceLimits{Cpu: l.Cpu, Memory: l.Memory}
}

func resourceLimitsFromProto(l *apiv1.ResourceLimits) *ResourceLimitsArgs {
	if l == nil {
		return nil
	}
	return &ResourceLimitsArgs{Cpu: l.Cpu, Memory: l.Memory}
}

// ---------- raw Karpenter helpers ----------

func rawKarpenterSpecsToProto(raw []RawKarpenterSpecArgs) []*apiv1.RawKarpenterSpec {
	if len(raw) == 0 {
		return nil
	}
	result := make([]*apiv1.RawKarpenterSpec, len(raw))
	for i, r := range raw {
		result[i] = &apiv1.RawKarpenterSpec{NodepoolYaml: r.NodepoolYaml, NodeclassYaml: r.NodeclassYaml}
	}
	return result
}

func rawKarpenterSpecsFromProto(raw []*apiv1.RawKarpenterSpec) []RawKarpenterSpecArgs {
	if len(raw) == 0 {
		return nil
	}
	result := make([]RawKarpenterSpecArgs, len(raw))
	for i, r := range raw {
		result[i] = RawKarpenterSpecArgs{NodepoolYaml: r.NodepoolYaml, NodeclassYaml: r.NodeclassYaml}
	}
	return result
}

// ---------- AWS NodeClass helpers ----------

func awsNodeClassSpecToProto(a *AWSNodeClassSpecArgs) *apiv1.AWSNodeClassSpec {
	if a == nil {
		return nil
	}
	spec := &apiv1.AWSNodeClassSpec{
		Tags:            a.Tags,
		Kubelet:         kubeletConfigToProto(a.Kubelet),
		MetadataOptions: metadataOptionsToProto(a.MetadataOptions),
	}
	if a.AssociatePublicIpAddress != nil {
		spec.AssociatePublicIpAddress = a.AssociatePublicIpAddress
	}
	if a.AmiFamily != nil {
		spec.AmiFamily = a.AmiFamily
	}
	if a.UserData != nil {
		spec.UserData = a.UserData
	}
	if a.Role != nil {
		spec.Role = a.Role
	}
	if a.InstanceProfile != nil {
		spec.InstanceProfile = a.InstanceProfile
	}
	if a.Context != nil {
		spec.Context = a.Context
	}
	if a.DetailedMonitoring != nil {
		spec.DetailedMonitoring = a.DetailedMonitoring
	}
	if a.InstanceStorePolicy != nil {
		v := apiv1.InstanceStorePolicy(apiv1.InstanceStorePolicy_value[*a.InstanceStorePolicy])
		spec.InstanceStorePolicy = &v
	}
	for _, t := range a.SubnetSelectorTerms {
		spec.SubnetSelectorTerms = append(spec.SubnetSelectorTerms, &apiv1.SubnetSelectorTerm{
			Tags: t.Tags,
			Id:   t.Id,
		})
	}
	for _, sg := range a.SecurityGroupSelectorTerms {
		spec.SecurityGroupSelectorTerms = append(spec.SecurityGroupSelectorTerms, &apiv1.SecurityGroupSelectorTerm{
			Tags: sg.Tags,
			Id:   sg.Id,
			Name: sg.Name,
		})
	}
	for _, cr := range a.CapacityReservationSelectorTerms {
		spec.CapacityReservationSelectorTerms = append(spec.CapacityReservationSelectorTerms, &apiv1.CapacityReservationSelectorTerm{
			Tags:    cr.Tags,
			Id:      cr.Id,
			OwnerId: cr.OwnerId,
		})
	}
	for _, ami := range a.AmiSelectorTerms {
		spec.AmiSelectorTerms = append(spec.AmiSelectorTerms, &apiv1.AMISelectorTerm{
			Alias:        ami.Alias,
			Tags:         ami.Tags,
			Id:           ami.Id,
			Name:         ami.Name,
			Owner:        ami.Owner,
			SsmParameter: ami.SsmParameter,
		})
	}
	for _, bdm := range a.BlockDeviceMappings {
		spec.BlockDeviceMappings = append(spec.BlockDeviceMappings, blockDeviceMappingToProto(bdm))
	}
	return spec
}

func awsNodeClassSpecFromProto(spec *apiv1.AWSNodeClassSpec) *AWSNodeClassSpecArgs {
	if spec == nil {
		return nil
	}
	a := &AWSNodeClassSpecArgs{
		Tags:            spec.Tags,
		Kubelet:         kubeletConfigFromProto(spec.Kubelet),
		MetadataOptions: metadataOptionsFromProto(spec.MetadataOptions),
	}
	if spec.AssociatePublicIpAddress != nil {
		a.AssociatePublicIpAddress = spec.AssociatePublicIpAddress
	}
	if spec.AmiFamily != nil {
		a.AmiFamily = spec.AmiFamily
	}
	if spec.UserData != nil {
		a.UserData = spec.UserData
	}
	if spec.Role != nil {
		a.Role = spec.Role
	}
	if spec.InstanceProfile != nil {
		a.InstanceProfile = spec.InstanceProfile
	}
	if spec.Context != nil {
		a.Context = spec.Context
	}
	if spec.DetailedMonitoring != nil {
		a.DetailedMonitoring = spec.DetailedMonitoring
	}
	if spec.InstanceStorePolicy != nil {
		s := spec.InstanceStorePolicy.String()
		a.InstanceStorePolicy = &s
	}
	for _, t := range spec.SubnetSelectorTerms {
		a.SubnetSelectorTerms = append(a.SubnetSelectorTerms, SubnetSelectorTermArgs{
			Tags: t.Tags,
			Id:   t.Id,
		})
	}
	for _, sg := range spec.SecurityGroupSelectorTerms {
		a.SecurityGroupSelectorTerms = append(a.SecurityGroupSelectorTerms, SecurityGroupSelectorTermArgs{
			Tags: sg.Tags,
			Id:   sg.Id,
			Name: sg.Name,
		})
	}
	for _, cr := range spec.CapacityReservationSelectorTerms {
		a.CapacityReservationSelectorTerms = append(a.CapacityReservationSelectorTerms, CapacityReservationSelectorTermArgs{
			Tags:    cr.Tags,
			Id:      cr.Id,
			OwnerId: cr.OwnerId,
		})
	}
	for _, ami := range spec.AmiSelectorTerms {
		a.AmiSelectorTerms = append(a.AmiSelectorTerms, AMISelectorTermArgs{
			Alias:        ami.Alias,
			Tags:         ami.Tags,
			Id:           ami.Id,
			Name:         ami.Name,
			Owner:        ami.Owner,
			SsmParameter: ami.SsmParameter,
		})
	}
	for _, bdm := range spec.BlockDeviceMappings {
		a.BlockDeviceMappings = append(a.BlockDeviceMappings, blockDeviceMappingFromProto(bdm))
	}
	return a
}

func blockDeviceMappingToProto(b BlockDeviceMappingArgs) *apiv1.BlockDeviceMapping {
	bdm := &apiv1.BlockDeviceMapping{}
	if b.DeviceName != nil {
		bdm.DeviceName = b.DeviceName
	}
	if b.RootVolume != nil {
		bdm.RootVolume = b.RootVolume
	}
	if b.Ebs != nil {
		ebs := &apiv1.BlockDevice{}
		if b.Ebs.DeleteOnTermination != nil {
			ebs.DeleteOnTermination = b.Ebs.DeleteOnTermination
		}
		if b.Ebs.Encrypted != nil {
			ebs.Encrypted = b.Ebs.Encrypted
		}
		if b.Ebs.Iops != nil {
			v := int64(*b.Ebs.Iops)
			ebs.Iops = &v
		}
		if b.Ebs.KmsKeyId != nil {
			ebs.KmsKeyId = b.Ebs.KmsKeyId
		}
		if b.Ebs.SnapshotId != nil {
			ebs.SnapshotId = b.Ebs.SnapshotId
		}
		if b.Ebs.Throughput != nil {
			v := int64(*b.Ebs.Throughput)
			ebs.Throughput = &v
		}
		if b.Ebs.VolumeInitializationRate != nil {
			v := int32(*b.Ebs.VolumeInitializationRate)
			ebs.VolumeInitializationRate = &v
		}
		if b.Ebs.VolumeSize != nil {
			ebs.VolumeSize = b.Ebs.VolumeSize
		}
		if b.Ebs.VolumeType != nil {
			ebs.VolumeType = b.Ebs.VolumeType
		}
		bdm.Ebs = ebs
	}
	return bdm
}

func blockDeviceMappingFromProto(b *apiv1.BlockDeviceMapping) BlockDeviceMappingArgs {
	if b == nil {
		return BlockDeviceMappingArgs{}
	}
	bdm := BlockDeviceMappingArgs{}
	if b.DeviceName != nil {
		bdm.DeviceName = b.DeviceName
	}
	if b.RootVolume != nil {
		bdm.RootVolume = b.RootVolume
	}
	if b.Ebs != nil {
		ebs := &BlockDeviceArgs{}
		if b.Ebs.DeleteOnTermination != nil {
			ebs.DeleteOnTermination = b.Ebs.DeleteOnTermination
		}
		if b.Ebs.Encrypted != nil {
			ebs.Encrypted = b.Ebs.Encrypted
		}
		if b.Ebs.Iops != nil {
			v := int(*b.Ebs.Iops)
			ebs.Iops = &v
		}
		if b.Ebs.KmsKeyId != nil {
			ebs.KmsKeyId = b.Ebs.KmsKeyId
		}
		if b.Ebs.SnapshotId != nil {
			ebs.SnapshotId = b.Ebs.SnapshotId
		}
		if b.Ebs.Throughput != nil {
			v := int(*b.Ebs.Throughput)
			ebs.Throughput = &v
		}
		if b.Ebs.VolumeInitializationRate != nil {
			v := int(*b.Ebs.VolumeInitializationRate)
			ebs.VolumeInitializationRate = &v
		}
		if b.Ebs.VolumeSize != nil {
			ebs.VolumeSize = b.Ebs.VolumeSize
		}
		if b.Ebs.VolumeType != nil {
			ebs.VolumeType = b.Ebs.VolumeType
		}
		bdm.Ebs = ebs
	}
	return bdm
}

func kubeletConfigToProto(k *KubeletConfigurationArgs) *apiv1.KubeletConfiguration {
	if k == nil {
		return nil
	}
	cfg := &apiv1.KubeletConfiguration{
		ClusterDns:              k.ClusterDns,
		SystemReserved:          k.SystemReserved,
		KubeReserved:            k.KubeReserved,
		EvictionHard:            k.EvictionHard,
		EvictionSoft:            k.EvictionSoft,
		EvictionSoftGracePeriod: k.EvictionSoftGracePeriod,
	}
	if k.MaxPods != nil {
		v := int32(*k.MaxPods)
		cfg.MaxPods = &v
	}
	if k.PodsPerCore != nil {
		v := int32(*k.PodsPerCore)
		cfg.PodsPerCore = &v
	}
	if k.EvictionMaxPodGracePeriod != nil {
		v := int32(*k.EvictionMaxPodGracePeriod)
		cfg.EvictionMaxPodGracePeriod = &v
	}
	if k.ImageGcHighThresholdPercent != nil {
		v := int32(*k.ImageGcHighThresholdPercent)
		cfg.ImageGcHighThresholdPercent = &v
	}
	if k.ImageGcLowThresholdPercent != nil {
		v := int32(*k.ImageGcLowThresholdPercent)
		cfg.ImageGcLowThresholdPercent = &v
	}
	if k.CpuCfsQuota != nil {
		cfg.CpuCfsQuota = k.CpuCfsQuota
	}
	return cfg
}

func kubeletConfigFromProto(cfg *apiv1.KubeletConfiguration) *KubeletConfigurationArgs {
	if cfg == nil {
		return nil
	}
	k := &KubeletConfigurationArgs{
		ClusterDns:              cfg.ClusterDns,
		SystemReserved:          cfg.SystemReserved,
		KubeReserved:            cfg.KubeReserved,
		EvictionHard:            cfg.EvictionHard,
		EvictionSoft:            cfg.EvictionSoft,
		EvictionSoftGracePeriod: cfg.EvictionSoftGracePeriod,
	}
	if cfg.MaxPods != nil {
		v := int(*cfg.MaxPods)
		k.MaxPods = &v
	}
	if cfg.PodsPerCore != nil {
		v := int(*cfg.PodsPerCore)
		k.PodsPerCore = &v
	}
	if cfg.EvictionMaxPodGracePeriod != nil {
		v := int(*cfg.EvictionMaxPodGracePeriod)
		k.EvictionMaxPodGracePeriod = &v
	}
	if cfg.ImageGcHighThresholdPercent != nil {
		v := int(*cfg.ImageGcHighThresholdPercent)
		k.ImageGcHighThresholdPercent = &v
	}
	if cfg.ImageGcLowThresholdPercent != nil {
		v := int(*cfg.ImageGcLowThresholdPercent)
		k.ImageGcLowThresholdPercent = &v
	}
	if cfg.CpuCfsQuota != nil {
		k.CpuCfsQuota = cfg.CpuCfsQuota
	}
	return k
}

func metadataOptionsToProto(m *MetadataOptionsArgs) *apiv1.MetadataOptions {
	if m == nil {
		return nil
	}
	opts := &apiv1.MetadataOptions{}
	if m.HttpEndpoint != "" {
		opts.HttpEndpoint = &m.HttpEndpoint
	}
	if m.HttpProtocolIpv6 != "" {
		opts.HttpProtocolIpv6 = &m.HttpProtocolIpv6
	}
	if m.HttpPutResponseHopLimit != 0 {
		v := int64(m.HttpPutResponseHopLimit)
		opts.HttpPutResponseHopLimit = &v
	}
	if m.HttpTokens != "" {
		opts.HttpTokens = &m.HttpTokens
	}
	return opts
}

func metadataOptionsFromProto(m *apiv1.MetadataOptions) *MetadataOptionsArgs {
	if m == nil {
		return nil
	}
	return &MetadataOptionsArgs{
		HttpEndpoint:            m.GetHttpEndpoint(),
		HttpProtocolIpv6:        m.GetHttpProtocolIpv6(),
		HttpPutResponseHopLimit: int(m.GetHttpPutResponseHopLimit()),
		HttpTokens:              m.GetHttpTokens(),
	}
}

// ---------- Azure NodeClass helpers ----------

func azureNodeClassSpecToProto(a *AzureNodeClassSpecArgs) *apiv1.AzureNodeClassSpec {
	if a == nil {
		return nil
	}
	spec := &apiv1.AzureNodeClassSpec{
		Tags:    a.Tags,
		Kubelet: azureKubeletConfigToProto(a.Kubelet),
	}
	if a.VnetSubnetId != "" {
		spec.VnetSubnetId = &a.VnetSubnetId
	}
	if a.OsDiskSizeGb != nil {
		v := int32(*a.OsDiskSizeGb)
		spec.OsDiskSizeGb = &v
	}
	if a.ImageFamily != nil {
		spec.ImageFamily = a.ImageFamily
	}
	if a.FipsMode != nil {
		spec.FipsMode = a.FipsMode
	}
	if a.MaxPods != nil {
		v := int32(*a.MaxPods)
		spec.MaxPods = &v
	}
	return spec
}

func azureNodeClassSpecFromProto(spec *apiv1.AzureNodeClassSpec) *AzureNodeClassSpecArgs {
	if spec == nil {
		return nil
	}
	a := &AzureNodeClassSpecArgs{
		Tags:    spec.Tags,
		Kubelet: azureKubeletConfigFromProto(spec.Kubelet),
	}
	if spec.VnetSubnetId != nil {
		a.VnetSubnetId = *spec.VnetSubnetId
	}
	if spec.OsDiskSizeGb != nil {
		v := int(*spec.OsDiskSizeGb)
		a.OsDiskSizeGb = &v
	}
	if spec.ImageFamily != nil {
		a.ImageFamily = spec.ImageFamily
	}
	if spec.FipsMode != nil {
		a.FipsMode = spec.FipsMode
	}
	if spec.MaxPods != nil {
		v := int(*spec.MaxPods)
		a.MaxPods = &v
	}
	return a
}

func azureKubeletConfigToProto(k *AzureKubeletConfigurationArgs) *apiv1.AzureKubeletConfiguration {
	if k == nil {
		return nil
	}
	cfg := &apiv1.AzureKubeletConfiguration{
		AllowedUnsafeSysctls: k.AllowedUnsafeSysctls,
	}
	if k.CpuManagerPolicy != nil {
		cfg.CpuManagerPolicy = k.CpuManagerPolicy
	}
	if k.CpuCfsQuota != nil {
		cfg.CpuCfsQuota = k.CpuCfsQuota
	}
	if k.CpuCfsQuotaPeriod != nil {
		cfg.CpuCfsQuotaPeriod = k.CpuCfsQuotaPeriod
	}
	if k.ImageGcHighThresholdPercent != nil {
		v := int32(*k.ImageGcHighThresholdPercent)
		cfg.ImageGcHighThresholdPercent = &v
	}
	if k.ImageGcLowThresholdPercent != nil {
		v := int32(*k.ImageGcLowThresholdPercent)
		cfg.ImageGcLowThresholdPercent = &v
	}
	if k.TopologyManagerPolicy != nil {
		cfg.TopologyManagerPolicy = k.TopologyManagerPolicy
	}
	if k.ContainerLogMaxSize != nil {
		cfg.ContainerLogMaxSize = k.ContainerLogMaxSize
	}
	if k.ContainerLogMaxFiles != nil {
		v := int32(*k.ContainerLogMaxFiles)
		cfg.ContainerLogMaxFiles = &v
	}
	if k.PodPidsLimit != nil {
		v := int64(*k.PodPidsLimit)
		cfg.PodPidsLimit = &v
	}
	return cfg
}

func azureKubeletConfigFromProto(cfg *apiv1.AzureKubeletConfiguration) *AzureKubeletConfigurationArgs {
	if cfg == nil {
		return nil
	}
	k := &AzureKubeletConfigurationArgs{
		AllowedUnsafeSysctls: cfg.AllowedUnsafeSysctls,
	}
	if cfg.CpuManagerPolicy != nil {
		k.CpuManagerPolicy = cfg.CpuManagerPolicy
	}
	if cfg.CpuCfsQuota != nil {
		k.CpuCfsQuota = cfg.CpuCfsQuota
	}
	if cfg.CpuCfsQuotaPeriod != nil {
		k.CpuCfsQuotaPeriod = cfg.CpuCfsQuotaPeriod
	}
	if cfg.ImageGcHighThresholdPercent != nil {
		v := int(*cfg.ImageGcHighThresholdPercent)
		k.ImageGcHighThresholdPercent = &v
	}
	if cfg.ImageGcLowThresholdPercent != nil {
		v := int(*cfg.ImageGcLowThresholdPercent)
		k.ImageGcLowThresholdPercent = &v
	}
	if cfg.TopologyManagerPolicy != nil {
		k.TopologyManagerPolicy = cfg.TopologyManagerPolicy
	}
	if cfg.ContainerLogMaxSize != nil {
		k.ContainerLogMaxSize = cfg.ContainerLogMaxSize
	}
	if cfg.ContainerLogMaxFiles != nil {
		v := int(*cfg.ContainerLogMaxFiles)
		k.ContainerLogMaxFiles = &v
	}
	if cfg.PodPidsLimit != nil {
		v := int(*cfg.PodPidsLimit)
		k.PodPidsLimit = &v
	}
	return k
}
