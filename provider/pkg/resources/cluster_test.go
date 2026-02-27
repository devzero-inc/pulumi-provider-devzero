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

// ---------- mock clients ----------

// mockMutationClient embeds the interface so unimplemented methods panic if called.
type mockMutationClient struct {
	apiv1connect.ClusterMutationServiceClient
	createFn     func(context.Context, *connect.Request[apiv1.CreateClusterRequest]) (*connect.Response[apiv1.CreateClusterResponse], error)
	updateFn     func(context.Context, *connect.Request[apiv1.UpdateClusterRequest]) (*connect.Response[apiv1.UpdateClusterResponse], error)
	deleteFn     func(context.Context, *connect.Request[apiv1.DeleteClusterRequest]) (*connect.Response[apiv1.DeleteClusterResponse], error)
	resetTokenFn func(context.Context, *connect.Request[apiv1.ResetClusterTokenRequest]) (*connect.Response[apiv1.ResetClusterTokenResponse], error)
}

func (m *mockMutationClient) CreateCluster(ctx context.Context, req *connect.Request[apiv1.CreateClusterRequest]) (*connect.Response[apiv1.CreateClusterResponse], error) {
	return m.createFn(ctx, req)
}
func (m *mockMutationClient) UpdateCluster(ctx context.Context, req *connect.Request[apiv1.UpdateClusterRequest]) (*connect.Response[apiv1.UpdateClusterResponse], error) {
	return m.updateFn(ctx, req)
}
func (m *mockMutationClient) DeleteCluster(ctx context.Context, req *connect.Request[apiv1.DeleteClusterRequest]) (*connect.Response[apiv1.DeleteClusterResponse], error) {
	return m.deleteFn(ctx, req)
}
func (m *mockMutationClient) ResetClusterToken(ctx context.Context, req *connect.Request[apiv1.ResetClusterTokenRequest]) (*connect.Response[apiv1.ResetClusterTokenResponse], error) {
	return m.resetTokenFn(ctx, req)
}

type mockK8SClient struct {
	apiv1connect.K8SServiceClient
	getClusterFn func(context.Context, *connect.Request[apiv1.GetClusterRequest]) (*connect.Response[apiv1.GetClusterResponse], error)
}

func (m *mockK8SClient) GetCluster(ctx context.Context, req *connect.Request[apiv1.GetClusterRequest]) (*connect.Response[apiv1.GetClusterResponse], error) {
	return m.getClusterFn(ctx, req)
}

// withMockClientSet injects a mock ClientSet and restores the previous one on cleanup.
func withMockClientSet(t *testing.T, mutation *mockMutationClient, k8s *mockK8SClient) {
	t.Helper()
	prev := clientset.Get()
	t.Cleanup(func() { clientset.Set(prev) })
	clientset.Set(&clientset.ClientSet{
		TeamID:                "team-test",
		ClusterMutationClient: mutation,
		K8SClient:             k8s,
	})
}

// ---------- Create ----------

func TestCluster_Create_Success(t *testing.T) {
	mut := &mockMutationClient{
		createFn: func(_ context.Context, req *connect.Request[apiv1.CreateClusterRequest]) (*connect.Response[apiv1.CreateClusterResponse], error) {
			if req.Msg.ClusterName != "my-cluster" {
				t.Errorf("ClusterName: got %q, want %q", req.Msg.ClusterName, "my-cluster")
			}
			if req.Msg.TeamId != "team-test" {
				t.Errorf("TeamId: got %q, want %q", req.Msg.TeamId, "team-test")
			}
			return connect.NewResponse(&apiv1.CreateClusterResponse{
				Cluster: &apiv1.Cluster{Id: "cluster-123"},
				Token:   "tok-abc",
			}), nil
		},
	}
	withMockClientSet(t, mut, &mockK8SClient{})

	c := &Cluster{}
	resp, err := c.Create(context.Background(), infer.CreateRequest[ClusterArgs]{
		Name:   "my-cluster",
		Inputs: ClusterArgs{Name: "my-cluster"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "cluster-123" {
		t.Errorf("id: got %q, want %q", resp.ID, "cluster-123")
	}
	if resp.Output.Token != "tok-abc" {
		t.Errorf("token: got %q, want %q", resp.Output.Token, "tok-abc")
	}
	if resp.Output.Name != "my-cluster" {
		t.Errorf("name: got %q, want %q", resp.Output.Name, "my-cluster")
	}
}

func TestCluster_Create_Preview_SkipsAPI(t *testing.T) {
	// No ClientSet set — would panic if API were called.
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	c := &Cluster{}
	resp, err := c.Create(context.Background(), infer.CreateRequest[ClusterArgs]{
		Name:   "preview-cluster",
		Inputs: ClusterArgs{Name: "preview-cluster"},
		DryRun: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("preview id should be empty, got %q", resp.ID)
	}
	if resp.Output.Token != "" {
		t.Errorf("preview token should be empty, got %q", resp.Output.Token)
	}
}

func TestCluster_Create_NilClientSet(t *testing.T) {
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	c := &Cluster{}
	_, err := c.Create(context.Background(), infer.CreateRequest[ClusterArgs]{
		Name:   "x",
		Inputs: ClusterArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error when ClientSet is nil")
	}
}

func TestCluster_Create_APIError(t *testing.T) {
	mut := &mockMutationClient{
		createFn: func(_ context.Context, _ *connect.Request[apiv1.CreateClusterRequest]) (*connect.Response[apiv1.CreateClusterResponse], error) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("server exploded"))
		},
	}
	withMockClientSet(t, mut, &mockK8SClient{})

	c := &Cluster{}
	_, err := c.Create(context.Background(), infer.CreateRequest[ClusterArgs]{
		Name:   "x",
		Inputs: ClusterArgs{Name: "x"},
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

// ---------- Read ----------

func TestCluster_Read_UsesCustomName(t *testing.T) {
	k8s := &mockK8SClient{
		getClusterFn: func(_ context.Context, _ *connect.Request[apiv1.GetClusterRequest]) (*connect.Response[apiv1.GetClusterResponse], error) {
			return connect.NewResponse(&apiv1.GetClusterResponse{
				Cluster: &apiv1.Cluster{Name: "system-name", CustomName: "custom-name"},
			}), nil
		},
	}
	withMockClientSet(t, &mockMutationClient{}, k8s)

	c := &Cluster{}
	resp, err := c.Read(context.Background(), infer.ReadRequest[ClusterArgs, ClusterState]{
		ID:     "cluster-123",
		Inputs: ClusterArgs{Name: "old"},
		State:  ClusterState{Token: "existing-token"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Inputs.Name != "custom-name" {
		t.Errorf("name: got %q, want %q", resp.Inputs.Name, "custom-name")
	}
	if resp.State.Token != "existing-token" {
		t.Errorf("token should be preserved, got %q", resp.State.Token)
	}
}

func TestCluster_Read_FallsBackToName(t *testing.T) {
	k8s := &mockK8SClient{
		getClusterFn: func(_ context.Context, _ *connect.Request[apiv1.GetClusterRequest]) (*connect.Response[apiv1.GetClusterResponse], error) {
			return connect.NewResponse(&apiv1.GetClusterResponse{
				Cluster: &apiv1.Cluster{Name: "system-name", CustomName: ""},
			}), nil
		},
	}
	withMockClientSet(t, &mockMutationClient{}, k8s)

	c := &Cluster{}
	resp, err := c.Read(context.Background(), infer.ReadRequest[ClusterArgs, ClusterState]{
		ID: "cluster-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Inputs.Name != "system-name" {
		t.Errorf("name fallback: got %q, want %q", resp.Inputs.Name, "system-name")
	}
}

// ---------- Update ----------

func TestCluster_Update_Success(t *testing.T) {
	mut := &mockMutationClient{
		updateFn: func(_ context.Context, req *connect.Request[apiv1.UpdateClusterRequest]) (*connect.Response[apiv1.UpdateClusterResponse], error) {
			return connect.NewResponse(&apiv1.UpdateClusterResponse{
				Cluster: &apiv1.Cluster{CustomName: req.Msg.ClusterName},
			}), nil
		},
	}
	withMockClientSet(t, mut, &mockK8SClient{})

	c := &Cluster{}
	resp, err := c.Update(context.Background(), infer.UpdateRequest[ClusterArgs, ClusterState]{
		ID:     "cluster-123",
		State:  ClusterState{ClusterArgs: ClusterArgs{Name: "old"}, Token: "existing-token"},
		Inputs: ClusterArgs{Name: "new-name"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Name != "new-name" {
		t.Errorf("name: got %q, want %q", resp.Output.Name, "new-name")
	}
	if resp.Output.Token != "existing-token" {
		t.Errorf("token should be preserved, got %q", resp.Output.Token)
	}
}

func TestCluster_Update_RotatesTokenWhenEmpty(t *testing.T) {
	mut := &mockMutationClient{
		updateFn: func(_ context.Context, req *connect.Request[apiv1.UpdateClusterRequest]) (*connect.Response[apiv1.UpdateClusterResponse], error) {
			return connect.NewResponse(&apiv1.UpdateClusterResponse{
				Cluster: &apiv1.Cluster{CustomName: req.Msg.ClusterName},
			}), nil
		},
		resetTokenFn: func(_ context.Context, _ *connect.Request[apiv1.ResetClusterTokenRequest]) (*connect.Response[apiv1.ResetClusterTokenResponse], error) {
			return connect.NewResponse(&apiv1.ResetClusterTokenResponse{Token: "new-rotated-token"}), nil
		},
	}
	withMockClientSet(t, mut, &mockK8SClient{})

	c := &Cluster{}
	// State.Token is empty — triggers rotation
	resp, err := c.Update(context.Background(), infer.UpdateRequest[ClusterArgs, ClusterState]{
		ID:     "cluster-123",
		State:  ClusterState{ClusterArgs: ClusterArgs{Name: "name"}, Token: ""},
		Inputs: ClusterArgs{Name: "name"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Token != "new-rotated-token" {
		t.Errorf("token: got %q, want %q", resp.Output.Token, "new-rotated-token")
	}
}

func TestCluster_Update_Preview_SkipsAPI(t *testing.T) {
	// No ClientSet — would panic if API were called.
	prev := clientset.Get()
	clientset.Set(nil)
	defer clientset.Set(prev)

	c := &Cluster{}
	resp, err := c.Update(context.Background(), infer.UpdateRequest[ClusterArgs, ClusterState]{
		ID:     "cluster-123",
		State:  ClusterState{ClusterArgs: ClusterArgs{Name: "old"}, Token: "tok"},
		Inputs: ClusterArgs{Name: "new"},
		DryRun: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Token != "tok" {
		t.Errorf("preview should preserve token, got %q", resp.Output.Token)
	}
}

// ---------- Delete ----------

func TestCluster_Delete_Success(t *testing.T) {
	mut := &mockMutationClient{
		deleteFn: func(_ context.Context, req *connect.Request[apiv1.DeleteClusterRequest]) (*connect.Response[apiv1.DeleteClusterResponse], error) {
			if req.Msg.ClusterId != "cluster-123" {
				t.Errorf("ClusterId: got %q, want %q", req.Msg.ClusterId, "cluster-123")
			}
			return connect.NewResponse(&apiv1.DeleteClusterResponse{}), nil
		},
	}
	withMockClientSet(t, mut, &mockK8SClient{})

	c := &Cluster{}
	_, err := c.Delete(context.Background(), infer.DeleteRequest[ClusterState]{
		ID: "cluster-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCluster_Delete_APIError(t *testing.T) {
	mut := &mockMutationClient{
		deleteFn: func(_ context.Context, _ *connect.Request[apiv1.DeleteClusterRequest]) (*connect.Response[apiv1.DeleteClusterResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("not found"))
		},
	}
	withMockClientSet(t, mut, &mockK8SClient{})

	c := &Cluster{}
	_, err := c.Delete(context.Background(), infer.DeleteRequest[ClusterState]{
		ID: "cluster-123",
	})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}
