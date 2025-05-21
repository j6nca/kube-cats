package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
   	"log"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	// "encoding/json"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubeInfo struct {
	Nodes interface{} `json:"nodes"`
	Pods  interface{} `json:"pods"`
}

var (
	cachedData KubeInfo
	dataMutex  sync.RWMutex
)

const KUBE_API_URL string = "http://localhost:8087"
// Store pod and node information in variable - we will only update these on a configured interval
var pods, nodes string

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
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

func podsHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.URL.Path != "/pods" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, getPods())
}

func nodesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.URL.Path != "/nodes" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	fmt.Fprintf(w, getNodes())
}

func getPods() string {
	resp, err := http.Get(KUBE_API_URL + "/api/v1/pods?limit=500")
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	pods := sb
	log.Printf(pods)
	return pods
}

func getNodes() string {
	resp, err := http.Get(KUBE_API_URL + "/api/v1/nodes?limit=500")
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	nodes := sb
	log.Printf(nodes)
	return nodes
}

// func getKubeInfo(path string) string {
// 	resp, err := http.Get(KUBE_API_URL + path)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	//Convert the body to type string
// 	sb := string(body)
// 	result := sb
// 	log.Printf(nodes)
// 	return result
// }

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/pods", podsHandler)
	http.HandleFunc("/nodes", nodesHandler)

	fmt.Printf("starting server at port 3030\n")
	if err := http.ListenAndServe(":3030", nil); err != nil {
		panic(err)
	}

	// Initial scrape
	getKubeInfo(clientset)

	// Scrape every 5 minutes
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			getKubeInfo(clientset)
		}
	}()
}

func getKubeInfo(clientset *kubernetes.Clientset) {
	log.Println(" Scraping Kubernetes API...")
	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Printf(" Failed to get nodes: %v", err)
		return
	}

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.Printf(" Failed to get pods: %v", err)
		return
	}

	dataMutex.Lock()
	cachedData = KubeInfo{
		Nodes: nodes.Items,
		Pods:  pods.Items,
	}
	dataMutex.Unlock()

	log.Println(" Kubernetes data cached")
}