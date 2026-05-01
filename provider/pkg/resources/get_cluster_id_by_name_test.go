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

// mockClusterServiceClient stubs ClusterServiceClient for testing.
type mockClusterServiceClient struct {
	apiv1connect.ClusterServiceClient
	getClusterIDByNameFn func(context.Context, *connect.Request[apiv1.GetClusterIDByNameRequest]) (*connect.Response[apiv1.GetClusterIDByNameResponse], error)
}

func (m *mockClusterServiceClient) GetClusterIDByName(ctx context.Context, req *connect.Request[apiv1.GetClusterIDByNameRequest]) (*connect.Response[apiv1.GetClusterIDByNameResponse], error) {
	return m.getClusterIDByNameFn(ctx, req)
}

// withMockClusterServiceClient injects a ClientSet with the given ClusterServiceClient mock.
func withMockClusterServiceClient(t *testing.T, svc *mockClusterServiceClient) {
	t.Helper()
	prev := clientset.Get()
	t.Cleanup(func() { clientset.Set(prev) })
	clientset.Set(&clientset.ClientSet{
		TeamID:               "team-test",
		ClusterServiceClient: svc,
	})
}

func TestGetClusterIDByName_Success(t *testing.T) {
	svc := &mockClusterServiceClient{
		getClusterIDByNameFn: func(_ context.Context, req *connect.Request[apiv1.GetClusterIDByNameRequest]) (*connect.Response[apiv1.GetClusterIDByNameResponse], error) {
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			if req.Msg.Name != "my-cluster" {
				t.Errorf("Name: got %q, want %q", req.Msg.Name, "my-cluster")
			}
			return connect.NewResponse(&apiv1.GetClusterIDByNameResponse{Id: "cluster-abc"}), nil
		},
	}
	withMockClusterServiceClient(t, svc)

	f := &GetClusterIdByName{}
	resp, err := f.Invoke(context.Background(), infer.FunctionRequest[GetClusterIDByNameArgs]{
		Input: GetClusterIDByNameArgs{TeamID: "team-test", Name: "my-cluster"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.ClusterID != "cluster-abc" {
		t.Errorf("clusterId: got %q, want %q", resp.Output.ClusterID, "cluster-abc")
	}
}

func TestGetClusterIDByName_UsesClientSetTeamIDWhenEmpty(t *testing.T) {
	svc := &mockClusterServiceClient{
		getClusterIDByNameFn: func(_ context.Context, req *connect.Request[apiv1.GetClusterIDByNameRequest]) (*connect.Response[apiv1.GetClusterIDByNameResponse], error) {
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q (should fall back to ClientSet.TeamID)", req.Msg.TeamId, "team-test")
			}
			return connect.NewResponse(&apiv1.GetClusterIDByNameResponse{Id: "cluster-xyz"}), nil
		},
	}
	withMockClusterServiceClient(t, svc)

	f := &GetClusterIdByName{}
	resp, err := f.Invoke(context.Background(), infer.FunctionRequest[GetClusterIDByNameArgs]{
		Input: GetClusterIDByNameArgs{TeamID: "", Name: "my-cluster"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.ClusterID != "cluster-xyz" {
		t.Errorf("clusterId: got %q, want %q", resp.Output.ClusterID, "cluster-xyz")
	}
}

func TestGetClusterIDByName_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	f := &GetClusterIdByName{}
	_, err := f.Invoke(context.Background(), infer.FunctionRequest[GetClusterIDByNameArgs]{
		Input: GetClusterIDByNameArgs{TeamID: "t", Name: "c"},
	})
	if err == nil {
		t.Fatal("expected error when ClientSet is nil")
	}
}

func TestGetClusterIDByName_APIError(t *testing.T) {
	svc := &mockClusterServiceClient{
		getClusterIDByNameFn: func(_ context.Context, _ *connect.Request[apiv1.GetClusterIDByNameRequest]) (*connect.Response[apiv1.GetClusterIDByNameResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("cluster not found"))
		},
	}
	withMockClusterServiceClient(t, svc)

	f := &GetClusterIdByName{}
	_, err := f.Invoke(context.Background(), infer.FunctionRequest[GetClusterIDByNameArgs]{
		Input: GetClusterIDByNameArgs{TeamID: "team-test", Name: "missing"},
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestGetClusterIDByName_EmptyID(t *testing.T) {
	svc := &mockClusterServiceClient{
		getClusterIDByNameFn: func(_ context.Context, _ *connect.Request[apiv1.GetClusterIDByNameRequest]) (*connect.Response[apiv1.GetClusterIDByNameResponse], error) {
			return connect.NewResponse(&apiv1.GetClusterIDByNameResponse{Id: ""}), nil
		},
	}
	withMockClusterServiceClient(t, svc)

	f := &GetClusterIdByName{}
	_, err := f.Invoke(context.Background(), infer.FunctionRequest[GetClusterIDByNameArgs]{
		Input: GetClusterIDByNameArgs{TeamID: "team-test", Name: "ghost"},
	})
	if err == nil {
		t.Fatal("expected error when response ID is empty")
	}
}

func ptr(s string) *string { return &s }

func TestGetClusterIDByName_WithRegionAndCloudProvider(t *testing.T) {
	svc := &mockClusterServiceClient{
		getClusterIDByNameFn: func(_ context.Context, req *connect.Request[apiv1.GetClusterIDByNameRequest]) (*connect.Response[apiv1.GetClusterIDByNameResponse], error) {
			if req.Msg.Region == nil || *req.Msg.Region != "us-east-1" {
				t.Errorf("Region: got %v, want %q", req.Msg.Region, "us-east-1")
			}
			if req.Msg.CloudProvider == nil || *req.Msg.CloudProvider != "aws" {
				t.Errorf("CloudProvider: got %v, want %q", req.Msg.CloudProvider, "aws")
			}
			return connect.NewResponse(&apiv1.GetClusterIDByNameResponse{Id: "cluster-regional"}), nil
		},
	}
	withMockClusterServiceClient(t, svc)

	f := &GetClusterIdByName{}
	resp, err := f.Invoke(context.Background(), infer.FunctionRequest[GetClusterIDByNameArgs]{
		Input: GetClusterIDByNameArgs{
			TeamID:        "team-test",
			Name:          "my-cluster",
			Region:        ptr("us-east-1"),
			CloudProvider: ptr("aws"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.ClusterID != "cluster-regional" {
		t.Errorf("clusterId: got %q, want %q", resp.Output.ClusterID, "cluster-regional")
	}
}

func TestGetClusterIDByName_WithLivenessPreferLive(t *testing.T) {
	svc := &mockClusterServiceClient{
		getClusterIDByNameFn: func(_ context.Context, req *connect.Request[apiv1.GetClusterIDByNameRequest]) (*connect.Response[apiv1.GetClusterIDByNameResponse], error) {
			if req.Msg.Liveness == nil {
				t.Fatal("Liveness: got nil, want PREFER_LIVE")
			}
			if *req.Msg.Liveness != apiv1.ClusterLivenessPreference_CLUSTER_LIVENESS_PREFERENCE_PREFER_LIVE {
				t.Errorf("Liveness: got %v, want PREFER_LIVE", *req.Msg.Liveness)
			}
			return connect.NewResponse(&apiv1.GetClusterIDByNameResponse{Id: "cluster-live"}), nil
		},
	}
	withMockClusterServiceClient(t, svc)

	f := &GetClusterIdByName{}
	resp, err := f.Invoke(context.Background(), infer.FunctionRequest[GetClusterIDByNameArgs]{
		Input: GetClusterIDByNameArgs{TeamID: "team-test", Name: "my-cluster", Liveness: ptr("PREFER_LIVE")},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.ClusterID != "cluster-live" {
		t.Errorf("clusterId: got %q, want %q", resp.Output.ClusterID, "cluster-live")
	}
}

func TestGetClusterIDByName_WithLivenessRequireLive(t *testing.T) {
	svc := &mockClusterServiceClient{
		getClusterIDByNameFn: func(_ context.Context, req *connect.Request[apiv1.GetClusterIDByNameRequest]) (*connect.Response[apiv1.GetClusterIDByNameResponse], error) {
			if req.Msg.Liveness == nil {
				t.Fatal("Liveness: got nil, want REQUIRE_LIVE")
			}
			if *req.Msg.Liveness != apiv1.ClusterLivenessPreference_CLUSTER_LIVENESS_PREFERENCE_REQUIRE_LIVE {
				t.Errorf("Liveness: got %v, want REQUIRE_LIVE", *req.Msg.Liveness)
			}
			return connect.NewResponse(&apiv1.GetClusterIDByNameResponse{Id: "cluster-required"}), nil
		},
	}
	withMockClusterServiceClient(t, svc)

	f := &GetClusterIdByName{}
	resp, err := f.Invoke(context.Background(), infer.FunctionRequest[GetClusterIDByNameArgs]{
		Input: GetClusterIDByNameArgs{TeamID: "team-test", Name: "my-cluster", Liveness: ptr("REQUIRE_LIVE")},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.ClusterID != "cluster-required" {
		t.Errorf("clusterId: got %q, want %q", resp.Output.ClusterID, "cluster-required")
	}
}

func TestGetClusterIDByName_InvalidLiveness(t *testing.T) {
	svc := &mockClusterServiceClient{
		getClusterIDByNameFn: func(_ context.Context, _ *connect.Request[apiv1.GetClusterIDByNameRequest]) (*connect.Response[apiv1.GetClusterIDByNameResponse], error) {
			t.Fatal("RPC should not be called with invalid liveness")
			return nil, nil
		},
	}
	withMockClusterServiceClient(t, svc)

	f := &GetClusterIdByName{}
	_, err := f.Invoke(context.Background(), infer.FunctionRequest[GetClusterIDByNameArgs]{
		Input: GetClusterIDByNameArgs{TeamID: "team-test", Name: "my-cluster", Liveness: ptr("INVALID_VALUE")},
	})
	if err == nil {
		t.Fatal("expected error for invalid liveness value")
	}
}