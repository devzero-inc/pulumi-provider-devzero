package provider

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// --- authInterceptor ---

func TestAuthInterceptor_SetsBearer(t *testing.T) {
	var capturedHeader string
	next := func(_ context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		capturedHeader = req.Header().Get("Authorization")
		return nil, nil
	}

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	connectReq := connect.NewRequest(req)
	_, _ = authInterceptor("my-secret-token")(next)(context.Background(), connectReq)

	want := "Bearer my-secret-token"
	if capturedHeader != want {
		t.Errorf("got %q, want %q", capturedHeader, want)
	}
}

// --- retryInterceptor ---

func TestRetryInterceptor_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	next := func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		calls++
		return nil, nil
	}

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	_, err := retryInterceptor(3, time.Millisecond)(next)(context.Background(), connect.NewRequest(req))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryInterceptor_RetriesOnUnavailable(t *testing.T) {
	calls := 0
	next := func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		calls++
		if calls < 3 {
			return nil, connect.NewError(connect.CodeUnavailable, errors.New("service down"))
		}
		return nil, nil
	}

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	_, err := retryInterceptor(3, time.Millisecond)(next)(context.Background(), connect.NewRequest(req))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryInterceptor_DoesNotRetryOnPermissionDenied(t *testing.T) {
	calls := 0
	next := func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		calls++
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("forbidden"))
	}

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	_, err := retryInterceptor(3, time.Millisecond)(next)(context.Background(), connect.NewRequest(req))

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retry), got %d", calls)
	}
}

func TestRetryInterceptor_ExhaustsAllAttempts(t *testing.T) {
	calls := 0
	next := func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		calls++
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("always down"))
	}

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	_, err := retryInterceptor(3, time.Millisecond)(next)(context.Background(), connect.NewRequest(req))

	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryInterceptor_RespectsContextCancellation(t *testing.T) {
	calls := 0
	next := func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		calls++
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("down"))
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	_, err := retryInterceptor(5, 50*time.Millisecond)(next)(ctx, connect.NewRequest(req))

	if err == nil {
		t.Fatal("expected error from context cancellation")
	}
	if calls >= 5 {
		t.Errorf("expected early exit, got %d calls", calls)
	}
}

func TestRetryInterceptor_RetriesOnDeadlineExceeded(t *testing.T) {
	calls := 0
	next := func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		calls++
		if calls < 2 {
			return nil, connect.NewError(connect.CodeDeadlineExceeded, errors.New("timeout"))
		}
		return nil, nil
	}

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	_, err := retryInterceptor(3, time.Millisecond)(next)(context.Background(), connect.NewRequest(req))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

// --- NewClientSet ---

func TestNewClientSet_NotNil(t *testing.T) {
	cs := NewClientSet("tok", "team1", "https://dakr.devzero.dev")
	if cs == nil {
		t.Fatal("expected non-nil ClientSet")
	}
	if cs.TeamID != "team1" {
		t.Errorf("TeamID: got %q, want %q", cs.TeamID, "team1")
	}
	if cs.K8SClient == nil {
		t.Error("K8SClient is nil")
	}
	if cs.ClusterMutationClient == nil {
		t.Error("ClusterMutationClient is nil")
	}
	if cs.RecommendationClient == nil {
		t.Error("RecommendationClient is nil")
	}
}

// --- clientset.Get/Set ---

func TestClientsetGetSet(t *testing.T) {
	prev := clientset.Get()
	defer clientset.Set(prev)

	clientset.Set(nil)
	if clientset.Get() != nil {
		t.Error("expected nil")
	}

	cs := NewClientSet("tok", "team1", "https://dakr.devzero.dev")
	clientset.Set(cs)
	if clientset.Get() != cs {
		t.Error("Get should return the set ClientSet")
	}
}
