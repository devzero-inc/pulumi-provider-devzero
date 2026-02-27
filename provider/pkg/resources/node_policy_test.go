package resources

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// mockRecommendationClientNodePolicy extends mockRecommendationClientFull with NodePolicy methods.
type mockRecommendationClientNodePolicy struct {
	mockRecommendationClientFull
	createNodePoliciesFn func(context.Context, *connect.Request[apiv1.CreateNodePoliciesRequest]) (*connect.Response[apiv1.CreateNodePoliciesResponse], error)
	listNodePoliciesFn   func(context.Context, *connect.Request[apiv1.ListNodePoliciesRequest]) (*connect.Response[apiv1.ListNodePoliciesResponse], error)
	updateNodePolicyFn   func(context.Context, *connect.Request[apiv1.UpdateNodePolicyRequest]) (*connect.Response[apiv1.UpdateNodePolicyResponse], error)
}

func (m *mockRecommendationClientNodePolicy) CreateNodePolicies(ctx context.Context, req *connect.Request[apiv1.CreateNodePoliciesRequest]) (*connect.Response[apiv1.CreateNodePoliciesResponse], error) {
	return m.createNodePoliciesFn(ctx, req)
}
func (m *mockRecommendationClientNodePolicy) ListNodePolicies(ctx context.Context, req *connect.Request[apiv1.ListNodePoliciesRequest]) (*connect.Response[apiv1.ListNodePoliciesResponse], error) {
	return m.listNodePoliciesFn(ctx, req)
}
func (m *mockRecommendationClientNodePolicy) UpdateNodePolicy(ctx context.Context, req *connect.Request[apiv1.UpdateNodePolicyRequest]) (*connect.Response[apiv1.UpdateNodePolicyResponse], error) {
	return m.updateNodePolicyFn(ctx, req)
}

func withMockNodePolicyClientSet(t *testing.T, rec *mockRecommendationClientNodePolicy) {
	t.Helper()
	prev := clientset.Get()
	t.Cleanup(func() { clientset.Set(prev) })
	clientset.Set(&clientset.ClientSet{
		TeamID:               "team-test",
		RecommendationClient: rec,
	})
}

// ---------- Create ----------

func TestNodePolicy_Create_Success(t *testing.T) {
	rec := &mockRecommendationClientNodePolicy{
		createNodePoliciesFn: func(_ context.Context, req *connect.Request[apiv1.CreateNodePoliciesRequest]) (*connect.Response[apiv1.CreateNodePoliciesResponse], error) {
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			if len(req.Msg.Policies) != 1 {
				t.Fatalf("expected 1 policy, got %d", len(req.Msg.Policies))
			}
			p := req.Msg.Policies[0]
			if p.Name != "my-node-policy" {
				t.Errorf("Name: got %q, want %q", p.Name, "my-node-policy")
			}
			if p.Weight != 10 {
				t.Errorf("Weight: got %d, want 10", p.Weight)
			}
			return connect.NewResponse(&apiv1.CreateNodePoliciesResponse{
				Policies: []*apiv1.NodePolicy{
					{Id: "np-123", Name: "my-node-policy", TeamId: "team-test", Weight: 10},
				},
			}), nil
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	resp, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyArgs]{
		Name:   "my-node-policy",
		Inputs: NodePolicyArgs{Name: "my-node-policy", Weight: 10},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "np-123" {
		t.Errorf("ID: got %q, want %q", resp.ID, "np-123")
	}
	if resp.Output.Name != "my-node-policy" {
		t.Errorf("Name: got %q, want %q", resp.Output.Name, "my-node-policy")
	}
}

func TestNodePolicy_Create_Preview(t *testing.T) {
	n := &NodePolicy{}
	resp, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyArgs]{
		DryRun: true,
		Inputs: NodePolicyArgs{Name: "preview-policy", Weight: 5},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID on preview, got %q", resp.ID)
	}
	if resp.Output.Name != "preview-policy" {
		t.Errorf("Name: got %q, want %q", resp.Output.Name, "preview-policy")
	}
}

func TestNodePolicy_Create_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	t.Cleanup(func() { clientset.Set(prev) })

	n := &NodePolicy{}
	_, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyArgs]{
		Inputs: NodePolicyArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodePolicy_Create_APIError(t *testing.T) {
	rec := &mockRecommendationClientNodePolicy{
		createNodePoliciesFn: func(_ context.Context, _ *connect.Request[apiv1.CreateNodePoliciesRequest]) (*connect.Response[apiv1.CreateNodePoliciesResponse], error) {
			return nil, errors.New("api error")
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	_, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyArgs]{
		Inputs: NodePolicyArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodePolicy_Create_EmptyResponse(t *testing.T) {
	rec := &mockRecommendationClientNodePolicy{
		createNodePoliciesFn: func(_ context.Context, _ *connect.Request[apiv1.CreateNodePoliciesRequest]) (*connect.Response[apiv1.CreateNodePoliciesResponse], error) {
			return connect.NewResponse(&apiv1.CreateNodePoliciesResponse{}), nil
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	_, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyArgs]{
		Inputs: NodePolicyArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error on empty response, got nil")
	}
}

// ---------- Read ----------

func TestNodePolicy_Read_Success(t *testing.T) {
	desc := "a policy"
	rec := &mockRecommendationClientNodePolicy{
		listNodePoliciesFn: func(_ context.Context, req *connect.Request[apiv1.ListNodePoliciesRequest]) (*connect.Response[apiv1.ListNodePoliciesResponse], error) {
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			return connect.NewResponse(&apiv1.ListNodePoliciesResponse{
				Policies: []*apiv1.NodePolicy{
					{Id: "np-123", Name: "my-policy", TeamId: "team-test", Description: desc},
					{Id: "np-999", Name: "other"},
				},
			}), nil
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	resp, err := n.Read(context.Background(), infer.ReadRequest[NodePolicyArgs, NodePolicyState]{
		ID:     "np-123",
		Inputs: NodePolicyArgs{Name: "my-policy"},
		State:  NodePolicyState{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "np-123" {
		t.Errorf("ID: got %q, want %q", resp.ID, "np-123")
	}
	if resp.State.Name != "my-policy" {
		t.Errorf("Name: got %q, want %q", resp.State.Name, "my-policy")
	}
	if resp.State.Description == nil || *resp.State.Description != desc {
		t.Errorf("Description: want %q", desc)
	}
}

func TestNodePolicy_Read_NotFound(t *testing.T) {
	rec := &mockRecommendationClientNodePolicy{
		listNodePoliciesFn: func(_ context.Context, _ *connect.Request[apiv1.ListNodePoliciesRequest]) (*connect.Response[apiv1.ListNodePoliciesResponse], error) {
			return connect.NewResponse(&apiv1.ListNodePoliciesResponse{
				Policies: []*apiv1.NodePolicy{{Id: "other-id", Name: "other"}},
			}), nil
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	_, err := n.Read(context.Background(), infer.ReadRequest[NodePolicyArgs, NodePolicyState]{
		ID: "np-missing",
	})
	if err == nil {
		t.Fatal("expected error for not-found, got nil")
	}
}

func TestNodePolicy_Read_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	t.Cleanup(func() { clientset.Set(prev) })

	n := &NodePolicy{}
	_, err := n.Read(context.Background(), infer.ReadRequest[NodePolicyArgs, NodePolicyState]{ID: "np-123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodePolicy_Read_APIError(t *testing.T) {
	rec := &mockRecommendationClientNodePolicy{
		listNodePoliciesFn: func(_ context.Context, _ *connect.Request[apiv1.ListNodePoliciesRequest]) (*connect.Response[apiv1.ListNodePoliciesResponse], error) {
			return nil, errors.New("api error")
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	_, err := n.Read(context.Background(), infer.ReadRequest[NodePolicyArgs, NodePolicyState]{ID: "np-123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------- Update ----------

func TestNodePolicy_Update_Success(t *testing.T) {
	rec := &mockRecommendationClientNodePolicy{
		updateNodePolicyFn: func(_ context.Context, req *connect.Request[apiv1.UpdateNodePolicyRequest]) (*connect.Response[apiv1.UpdateNodePolicyResponse], error) {
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			if req.Msg.Policy == nil {
				t.Fatal("policy is nil")
			}
			if req.Msg.Policy.Id != "np-123" {
				t.Errorf("Policy.Id: got %q, want %q", req.Msg.Policy.Id, "np-123")
			}
			if req.Msg.Policy.Name != "updated-policy" {
				t.Errorf("Name: got %q, want %q", req.Msg.Policy.Name, "updated-policy")
			}
			return connect.NewResponse(&apiv1.UpdateNodePolicyResponse{
				Policy: &apiv1.NodePolicy{Id: "np-123", Name: "updated-policy", TeamId: "team-test"},
			}), nil
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	resp, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyArgs, NodePolicyState]{
		ID:     "np-123",
		Inputs: NodePolicyArgs{Name: "updated-policy"},
		State:  NodePolicyState{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Name != "updated-policy" {
		t.Errorf("Name: got %q, want %q", resp.Output.Name, "updated-policy")
	}
}

func TestNodePolicy_Update_Preview(t *testing.T) {
	n := &NodePolicy{}
	resp, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyArgs, NodePolicyState]{
		ID:     "np-123",
		DryRun: true,
		Inputs: NodePolicyArgs{Name: "preview-update"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Name != "preview-update" {
		t.Errorf("Name: got %q, want %q", resp.Output.Name, "preview-update")
	}
}

func TestNodePolicy_Update_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	t.Cleanup(func() { clientset.Set(prev) })

	n := &NodePolicy{}
	_, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyArgs, NodePolicyState]{
		ID:     "np-123",
		Inputs: NodePolicyArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodePolicy_Update_APIError(t *testing.T) {
	rec := &mockRecommendationClientNodePolicy{
		updateNodePolicyFn: func(_ context.Context, _ *connect.Request[apiv1.UpdateNodePolicyRequest]) (*connect.Response[apiv1.UpdateNodePolicyResponse], error) {
			return nil, errors.New("api error")
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	_, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyArgs, NodePolicyState]{
		ID:     "np-123",
		Inputs: NodePolicyArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodePolicy_Update_EmptyResponse(t *testing.T) {
	rec := &mockRecommendationClientNodePolicy{
		updateNodePolicyFn: func(_ context.Context, _ *connect.Request[apiv1.UpdateNodePolicyRequest]) (*connect.Response[apiv1.UpdateNodePolicyResponse], error) {
			return connect.NewResponse(&apiv1.UpdateNodePolicyResponse{}), nil
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	_, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyArgs, NodePolicyState]{
		ID:     "np-123",
		Inputs: NodePolicyArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error on empty response, got nil")
	}
}

// ---------- Delete ----------

func TestNodePolicy_Delete_StateOnly(t *testing.T) {
	// Delete should always succeed without calling any API.
	n := &NodePolicy{}
	_, err := n.Delete(context.Background(), infer.DeleteRequest[NodePolicyState]{
		ID:    "np-123",
		State: NodePolicyState{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodePolicy_Delete_NilClientSet(t *testing.T) {
	// Even with nil clientset, delete must succeed (state-only).
	prev := clientset.Get()
	clientset.Set(nil)
	t.Cleanup(func() { clientset.Set(prev) })

	n := &NodePolicy{}
	_, err := n.Delete(context.Background(), infer.DeleteRequest[NodePolicyState]{ID: "np-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- proto conversion roundtrips ----------

func TestNodePolicy_TaintsRoundtrip(t *testing.T) {
	input := []TaintArgs{
		{Key: "dedicated", Value: "gpu", Effect: "NoSchedule"},
		{Key: "maintenance", Effect: "NoExecute"},
	}
	proto := taintsToProto(input)
	if len(proto) != 2 {
		t.Fatalf("expected 2 taints, got %d", len(proto))
	}
	got := taintsFromProto(proto)
	for i, want := range input {
		if got[i] != want {
			t.Errorf("taint[%d]: got %+v, want %+v", i, got[i], want)
		}
	}
}

func TestNodePolicy_DisruptionPolicyRoundtrip(t *testing.T) {
	input := &DisruptionPolicyArgs{
		ConsolidateAfter:              "30m",
		ConsolidationPolicy:           "WhenEmptyOrUnderutilized",
		ExpireAfter:                   "720h",
		TtlSecondsAfterEmpty:          300,
		TerminationGracePeriodSeconds: 60,
		Budgets: []DisruptionBudgetArgs{
			{Reasons: []string{"Underutilized", "Empty"}, Nodes: "10%", Duration: "1h"},
		},
	}
	proto := disruptionPolicyToProto(input)
	if proto.ConsolidateAfter != "30m" {
		t.Errorf("ConsolidateAfter: got %q", proto.ConsolidateAfter)
	}
	if len(proto.Budgets) != 1 {
		t.Fatalf("expected 1 budget, got %d", len(proto.Budgets))
	}

	got := disruptionPolicyFromProto(proto)
	if got.ConsolidateAfter != input.ConsolidateAfter {
		t.Errorf("ConsolidateAfter: got %q, want %q", got.ConsolidateAfter, input.ConsolidateAfter)
	}
	if got.TtlSecondsAfterEmpty != input.TtlSecondsAfterEmpty {
		t.Errorf("TtlSecondsAfterEmpty: got %d, want %d", got.TtlSecondsAfterEmpty, input.TtlSecondsAfterEmpty)
	}
	if len(got.Budgets) != 1 || got.Budgets[0].Nodes != "10%" {
		t.Errorf("Budgets: got %+v", got.Budgets)
	}
}

func TestNodePolicy_ResourceLimitsRoundtrip(t *testing.T) {
	input := &ResourceLimitsArgs{Cpu: "1000", Memory: "4Gi"}
	got := resourceLimitsFromProto(resourceLimitsToProto(input))
	if got.Cpu != "1000" || got.Memory != "4Gi" {
		t.Errorf("got %+v", got)
	}
}

func TestNodePolicy_RawKarpenterRoundtrip(t *testing.T) {
	input := []RawKarpenterSpecArgs{
		{NodepoolYaml: "apiVersion: karpenter.sh/v1\nkind: NodePool", NodeclassYaml: "apiVersion: karpenter.k8s.aws/v1\nkind: EC2NodeClass"},
	}
	got := rawKarpenterSpecsFromProto(rawKarpenterSpecsToProto(input))
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if got[0].NodepoolYaml != input[0].NodepoolYaml {
		t.Errorf("NodepoolYaml mismatch")
	}
	if got[0].NodeclassYaml != input[0].NodeclassYaml {
		t.Errorf("NodeclassYaml mismatch")
	}
}

func TestNodePolicy_AWSNodeClassSpecRoundtrip(t *testing.T) {
	maxPods := 110
	enc := true
	volSize := "100Gi"
	volType := "gp3"
	input := &AWSNodeClassSpecArgs{
		AmiFamily: strPtr("AL2"),
		Tags:      map[string]string{"env": "prod"},
		SubnetSelectorTerms: []SubnetSelectorTermArgs{
			{Tags: map[string]string{"Name": "private"}, Id: "subnet-abc"},
		},
		SecurityGroupSelectorTerms: []SecurityGroupSelectorTermArgs{
			{Name: "default-sg"},
		},
		AmiSelectorTerms: []AMISelectorTermArgs{
			{Alias: "al2@latest"},
		},
		Kubelet: &KubeletConfigurationArgs{
			MaxPods: &maxPods,
		},
		BlockDeviceMappings: []BlockDeviceMappingArgs{
			{
				DeviceName: strPtr("/dev/xvda"),
				Ebs: &BlockDeviceArgs{
					Encrypted:  &enc,
					VolumeSize: &volSize,
					VolumeType: &volType,
				},
			},
		},
		MetadataOptions: &MetadataOptionsArgs{
			HttpTokens:   "required",
			HttpEndpoint: "enabled",
		},
	}

	proto := awsNodeClassSpecToProto(input)
	if proto.AmiFamily == nil || *proto.AmiFamily != "AL2" {
		t.Errorf("AmiFamily: got %v", proto.AmiFamily)
	}
	if len(proto.SubnetSelectorTerms) != 1 {
		t.Errorf("SubnetSelectorTerms: got %d", len(proto.SubnetSelectorTerms))
	}
	if proto.MetadataOptions == nil {
		t.Fatal("MetadataOptions is nil")
	}

	got := awsNodeClassSpecFromProto(proto)
	if got.AmiFamily == nil || *got.AmiFamily != "AL2" {
		t.Errorf("AmiFamily roundtrip failed")
	}
	if len(got.SubnetSelectorTerms) != 1 || got.SubnetSelectorTerms[0].Id != "subnet-abc" {
		t.Errorf("SubnetSelectorTerms roundtrip failed")
	}
	if got.Kubelet == nil || got.Kubelet.MaxPods == nil || *got.Kubelet.MaxPods != 110 {
		t.Errorf("Kubelet.MaxPods roundtrip failed")
	}
	if len(got.BlockDeviceMappings) != 1 {
		t.Fatalf("BlockDeviceMappings: got %d", len(got.BlockDeviceMappings))
	}
	bdm := got.BlockDeviceMappings[0]
	if bdm.Ebs == nil || bdm.Ebs.VolumeSize == nil || *bdm.Ebs.VolumeSize != "100Gi" {
		t.Errorf("BlockDeviceMapping.Ebs.VolumeSize roundtrip failed")
	}
	if got.MetadataOptions == nil || got.MetadataOptions.HttpTokens != "required" {
		t.Errorf("MetadataOptions roundtrip failed")
	}
}

func TestNodePolicy_AzureNodeClassSpecRoundtrip(t *testing.T) {
	osDisk := 128
	maxPods := 50
	input := &AzureNodeClassSpecArgs{
		VnetSubnetId: "subnet-azure",
		OsDiskSizeGb: &osDisk,
		ImageFamily:  strPtr("Ubuntu2204"),
		Tags:         map[string]string{"cluster": "dev"},
		MaxPods:      &maxPods,
		Kubelet: &AzureKubeletConfigurationArgs{
			CpuManagerPolicy:     strPtr("static"),
			AllowedUnsafeSysctls: []string{"net.core.somaxconn"},
		},
	}

	proto := azureNodeClassSpecToProto(input)
	if proto.VnetSubnetId == nil || *proto.VnetSubnetId != "subnet-azure" {
		t.Errorf("VnetSubnetId: got %v", proto.VnetSubnetId)
	}
	if proto.ImageFamily == nil || *proto.ImageFamily != "Ubuntu2204" {
		t.Errorf("ImageFamily: got %v", proto.ImageFamily)
	}

	got := azureNodeClassSpecFromProto(proto)
	if got.VnetSubnetId != "subnet-azure" {
		t.Errorf("VnetSubnetId roundtrip: got %q", got.VnetSubnetId)
	}
	if got.OsDiskSizeGb == nil || *got.OsDiskSizeGb != 128 {
		t.Errorf("OsDiskSizeGb roundtrip failed")
	}
	if got.MaxPods == nil || *got.MaxPods != 50 {
		t.Errorf("MaxPods roundtrip failed")
	}
	if got.Kubelet == nil || got.Kubelet.CpuManagerPolicy == nil || *got.Kubelet.CpuManagerPolicy != "static" {
		t.Errorf("Kubelet.CpuManagerPolicy roundtrip failed")
	}
}

func TestNodePolicy_Create_WithSelectorsAndDisruption(t *testing.T) {
	rec := &mockRecommendationClientNodePolicy{
		createNodePoliciesFn: func(_ context.Context, req *connect.Request[apiv1.CreateNodePoliciesRequest]) (*connect.Response[apiv1.CreateNodePoliciesResponse], error) {
			p := req.Msg.Policies[0]
			if p.InstanceTypes == nil {
				t.Error("InstanceTypes selector missing")
			}
			if p.CapacityTypes == nil {
				t.Error("CapacityTypes selector missing")
			}
			if p.Disruption == nil {
				t.Error("Disruption missing")
			}
			return connect.NewResponse(&apiv1.CreateNodePoliciesResponse{
				Policies: []*apiv1.NodePolicy{{Id: "np-full", Name: "full-policy"}},
			}), nil
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	_, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyArgs]{
		Inputs: NodePolicyArgs{
			Name: "full-policy",
			InstanceTypes: &LabelSelectorArgs{
				MatchLabels: map[string]string{"karpenter.k8s.aws/instance-type": "m5.large"},
			},
			CapacityTypes: &LabelSelectorArgs{
				MatchLabels: map[string]string{"karpenter.sh/capacity-type": "spot"},
			},
			Disruption: &DisruptionPolicyArgs{
				ConsolidationPolicy: "WhenEmpty",
				ConsolidateAfter:    "30m",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodePolicy_Create_WithRawYAML(t *testing.T) {
	rec := &mockRecommendationClientNodePolicy{
		createNodePoliciesFn: func(_ context.Context, req *connect.Request[apiv1.CreateNodePoliciesRequest]) (*connect.Response[apiv1.CreateNodePoliciesResponse], error) {
			p := req.Msg.Policies[0]
			if len(p.Raw) != 1 {
				t.Errorf("expected 1 raw spec, got %d", len(p.Raw))
			}
			if p.Raw[0].NodepoolYaml == "" {
				t.Error("NodepoolYaml should not be empty")
			}
			return connect.NewResponse(&apiv1.CreateNodePoliciesResponse{
				Policies: []*apiv1.NodePolicy{{Id: "np-raw", Name: "raw-policy"}},
			}), nil
		},
	}
	withMockNodePolicyClientSet(t, rec)

	n := &NodePolicy{}
	_, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyArgs]{
		Inputs: NodePolicyArgs{
			Name: "raw-policy",
			Raw: []RawKarpenterSpecArgs{
				{NodepoolYaml: "apiVersion: karpenter.sh/v1\nkind: NodePool\nmetadata:\n  name: custom"},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// strPtr is a test helper that returns a pointer to a string value.
func strPtr(s string) *string { return &s }
