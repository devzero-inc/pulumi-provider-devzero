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

// mockRecommendationClientWithTarget extends mockRecommendationClient with target CRUD.
// We add methods to the existing mock via a separate type so the two test files don't conflict.
type mockRecommendationClientFull struct {
	mockRecommendationClient
	createTargetFn func(context.Context, *connect.Request[apiv1.CreateWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.CreateWorkloadPolicyTargetResponse], error)
	getTargetFn    func(context.Context, *connect.Request[apiv1.GetWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.GetWorkloadPolicyTargetResponse], error)
	updateTargetFn func(context.Context, *connect.Request[apiv1.UpdateWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.UpdateWorkloadPolicyTargetResponse], error)
	deleteTargetFn func(context.Context, *connect.Request[apiv1.DeleteWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.DeleteWorkloadPolicyTargetResponse], error)
}

func (m *mockRecommendationClientFull) CreateWorkloadPolicyTarget(ctx context.Context, req *connect.Request[apiv1.CreateWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.CreateWorkloadPolicyTargetResponse], error) {
	return m.createTargetFn(ctx, req)
}
func (m *mockRecommendationClientFull) GetWorkloadPolicyTarget(ctx context.Context, req *connect.Request[apiv1.GetWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.GetWorkloadPolicyTargetResponse], error) {
	return m.getTargetFn(ctx, req)
}
func (m *mockRecommendationClientFull) UpdateWorkloadPolicyTarget(ctx context.Context, req *connect.Request[apiv1.UpdateWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.UpdateWorkloadPolicyTargetResponse], error) {
	return m.updateTargetFn(ctx, req)
}
func (m *mockRecommendationClientFull) DeleteWorkloadPolicyTarget(ctx context.Context, req *connect.Request[apiv1.DeleteWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.DeleteWorkloadPolicyTargetResponse], error) {
	return m.deleteTargetFn(ctx, req)
}

func withMockTargetClientSet(t *testing.T, rec *mockRecommendationClientFull) {
	t.Helper()
	prev := clientset.Get()
	t.Cleanup(func() { clientset.Set(prev) })
	clientset.Set(&clientset.ClientSet{
		TeamID:               "team-test",
		RecommendationClient: rec,
	})
}

// ---------- Create ----------

func TestWorkloadPolicyTarget_Create_Success(t *testing.T) {
	rec := &mockRecommendationClientFull{
		createTargetFn: func(_ context.Context, req *connect.Request[apiv1.CreateWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.CreateWorkloadPolicyTargetResponse], error) {
			if req.Msg.Name != "my-target" {
				t.Errorf("Name: got %q, want %q", req.Msg.Name, "my-target")
			}
			if req.Msg.PolicyId != "policy-abc" {
				t.Errorf("PolicyId: got %q, want %q", req.Msg.PolicyId, "policy-abc")
			}
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			if len(req.Msg.ClusterIds) != 1 || req.Msg.ClusterIds[0] != "cluster-1" {
				t.Errorf("ClusterIds: got %v", req.Msg.ClusterIds)
			}
			return connect.NewResponse(&apiv1.CreateWorkloadPolicyTargetResponse{
				Target: &apiv1.WorkloadPolicyTarget{
					TargetId:   "target-123",
					Name:       "my-target",
					PolicyId:   "policy-abc",
					ClusterIds: []string{"cluster-1"},
				},
			}), nil
		},
	}
	withMockTargetClientSet(t, rec)

	w := &WorkloadPolicyTarget{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyTargetArgs]{
		Name: "my-target",
		Inputs: WorkloadPolicyTargetArgs{
			Name:       "my-target",
			PolicyId:   "policy-abc",
			ClusterIds: []string{"cluster-1"},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "target-123" {
		t.Errorf("id: got %q, want %q", resp.ID, "target-123")
	}
	if resp.Output.Name != "my-target" {
		t.Errorf("name: got %q, want %q", resp.Output.Name, "my-target")
	}
	if resp.Output.PolicyId != "policy-abc" {
		t.Errorf("policyId: got %q, want %q", resp.Output.PolicyId, "policy-abc")
	}
}

func TestWorkloadPolicyTarget_Create_Preview_SkipsAPI(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadPolicyTarget{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyTargetArgs]{
		Name: "preview-target",
		Inputs: WorkloadPolicyTargetArgs{
			Name:       "preview-target",
			PolicyId:   "policy-abc",
			ClusterIds: []string{"cluster-1"},
		},
		DryRun: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("preview id should be empty, got %q", resp.ID)
	}
	if resp.Output.Name != "preview-target" {
		t.Errorf("preview name: got %q, want %q", resp.Output.Name, "preview-target")
	}
}

func TestWorkloadPolicyTarget_Create_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadPolicyTarget{}
	_, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyTargetArgs]{
		Inputs: WorkloadPolicyTargetArgs{Name: "x", PolicyId: "p", ClusterIds: []string{"c"}},
	})
	if err == nil {
		t.Fatal("expected error when ClientSet is nil")
	}
}

func TestWorkloadPolicyTarget_Create_APIError(t *testing.T) {
	rec := &mockRecommendationClientFull{
		createTargetFn: func(_ context.Context, _ *connect.Request[apiv1.CreateWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.CreateWorkloadPolicyTargetResponse], error) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("server error"))
		},
	}
	withMockTargetClientSet(t, rec)

	w := &WorkloadPolicyTarget{}
	_, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyTargetArgs]{
		Inputs: WorkloadPolicyTargetArgs{Name: "x", PolicyId: "p", ClusterIds: []string{"c"}},
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestWorkloadPolicyTarget_Create_WithAllSelectors(t *testing.T) {
	rec := &mockRecommendationClientFull{
		createTargetFn: func(_ context.Context, req *connect.Request[apiv1.CreateWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.CreateWorkloadPolicyTargetResponse], error) {
			if req.Msg.NamePattern == nil || req.Msg.NamePattern.Pattern != "^api-" {
				t.Errorf("NamePattern: got %v", req.Msg.NamePattern)
			}
			if req.Msg.NamespaceSelector == nil {
				t.Error("NamespaceSelector should be set")
			} else if req.Msg.NamespaceSelector.MatchLabels["env"] != "prod" {
				t.Errorf("NamespaceSelector.MatchLabels: got %v", req.Msg.NamespaceSelector.MatchLabels)
			}
			if len(req.Msg.KindFilter) != 2 {
				t.Errorf("KindFilter: got %d items, want 2", len(req.Msg.KindFilter))
			}
			return connect.NewResponse(&apiv1.CreateWorkloadPolicyTargetResponse{
				Target: &apiv1.WorkloadPolicyTarget{
					TargetId:          "target-xyz",
					Name:              "full-target",
					PolicyId:          "policy-abc",
					ClusterIds:        []string{"cluster-1"},
					NamePattern:       req.Msg.NamePattern,
					NamespaceSelector: req.Msg.NamespaceSelector,
					KindFilter:        req.Msg.KindFilter,
				},
			}), nil
		},
	}
	withMockTargetClientSet(t, rec)

	flags := "i"
	w := &WorkloadPolicyTarget{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyTargetArgs]{
		Inputs: WorkloadPolicyTargetArgs{
			Name:       "full-target",
			PolicyId:   "policy-abc",
			ClusterIds: []string{"cluster-1"},
			KindFilter: []string{"Deployment", "StatefulSet"},
			NamePattern: &NamePatternArgs{
				Pattern: "^api-",
				Flags:   &flags,
			},
			NamespaceSelector: &LabelSelectorArgs{
				MatchLabels: map[string]string{"env": "prod"},
			},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "target-xyz" {
		t.Errorf("id: got %q, want %q", resp.ID, "target-xyz")
	}
	if resp.Output.NamePattern == nil || resp.Output.NamePattern.Pattern != "^api-" {
		t.Errorf("NamePattern roundtrip: got %v", resp.Output.NamePattern)
	}
	if len(resp.Output.KindFilter) != 2 {
		t.Errorf("KindFilter roundtrip: got %v", resp.Output.KindFilter)
	}
}

// ---------- Read ----------

func TestWorkloadPolicyTarget_Read_Success(t *testing.T) {
	desc := "my description"
	rec := &mockRecommendationClientFull{
		getTargetFn: func(_ context.Context, req *connect.Request[apiv1.GetWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.GetWorkloadPolicyTargetResponse], error) {
			if req.Msg.TargetId != "target-123" {
				t.Errorf("TargetId: got %q, want %q", req.Msg.TargetId, "target-123")
			}
			return connect.NewResponse(&apiv1.GetWorkloadPolicyTargetResponse{
				Target: &apiv1.WorkloadPolicyTarget{
					TargetId:    "target-123",
					Name:        "my-target",
					PolicyId:    "policy-abc",
					ClusterIds:  []string{"cluster-1"},
					Description: desc,
					Enabled:     true,
					Priority:    5,
				},
			}), nil
		},
	}
	withMockTargetClientSet(t, rec)

	w := &WorkloadPolicyTarget{}
	resp, err := w.Read(context.Background(), infer.ReadRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{
		ID: "target-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Inputs.Name != "my-target" {
		t.Errorf("name: got %q, want %q", resp.Inputs.Name, "my-target")
	}
	if resp.Inputs.Description == nil || *resp.Inputs.Description != desc {
		t.Errorf("description: got %v, want %q", resp.Inputs.Description, desc)
	}
	if !resp.Inputs.Enabled {
		t.Error("enabled should be true")
	}
	if resp.Inputs.Priority != 5 {
		t.Errorf("priority: got %d, want 5", resp.Inputs.Priority)
	}
}

func TestWorkloadPolicyTarget_Read_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadPolicyTarget{}
	_, err := w.Read(context.Background(), infer.ReadRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{
		ID: "target-123",
	})
	if err == nil {
		t.Fatal("expected error when ClientSet is nil")
	}
}

func TestWorkloadPolicyTarget_Read_APIError(t *testing.T) {
	rec := &mockRecommendationClientFull{
		getTargetFn: func(_ context.Context, _ *connect.Request[apiv1.GetWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.GetWorkloadPolicyTargetResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("not found"))
		},
	}
	withMockTargetClientSet(t, rec)

	w := &WorkloadPolicyTarget{}
	_, err := w.Read(context.Background(), infer.ReadRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{
		ID: "target-123",
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

// ---------- Update ----------

func TestWorkloadPolicyTarget_Update_Success(t *testing.T) {
	rec := &mockRecommendationClientFull{
		updateTargetFn: func(_ context.Context, req *connect.Request[apiv1.UpdateWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.UpdateWorkloadPolicyTargetResponse], error) {
			if req.Msg.TargetId != "target-123" {
				t.Errorf("TargetId: got %q, want %q", req.Msg.TargetId, "target-123")
			}
			return connect.NewResponse(&apiv1.UpdateWorkloadPolicyTargetResponse{
				Target: &apiv1.WorkloadPolicyTarget{
					TargetId:   "target-123",
					Name:       req.Msg.Name,
					PolicyId:   "policy-abc",
					ClusterIds: req.Msg.ClusterIds,
				},
			}), nil
		},
	}
	withMockTargetClientSet(t, rec)

	w := &WorkloadPolicyTarget{}
	resp, err := w.Update(context.Background(), infer.UpdateRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{
		ID:    "target-123",
		State: WorkloadPolicyTargetState{WorkloadPolicyTargetArgs: WorkloadPolicyTargetArgs{Name: "old", PolicyId: "policy-abc", ClusterIds: []string{"cluster-1"}}},
		Inputs: WorkloadPolicyTargetArgs{
			Name:       "new-name",
			PolicyId:   "policy-abc",
			ClusterIds: []string{"cluster-1"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Name != "new-name" {
		t.Errorf("name: got %q, want %q", resp.Output.Name, "new-name")
	}
}

func TestWorkloadPolicyTarget_Update_Preview_SkipsAPI(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadPolicyTarget{}
	resp, err := w.Update(context.Background(), infer.UpdateRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{
		ID:     "target-123",
		State:  WorkloadPolicyTargetState{WorkloadPolicyTargetArgs: WorkloadPolicyTargetArgs{Name: "old", PolicyId: "p", ClusterIds: []string{"c"}}},
		Inputs: WorkloadPolicyTargetArgs{Name: "new", PolicyId: "p", ClusterIds: []string{"c"}},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Name != "new" {
		t.Errorf("preview name: got %q, want %q", resp.Output.Name, "new")
	}
}

func TestWorkloadPolicyTarget_Update_APIError(t *testing.T) {
	rec := &mockRecommendationClientFull{
		updateTargetFn: func(_ context.Context, _ *connect.Request[apiv1.UpdateWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.UpdateWorkloadPolicyTargetResponse], error) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("server error"))
		},
	}
	withMockTargetClientSet(t, rec)

	w := &WorkloadPolicyTarget{}
	_, err := w.Update(context.Background(), infer.UpdateRequest[WorkloadPolicyTargetArgs, WorkloadPolicyTargetState]{
		ID:     "target-123",
		Inputs: WorkloadPolicyTargetArgs{Name: "x", PolicyId: "p", ClusterIds: []string{"c"}},
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

// ---------- Delete ----------

func TestWorkloadPolicyTarget_Delete_Success(t *testing.T) {
	rec := &mockRecommendationClientFull{
		deleteTargetFn: func(_ context.Context, req *connect.Request[apiv1.DeleteWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.DeleteWorkloadPolicyTargetResponse], error) {
			if len(req.Msg.TargetIds) != 1 || req.Msg.TargetIds[0] != "target-123" {
				t.Errorf("TargetIds: got %v", req.Msg.TargetIds)
			}
			return connect.NewResponse(&apiv1.DeleteWorkloadPolicyTargetResponse{Success: true}), nil
		},
	}
	withMockTargetClientSet(t, rec)

	w := &WorkloadPolicyTarget{}
	_, err := w.Delete(context.Background(), infer.DeleteRequest[WorkloadPolicyTargetState]{
		ID: "target-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkloadPolicyTarget_Delete_APIError(t *testing.T) {
	rec := &mockRecommendationClientFull{
		deleteTargetFn: func(_ context.Context, _ *connect.Request[apiv1.DeleteWorkloadPolicyTargetRequest]) (*connect.Response[apiv1.DeleteWorkloadPolicyTargetResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("not found"))
		},
	}
	withMockTargetClientSet(t, rec)

	w := &WorkloadPolicyTarget{}
	_, err := w.Delete(context.Background(), infer.DeleteRequest[WorkloadPolicyTargetState]{
		ID: "target-123",
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

// ---------- conversion helpers ----------

func TestKindFilterRoundTrip(t *testing.T) {
	kinds := []string{"Pod", "Job", "Deployment", "StatefulSet", "DaemonSet", "ReplicaSet", "CronJob", "ReplicationController", "Rollout"}
	back := kindFilterFromProto(kindFilterToProto(kinds))
	if len(back) != len(kinds) {
		t.Fatalf("len: got %d, want %d", len(back), len(kinds))
	}
	for i, want := range kinds {
		if back[i] != want {
			t.Errorf("[%d]: got %q, want %q", i, back[i], want)
		}
	}
}

func TestNamePatternRoundTrip(t *testing.T) {
	flags := "i"
	n := &NamePatternArgs{Pattern: "^api-.*", Flags: &flags}
	back := namePatternFromProto(namePatternToProto(n))
	if back == nil {
		t.Fatal("expected non-nil")
	}
	if back.Pattern != "^api-.*" {
		t.Errorf("Pattern: got %q, want %q", back.Pattern, "^api-.*")
	}
	if back.Flags == nil || *back.Flags != flags {
		t.Errorf("Flags: got %v, want %q", back.Flags, flags)
	}
}

func TestLabelSelectorRoundTrip(t *testing.T) {
	s := &LabelSelectorArgs{
		MatchLabels: map[string]string{"app": "api", "env": "prod"},
		MatchExpressions: []LabelSelectorRequirementArgs{
			{Key: "tier", Operator: "In", Values: []string{"frontend", "backend"}},
			{Key: "legacy", Operator: "DoesNotExist"},
		},
	}
	back := labelSelectorFromProto(labelSelectorToProto(s))
	if back == nil {
		t.Fatal("expected non-nil")
	}
	if back.MatchLabels["app"] != "api" {
		t.Errorf("MatchLabels[app]: got %q, want %q", back.MatchLabels["app"], "api")
	}
	if len(back.MatchExpressions) != 2 {
		t.Fatalf("MatchExpressions len: got %d, want 2", len(back.MatchExpressions))
	}
	if back.MatchExpressions[0].Operator != "In" {
		t.Errorf("Operator: got %q, want In", back.MatchExpressions[0].Operator)
	}
	if back.MatchExpressions[1].Operator != "DoesNotExist" {
		t.Errorf("Operator: got %q, want DoesNotExist", back.MatchExpressions[1].Operator)
	}
}

func TestNamePattern_Nil(t *testing.T) {
	if namePatternToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
	if namePatternFromProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestLabelSelector_Nil(t *testing.T) {
	if labelSelectorToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
	if labelSelectorFromProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
}
