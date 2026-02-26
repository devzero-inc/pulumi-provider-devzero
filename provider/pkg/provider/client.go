package provider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"

	apiv1connect "github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1/apiv1connect"
)

// ClientSet holds the three Connect service clients used by all resources.
type ClientSet struct {
	TeamID                string
	ClusterMutationClient apiv1connect.ClusterMutationServiceClient
	K8SClient             apiv1connect.K8SServiceClient
	RecommendationClient  apiv1connect.K8SRecommendationServiceClient
}

// activeClientSet is the package-level singleton initialized during ProviderConfig.Configure.
// Resources call GetClientSet() to obtain it.
var activeClientSet *ClientSet

// GetClientSet returns the initialized ClientSet.
// Returns nil if ProviderConfig.Configure has not been called yet.
func GetClientSet() *ClientSet {
	return activeClientSet
}

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

// retryInterceptor returns a Connect interceptor that retries transient errors
// (CodeUnavailable, CodeDeadlineExceeded) with exponential backoff.
// maxAttempts is the total number of attempts (including the first).
// initialDelay is the sleep duration before the second attempt; it doubles each retry.
func retryInterceptor(maxAttempts int, initialDelay time.Duration) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			delay := initialDelay
			for attempt := 0; attempt < maxAttempts; attempt++ {
				resp, err := next(ctx, req)
				if err == nil {
					return resp, nil
				}
				// Only retry on transient error codes.
				code := connect.CodeOf(err)
				if code != connect.CodeUnavailable && code != connect.CodeDeadlineExceeded {
					return nil, err
				}
				// Last attempt — return the error without sleeping.
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
			// Unreachable, but required by the compiler.
			return nil, nil
		}
	}
}

// NewClientSet constructs a ClientSet from validated provider credentials.
// Called once during ProviderConfig.Configure; result stored in activeClientSet.
func NewClientSet(token, teamID, baseURL string) *ClientSet {
	httpClient := newHTTPClient()
	opts := []connect.ClientOption{
		connect.WithGRPC(),
		connect.WithInterceptors(
			authInterceptor(token),
			retryInterceptor(3, 100*time.Millisecond),
		),
	}

	return &ClientSet{
		TeamID: teamID,
		ClusterMutationClient: apiv1connect.NewClusterMutationServiceClient(
			httpClient, baseURL, opts...,
		),
		K8SClient: apiv1connect.NewK8SServiceClient(
			httpClient, baseURL, opts...,
		),
		RecommendationClient: apiv1connect.NewK8SRecommendationServiceClient(
			httpClient, baseURL, opts...,
		),
	}
}
