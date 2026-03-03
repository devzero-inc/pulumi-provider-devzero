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

// Annotate provides SDK documentation for TaintArgs fields.
func (t *TaintArgs) Annotate(a infer.Annotator) {
	a.Describe(&t.Key, "Taint key to apply to provisioned nodes.")
	a.Describe(&t.Value, "Taint value associated with the key.")
	a.Describe(&t.Effect, "Taint effect. One of: 'NoSchedule', 'PreferNoSchedule', 'NoExecute'.")
}

// DisruptionBudgetArgs represents a single disruption budget entry.
type DisruptionBudgetArgs struct {
	Reasons  []string `pulumi:"reasons,optional"`
	Nodes    string   `pulumi:"nodes,optional"`
	Schedule string   `pulumi:"schedule,optional"`
	Duration string   `pulumi:"duration,optional"`
}

// Annotate provides SDK documentation for DisruptionBudgetArgs fields.
func (d *DisruptionBudgetArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.Reasons, "Disruption reasons this budget applies to (e.g. 'Underutilized', 'Empty', 'Drifted').")
	a.Describe(&d.Nodes, "Maximum nodes that may be disrupted, as an absolute count or percentage (e.g. '10%').")
	a.Describe(&d.Schedule, "Cron schedule during which this budget is active (5-field format).")
	a.Describe(&d.Duration, "Duration the budget window stays active (e.g. '1h', '30m').")
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

// Annotate provides SDK documentation for DisruptionPolicyArgs fields.
func (d *DisruptionPolicyArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.ConsolidateAfter, "Duration to wait after a node becomes empty before consolidating (e.g. '30s').")
	a.Describe(&d.ConsolidationPolicy, "When to consolidate nodes. One of: 'WhenEmpty', 'WhenUnderutilized'.")
	a.Describe(&d.ExpireAfter, "Duration after which provisioned nodes are replaced regardless of load (e.g. '720h').")
	a.Describe(&d.TtlSecondsAfterEmpty, "Seconds before an empty node is terminated (deprecated; prefer ConsolidateAfter).")
	a.Describe(&d.TerminationGracePeriodSeconds, "Grace period in seconds before forcefully terminating a draining node.")
	a.Describe(&d.Budgets, "Disruption budgets limiting how many nodes can be disrupted at once.")
}

// ResourceLimitsArgs sets resource limits on the node pool.
type ResourceLimitsArgs struct {
	Cpu    string `pulumi:"cpu,optional"`
	Memory string `pulumi:"memory,optional"`
}

// Annotate provides SDK documentation for ResourceLimitsArgs fields.
func (r *ResourceLimitsArgs) Annotate(a infer.Annotator) {
	a.Describe(&r.Cpu, "Maximum total CPU that may be provisioned (e.g. '1000' for 1000 vCPUs).")
	a.Describe(&r.Memory, "Maximum total memory that may be provisioned (e.g. '1000Gi').")
}

// SubnetSelectorTermArgs selects subnets by tag or ID.
type SubnetSelectorTermArgs struct {
	Tags map[string]string `pulumi:"tags,optional"`
	Id   string            `pulumi:"id,optional"`
}

// Annotate provides SDK documentation for SubnetSelectorTermArgs fields.
func (s *SubnetSelectorTermArgs) Annotate(a infer.Annotator) {
	a.Describe(&s.Tags, "Map of AWS tags used to select subnets.")
	a.Describe(&s.Id, "Explicit AWS subnet ID.")
}

// SecurityGroupSelectorTermArgs selects security groups.
type SecurityGroupSelectorTermArgs struct {
	Tags map[string]string `pulumi:"tags,optional"`
	Id   string            `pulumi:"id,optional"`
	Name string            `pulumi:"name,optional"`
}

// Annotate provides SDK documentation for SecurityGroupSelectorTermArgs fields.
func (s *SecurityGroupSelectorTermArgs) Annotate(a infer.Annotator) {
	a.Describe(&s.Tags, "Map of AWS tags used to select security groups.")
	a.Describe(&s.Id, "Explicit AWS security group ID.")
	a.Describe(&s.Name, "Security group name filter.")
}

// CapacityReservationSelectorTermArgs selects capacity reservations.
type CapacityReservationSelectorTermArgs struct {
	Tags    map[string]string `pulumi:"tags,optional"`
	Id      string            `pulumi:"id,optional"`
	OwnerId string            `pulumi:"ownerId,optional"`
}

// Annotate provides SDK documentation for CapacityReservationSelectorTermArgs fields.
func (c *CapacityReservationSelectorTermArgs) Annotate(a infer.Annotator) {
	a.Describe(&c.Tags, "Map of AWS tags used to select capacity reservations.")
	a.Describe(&c.Id, "Explicit capacity reservation ID.")
	a.Describe(&c.OwnerId, "AWS account ID that owns the capacity reservation.")
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

// Annotate provides SDK documentation for AMISelectorTermArgs fields.
func (a *AMISelectorTermArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Alias, "Well-known alias for the AMI family (e.g. 'al2@latest').")
	ann.Describe(&a.Tags, "Map of AWS tags used to select AMIs.")
	ann.Describe(&a.Id, "Explicit AMI ID.")
	ann.Describe(&a.Name, "AMI name filter (supports wildcards).")
	ann.Describe(&a.Owner, "AWS account ID or alias that owns the AMI.")
	ann.Describe(&a.SsmParameter, "SSM parameter path that stores the AMI ID.")
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

// Annotate provides SDK documentation for BlockDeviceArgs fields.
func (b *BlockDeviceArgs) Annotate(a infer.Annotator) {
	a.Describe(&b.DeleteOnTermination, "Whether to delete the EBS volume when the instance terminates.")
	a.Describe(&b.Encrypted, "Whether to encrypt the EBS volume.")
	a.Describe(&b.Iops, "IOPS to provision for io1/io2 volume types.")
	a.Describe(&b.KmsKeyId, "KMS key ID or ARN used to encrypt the volume.")
	a.Describe(&b.SnapshotId, "EBS snapshot ID to restore the volume from.")
	a.Describe(&b.Throughput, "Throughput in MiB/s for gp3 volumes.")
	a.Describe(&b.VolumeInitializationRate, "Rate in MiB/s for initializing volumes from snapshots.")
	a.Describe(&b.VolumeSize, "Volume size (e.g. '20Gi').")
	a.Describe(&b.VolumeType, "EBS volume type (e.g. 'gp3', 'io1', 'st1').")
}

// BlockDeviceMappingArgs maps an EBS block device to a device name.
type BlockDeviceMappingArgs struct {
	DeviceName *string          `pulumi:"deviceName,optional"`
	Ebs        *BlockDeviceArgs `pulumi:"ebs,optional"`
	RootVolume *bool            `pulumi:"rootVolume,optional"`
}

// Annotate provides SDK documentation for BlockDeviceMappingArgs fields.
func (b *BlockDeviceMappingArgs) Annotate(a infer.Annotator) {
	a.Describe(&b.DeviceName, "Device name to map the volume to (e.g. '/dev/xvda').")
	a.Describe(&b.Ebs, "EBS volume configuration for this device.")
	a.Describe(&b.RootVolume, "Whether this mapping is for the root volume.")
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

// Annotate provides SDK documentation for KubeletConfigurationArgs fields.
func (k *KubeletConfigurationArgs) Annotate(a infer.Annotator) {
	a.Describe(&k.ClusterDns, "List of DNS server IP addresses used by kubelet.")
	a.Describe(&k.MaxPods, "Maximum number of pods per node.")
	a.Describe(&k.PodsPerCore, "Maximum pods per CPU core (multiplied by core count for max pods).")
	a.Describe(&k.SystemReserved, "Resources reserved for OS system daemons (e.g. {'cpu': '100m'}).")
	a.Describe(&k.KubeReserved, "Resources reserved for Kubernetes system components.")
	a.Describe(&k.EvictionHard, "Hard eviction thresholds that trigger immediate pod eviction.")
	a.Describe(&k.EvictionSoft, "Soft eviction thresholds that trigger eviction after a grace period.")
	a.Describe(&k.EvictionSoftGracePeriod, "Grace period for each soft eviction threshold.")
	a.Describe(&k.EvictionMaxPodGracePeriod, "Maximum grace period in seconds when evicting pods.")
	a.Describe(&k.ImageGcHighThresholdPercent, "Disk usage percentage that triggers image garbage collection.")
	a.Describe(&k.ImageGcLowThresholdPercent, "Disk usage percentage below which image GC stops freeing space.")
	a.Describe(&k.CpuCfsQuota, "Whether to enforce CPU CFS quota limits for containers.")
}

// MetadataOptionsArgs configures EC2 instance metadata options.
type MetadataOptionsArgs struct {
	HttpEndpoint            string `pulumi:"httpEndpoint,optional"`
	HttpProtocolIpv6        string `pulumi:"httpProtocolIpv6,optional"`
	HttpPutResponseHopLimit int    `pulumi:"httpPutResponseHopLimit,optional"`
	HttpTokens              string `pulumi:"httpTokens,optional"`
}

// Annotate provides SDK documentation for MetadataOptionsArgs fields.
func (m *MetadataOptionsArgs) Annotate(a infer.Annotator) {
	a.Describe(&m.HttpEndpoint, "Enable or disable the instance metadata endpoint. One of: 'enabled', 'disabled'.")
	a.Describe(&m.HttpProtocolIpv6, "Enable IPv6 for the metadata endpoint. One of: 'enabled', 'disabled'.")
	a.Describe(&m.HttpPutResponseHopLimit, "HTTP PUT response hop limit for metadata requests (1-64).")
	a.Describe(&m.HttpTokens, "Whether to require IMDSv2 tokens. One of: 'optional', 'required'.")
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

// Annotate provides SDK documentation for AWSNodeClassSpecArgs fields.
func (a *AWSNodeClassSpecArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.SubnetSelectorTerms, "Selectors for the subnets nodes will be launched into.")
	ann.Describe(&a.SecurityGroupSelectorTerms, "Selectors for security groups attached to nodes.")
	ann.Describe(&a.CapacityReservationSelectorTerms, "Selectors for EC2 capacity reservations to prioritize.")
	ann.Describe(&a.AssociatePublicIpAddress, "Whether to assign a public IP address to provisioned nodes.")
	ann.Describe(&a.AmiSelectorTerms, "Selectors for the AMIs used to launch nodes.")
	ann.Describe(&a.AmiFamily, "AMI family shorthand (e.g. 'AL2', 'Bottlerocket', 'Windows2022').")
	ann.Describe(&a.UserData, "Custom user data script injected into the node launch template.")
	ann.Describe(&a.Role, "IAM role name assigned to nodes.")
	ann.Describe(&a.InstanceProfile, "IAM instance profile name (use instead of Role when profile already exists).")
	ann.Describe(&a.Tags, "AWS tags applied to all resources created by this node class.")
	ann.Describe(&a.Kubelet, "Kubelet configuration overrides for AWS nodes.")
	ann.Describe(&a.BlockDeviceMappings, "EBS block device mappings for nodes.")
	ann.Describe(&a.InstanceStorePolicy, "Policy for handling instance store volumes. One of: 'RAID0'.")
	ann.Describe(&a.DetailedMonitoring, "Enable detailed CloudWatch monitoring for instances.")
	ann.Describe(&a.MetadataOptions, "EC2 instance metadata (IMDS) options.")
	ann.Describe(&a.Context, "EC2 launch template context ARN.")
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

// Annotate provides SDK documentation for AzureKubeletConfigurationArgs fields.
func (a *AzureKubeletConfigurationArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.CpuManagerPolicy, "CPU manager policy. One of: 'none', 'static'.")
	ann.Describe(&a.CpuCfsQuota, "Whether to enforce CPU CFS quota for containers.")
	ann.Describe(&a.CpuCfsQuotaPeriod, "CPU CFS quota period (e.g. '100ms').")
	ann.Describe(&a.ImageGcHighThresholdPercent, "Disk usage percentage triggering image GC.")
	ann.Describe(&a.ImageGcLowThresholdPercent, "Disk usage percentage below which image GC stops.")
	ann.Describe(&a.TopologyManagerPolicy, "Topology manager policy for NUMA-aware scheduling.")
	ann.Describe(&a.AllowedUnsafeSysctls, "Unsafe sysctl patterns that are allowed (e.g. 'net.ipv4.*').")
	ann.Describe(&a.ContainerLogMaxSize, "Maximum container log file size before rotation (e.g. '10Mi').")
	ann.Describe(&a.ContainerLogMaxFiles, "Maximum number of container log files to retain.")
	ann.Describe(&a.PodPidsLimit, "Maximum number of process IDs per pod.")
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

// Annotate provides SDK documentation for AzureNodeClassSpecArgs fields.
func (a *AzureNodeClassSpecArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.VnetSubnetId, "Azure VNet subnet resource ID for node networking.")
	ann.Describe(&a.OsDiskSizeGb, "OS disk size in GB.")
	ann.Describe(&a.ImageFamily, "Azure node image family (e.g. 'AzureLinux', 'Ubuntu2204').")
	ann.Describe(&a.FipsMode, "FIPS compliance mode. One of: 'Enabled', 'Disabled'.")
	ann.Describe(&a.Tags, "Azure tags applied to all resources created by this node class.")
	ann.Describe(&a.Kubelet, "Kubelet configuration overrides for Azure nodes.")
	ann.Describe(&a.MaxPods, "Maximum pods per node (overrides AKS default).")
}

// RawKarpenterSpecArgs provides raw YAML for a custom Karpenter node pool / node class.
type RawKarpenterSpecArgs struct {
	NodepoolYaml  string `pulumi:"nodepoolYaml,optional"`
	NodeclassYaml string `pulumi:"nodeclassYaml,optional"`
}

// Annotate provides SDK documentation for RawKarpenterSpecArgs fields.
func (r *RawKarpenterSpecArgs) Annotate(a infer.Annotator) {
	a.Describe(&r.NodepoolYaml, "Raw YAML for a complete Karpenter NodePool resource (escape hatch).")
	a.Describe(&r.NodeclassYaml, "Raw YAML for a complete Karpenter NodeClass resource (escape hatch).")
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
