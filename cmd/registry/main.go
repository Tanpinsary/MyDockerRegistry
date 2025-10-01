package main

import (
	"log"
	"net/http"

	"my_docker_registry/internal/handler"
	"my_docker_registry/internal/storage"

	"github.com/gorilla/mux"
)

func main() {
	// 初始化存储层
	storageDriver, err := storage.NewFileSystemDriver("./registry_data")
	if err != nil {
		log.Fatalf("Failed to initialize storage driver: %v", err)
	}

	// 初始化处理层
	registryHandler := handler.NewRegistryHandler(storageDriver)

	// 创建路由器
	r := mux.NewRouter()

	// 基础 API 版本检查
	r.HandleFunc("/v2/", registryHandler.APIVersionHandler).Methods("GET")

	// Manifests 相关路由
	// GET, PUT, HEAD, DELETE /v2/{name}/manifests/{reference}
	r.HandleFunc("/v2/{name:.+}/manifests/{reference}", registryHandler.ManifestHandler).Methods("GET", "PUT", "HEAD", "DELETE")

	// Blobs 相关路由
	// HEAD, GET /v2/{name}/blobs/{digest}
	r.HandleFunc("/v2/{name:.+}/blobs/{digest}", registryHandler.BlobHandler).Methods("HEAD", "GET")

	// POST /v2/{name}/blobs/uploads/
	r.HandleFunc("/v2/{name:.+}/blobs/uploads/", registryHandler.InitiateBlobUploadHandler).Methods("POST")

	// GET, PATCH, PUT, DELETE /v2/{name}/blobs/uploads/{uuid}
	r.HandleFunc("/v2/{name:.+}/blobs/uploads/{uuid}", registryHandler.BlobUploadHandler).Methods("GET", "PATCH", "PUT", "DELETE")

	port := "5000"
	log.Printf("Starting Docker Registry backend on port %s...", port)
	log.Printf("Registry data will be stored in: ./registry_data")

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
