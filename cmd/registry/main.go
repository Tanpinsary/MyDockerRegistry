package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	notImplementedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		log.Printf("Received request for unimplemented endpoint: %s %s, Vars: %v", r.Method, r.URL.Path, vars)
		http.Error(w, "API endpoint not implemented yet.", http.StatusNotImplemented)
	})

	// 基础 API 版本检查
	r.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Docker-Distribution-API-Version", "registry/2")
		w.WriteHeader(http.StatusOK)
		log.Println("Responded to /v2/ API version check")
	}).Methods("GET")

	// Manifests 相关路由
	// GET, PUT, HEAD, DELETE /v2/{name}/manifests/{reference}
	r.HandleFunc("/v2/{name:.+}/manifests/{reference}", notImplementedHandler).Methods("GET", "PUT", "HEAD", "DELETE")

	// Blobs 相关路由
	// HEAD, GET /v2/{name}/blobs/{digest}
	r.HandleFunc("/v2/{name:.+}/blobs/{digest}", notImplementedHandler).Methods("HEAD", "GET")

	// POST /v2/{name}/blobs/uploads/
	r.HandleFunc("/v2/{name:.+}/blobs/uploads/", notImplementedHandler).Methods("POST")

	// GET, PATCH, PUT, DELETE /v2/{name}/blobs/uploads/{uuid}
	r.HandleFunc("/v2/{name:.+}/blobs/uploads/{uuid}", notImplementedHandler).Methods("GET", "PATCH", "PUT", "DELETE")

	port := "5000"
	log.Printf("Starting Docker Registry backend on port %s...", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
