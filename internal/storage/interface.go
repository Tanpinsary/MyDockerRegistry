package storage

import (
	"my_docker_registry/internal/types"
)

type StorageDriver interface {
	// Manifest API
	GetManifest(name, reference string) (*types.ManifestData, error)
	PutManifest(name, reference string) (string, string, int, error)
	CheckManifestExists(name, reference string) (int, string, string)
	DeleteManifest(name, reference string)

	// Blob API
	InitiateBlobUpload(name, mount, from string) (*types.BlobUploadResult, error)
	BlobExists(repository, digest string) (*types.BlobDescriptor, error)
	RetrieveBlob()
	GetBlobUploadStatus()
	CompleteBlobUpload()
	UploadBlobChunk()
	CancelBlobUpload()
}
