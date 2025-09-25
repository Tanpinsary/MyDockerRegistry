package types

// 查询 Blob 通过 name 和 reference
type GetBlobParams struct {
	RepositoryName string
	Reference      string
	UUID           string
	Digest         string
}

// BlobExists 和 RetrieveBlob 返回参数
type BlobStatus struct {
	Content_Length int
	Digest         string
	Content_Type   string
}

// GetBlobUploadStatus 返回参数
type BlobUploadStatus struct {
	Range    string
	UUID     string
	Location string
}

// CompleteBlobUpload 返回参数
type CompleteBlobUploadResponse struct {
	Digest         string
	Location       string
	Content_Length int
}

// InitiateBlobUpload 接受参数
type InitiateBlobUploadParams struct {
	RepositoryName string
	Mount          string
	From           string
}

// UploadBlobChunk 返回参数
type UploadBlobChunkResponse struct {
	Location string
	Range    string
	UUID     string
}

// blob 标识符
type BlobDescriptor struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

// 201 when mouted, 202 when initiated
type InitiateBlobUploadResponse struct {
	// 通用状态接口
	Status interface {
		GetStatusCode() int
	}

	// 具体状态
	MountedStatus   *BlobUploadMountedStatus
	InitiatedStatus *BlobUploadInitiatedStatus
}

// when 201
type BlobUploadMountedStatus struct {
	Location       string
	Digest         string
	Content_Length int
}

// when 202
type BlobUploadInitiatedStatus struct {
	Range   string
	UUID    string
	Locaton string
	Digest  string
}
