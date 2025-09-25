package storage

import (
	"my_docker_registry/internal/types"
)

type StorageDriver interface {
	// Manifest API
	GetImageManifest(params types.GetManifestParams) (*types.Manifest, error)
	PutImageManifest(params types.GetManifestParams) (*types.ManifestData, error)
	CheckImageManifestExists(params types.GetManifestParams) (*types.ManifestData, error)
	DeleteImageManifest(params types.GetManifestParams)

	// Blob API
	InitiateBlobUpload(params types.InitiateBlobUploadParams) (*types.InitiateBlobUploadResponse, error)
	BlobExists(params types.GetBlobParams) (*types.BlobStatus, error)
	RetrieveBlob(params types.GetBlobParams) (*types.BlobStatus, error)
	GetBlobUploadStatus(params types.GetBlobParams) (*types.BlobUploadStatus, error)
	CompleteBlobUpload(params types.GetBlobParams) (*types.CompleteBlobUploadResponse, error)
	UploadBlobChunk(params types.GetBlobParams) (*types.UploadBlobChunkResponse, error)
	CancelBlobUpload(params types.GetBlobParams) (int, error)
}
