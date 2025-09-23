package types

// Content-Type
const (
	ManifestV2MediaType     = "application/vnd.docker.distribution.manifest.v2+json"
	ManifestListV2MediaType = "application/vnd.docker.distribution.manifest.list.v2+json"
)

// ManifestV2

type ManifestV2 struct {
	SchemaVersion int              `json:"schemaVersion"`
	MediaType     string           `json:"mediaType"`
	Config        BlobDescriptor   `json:"config"`
	Layers        []BlobDescriptor `json:"layers"`
}

// ManifestListV2
type ManifestListV2 struct {
	SchemaVersion int                      `json:"schemaversion"`
	MediaType     string                   `json:"mediaType"`
	Manifests     []ManifestListDescriptor `json:"manifests"`
}

type ManifestListDescriptor struct {
	MediaType string   `json:"mediaType"`
	Size      int64    `json:"size"`
	Digest    string   `json:"digest"`
	Platform  Platform `json:"platform"`
}

type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
}

// 返回对象打包

type ManifestData struct {
	Digest    string
	MediaType string
	Schema    []byte
}
