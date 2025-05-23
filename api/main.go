package main

import (
	"fmt"
	"net/http"
   	"log"
	"time"
	"os"
)

func main() {
	cache := &Cache{}
	clientset, err := getClientset()
	if err != nil {
		log.Fatalf("Failed to get Kubernetes client: %v", err)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			getKubeInfo(clientset, cache)
			<-ticker.C
		}
	}()

	http.HandleFunc("/health", healthHandler)
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
