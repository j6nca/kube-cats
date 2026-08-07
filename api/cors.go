package main

import (
	"net/http"
	"os"
	"strings"
)

// corsPolicy decides which browser origins may read the API.
//
// Configured with CORS_ALLOW_ORIGINS as a comma-separated list, e.g.
//
//	CORS_ALLOW_ORIGINS=https://meow.j6n.dev,http://kube-cats.j6n.internal
//
// A matching request Origin is echoed back, which is what lets several
// hostnames share one deployment — a fixed header can only ever satisfy one.
// "*" allows everything. When unset the previous behaviour applies: "*" in dev
// and the public frontend origin in production.
type corsPolicy struct {
	allowAll bool
	origins  map[string]bool
}

func newCORSPolicy(raw, environment string) corsPolicy {
	if strings.TrimSpace(raw) == "" {
		if environment == "production" {
			raw = "https://meow.j6n.dev"
		} else {
			raw = "*"
		}
	}

	policy := corsPolicy{origins: map[string]bool{}}
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		switch {
		case origin == "":
		case origin == "*":
			policy.allowAll = true
		default:
			// Origins are compared case-insensitively; browsers send the
			// scheme and host lowercased, but config is hand-written.
			policy.origins[strings.ToLower(strings.TrimSuffix(origin, "/"))] = true
		}
	}

	return policy
}

// allowed reports the value to send in Access-Control-Allow-Origin, and
// whether anything should be sent at all. A non-matching origin gets no
// header, so the browser blocks the read.
func (p corsPolicy) allowed(origin string) (string, bool) {
	if p.allowAll {
		return "*", true
	}
	if origin == "" {
		return "", false
	}
	if p.origins[strings.ToLower(strings.TrimSuffix(origin, "/"))] {
		return origin, true
	}
	return "", false
}

func (p corsPolicy) apply(w http.ResponseWriter, r *http.Request) {
	header := w.Header()

	// The response body is the same for everyone, but the CORS header is not,
	// so caches must key on Origin.
	if !p.allowAll {
		header.Add("Vary", "Origin")
	}

	if value, ok := p.allowed(r.Header.Get("Origin")); ok {
		header.Set("Access-Control-Allow-Origin", value)
	}

	header.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
}

// cors is resolved once at start-up. Tests replace it directly.
var cors = newCORSPolicy(os.Getenv("CORS_ALLOW_ORIGINS"), os.Getenv("ENVIRONMENT"))
