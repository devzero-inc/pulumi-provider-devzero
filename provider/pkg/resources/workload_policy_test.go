package resources

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	apiv1connect "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1/apiv1connect"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// mockRecommendationClient stubs only the policy CRUD methods we test.
type mockRecommendationClient struct {
	apiv1connect.K8SRecommendationServiceClient
	createFn func(context.Context, *connect.Request[apiv1.CreateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.CreateWorkloadRecommendationPolicyResponse], error)
	getFn    func(context.Context, *connect.Request[apiv1.GetWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.GetWorkloadRecommendationPolicyResponse], error)
	updateFn func(context.Context, *connect.Request[apiv1.UpdateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.UpdateWorkloadRecommendationPolicyResponse], error)
	deleteFn func(context.Context, *connect.Request[apiv1.DeleteWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.DeleteWorkloadRecommendationPolicyResponse], error)
}

func (m *mockRecommendationClient) CreateWorkloadRecommendationPolicy(ctx context.Context, req *connect.Request[apiv1.CreateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.CreateWorkloadRecommendationPolicyResponse], error) {
	return m.createFn(ctx, req)
}
func (m *mockRecommendationClient) GetWorkloadRecommendationPolicy(ctx context.Context, req *connect.Request[apiv1.GetWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.GetWorkloadRecommendationPolicyResponse], error) {
	return m.getFn(ctx, req)
}
func (m *mockRecommendationClient) UpdateWorkloadRecommendationPolicy(ctx context.Context, req *connect.Request[apiv1.UpdateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.UpdateWorkloadRecommendationPolicyResponse], error) {
	return m.updateFn(ctx, req)
}
func (m *mockRecommendationClient) DeleteWorkloadRecommendationPolicy(ctx context.Context, req *connect.Request[apiv1.DeleteWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.DeleteWorkloadRecommendationPolicyResponse], error) {
	return m.deleteFn(ctx, req)
}

func withMockRecommendationClientSet(t *testing.T, rec *mockRecommendationClient) {
	t.Helper()
	prev := clientset.Get()
	t.Cleanup(func() { clientset.Set(prev) })
	clientset.Set(&clientset.ClientSet{
		TeamID:               "team-test",
		RecommendationClient: rec,
	})
}

// ---------- Create ----------

func TestWorkloadPolicy_Create_Success(t *testing.T) {
	rec := &mockRecommendationClient{
		createFn: func(_ context.Context, req *connect.Request[apiv1.CreateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.CreateWorkloadRecommendationPolicyResponse], error) {
			if req.Msg.Policy.Name != "my-policy" {
				t.Errorf("Name: got %q, want %q", req.Msg.Policy.Name, "my-policy")
			}
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			return connect.NewResponse(&apiv1.CreateWorkloadRecommendationPolicyResponse{
				Policy: &apiv1.WorkloadRecommendationPolicy{
					PolicyId: "policy-123",
					Name:     "my-policy",
				},
			}), nil
		},
	}
	withMockRecommendationClientSet(t, rec)

	w := &WorkloadPolicy{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyArgs]{
		Name:   "my-policy",
		Inputs: WorkloadPolicyArgs{Name: "my-policy"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "policy-123" {
		t.Errorf("id: got %q, want %q", resp.ID, "policy-123")
	}
	if resp.Output.Name != "my-policy" {
		t.Errorf("name: got %q, want %q", resp.Output.Name, "my-policy")
	}
}

func TestWorkloadPolicy_Create_Preview_SkipsAPI(t *testing.T) {
	// Nil ClientSet — would panic if API were called.
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadPolicy{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyArgs]{
		Name:   "preview-policy",
		Inputs: WorkloadPolicyArgs{Name: "preview-policy"},
		DryRun: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("preview id should be empty, got %q", resp.ID)
	}
	if resp.Output.Name != "preview-policy" {
		t.Errorf("preview name: got %q, want %q", resp.Output.Name, "preview-policy")
	}
}

func TestWorkloadPolicy_Create_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadPolicy{}
	_, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyArgs]{
		Inputs: WorkloadPolicyArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error when ClientSet is nil")
	}
}

func TestWorkloadPolicy_Create_APIError(t *testing.T) {
	rec := &mockRecommendationClient{
		createFn: func(_ context.Context, _ *connect.Request[apiv1.CreateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.CreateWorkloadRecommendationPolicyResponse], error) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("server error"))
		},
	}
	withMockRecommendationClientSet(t, rec)

	w := &WorkloadPolicy{}
	_, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyArgs]{
		Inputs: WorkloadPolicyArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestWorkloadPolicy_Create_WithScaling(t *testing.T) {
	minReq := 500
	maxReq := 2000
	overhead := 0.1
	loopback := 3600
	minReplicas := 1
	maxReplicas := 5

	rec := &mockRecommendationClient{
		createFn: func(_ context.Context, req *connect.Request[apiv1.CreateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.CreateWorkloadRecommendationPolicyResponse], error) {
			cpu := req.Msg.Policy.CpuVerticalScaling
			if cpu == nil {
				t.Fatal("expected cpu_vertical_scaling to be set")
			}
			if !cpu.Enabled {
				t.Error("cpu scaling should be enabled")
			}
			if cpu.MinRequest == nil || *cpu.MinRequest != int64(minReq) {
				t.Errorf("cpu MinRequest: got %v, want %d", cpu.MinRequest, minReq)
			}
			if cpu.MaxRequest == nil || *cpu.MaxRequest != int64(maxReq) {
				t.Errorf("cpu MaxRequest: got %v, want %d", cpu.MaxRequest, maxReq)
			}
			hs := req.Msg.Policy.HorizontalScaling
			if hs == nil {
				t.Fatal("expected horizontal_scaling to be set")
			}
			if hs.MinReplicas == nil || *hs.MinReplicas != int32(minReplicas) {
				t.Errorf("MinReplicas: got %v, want %d", hs.MinReplicas, minReplicas)
			}
			lp := req.Msg.Policy.LoopbackPeriodSeconds
			if lp == nil || *lp != int32(loopback) {
				t.Errorf("LoopbackPeriodSeconds: got %v, want %d", lp, loopback)
			}
			return connect.NewResponse(&apiv1.CreateWorkloadRecommendationPolicyResponse{
				Policy: &apiv1.WorkloadRecommendationPolicy{
					PolicyId:           "policy-xyz",
					Name:               "scaling-policy",
					CpuVerticalScaling: cpu,
					HorizontalScaling:  hs,
				},
			}), nil
		},
	}
	withMockRecommendationClientSet(t, rec)

	w := &WorkloadPolicy{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyArgs]{
		Inputs: WorkloadPolicyArgs{
			Name:                  "scaling-policy",
			LoopbackPeriodSeconds: &loopback,
			CpuVerticalScaling: &VerticalScalingArgs{
				Enabled:            true,
				MinRequest:         &minReq,
				MaxRequest:         &maxReq,
				OverheadMultiplier: &overhead,
			},
			HorizontalScaling: &HorizontalScalingArgs{
				Enabled:     true,
				MinReplicas: &minReplicas,
				MaxReplicas: &maxReplicas,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "policy-xyz" {
		t.Errorf("id: got %q, want %q", resp.ID, "policy-xyz")
	}
}

func TestWorkloadPolicy_Create_WithActionAndDetectionTriggers(t *testing.T) {
	rec := &mockRecommendationClient{
		createFn: func(_ context.Context, req *connect.Request[apiv1.CreateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.CreateWorkloadRecommendationPolicyResponse], error) {
			if len(req.Msg.Policy.ActionTriggers) != 1 ||
				req.Msg.Policy.ActionTriggers[0] != apiv1.ActionTrigger_ACTION_TRIGGER_ON_DETECTION {
				t.Errorf("ActionTriggers: got %v", req.Msg.Policy.ActionTriggers)
			}
			if len(req.Msg.Policy.DetectionTriggers) != 2 {
				t.Errorf("DetectionTriggers: got %d, want 2", len(req.Msg.Policy.DetectionTriggers))
			}
			return connect.NewResponse(&apiv1.CreateWorkloadRecommendationPolicyResponse{
				Policy: &apiv1.WorkloadRecommendationPolicy{
					PolicyId:          "policy-triggers",
					Name:              "trigger-policy",
					ActionTriggers:    req.Msg.Policy.ActionTriggers,
					DetectionTriggers: req.Msg.Policy.DetectionTriggers,
				},
			}), nil
		},
	}
	withMockRecommendationClientSet(t, rec)

	w := &WorkloadPolicy{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadPolicyArgs]{
		Inputs: WorkloadPolicyArgs{
			Name:              "trigger-policy",
			ActionTriggers:    []string{"on_detection"},
			DetectionTriggers: []string{"pod_creation", "pod_update"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Output.ActionTriggers) != 1 || resp.Output.ActionTriggers[0] != "on_detection" {
		t.Errorf("action triggers roundtrip: got %v", resp.Output.ActionTriggers)
	}
	if len(resp.Output.DetectionTriggers) != 2 {
		t.Errorf("detection triggers roundtrip: got %v", resp.Output.DetectionTriggers)
	}
}

// ---------- Read ----------

func TestWorkloadPolicy_Read_Success(t *testing.T) {
	desc := "a description"
	rec := &mockRecommendationClient{
		getFn: func(_ context.Context, req *connect.Request[apiv1.GetWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.GetWorkloadRecommendationPolicyResponse], error) {
			if req.Msg.PolicyId != "policy-123" {
				t.Errorf("PolicyId: got %q, want %q", req.Msg.PolicyId, "policy-123")
			}
			return connect.NewResponse(&apiv1.GetWorkloadRecommendationPolicyResponse{
				Policy: &apiv1.WorkloadRecommendationPolicy{
					PolicyId:    "policy-123",
					Name:        "my-policy",
					Description: desc,
					ActionTriggers: []apiv1.ActionTrigger{
						apiv1.ActionTrigger_ACTION_TRIGGER_ON_DETECTION,
					},
				},
			}), nil
		},
	}
	withMockRecommendationClientSet(t, rec)

	w := &WorkloadPolicy{}
	resp, err := w.Read(context.Background(), infer.ReadRequest[WorkloadPolicyArgs, WorkloadPolicyState]{
		ID: "policy-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Inputs.Name != "my-policy" {
		t.Errorf("name: got %q, want %q", resp.Inputs.Name, "my-policy")
	}
	if resp.Inputs.Description == nil || *resp.Inputs.Description != desc {
		t.Errorf("description: got %v, want %q", resp.Inputs.Description, desc)
	}
	if len(resp.Inputs.ActionTriggers) != 1 || resp.Inputs.ActionTriggers[0] != "on_detection" {
		t.Errorf("action triggers: got %v", resp.Inputs.ActionTriggers)
	}
}

func TestWorkloadPolicy_Read_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadPolicy{}
	_, err := w.Read(context.Background(), infer.ReadRequest[WorkloadPolicyArgs, WorkloadPolicyState]{
		ID: "policy-123",
	})
	if err == nil {
		t.Fatal("expected error when ClientSet is nil")
	}
}

func TestWorkloadPolicy_Read_APIError(t *testing.T) {
	rec := &mockRecommendationClient{
		getFn: func(_ context.Context, _ *connect.Request[apiv1.GetWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.GetWorkloadRecommendationPolicyResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("not found"))
		},
	}
	withMockRecommendationClientSet(t, rec)

	w := &WorkloadPolicy{}
	_, err := w.Read(context.Background(), infer.ReadRequest[WorkloadPolicyArgs, WorkloadPolicyState]{
		ID: "policy-123",
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

// ---------- Update ----------

func TestWorkloadPolicy_Update_Success(t *testing.T) {
	rec := &mockRecommendationClient{
		updateFn: func(_ context.Context, req *connect.Request[apiv1.UpdateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.UpdateWorkloadRecommendationPolicyResponse], error) {
			if req.Msg.Policy.PolicyId != "policy-123" {
				t.Errorf("PolicyId: got %q, want %q", req.Msg.Policy.PolicyId, "policy-123")
			}
			return connect.NewResponse(&apiv1.UpdateWorkloadRecommendationPolicyResponse{
				Policy: &apiv1.WorkloadRecommendationPolicy{
					PolicyId: "policy-123",
					Name:     req.Msg.Policy.Name,
				},
			}), nil
		},
	}
	withMockRecommendationClientSet(t, rec)

	w := &WorkloadPolicy{}
	resp, err := w.Update(context.Background(), infer.UpdateRequest[WorkloadPolicyArgs, WorkloadPolicyState]{
		ID:     "policy-123",
		State:  WorkloadPolicyState{WorkloadPolicyArgs: WorkloadPolicyArgs{Name: "old-name"}},
		Inputs: WorkloadPolicyArgs{Name: "new-name"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Name != "new-name" {
		t.Errorf("name: got %q, want %q", resp.Output.Name, "new-name")
	}
}

func TestWorkloadPolicy_Update_Preview_SkipsAPI(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadPolicy{}
	resp, err := w.Update(context.Background(), infer.UpdateRequest[WorkloadPolicyArgs, WorkloadPolicyState]{
		ID:     "policy-123",
		State:  WorkloadPolicyState{WorkloadPolicyArgs: WorkloadPolicyArgs{Name: "old"}},
		Inputs: WorkloadPolicyArgs{Name: "new"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Name != "new" {
		t.Errorf("preview name: got %q, want %q", resp.Output.Name, "new")
	}
}

func TestWorkloadPolicy_Update_APIError(t *testing.T) {
	rec := &mockRecommendationClient{
		updateFn: func(_ context.Context, _ *connect.Request[apiv1.UpdateWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.UpdateWorkloadRecommendationPolicyResponse], error) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("server error"))
		},
	}
	withMockRecommendationClientSet(t, rec)

	w := &WorkloadPolicy{}
	_, err := w.Update(context.Background(), infer.UpdateRequest[WorkloadPolicyArgs, WorkloadPolicyState]{
		ID:     "policy-123",
		Inputs: WorkloadPolicyArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

// ---------- Delete ----------

func TestWorkloadPolicy_Delete_Success(t *testing.T) {
	rec := &mockRecommendationClient{
		deleteFn: func(_ context.Context, req *connect.Request[apiv1.DeleteWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.DeleteWorkloadRecommendationPolicyResponse], error) {
			if req.Msg.PolicyId != "policy-123" {
				t.Errorf("PolicyId: got %q, want %q", req.Msg.PolicyId, "policy-123")
			}
			return connect.NewResponse(&apiv1.DeleteWorkloadRecommendationPolicyResponse{Success: true}), nil
		},
	}
	withMockRecommendationClientSet(t, rec)

	w := &WorkloadPolicy{}
	_, err := w.Delete(context.Background(), infer.DeleteRequest[WorkloadPolicyState]{
		ID: "policy-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkloadPolicy_Delete_APIError(t *testing.T) {
	rec := &mockRecommendationClient{
		deleteFn: func(_ context.Context, _ *connect.Request[apiv1.DeleteWorkloadRecommendationPolicyRequest]) (*connect.Response[apiv1.DeleteWorkloadRecommendationPolicyResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("not found"))
		},
	}
	withMockRecommendationClientSet(t, rec)

	w := &WorkloadPolicy{}
	_, err := w.Delete(context.Background(), infer.DeleteRequest[WorkloadPolicyState]{
		ID: "policy-123",
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

// ---------- conversion helpers ----------

func TestActionTriggerRoundTrip(t *testing.T) {
	triggers := []string{"on_detection", "on_schedule"}
	back := actionTriggersFromProto(actionTriggersToProto(triggers))
	if len(back) != len(triggers) {
		t.Fatalf("len: got %d, want %d", len(back), len(triggers))
	}
	for i, want := range triggers {
		if back[i] != want {
			t.Errorf("[%d]: got %q, want %q", i, back[i], want)
		}
	}
}

func TestDetectionTriggerRoundTrip(t *testing.T) {
	triggers := []string{"pod_creation", "pod_update", "pod_reschedule"}
	back := detectionTriggersFromProto(detectionTriggersToProto(triggers))
	if len(back) != len(triggers) {
		t.Fatalf("len: got %d, want %d", len(back), len(triggers))
	}
	for i, want := range triggers {
		if back[i] != want {
			t.Errorf("[%d]: got %q, want %q", i, back[i], want)
		}
	}
}

func TestVerticalScalingRoundTrip(t *testing.T) {
	minReq := 100
	maxReq := 2000
	overhead := 0.05
	enabled := true
	v := &VerticalScalingArgs{
		Enabled:                 true,
		MinRequest:              &minReq,
		MaxRequest:              &maxReq,
		OverheadMultiplier:      &overhead,
		LimitsAdjustmentEnabled: &enabled,
	}
	back := verticalScalingFromProto(verticalScalingToProto(v))
	if back == nil {
		t.Fatal("expected non-nil")
	}
	if !back.Enabled {
		t.Error("Enabled should be true")
	}
	if back.MinRequest == nil || *back.MinRequest != minReq {
		t.Errorf("MinRequest: got %v, want %d", back.MinRequest, minReq)
	}
	if back.MaxRequest == nil || *back.MaxRequest != maxReq {
		t.Errorf("MaxRequest: got %v, want %d", back.MaxRequest, maxReq)
	}
	if back.LimitsAdjustmentEnabled == nil || !*back.LimitsAdjustmentEnabled {
		t.Error("LimitsAdjustmentEnabled should be true")
	}
}

func TestHorizontalScalingRoundTrip(t *testing.T) {
	minR := 1
	maxR := 10
	util := 0.75
	metric := "cpu"
	h := &HorizontalScalingArgs{
		Enabled:           true,
		MinReplicas:       &minR,
		MaxReplicas:       &maxR,
		TargetUtilization: &util,
		PrimaryMetric:     &metric,
	}
	back := horizontalScalingFromProto(horizontalScalingToProto(h))
	if back == nil {
		t.Fatal("expected non-nil")
	}
	if back.MinReplicas == nil || *back.MinReplicas != minR {
		t.Errorf("MinReplicas: got %v, want %d", back.MinReplicas, minR)
	}
	if back.PrimaryMetric == nil || *back.PrimaryMetric != metric {
		t.Errorf("PrimaryMetric: got %v, want %q", back.PrimaryMetric, metric)
	}
}

func TestVerticalScaling_Nil(t *testing.T) {
	if verticalScalingToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
	if verticalScalingFromProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestHorizontalScaling_Nil(t *testing.T) {
	if horizontalScalingToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
	if horizontalScalingFromProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
}
