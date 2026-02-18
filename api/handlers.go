package main

import (
	"net/http"
	"encoding/json"
	"fmt"
  "os"
)

func enableCors(w *http.ResponseWriter) {
  allowOrigin := "*"

  if os.Getenv("ENVIRONMENT") == "production" {
    allowOrigin = "https://meow.j6n.dev"
  }
    
	(*w).Header().Set("Access-Control-Allow-Origin", allowOrigin)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
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
	enableCors(&w)
	if r.URL.Path != "/test" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "This is a test")
}

func podsHandler(cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)

		cache.RLock()
		defer cache.RUnlock()
		json.NewEncoder(w).Encode(cache.Pods)
	}
}

func nodesHandler(cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)

		cache.RLock()
		defer cache.RUnlock()
		json.NewEncoder(w).Encode(cache.Nodes)
	}
}
