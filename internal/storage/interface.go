package storage

import (
	"my_docker_registry/internal/types"
)

type StorageDriver interface {
	// Manifest API
	GetManifest(params types.GetManifestParams) (*types.Manifest, error)
	PutManifest(params types.PutManifestParams) (*types.ManifestData, error)
	ManifestExists(params types.GetManifestParams) (*types.ManifestData, error)
	DeleteManifest(params types.GetManifestParams) error

	// Blob API
	InitiateBlobUpload(params types.InitiateBlobUploadParams) (*types.InitiateBlobUploadResponse, error)
	BlobExists(params types.GetBlobParams) (*types.BlobStatus, error)
	RetrieveBlob(params types.GetBlobParams) (*types.BlobStatus, error)
	GetBlobUploadStatus(params types.GetBlobParams) (*types.BlobUploadStatus, error)
	CompleteBlobUpload(params types.GetBlobParams) (*types.CompleteBlobUploadResponse, error)
	UploadBlobChunk(params types.UploadBlobChunkParams) (*types.UploadBlobChunkResponse, error)
	CancelBlobUpload(params types.GetBlobParams) (int, error)
}
