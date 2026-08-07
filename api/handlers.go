package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// enableCors applies the configured policy and reports whether the request was
// a preflight that has already been answered.
func enableCors(w http.ResponseWriter, r *http.Request) bool {
	cors.apply(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}

	return false
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if enableCors(w, r) {
		return
	}
	if r.URL.Path != "/health" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "OK")
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	if enableCors(w, r) {
		return
	}
	if r.URL.Path != "/test" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "This is a test.")
}

func podsHandler(cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if enableCors(w, r) {
			return
		}

		cache.RLock()
		defer cache.RUnlock()
		json.NewEncoder(w).Encode(cache.Pods)
	}
}

func nodesHandler(cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if enableCors(w, r) {
			return
		}

		cache.RLock()
		defer cache.RUnlock()
		json.NewEncoder(w).Encode(cache.Nodes)
	}
}
