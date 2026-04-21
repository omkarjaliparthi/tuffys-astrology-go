package tuffys

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNatalChart_HappyPath(t *testing.T) {
	var gotPath, gotMethod, gotAuth, gotContentType string
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotAuth = r.Header.Get("x-api-key")
		gotContentType = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"planets":{"sun":{"longitude":84.123}}}`))
	}))
	defer srv.Close()

	client := New(srv.URL, WithAPIKey("test-key"))

	chart, err := client.NatalChart(context.Background(), Person{
		Datetime:  "1990-06-15T12:00:00Z",
		Latitude:  51.5,
		Longitude: 0,
	}, NatalChartOpts{HouseSystem: "placidus"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/chart/natal" {
		t.Errorf("path = %q, want /api/v1/chart/natal", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotAuth != "test-key" {
		t.Errorf("x-api-key = %q, want test-key", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	if gotBody["houseSystem"] != "placidus" {
		t.Errorf("houseSystem in body = %v, want placidus", gotBody["houseSystem"])
	}
	if chart["planets"] == nil {
		t.Errorf("expected planets in response, got %v", chart)
	}
}

func TestAPIError_StructuredEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"code":"rate_limited","message":"slow down"}}`))
	}))
	defer srv.Close()

	client := New(srv.URL, WithAPIKey("k"))

	_, err := client.NatalChart(context.Background(), Person{Datetime: "2000-01-01T00:00:00Z"})

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != http.StatusTooManyRequests {
		t.Errorf("Status = %d, want 429", apiErr.Status)
	}
	if apiErr.Code != "rate_limited" {
		t.Errorf("Code = %q, want rate_limited", apiErr.Code)
	}
	if apiErr.Message != "slow down" {
		t.Errorf("Message = %q, want 'slow down'", apiErr.Message)
	}
	if !strings.Contains(apiErr.Error(), "429") {
		t.Errorf("Error() should contain status code, got %q", apiErr.Error())
	}
}

func TestAPIError_RawFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`upstream down`))
	}))
	defer srv.Close()

	client := New(srv.URL, WithAPIKey("k"))

	_, err := client.NatalChart(context.Background(), Person{Datetime: "2000-01-01T00:00:00Z"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		t.Errorf("expected generic error for unstructured body, got *APIError: %v", err)
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("error should mention 502, got %q", err.Error())
	}
}

func TestNoAPIKey_OmitsHeader(t *testing.T) {
	var sawHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawHeader = r.Header.Get("x-api-key")
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	client := New(srv.URL) // no WithAPIKey
	_, _ = client.Positions(context.Background(), "2025-01-01T00:00:00Z")

	if sawHeader != "" {
		t.Errorf("x-api-key should not be set when no key configured, got %q", sawHeader)
	}
}

func TestWithHTTPClient_OverridesTransport(t *testing.T) {
	called := false
	custom := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{}`)),
			Header:     make(http.Header),
		}, nil
	})}

	client := New("https://does-not-matter.example", WithHTTPClient(custom))
	_, err := client.Positions(context.Background(), "2025-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("custom http.Client transport was not invoked")
	}
}

func TestContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block long enough that ctx cancels first.
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := New(srv.URL, WithAPIKey("k"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.Positions(ctx, "2025-01-01T00:00:00Z")
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context") {
		t.Errorf("expected context-related error, got %v", err)
	}
}

func TestBaseURL_TrailingSlashStripped(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	// Trailing slash on baseURL — client should not emit //api/...
	client := New(srv.URL+"/", WithAPIKey("k"))
	_, _ = client.Positions(context.Background(), "2025-01-01T00:00:00Z")

	if strings.HasPrefix(gotPath, "//") {
		t.Errorf("path has double slash prefix: %q", gotPath)
	}
	if gotPath != "/api/v1/positions" {
		t.Errorf("path = %q, want /api/v1/positions", gotPath)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
