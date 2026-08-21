package provider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"

	apiv1connect "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1/apiv1connect"
	"github.com/devzero-inc/pulumi-provider-devzero/provider/pkg/clientset"
)

// newHTTPClient returns an *http.Client with production-ready timeouts and connection pooling.
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   5,
		},
	}
}

// authInterceptor returns a Connect interceptor that injects a Bearer token
// into every outgoing request header.
func authInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
			return next(ctx, req)
		}
	}
}

// isIdempotentProcedure reports whether an RPC is safe to retry after a
// transport-level failure. Only read-style RPCs qualify: a Create/Upsert whose
// first attempt actually landed server-side would be duplicated on retry
// (CreateNodePolicies without a client-set id always inserts a new row, and
// there is no name uniqueness).
func isIdempotentProcedure(procedure string) bool {
	name := procedure
	if i := strings.LastIndex(procedure, "/"); i >= 0 {
		name = procedure[i+1:]
	}
	return strings.HasPrefix(name, "Get") || strings.HasPrefix(name, "List")
}

// retryInterceptor returns a Connect interceptor that retries transient errors
// (CodeUnavailable, CodeDeadlineExceeded) with exponential backoff. Only
// idempotent (read-style) RPCs are retried.
func retryInterceptor(maxAttempts int, initialDelay time.Duration) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if maxAttempts < 1 || !isIdempotentProcedure(req.Spec().Procedure) {
				return next(ctx, req)
			}
			delay := initialDelay
			for attempt := 0; attempt < maxAttempts; attempt++ {
				resp, err := next(ctx, req)
				if err == nil {
					return resp, nil
				}
				code := connect.CodeOf(err)
				if code != connect.CodeUnavailable && code != connect.CodeDeadlineExceeded {
					return nil, err
				}
				if attempt == maxAttempts-1 {
					return nil, err
				}
				select {
				case <-time.After(delay):
					delay *= 2
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			return nil, nil
		}
	}
}

// NewClientSet constructs a *clientset.ClientSet from validated provider credentials.
func NewClientSet(token, teamID, baseURL string) *clientset.ClientSet {
	httpClient := newHTTPClient()
	opts := []connect.ClientOption{
		connect.WithGRPC(),
		connect.WithInterceptors(
			authInterceptor(token),
			retryInterceptor(3, 100*time.Millisecond),
		),
	}

	return &clientset.ClientSet{
		TeamID: teamID,
		Token:  token,
		ClusterMutationClient: apiv1connect.NewClusterMutationServiceClient(
			httpClient, baseURL, opts...,
		),
		K8SClient: apiv1connect.NewK8SServiceClient(
			httpClient, baseURL, opts...,
		),
		RecommendationClient: apiv1connect.NewK8SRecommendationServiceClient(
			httpClient, baseURL, opts...,
		),
		ClusterServiceClient: apiv1connect.NewClusterServiceClient(
			httpClient, baseURL, opts...,
		),
	}
}
