package types

// blob 标识符
type BlobDescriptor struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

// blob 上传
type BlobUpload struct {
	UUID          string
	Repository    string
	TemporaryPath string
	CurrentSize   int64
}

type BlobUploadResult struct {
	// 新上传
	UploadUUID string `json:"uuid,omitempty"`
	// 成功挂载
	MountedBlob *BlobDescriptor `json:"mounted_blob,omitempty"`
}
