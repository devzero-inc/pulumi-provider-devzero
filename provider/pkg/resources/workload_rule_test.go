package resources

import (
	"context"
	"errors"
	"math"
	"testing"

	"connectrpc.com/connect"
	"github.com/pulumi/pulumi-go-provider/infer"

	apiv1 "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1"
	apiv1connect "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1/apiv1connect"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// mockWorkloadRuleClient stubs only the workload rule CRUD methods we test.
type mockWorkloadRuleClient struct {
	apiv1connect.K8SRecommendationServiceClient
	upsertFn func(context.Context, *connect.Request[apiv1.UpsertManualWorkloadRuleRequest]) (*connect.Response[apiv1.UpsertManualWorkloadRuleResponse], error)
	getFn    func(context.Context, *connect.Request[apiv1.GetWorkloadRuleByIDRequest]) (*connect.Response[apiv1.GetWorkloadRuleByIDResponse], error)
	deleteFn func(context.Context, *connect.Request[apiv1.DeleteWorkloadRuleRequest]) (*connect.Response[apiv1.DeleteWorkloadRuleResponse], error)
}

func (m *mockWorkloadRuleClient) UpsertManualWorkloadRule(ctx context.Context, req *connect.Request[apiv1.UpsertManualWorkloadRuleRequest]) (*connect.Response[apiv1.UpsertManualWorkloadRuleResponse], error) {
	return m.upsertFn(ctx, req)
}
func (m *mockWorkloadRuleClient) GetWorkloadRuleByID(ctx context.Context, req *connect.Request[apiv1.GetWorkloadRuleByIDRequest]) (*connect.Response[apiv1.GetWorkloadRuleByIDResponse], error) {
	return m.getFn(ctx, req)
}
func (m *mockWorkloadRuleClient) DeleteWorkloadRule(ctx context.Context, req *connect.Request[apiv1.DeleteWorkloadRuleRequest]) (*connect.Response[apiv1.DeleteWorkloadRuleResponse], error) {
	return m.deleteFn(ctx, req)
}

func withMockWorkloadRuleClientSet(t *testing.T, rec *mockWorkloadRuleClient) {
	t.Helper()
	prev := clientset.Get()
	t.Cleanup(func() { clientset.Set(prev) })
	clientset.Set(&clientset.ClientSet{
		TeamID:               "team-test",
		RecommendationClient: rec,
	})
}

// ---------- Create ----------

func TestWorkloadRule_Create_Success(t *testing.T) {
	rec := &mockWorkloadRuleClient{
		upsertFn: func(_ context.Context, req *connect.Request[apiv1.UpsertManualWorkloadRuleRequest]) (*connect.Response[apiv1.UpsertManualWorkloadRuleResponse], error) {
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			if req.Msg.ClusterId != "cluster-1" {
				t.Errorf("ClusterId: got %q, want %q", req.Msg.ClusterId, "cluster-1")
			}
			if req.Msg.Namespace != "default" {
				t.Errorf("Namespace: got %q, want %q", req.Msg.Namespace, "default")
			}
			if req.Msg.Kind != "Deployment" {
				t.Errorf("Kind: got %q, want %q", req.Msg.Kind, "Deployment")
			}
			if req.Msg.Name != "my-app" {
				t.Errorf("Name: got %q, want %q", req.Msg.Name, "my-app")
			}
			return connect.NewResponse(&apiv1.UpsertManualWorkloadRuleResponse{
				Rule: &apiv1.WorkloadRule{
					RuleId:    "rule-123",
					ClusterId: "cluster-1",
					Namespace: "default",
					Kind:      "Deployment",
					Name:      "my-app",
				},
			}), nil
		},
	}
	withMockWorkloadRuleClientSet(t, rec)

	w := &WorkloadRule{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadRuleArgs]{
		Inputs: WorkloadRuleArgs{
			ClusterID: "cluster-1",
			Namespace: "default",
			Kind:      "Deployment",
			Name:      "my-app",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "rule-123" {
		t.Errorf("id: got %q, want %q", resp.ID, "rule-123")
	}
	if resp.Output.ClusterID != "cluster-1" {
		t.Errorf("ClusterID: got %q, want %q", resp.Output.ClusterID, "cluster-1")
	}
}

func TestWorkloadRule_Create_Preview_SkipsAPI(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadRule{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadRuleArgs]{
		Inputs: WorkloadRuleArgs{
			ClusterID: "cluster-1",
			Namespace: "default",
			Kind:      "Deployment",
			Name:      "my-app",
		},
		DryRun: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("preview id should be empty, got %q", resp.ID)
	}
	if resp.Output.Name != "my-app" {
		t.Errorf("preview name: got %q, want %q", resp.Output.Name, "my-app")
	}
}

func TestWorkloadRule_Create_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadRule{}
	_, err := w.Create(context.Background(), infer.CreateRequest[WorkloadRuleArgs]{
		Inputs: WorkloadRuleArgs{ClusterID: "c", Namespace: "ns", Kind: "Deployment", Name: "app"},
	})
	if err == nil {
		t.Fatal("expected error when ClientSet is nil")
	}
}

func TestWorkloadRule_Create_APIError(t *testing.T) {
	rec := &mockWorkloadRuleClient{
		upsertFn: func(_ context.Context, _ *connect.Request[apiv1.UpsertManualWorkloadRuleRequest]) (*connect.Response[apiv1.UpsertManualWorkloadRuleResponse], error) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("server error"))
		},
	}
	withMockWorkloadRuleClientSet(t, rec)

	w := &WorkloadRule{}
	_, err := w.Create(context.Background(), infer.CreateRequest[WorkloadRuleArgs]{
		Inputs: WorkloadRuleArgs{ClusterID: "c", Namespace: "ns", Kind: "Deployment", Name: "app"},
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestWorkloadRule_Create_WithAutoGenerate(t *testing.T) {
	rec := &mockWorkloadRuleClient{
		upsertFn: func(_ context.Context, req *connect.Request[apiv1.UpsertManualWorkloadRuleRequest]) (*connect.Response[apiv1.UpsertManualWorkloadRuleResponse], error) {
			if !req.Msg.AutoGenerate {
				t.Error("AutoGenerate: expected true")
			}
			if req.Msg.Fields != nil {
				t.Error("Fields: expected nil when auto_generate=true")
			}
			return connect.NewResponse(&apiv1.UpsertManualWorkloadRuleResponse{
				Rule: &apiv1.WorkloadRule{RuleId: "rule-auto"},
			}), nil
		},
	}
	withMockWorkloadRuleClientSet(t, rec)

	autoGen := true
	w := &WorkloadRule{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadRuleArgs]{
		Inputs: WorkloadRuleArgs{
			ClusterID:    "c",
			Namespace:    "ns",
			Kind:         "Deployment",
			Name:         "app",
			AutoGenerate: &autoGen,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "rule-auto" {
		t.Errorf("id: got %q, want %q", resp.ID, "rule-auto")
	}
}

func TestWorkloadRule_Create_WithCpuRule(t *testing.T) {
	minReq := 100
	maxReq := 4000

	rec := &mockWorkloadRuleClient{
		upsertFn: func(_ context.Context, req *connect.Request[apiv1.UpsertManualWorkloadRuleRequest]) (*connect.Response[apiv1.UpsertManualWorkloadRuleResponse], error) {
			if req.Msg.Fields == nil {
				t.Fatal("expected Fields to be set")
			}
			cpu := req.Msg.Fields.CpuRule
			if cpu == nil {
				t.Fatal("expected cpu_rule to be set")
			}
			if !cpu.Enabled {
				t.Error("cpu Enabled: expected true")
			}
			if cpu.MinRequest == nil || *cpu.MinRequest != int64(minReq) {
				t.Errorf("cpu MinRequest: got %v, want %d", cpu.MinRequest, minReq)
			}
			if cpu.MaxRequest == nil || *cpu.MaxRequest != int64(maxReq) {
				t.Errorf("cpu MaxRequest: got %v, want %d", cpu.MaxRequest, maxReq)
			}
			return connect.NewResponse(&apiv1.UpsertManualWorkloadRuleResponse{
				Rule: &apiv1.WorkloadRule{
					RuleId:  "rule-cpu",
					CpuRule: cpu,
				},
			}), nil
		},
	}
	withMockWorkloadRuleClientSet(t, rec)

	w := &WorkloadRule{}
	resp, err := w.Create(context.Background(), infer.CreateRequest[WorkloadRuleArgs]{
		Inputs: WorkloadRuleArgs{
			ClusterID: "c",
			Namespace: "ns",
			Kind:      "Deployment",
			Name:      "app",
			CpuRule: &ResourceRuleConfigArgs{
				Enabled:    true,
				MinRequest: &minReq,
				MaxRequest: &maxReq,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "rule-cpu" {
		t.Errorf("id: got %q, want %q", resp.ID, "rule-cpu")
	}
	if resp.Output.CpuRule == nil {
		t.Fatal("expected CpuRule in output")
	}
	if resp.Output.CpuRule.MinRequest == nil || *resp.Output.CpuRule.MinRequest != minReq {
		t.Errorf("output CpuRule.MinRequest: got %v, want %d", resp.Output.CpuRule.MinRequest, minReq)
	}
}

// ---------- Read ----------

func TestWorkloadRule_Read_Success(t *testing.T) {
	rec := &mockWorkloadRuleClient{
		getFn: func(_ context.Context, req *connect.Request[apiv1.GetWorkloadRuleByIDRequest]) (*connect.Response[apiv1.GetWorkloadRuleByIDResponse], error) {
			if req.Msg.RuleId != "rule-123" {
				t.Errorf("RuleId: got %q, want %q", req.Msg.RuleId, "rule-123")
			}
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			return connect.NewResponse(&apiv1.GetWorkloadRuleByIDResponse{
				Rule: &apiv1.WorkloadRule{
					RuleId:    "rule-123",
					ClusterId: "cluster-1",
					Namespace: "default",
					Kind:      "Deployment",
					Name:      "my-app",
				},
			}), nil
		},
	}
	withMockWorkloadRuleClientSet(t, rec)

	w := &WorkloadRule{}
	resp, err := w.Read(context.Background(), infer.ReadRequest[WorkloadRuleArgs, WorkloadRuleState]{
		ID: "rule-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "rule-123" {
		t.Errorf("id: got %q, want %q", resp.ID, "rule-123")
	}
	if resp.Inputs.ClusterID != "cluster-1" {
		t.Errorf("ClusterID: got %q, want %q", resp.Inputs.ClusterID, "cluster-1")
	}
	if resp.Inputs.Namespace != "default" {
		t.Errorf("Namespace: got %q, want %q", resp.Inputs.Namespace, "default")
	}
	if resp.Inputs.Name != "my-app" {
		t.Errorf("Name: got %q, want %q", resp.Inputs.Name, "my-app")
	}
}

func TestWorkloadRule_Read_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadRule{}
	_, err := w.Read(context.Background(), infer.ReadRequest[WorkloadRuleArgs, WorkloadRuleState]{
		ID: "rule-123",
	})
	if err == nil {
		t.Fatal("expected error when ClientSet is nil")
	}
}

func TestWorkloadRule_Read_APIError(t *testing.T) {
	rec := &mockWorkloadRuleClient{
		getFn: func(_ context.Context, _ *connect.Request[apiv1.GetWorkloadRuleByIDRequest]) (*connect.Response[apiv1.GetWorkloadRuleByIDResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("not found"))
		},
	}
	withMockWorkloadRuleClientSet(t, rec)

	w := &WorkloadRule{}
	_, err := w.Read(context.Background(), infer.ReadRequest[WorkloadRuleArgs, WorkloadRuleState]{
		ID: "rule-123",
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

// ---------- Update ----------

func TestWorkloadRule_Update_Success(t *testing.T) {
	rec := &mockWorkloadRuleClient{
		upsertFn: func(_ context.Context, req *connect.Request[apiv1.UpsertManualWorkloadRuleRequest]) (*connect.Response[apiv1.UpsertManualWorkloadRuleResponse], error) {
			if req.Msg.ClusterId != "cluster-1" {
				t.Errorf("ClusterId: got %q, want %q", req.Msg.ClusterId, "cluster-1")
			}
			if req.Msg.Name != "my-app" {
				t.Errorf("Name: got %q, want %q", req.Msg.Name, "my-app")
			}
			return connect.NewResponse(&apiv1.UpsertManualWorkloadRuleResponse{
				Rule: &apiv1.WorkloadRule{
					RuleId:    "rule-123",
					ClusterId: "cluster-1",
					Namespace: "default",
					Kind:      "Deployment",
					Name:      "my-app",
				},
			}), nil
		},
	}
	withMockWorkloadRuleClientSet(t, rec)

	w := &WorkloadRule{}
	resp, err := w.Update(context.Background(), infer.UpdateRequest[WorkloadRuleArgs, WorkloadRuleState]{
		ID: "rule-123",
		State: WorkloadRuleState{WorkloadRuleArgs: WorkloadRuleArgs{
			ClusterID: "cluster-1",
			Namespace: "default",
			Kind:      "Deployment",
			Name:      "my-app",
		}},
		Inputs: WorkloadRuleArgs{
			ClusterID: "cluster-1",
			Namespace: "default",
			Kind:      "Deployment",
			Name:      "my-app",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.ClusterID != "cluster-1" {
		t.Errorf("ClusterID: got %q, want %q", resp.Output.ClusterID, "cluster-1")
	}
}

func TestWorkloadRule_Update_Preview_SkipsAPI(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadRule{}
	resp, err := w.Update(context.Background(), infer.UpdateRequest[WorkloadRuleArgs, WorkloadRuleState]{
		ID: "rule-123",
		State: WorkloadRuleState{WorkloadRuleArgs: WorkloadRuleArgs{
			ClusterID: "cluster-1",
			Namespace: "default",
			Kind:      "Deployment",
			Name:      "my-app",
		}},
		Inputs: WorkloadRuleArgs{
			ClusterID: "cluster-1",
			Namespace: "default",
			Kind:      "Deployment",
			Name:      "my-app",
		},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Name != "my-app" {
		t.Errorf("preview name: got %q, want %q", resp.Output.Name, "my-app")
	}
}

func TestWorkloadRule_Update_APIError(t *testing.T) {
	rec := &mockWorkloadRuleClient{
		upsertFn: func(_ context.Context, _ *connect.Request[apiv1.UpsertManualWorkloadRuleRequest]) (*connect.Response[apiv1.UpsertManualWorkloadRuleResponse], error) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("server error"))
		},
	}
	withMockWorkloadRuleClientSet(t, rec)

	w := &WorkloadRule{}
	_, err := w.Update(context.Background(), infer.UpdateRequest[WorkloadRuleArgs, WorkloadRuleState]{
		ID:     "rule-123",
		Inputs: WorkloadRuleArgs{ClusterID: "c", Namespace: "ns", Kind: "Deployment", Name: "app"},
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

// ---------- Delete ----------

func TestWorkloadRule_Delete_Success(t *testing.T) {
	rec := &mockWorkloadRuleClient{
		deleteFn: func(_ context.Context, req *connect.Request[apiv1.DeleteWorkloadRuleRequest]) (*connect.Response[apiv1.DeleteWorkloadRuleResponse], error) {
			if req.Msg.RuleId != "rule-123" {
				t.Errorf("RuleId: got %q, want %q", req.Msg.RuleId, "rule-123")
			}
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			return connect.NewResponse(&apiv1.DeleteWorkloadRuleResponse{}), nil
		},
	}
	withMockWorkloadRuleClientSet(t, rec)

	w := &WorkloadRule{}
	_, err := w.Delete(context.Background(), infer.DeleteRequest[WorkloadRuleState]{
		ID: "rule-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkloadRule_Delete_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	w := &WorkloadRule{}
	_, err := w.Delete(context.Background(), infer.DeleteRequest[WorkloadRuleState]{
		ID: "rule-123",
	})
	if err == nil {
		t.Fatal("expected error when ClientSet is nil")
	}
}

func TestWorkloadRule_Delete_APIError(t *testing.T) {
	rec := &mockWorkloadRuleClient{
		deleteFn: func(_ context.Context, _ *connect.Request[apiv1.DeleteWorkloadRuleRequest]) (*connect.Response[apiv1.DeleteWorkloadRuleResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("not found"))
		},
	}
	withMockWorkloadRuleClientSet(t, rec)

	w := &WorkloadRule{}
	_, err := w.Delete(context.Background(), infer.DeleteRequest[WorkloadRuleState]{
		ID: "rule-123",
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

// ---------- conversion round-trips ----------

func TestResourceRuleConfig_RoundTrip(t *testing.T) {
	minReq := 100
	maxReq := 4000
	limitMul := 1.5
	targetPct := 0.95
	maxUp := 50.0
	maxDown := 20.0

	a := &ResourceRuleConfigArgs{
		Enabled:                 true,
		MinRequest:              &minReq,
		MaxRequest:              &maxReq,
		LimitMultiplier:         &limitMul,
		LimitsAdjustmentEnabled: true,
		TargetPercentile:        &targetPct,
		MaxScaleUpPercent:       &maxUp,
		MaxScaleDownPercent:     &maxDown,
		LimitsRemovalEnabled:    true,
	}

	back := resourceRuleConfigFromProto(resourceRuleConfigToProto(a))
	if back == nil {
		t.Fatal("expected non-nil")
	}
	if !back.Enabled {
		t.Error("Enabled: expected true")
	}
	if back.MinRequest == nil || *back.MinRequest != minReq {
		t.Errorf("MinRequest: got %v, want %d", back.MinRequest, minReq)
	}
	if back.MaxRequest == nil || *back.MaxRequest != maxReq {
		t.Errorf("MaxRequest: got %v, want %d", back.MaxRequest, maxReq)
	}
	if back.LimitMultiplier == nil || math.Abs(*back.LimitMultiplier-limitMul) > 1e-5 {
		t.Errorf("LimitMultiplier: got %v, want %f", ptrVal(back.LimitMultiplier), limitMul)
	}
	if !back.LimitsAdjustmentEnabled {
		t.Error("LimitsAdjustmentEnabled: expected true")
	}
	if back.TargetPercentile == nil || math.Abs(*back.TargetPercentile-targetPct) > 1e-5 {
		t.Errorf("TargetPercentile: got %v, want %f", ptrVal(back.TargetPercentile), targetPct)
	}
	if !back.LimitsRemovalEnabled {
		t.Error("LimitsRemovalEnabled: expected true")
	}
}

func TestResourceRuleConfig_Nil(t *testing.T) {
	if resourceRuleConfigToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
	if resourceRuleConfigFromProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestHPARuleConfig_RoundTrip(t *testing.T) {
	minR := 1
	maxR := 10
	util := 0.7
	metric := "cpu"
	maxChange := 50.0

	h := &HPARuleConfigArgs{
		Enabled:                true,
		MinReplicas:            &minR,
		MaxReplicas:            &maxR,
		TargetUtilization:      &util,
		PrimaryMetric:          &metric,
		MaxReplicaChangePercent: &maxChange,
	}

	back := hpaRuleConfigFromProto(hpaRuleConfigToProto(h))
	if back == nil {
		t.Fatal("expected non-nil")
	}
	if !back.Enabled {
		t.Error("Enabled: expected true")
	}
	if back.MinReplicas == nil || *back.MinReplicas != minR {
		t.Errorf("MinReplicas: got %v, want %d", back.MinReplicas, minR)
	}
	if back.MaxReplicas == nil || *back.MaxReplicas != maxR {
		t.Errorf("MaxReplicas: got %v, want %d", back.MaxReplicas, maxR)
	}
	if back.PrimaryMetric == nil || *back.PrimaryMetric != metric {
		t.Errorf("PrimaryMetric: got %v, want %q", back.PrimaryMetric, metric)
	}
	if back.MaxReplicaChangePercent == nil || *back.MaxReplicaChangePercent != maxChange {
		t.Errorf("MaxReplicaChangePercent: got %v, want %f", back.MaxReplicaChangePercent, maxChange)
	}
}

func TestHPARuleConfig_Nil(t *testing.T) {
	if hpaRuleConfigToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
	if hpaRuleConfigFromProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestEmergencyResponseConfig_RoundTrip(t *testing.T) {
	e := &EmergencyResponseConfigArgs{
		OomEnabled:              true,
		OomMemoryMultiplier:     2.0,
		OomMaxReactions:         3,
		OomCooldownSeconds:      60,
		CpuThrottlingEnabled:    true,
		CpuThrottlingThreshold:  0.8,
		CpuThrottlingMultiplier: 1.5,
	}

	back := emergencyResponseFromProto(emergencyResponseToProto(e))
	if back == nil {
		t.Fatal("expected non-nil")
	}
	if !back.OomEnabled {
		t.Error("OomEnabled: expected true")
	}
	if back.OomMemoryMultiplier != e.OomMemoryMultiplier {
		t.Errorf("OomMemoryMultiplier: got %f, want %f", back.OomMemoryMultiplier, e.OomMemoryMultiplier)
	}
	if back.OomMaxReactions != e.OomMaxReactions {
		t.Errorf("OomMaxReactions: got %d, want %d", back.OomMaxReactions, e.OomMaxReactions)
	}
	if !back.CpuThrottlingEnabled {
		t.Error("CpuThrottlingEnabled: expected true")
	}
	if back.CpuThrottlingMultiplier != e.CpuThrottlingMultiplier {
		t.Errorf("CpuThrottlingMultiplier: got %f, want %f", back.CpuThrottlingMultiplier, e.CpuThrottlingMultiplier)
	}
}

func TestEmergencyResponseConfig_Nil(t *testing.T) {
	if emergencyResponseToProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
	if emergencyResponseFromProto(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func ptrVal(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func TestContainerResourceRuleConfig_RoundTrip(t *testing.T) {
	minReq := 100
	containers := []ContainerResourceRuleConfigArgs{
		{
			ContainerName: "main",
			CpuRule: &ResourceRuleConfigArgs{
				Enabled:    true,
				MinRequest: &minReq,
			},
		},
		{
			ContainerName: "sidecar",
			MemoryRule: &ResourceRuleConfigArgs{
				Enabled: true,
			},
		},
	}

	back := containerRuleConfigsFromProto(containerRuleConfigsToProto(containers))
	if len(back) != 2 {
		t.Fatalf("len: got %d, want 2", len(back))
	}
	if back[0].ContainerName != "main" {
		t.Errorf("containers[0].ContainerName: got %q, want %q", back[0].ContainerName, "main")
	}
	if back[0].CpuRule == nil {
		t.Fatal("containers[0].CpuRule: expected non-nil")
	}
	if back[0].CpuRule.MinRequest == nil || *back[0].CpuRule.MinRequest != minReq {
		t.Errorf("containers[0].CpuRule.MinRequest: got %v, want %d", back[0].CpuRule.MinRequest, minReq)
	}
	if back[1].ContainerName != "sidecar" {
		t.Errorf("containers[1].ContainerName: got %q, want %q", back[1].ContainerName, "sidecar")
	}
}
