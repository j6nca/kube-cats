package main

import (
	"fmt"
	"net/http"
   	"log"
	"time"
	"os"
)

// scrapeInterval controls how often the Kubernetes API is polled.
// Override with SCRAPE_INTERVAL as a Go duration, e.g. "30s".
func scrapeInterval() time.Duration {
	raw := os.Getenv("SCRAPE_INTERVAL")
	if raw == "" {
		return time.Minute
	}

	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		log.Printf("Invalid SCRAPE_INTERVAL %q, defaulting to 1m", raw)
		return time.Minute
	}

	return d
}

func main() {
	cache := &Cache{}
	clients, err := getClients()
	if err != nil {
		log.Fatalf("Failed to get Kubernetes client: %v", err)
	}

	go func() {
		ticker := time.NewTicker(scrapeInterval())
		defer ticker.Stop()
		for {
			getKubeInfo(clients, cache)
			<-ticker.C
		}
	}()

	http.HandleFunc("/health", healthHandler)
  http.HandleFunc("/test", testHandler)
	http.HandleFunc("/pods", podsHandler(cache))
	http.HandleFunc("/nodes", nodesHandler(cache))
	
	port := os.Getenv("API_PORT")
	if port == "" {
			port = "8001"
	}

	fmt.Printf("starting server at port %s\n", port)
	if err := http.ListenAndServe(":" + port, nil); err != nil {
		panic(err)
	}
}
