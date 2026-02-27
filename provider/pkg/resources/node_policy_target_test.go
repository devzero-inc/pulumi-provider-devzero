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

// mockRecommendationClientNPT extends mockRecommendationClientNodePolicy with NodePolicyTarget methods.
type mockRecommendationClientNPT struct {
	mockRecommendationClientNodePolicy
	createNodePolicyTargetsFn func(context.Context, *connect.Request[apiv1.CreateNodePolicyTargetsRequest]) (*connect.Response[apiv1.CreateNodePolicyTargetsResponse], error)
	listNodePolicyTargetsFn   func(context.Context, *connect.Request[apiv1.ListNodePolicyTargetsRequest]) (*connect.Response[apiv1.ListNodePolicyTargetsResponse], error)
	updateNodePolicyTargetFn  func(context.Context, *connect.Request[apiv1.UpdateNodePolicyTargetRequest]) (*connect.Response[apiv1.UpdateNodePolicyTargetResponse], error)
}

func (m *mockRecommendationClientNPT) CreateNodePolicyTargets(ctx context.Context, req *connect.Request[apiv1.CreateNodePolicyTargetsRequest]) (*connect.Response[apiv1.CreateNodePolicyTargetsResponse], error) {
	return m.createNodePolicyTargetsFn(ctx, req)
}
func (m *mockRecommendationClientNPT) ListNodePolicyTargets(ctx context.Context, req *connect.Request[apiv1.ListNodePolicyTargetsRequest]) (*connect.Response[apiv1.ListNodePolicyTargetsResponse], error) {
	return m.listNodePolicyTargetsFn(ctx, req)
}
func (m *mockRecommendationClientNPT) UpdateNodePolicyTarget(ctx context.Context, req *connect.Request[apiv1.UpdateNodePolicyTargetRequest]) (*connect.Response[apiv1.UpdateNodePolicyTargetResponse], error) {
	return m.updateNodePolicyTargetFn(ctx, req)
}

func withMockNPTClientSet(t *testing.T, rec *mockRecommendationClientNPT) {
	t.Helper()
	prev := clientset.Get()
	t.Cleanup(func() { clientset.Set(prev) })
	clientset.Set(&clientset.ClientSet{
		TeamID:               "team-test",
		RecommendationClient: rec,
	})
}

// ---------- Create ----------

func TestNodePolicyTarget_Create_Success(t *testing.T) {
	rec := &mockRecommendationClientNPT{
		createNodePolicyTargetsFn: func(_ context.Context, req *connect.Request[apiv1.CreateNodePolicyTargetsRequest]) (*connect.Response[apiv1.CreateNodePolicyTargetsResponse], error) {
			if len(req.Msg.Targets) != 1 {
				t.Fatalf("expected 1 target, got %d", len(req.Msg.Targets))
			}
			tgt := req.Msg.Targets[0]
			if tgt.Name != "prod-target" {
				t.Errorf("Name: got %q, want %q", tgt.Name, "prod-target")
			}
			if tgt.PolicyId != "np-123" {
				t.Errorf("PolicyId: got %q, want %q", tgt.PolicyId, "np-123")
			}
			if tgt.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", tgt.TeamId, "team-test")
			}
			if len(tgt.ClusterIds) != 2 {
				t.Errorf("ClusterIds: got %v", tgt.ClusterIds)
			}
			if !tgt.Enabled {
				t.Error("Enabled should be true")
			}
			return connect.NewResponse(&apiv1.CreateNodePolicyTargetsResponse{
				Targets: []*apiv1.NodePolicyTarget{
					{
						TargetId:   "npt-456",
						Name:       "prod-target",
						PolicyId:   "np-123",
						TeamId:     "team-test",
						ClusterIds: []string{"c-1", "c-2"},
						Enabled:    true,
					},
				},
			}), nil
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	resp, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyTargetArgs]{
		Name: "prod-target",
		Inputs: NodePolicyTargetArgs{
			Name:       "prod-target",
			PolicyId:   "np-123",
			ClusterIds: []string{"c-1", "c-2"},
			Enabled:    true,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "npt-456" {
		t.Errorf("ID: got %q, want %q", resp.ID, "npt-456")
	}
	if resp.Output.Name != "prod-target" {
		t.Errorf("Name: got %q, want %q", resp.Output.Name, "prod-target")
	}
	if resp.Output.PolicyId != "np-123" {
		t.Errorf("PolicyId: got %q, want %q", resp.Output.PolicyId, "np-123")
	}
}

func TestNodePolicyTarget_Create_WithDescription(t *testing.T) {
	desc := "applies to all prod clusters"
	rec := &mockRecommendationClientNPT{
		createNodePolicyTargetsFn: func(_ context.Context, req *connect.Request[apiv1.CreateNodePolicyTargetsRequest]) (*connect.Response[apiv1.CreateNodePolicyTargetsResponse], error) {
			tgt := req.Msg.Targets[0]
			if tgt.Description != desc {
				t.Errorf("Description: got %q, want %q", tgt.Description, desc)
			}
			return connect.NewResponse(&apiv1.CreateNodePolicyTargetsResponse{
				Targets: []*apiv1.NodePolicyTarget{
					{TargetId: "npt-789", Name: "desc-target", PolicyId: "np-abc", Description: desc, ClusterIds: []string{"c-1"}},
				},
			}), nil
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	resp, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyTargetArgs]{
		Inputs: NodePolicyTargetArgs{
			Name:        "desc-target",
			PolicyId:    "np-abc",
			ClusterIds:  []string{"c-1"},
			Description: &desc,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Description == nil || *resp.Output.Description != desc {
		t.Errorf("Description roundtrip failed")
	}
}

func TestNodePolicyTarget_Create_Preview(t *testing.T) {
	n := &NodePolicyTarget{}
	resp, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyTargetArgs]{
		DryRun: true,
		Inputs: NodePolicyTargetArgs{
			Name:       "preview-target",
			PolicyId:   "np-preview",
			ClusterIds: []string{"c-1"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID on preview, got %q", resp.ID)
	}
	if resp.Output.Name != "preview-target" {
		t.Errorf("Name: got %q, want %q", resp.Output.Name, "preview-target")
	}
}

func TestNodePolicyTarget_Create_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	t.Cleanup(func() { clientset.Set(prev) })

	n := &NodePolicyTarget{}
	_, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyTargetArgs]{
		Inputs: NodePolicyTargetArgs{Name: "x", PolicyId: "p", ClusterIds: []string{"c"}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodePolicyTarget_Create_APIError(t *testing.T) {
	rec := &mockRecommendationClientNPT{
		createNodePolicyTargetsFn: func(_ context.Context, _ *connect.Request[apiv1.CreateNodePolicyTargetsRequest]) (*connect.Response[apiv1.CreateNodePolicyTargetsResponse], error) {
			return nil, errors.New("api error")
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	_, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyTargetArgs]{
		Inputs: NodePolicyTargetArgs{Name: "x", PolicyId: "p", ClusterIds: []string{"c"}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodePolicyTarget_Create_EmptyResponse(t *testing.T) {
	rec := &mockRecommendationClientNPT{
		createNodePolicyTargetsFn: func(_ context.Context, _ *connect.Request[apiv1.CreateNodePolicyTargetsRequest]) (*connect.Response[apiv1.CreateNodePolicyTargetsResponse], error) {
			return connect.NewResponse(&apiv1.CreateNodePolicyTargetsResponse{}), nil
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	_, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyTargetArgs]{
		Inputs: NodePolicyTargetArgs{Name: "x", PolicyId: "p", ClusterIds: []string{"c"}},
	})
	if err == nil {
		t.Fatal("expected error on empty response, got nil")
	}
}

// ---------- Read ----------

func TestNodePolicyTarget_Read_Success(t *testing.T) {
	rec := &mockRecommendationClientNPT{
		listNodePolicyTargetsFn: func(_ context.Context, req *connect.Request[apiv1.ListNodePolicyTargetsRequest]) (*connect.Response[apiv1.ListNodePolicyTargetsResponse], error) {
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			return connect.NewResponse(&apiv1.ListNodePolicyTargetsResponse{
				Targets: []*apiv1.NodePolicyTarget{
					{TargetId: "npt-456", Name: "prod-target", PolicyId: "np-123", ClusterIds: []string{"c-1"}, Enabled: true},
					{TargetId: "npt-999", Name: "other-target"},
				},
			}), nil
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	resp, err := n.Read(context.Background(), infer.ReadRequest[NodePolicyTargetArgs, NodePolicyTargetState]{
		ID:     "npt-456",
		Inputs: NodePolicyTargetArgs{Name: "prod-target"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "npt-456" {
		t.Errorf("ID: got %q, want %q", resp.ID, "npt-456")
	}
	if resp.State.Name != "prod-target" {
		t.Errorf("Name: got %q, want %q", resp.State.Name, "prod-target")
	}
	if resp.State.PolicyId != "np-123" {
		t.Errorf("PolicyId: got %q, want %q", resp.State.PolicyId, "np-123")
	}
	if !resp.State.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestNodePolicyTarget_Read_NotFound(t *testing.T) {
	rec := &mockRecommendationClientNPT{
		listNodePolicyTargetsFn: func(_ context.Context, _ *connect.Request[apiv1.ListNodePolicyTargetsRequest]) (*connect.Response[apiv1.ListNodePolicyTargetsResponse], error) {
			return connect.NewResponse(&apiv1.ListNodePolicyTargetsResponse{
				Targets: []*apiv1.NodePolicyTarget{{TargetId: "other", Name: "other"}},
			}), nil
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	_, err := n.Read(context.Background(), infer.ReadRequest[NodePolicyTargetArgs, NodePolicyTargetState]{ID: "npt-missing"})
	if err == nil {
		t.Fatal("expected error for not-found, got nil")
	}
}

func TestNodePolicyTarget_Read_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	t.Cleanup(func() { clientset.Set(prev) })

	n := &NodePolicyTarget{}
	_, err := n.Read(context.Background(), infer.ReadRequest[NodePolicyTargetArgs, NodePolicyTargetState]{ID: "npt-456"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodePolicyTarget_Read_APIError(t *testing.T) {
	rec := &mockRecommendationClientNPT{
		listNodePolicyTargetsFn: func(_ context.Context, _ *connect.Request[apiv1.ListNodePolicyTargetsRequest]) (*connect.Response[apiv1.ListNodePolicyTargetsResponse], error) {
			return nil, errors.New("api error")
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	_, err := n.Read(context.Background(), infer.ReadRequest[NodePolicyTargetArgs, NodePolicyTargetState]{ID: "npt-456"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------- Update ----------

func TestNodePolicyTarget_Update_Success(t *testing.T) {
	rec := &mockRecommendationClientNPT{
		updateNodePolicyTargetFn: func(_ context.Context, req *connect.Request[apiv1.UpdateNodePolicyTargetRequest]) (*connect.Response[apiv1.UpdateNodePolicyTargetResponse], error) {
			tgt := req.Msg.Target
			if tgt == nil {
				t.Fatal("target is nil")
			}
			if tgt.TargetId != "npt-456" {
				t.Errorf("TargetId: got %q, want %q", tgt.TargetId, "npt-456")
			}
			if tgt.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", tgt.TeamId, "team-test")
			}
			if tgt.Name != "updated-target" {
				t.Errorf("Name: got %q, want %q", tgt.Name, "updated-target")
			}
			return connect.NewResponse(&apiv1.UpdateNodePolicyTargetResponse{
				Target: &apiv1.NodePolicyTarget{
					TargetId: "npt-456",
					Name:     "updated-target",
					PolicyId: "np-123",
					TeamId:   "team-test",
					Enabled:  true,
				},
			}), nil
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	resp, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyTargetArgs, NodePolicyTargetState]{
		ID:     "npt-456",
		Inputs: NodePolicyTargetArgs{Name: "updated-target", PolicyId: "np-123", ClusterIds: []string{"c-1"}, Enabled: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Name != "updated-target" {
		t.Errorf("Name: got %q, want %q", resp.Output.Name, "updated-target")
	}
}

func TestNodePolicyTarget_Update_Disable(t *testing.T) {
	rec := &mockRecommendationClientNPT{
		updateNodePolicyTargetFn: func(_ context.Context, req *connect.Request[apiv1.UpdateNodePolicyTargetRequest]) (*connect.Response[apiv1.UpdateNodePolicyTargetResponse], error) {
			if req.Msg.Target.Enabled {
				t.Error("Enabled should be false")
			}
			return connect.NewResponse(&apiv1.UpdateNodePolicyTargetResponse{
				Target: &apiv1.NodePolicyTarget{TargetId: "npt-456", Name: "staging", PolicyId: "np-123", Enabled: false},
			}), nil
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	resp, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyTargetArgs, NodePolicyTargetState]{
		ID:     "npt-456",
		Inputs: NodePolicyTargetArgs{Name: "staging", PolicyId: "np-123", ClusterIds: []string{"c-1"}, Enabled: false},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Enabled {
		t.Error("Enabled should be false after update")
	}
}

func TestNodePolicyTarget_Update_Preview(t *testing.T) {
	n := &NodePolicyTarget{}
	resp, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyTargetArgs, NodePolicyTargetState]{
		ID:     "npt-456",
		DryRun: true,
		Inputs: NodePolicyTargetArgs{Name: "preview-update", PolicyId: "np-123", ClusterIds: []string{"c-1"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Name != "preview-update" {
		t.Errorf("Name: got %q, want %q", resp.Output.Name, "preview-update")
	}
}

func TestNodePolicyTarget_Update_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	t.Cleanup(func() { clientset.Set(prev) })

	n := &NodePolicyTarget{}
	_, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyTargetArgs, NodePolicyTargetState]{
		ID:     "npt-456",
		Inputs: NodePolicyTargetArgs{Name: "x", PolicyId: "p", ClusterIds: []string{"c"}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodePolicyTarget_Update_APIError(t *testing.T) {
	rec := &mockRecommendationClientNPT{
		updateNodePolicyTargetFn: func(_ context.Context, _ *connect.Request[apiv1.UpdateNodePolicyTargetRequest]) (*connect.Response[apiv1.UpdateNodePolicyTargetResponse], error) {
			return nil, errors.New("api error")
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	_, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyTargetArgs, NodePolicyTargetState]{
		ID:     "npt-456",
		Inputs: NodePolicyTargetArgs{Name: "x", PolicyId: "p", ClusterIds: []string{"c"}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNodePolicyTarget_Update_EmptyResponse(t *testing.T) {
	rec := &mockRecommendationClientNPT{
		updateNodePolicyTargetFn: func(_ context.Context, _ *connect.Request[apiv1.UpdateNodePolicyTargetRequest]) (*connect.Response[apiv1.UpdateNodePolicyTargetResponse], error) {
			return connect.NewResponse(&apiv1.UpdateNodePolicyTargetResponse{}), nil
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	_, err := n.Update(context.Background(), infer.UpdateRequest[NodePolicyTargetArgs, NodePolicyTargetState]{
		ID:     "npt-456",
		Inputs: NodePolicyTargetArgs{Name: "x", PolicyId: "p", ClusterIds: []string{"c"}},
	})
	if err == nil {
		t.Fatal("expected error on empty response, got nil")
	}
}

// ---------- Delete ----------

func TestNodePolicyTarget_Delete_StateOnly(t *testing.T) {
	n := &NodePolicyTarget{}
	_, err := n.Delete(context.Background(), infer.DeleteRequest[NodePolicyTargetState]{
		ID:    "npt-456",
		State: NodePolicyTargetState{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodePolicyTarget_Delete_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	t.Cleanup(func() { clientset.Set(prev) })

	n := &NodePolicyTarget{}
	_, err := n.Delete(context.Background(), infer.DeleteRequest[NodePolicyTargetState]{ID: "npt-456"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- proto conversion roundtrip ----------

func TestNodePolicyTarget_ProtoRoundtrip(t *testing.T) {
	desc := "all prod clusters"
	input := NodePolicyTargetArgs{
		Name:        "prod-target",
		PolicyId:    "np-123",
		ClusterIds:  []string{"c-1", "c-2", "c-3"},
		Description: &desc,
		Enabled:     true,
	}

	proto := nodePolicyTargetArgsToProto("team-abc", "npt-456", input)
	if proto.TargetId != "npt-456" {
		t.Errorf("TargetId: got %q", proto.TargetId)
	}
	if proto.TeamId != "team-abc" {
		t.Errorf("TeamId: got %q", proto.TeamId)
	}
	if proto.Name != "prod-target" {
		t.Errorf("Name: got %q", proto.Name)
	}
	if proto.Description != desc {
		t.Errorf("Description: got %q", proto.Description)
	}
	if len(proto.ClusterIds) != 3 {
		t.Errorf("ClusterIds: got %v", proto.ClusterIds)
	}
	if !proto.Enabled {
		t.Error("Enabled should be true")
	}

	got := nodePolicyTargetProtoToArgs(proto)
	if got.Name != input.Name {
		t.Errorf("Name roundtrip: got %q", got.Name)
	}
	if got.PolicyId != input.PolicyId {
		t.Errorf("PolicyId roundtrip: got %q", got.PolicyId)
	}
	if len(got.ClusterIds) != 3 {
		t.Errorf("ClusterIds roundtrip: got %v", got.ClusterIds)
	}
	if got.Description == nil || *got.Description != desc {
		t.Errorf("Description roundtrip failed")
	}
	if !got.Enabled {
		t.Error("Enabled roundtrip failed")
	}
}

func TestNodePolicyTarget_MultipleClusterIds(t *testing.T) {
	clusterIDs := []string{"c-us-east-1", "c-us-west-2", "c-eu-west-1"}
	rec := &mockRecommendationClientNPT{
		createNodePolicyTargetsFn: func(_ context.Context, req *connect.Request[apiv1.CreateNodePolicyTargetsRequest]) (*connect.Response[apiv1.CreateNodePolicyTargetsResponse], error) {
			tgt := req.Msg.Targets[0]
			if len(tgt.ClusterIds) != 3 {
				t.Errorf("expected 3 cluster IDs, got %d: %v", len(tgt.ClusterIds), tgt.ClusterIds)
			}
			return connect.NewResponse(&apiv1.CreateNodePolicyTargetsResponse{
				Targets: []*apiv1.NodePolicyTarget{
					{TargetId: "npt-multi", Name: "multi-cluster", PolicyId: "np-123", ClusterIds: clusterIDs},
				},
			}), nil
		},
	}
	withMockNPTClientSet(t, rec)

	n := &NodePolicyTarget{}
	resp, err := n.Create(context.Background(), infer.CreateRequest[NodePolicyTargetArgs]{
		Inputs: NodePolicyTargetArgs{
			Name:       "multi-cluster",
			PolicyId:   "np-123",
			ClusterIds: clusterIDs,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Output.ClusterIds) != 3 {
		t.Errorf("ClusterIds: got %v", resp.Output.ClusterIds)
	}
}
