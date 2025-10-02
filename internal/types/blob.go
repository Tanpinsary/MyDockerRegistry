package types

import "io"

// 查询 Blob 通过 name 和 reference
type GetBlobParams struct {
	RepositoryName string
	Reference      string
	UUID           string
	Digest         string
}

// BlobExists 和 RetrieveBlob 返回参数
type BlobStatus struct {
	ContentLength int
	Digest        string
	ContentType   string
	Reader        io.ReadCloser // 用于 GET 请求时输出文件内容
}

// GetBlobUploadStatus 返回参数
type BlobUploadStatus struct {
	Range    string
	UUID     string
	Location string
}

// CompleteBlobUpload 接受参数
type CompleteBlobUploadParams struct {
	RepositoryName string
	UUID           string
	Digest         string
	Data           []byte
}

// CompleteBlobUpload 返回参数
type CompleteBlobUploadResponse struct {
	Digest        string
	Location      string
	ContentLength int
}

// InitiateBlobUpload 接受参数
type InitiateBlobUploadParams struct {
	RepositoryName string
	Mount          string
	From           string
}

// UploadBlobChunk 接受参数
type UploadBlobChunkParams struct {
	RepositoryName string
	UUID           string
	Content        []byte // 来自请求体
	RangeFrom      int64  // 从 Content-Range 解析出的起始字节
	RangeTo        int64  // 从 Content-Range 解析出的结束字节
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
	Location      string
	Digest        string
	ContentLength int
}

func (s *BlobUploadMountedStatus) GetStatusCode() int {
	return 201 // 201 Created
}

// when 202
type BlobUploadInitiatedStatus struct {
	Range    string
	UUID     string
	Location string
	Digest   string
}

func (s *BlobUploadInitiatedStatus) GetStatusCode() int {
	return 202 // 202 Accepted
}
