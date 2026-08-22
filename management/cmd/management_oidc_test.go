package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	nbconfig "github.com/netbirdio/netbird/management/internals/server/config"
)

const testOIDCIssuer = "https://idp.example.com/"

func validOIDCConfigJSON() []byte {
	body, _ := json.Marshal(OIDCConfigResponse{
		Issuer:                testOIDCIssuer,
		TokenEndpoint:         testOIDCIssuer + "token",
		DeviceAuthEndpoint:    testOIDCIssuer + "device",
		JwksURI:               testOIDCIssuer + "jwks",
		AuthorizationEndpoint: testOIDCIssuer + "auth",
	})
	return body
}

func withOIDCRetry(t *testing.T, cfg oidcRetryConfig) {
	t.Helper()
	orig := oidcRetry
	oidcRetry = cfg
	t.Cleanup(func() { oidcRetry = orig })
}

func instantOIDCRetry() oidcRetryConfig {
	return oidcRetryConfig{
		HTTPTimeout:    2 * time.Second,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		TotalBudget:    time.Second,
		JitterFraction: 0,
		Sleep: func(ctx context.Context, d time.Duration) error {
			return ctx.Err()
		},
		Now: time.Now,
	}
}

func TestOIDCDiscoveryRetriesThenSucceeds(t *testing.T) {
	const failCount = 3
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) <= failCount {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("unavailable"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(validOIDCConfigJSON())
	}))
	defer srv.Close()

	withOIDCRetry(t, instantOIDCRetry())

	cfg := &nbconfig.Config{
		HttpConfig: &nbconfig.HttpServerConfig{
			AuthIssuer:         "old-issuer",
			AuthKeysLocation:   "old-jwks",
			OIDCConfigEndpoint: srv.URL,
		},
	}
	if err := ApplyOIDCConfig(context.Background(), cfg); err != nil {
		t.Fatalf("ApplyOIDCConfig: %v", err)
	}
	if got := hits.Load(); got != failCount+1 {
		t.Fatalf("got %d requests, want %d", got, failCount+1)
	}
	if cfg.HttpConfig.AuthIssuer != testOIDCIssuer {
		t.Fatalf("AuthIssuer = %q, want %q", cfg.HttpConfig.AuthIssuer, testOIDCIssuer)
	}
	if cfg.HttpConfig.AuthKeysLocation != testOIDCIssuer+"jwks" {
		t.Fatalf("AuthKeysLocation = %q, want %q", cfg.HttpConfig.AuthKeysLocation, testOIDCIssuer+"jwks")
	}
}

func TestOIDCDiscoveryPermanent4xxNoRetry(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("missing"))
	}))
	defer srv.Close()

	withOIDCRetry(t, instantOIDCRetry())

	_, err := fetchOIDCConfig(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error on permanent 4xx")
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("retried permanent 4xx: got %d requests, want 1", got)
	}
}

func TestOIDCDiscoveryContextCanceledDuringBackoff(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("down"))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	enteredBackoff := make(chan struct{})
	withOIDCRetry(t, oidcRetryConfig{
		HTTPTimeout:    2 * time.Second,
		InitialBackoff: time.Minute,
		MaxBackoff:     time.Minute,
		TotalBudget:    10 * time.Minute,
		JitterFraction: 0,
		Sleep: func(ctx context.Context, d time.Duration) error {
			close(enteredBackoff)
			return sleepWithContext(ctx, d)
		},
		Now: time.Now,
	})

	go func() {
		<-enteredBackoff
		cancel()
	}()

	start := time.Now()
	_, err := fetchOIDCConfig(ctx, srv.URL)
	elapsed := time.Since(start)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("did not return promptly after cancel: %s", elapsed)
	}
}

func TestOIDCDiscoveryPermanentTransportErrorsNoRetry(t *testing.T) {
	withOIDCRetry(t, instantOIDCRetry())

	// Unsupported scheme: a configuration error, must not burn the retry budget.
	if _, err := fetchOIDCConfig(context.Background(), "idp.example.com/.well-known/openid-configuration"); err == nil {
		t.Fatal("expected error for non-absolute endpoint")
	}

	// TLS certificate verification failure is deterministic: fail on the first attempt.
	var hits atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write(validOIDCConfigJSON())
	}))
	defer srv.Close()
	_, err := fetchOIDCConfig(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected certificate verification error")
	}
	if isRetryableOIDCFetchErr(err) {
		t.Fatalf("certificate error must be permanent, got retryable: %v", err)
	}
}
