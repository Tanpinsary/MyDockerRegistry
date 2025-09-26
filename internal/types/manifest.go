package types

// Content-Type
const (
	ManifestV2MediaType     = "application/vnd.docker.distribution.manifest.v2+json"
	ManifestListV2MediaType = "application/vnd.docker.distribution.manifest.list.v2+json"
)

// 查询 Manifest 通过 name 和 reference
type GetManifestParams struct {
	RepositoryName string
	Reference      string
}

// Manifest 代表一个 Docker 镜像清单 (v2)。
// 这是与 Registry API 交互时的核心数据结构。
type Manifest struct {
	SchemaVersion int              `json:"schemaVersion"`
	MediaType     string           `json:"mediaType"`
	Config        BlobDescriptor   `json:"config"`
	Layers        []BlobDescriptor `json:"layers"`
}

// PutManifestParams 封装了 PutManifest 方法所需的所有参数。
type PutManifestParams struct {
	RepositoryName string
	Reference      string
	MediaType      string
	Content        []byte
}

// ManifestLDescriptor 描述了一个可通过内容寻址的组件（如 config 或 layer）。
// 它包含了媒体类型、大小和内容的摘要 (digest)。
type ManifestListDescriptor struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
}

// 接受 PutImageManifest 和 CheckImageManifest 返回参数
type ManifestData struct {
	Digest        string
	Location      string
	ContentLength int //always 0
}
