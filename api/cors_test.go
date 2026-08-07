package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func headersFor(t *testing.T, policy corsPolicy, origin string) http.Header {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/nodes", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	policy.apply(rec, req)

	return rec.Header()
}

// Unset config must behave exactly as it did before it was configurable.
func TestDefaultsPreserveOldBehaviour(t *testing.T) {
	dev := newCORSPolicy("", "")
	if got := headersFor(t, dev, "http://localhost:3000").Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("dev default = %q, want *", got)
	}

	prod := newCORSPolicy("", "production")
	if got := headersFor(t, prod, "https://meow.j6n.dev").Get("Access-Control-Allow-Origin"); got != "https://meow.j6n.dev" {
		t.Errorf("prod default = %q, want the public frontend origin", got)
	}
	if got := headersFor(t, prod, "https://evil.example").Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("prod default allowed %q, want no header", got)
	}
}

// The whole point: several hostnames sharing one deployment.
func TestMultipleOriginsEachEchoed(t *testing.T) {
	policy := newCORSPolicy("https://meow.j6n.dev, http://kube-cats.j6n.internal ,http://localhost:3000", "production")

	for _, origin := range []string{
		"https://meow.j6n.dev",
		"http://kube-cats.j6n.internal",
		"http://localhost:3000",
	} {
		if got := headersFor(t, policy, origin).Get("Access-Control-Allow-Origin"); got != origin {
			t.Errorf("origin %q got %q, want it echoed back", origin, got)
		}
	}

	if got := headersFor(t, policy, "https://not-listed.example").Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("unlisted origin got %q, want no header", got)
	}
}

func TestWildcardInList(t *testing.T) {
	policy := newCORSPolicy("https://meow.j6n.dev,*", "production")
	if got := headersFor(t, policy, "https://anything.example").Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("wildcard = %q, want *", got)
	}
}

// Caches must not serve one origin's CORS header to another.
func TestVaryOriginSetWhenOriginDependent(t *testing.T) {
	policy := newCORSPolicy("https://meow.j6n.dev", "production")
	if got := headersFor(t, policy, "https://meow.j6n.dev").Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin", got)
	}

	// "*" is the same for everyone, so it does not vary.
	if got := headersFor(t, newCORSPolicy("*", ""), "https://x.example").Get("Vary"); got != "" {
		t.Errorf("Vary = %q, want empty for wildcard", got)
	}
}

func TestOriginMatchingIsCaseAndSlashInsensitive(t *testing.T) {
	policy := newCORSPolicy("https://Meow.J6N.dev/", "production")
	if got := headersFor(t, policy, "https://meow.j6n.dev").Get("Access-Control-Allow-Origin"); got != "https://meow.j6n.dev" {
		t.Errorf("got %q, want the request origin echoed", got)
	}
}

func TestRequestWithoutOriginGetsNoHeader(t *testing.T) {
	policy := newCORSPolicy("https://meow.j6n.dev", "production")
	if got := headersFor(t, policy, "").Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("got %q, want no header for a non-browser request", got)
	}
}

func TestEmptyAndWhitespaceEntriesIgnored(t *testing.T) {
	policy := newCORSPolicy("  ,https://meow.j6n.dev,  ,", "production")
	if len(policy.origins) != 1 {
		t.Errorf("origins = %v, want exactly one", policy.origins)
	}
	if policy.allowAll {
		t.Error("allowAll = true, want false")
	}
}

// A blank value must not be read as "allow nothing" and lock everyone out.
func TestWhitespaceOnlyFallsBackToDefault(t *testing.T) {
	policy := newCORSPolicy("   ", "production")
	if got := headersFor(t, policy, "https://meow.j6n.dev").Get("Access-Control-Allow-Origin"); got != "https://meow.j6n.dev" {
		t.Errorf("got %q, want the production default", got)
	}
}

func TestPreflightAnsweredAndShortCircuits(t *testing.T) {
	original := cors
	defer func() { cors = original }()
	cors = newCORSPolicy("https://meow.j6n.dev", "production")

	req := httptest.NewRequest(http.MethodOptions, "/nodes", nil)
	req.Header.Set("Origin", "https://meow.j6n.dev")
	rec := httptest.NewRecorder()

	nodesHandler(&Cache{})(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if body := rec.Body.String(); body != "" {
		t.Errorf("body = %q, want empty for a preflight", body)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://meow.j6n.dev" {
		t.Errorf("preflight ACAO = %q, want the origin", got)
	}
}

func TestHandlersApplyPolicy(t *testing.T) {
	original := cors
	defer func() { cors = original }()
	cors = newCORSPolicy("http://localhost:3000", "production")

	handlers := map[string]http.HandlerFunc{
		"/health": healthHandler,
		"/test":   testHandler,
		"/pods":   podsHandler(&Cache{}),
		"/nodes":  nodesHandler(&Cache{}),
	}

	for path, handler := range handlers {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
			t.Errorf("%s ACAO = %q, want the allowed origin", path, got)
		}
	}
}
