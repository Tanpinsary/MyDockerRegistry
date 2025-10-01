package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// Content-Type
const (
	ManifestV2MediaType     = "application/vnd.docker.distribution.manifest.v2+json"
	ManifestListV2MediaType = "application/vnd.docker.distribution.manifest.list.v2+json"
	ConfigV1MediaType       = "application/vnd.docker.container.image.v1+json"
	LayerMediaType          = "application/vnd.docker.image.rootfs.diff.tar.gzip"
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

type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Variant      string `json:"variant,omitempty"`
}

// ManifestList 代表一个多架构 Docker 镜像清单列表 (v2)
type ManifestList struct {
	SchemaVersion int                      `json:"schemaVersion"`
	MediaType     string                   `json:"mediaType"`
	Manifests     []ManifestListDescriptor `json:"manifests"`
}

// ManifestListDescriptor 描述了清单列表中的一个清单
type ManifestListDescriptor struct {
	MediaType string    `json:"mediaType"`
	Size      int64     `json:"size"`
	Digest    string    `json:"digest"`
	Platform  *Platform `json:"platform,omitempty"`
}

// ManifestResponse 包含 manifest 的原始内容和解析后的数据
type ManifestResponse struct {
	Content      []byte        // 原始 JSON 内容
	MediaType    string        // 检测到的媒体类型
	Manifest     *Manifest     // 解析后的 v2 manifest (如果适用)
	ManifestList *ManifestList // 解析后的 manifest list (如果适用)
}

// 接受 PutImageManifest 和 CheckImageManifest 返回参数
type ManifestData struct {
	Digest        string
	Location      string
	ContentLength int    //always 0
	MediaType     string // 添加媒体类型字段
}

// DetectManifestMediaType 根据 manifest 内容检测并返回正确的 Content-Type
func DetectManifestMediaType(content []byte) string {
	// 尝试解析为通用结构来检查 mediaType 字段
	var base struct {
		MediaType string `json:"mediaType"`
	}

	if err := json.Unmarshal(content, &base); err != nil {
		// 如果解析失败，默认返回 v2 manifest
		return ManifestV2MediaType
	}

	// 根据 mediaType 字段判断类型
	switch base.MediaType {
	case ManifestListV2MediaType:
		return ManifestListV2MediaType
	case ManifestV2MediaType:
		return ManifestV2MediaType
	default:
		// 对于没有明确 mediaType 或未知类型，尝试解析结构判断
		var manifestV2 Manifest
		var manifestList ManifestList

		if err := json.Unmarshal(content, &manifestList); err == nil && len(manifestList.Manifests) > 0 {
			return ManifestListV2MediaType
		}

		if err := json.Unmarshal(content, &manifestV2); err == nil && len(manifestV2.Layers) > 0 {
			return ManifestV2MediaType
		}

		// 默认情况
		return ManifestV2MediaType
	}
}

// CalculateDigest 计算内容的 SHA256 摘要
func CalculateDigest(content []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(content))
}
