package storage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"my_docker_registry/internal/types"
	"os"
	"path/filepath"
	"strings"
)

// fileSystemDriver 实现了 StorageDriver 接口，使用本地文件系统作为后端。
type fileSystemDriver struct {
	rootDirectory string
}

// NewFileSystemDriver 创建一个新的 fileSystemDriver 实例。
// rootDirectory 是用于存储所有数据的根目录。
func NewFileSystemDriver(rootDirectory string) (StorageDriver, error) {
	// 确保根目录存在
	if err := os.MkdirAll(rootDirectory, 0755); err != nil {
		return nil, err
	}
	return &fileSystemDriver{rootDirectory: rootDirectory}, nil
}

// 预处理函数

// manifestPath 根据摘要值构建清单内容文件的路径。
// 路径格式: <root>/repositories/<name>/_manifests/revisions/sha256/<hash>
func (d *fileSystemDriver) manifestPath(repoName, digest string) string {
	hash := strings.TrimPrefix(digest, "sha256:")
	return filepath.Join(d.rootDirectory, "repositories", repoName, "_manifests", "revisions", "sha256", hash)
}

// tagPath 构建标签链接文件的路径。
// 路径格式: <root>/repositories/<name>/_manifests/tags/<tag>/current/link
func (d *fileSystemDriver) tagPath(repoName, tagName string) string {
	return filepath.Join(d.rootDirectory, "repositories", repoName, "_manifests", "tags", tagName, "current", "link")
}

// resolveReference 接受一个引用（标签或摘要）并返回摘要值。
// 如果引用是标签，则读取链接文件来查找摘要。
// 如果引用是摘要，则直接返回。
func (d *fileSystemDriver) resolveReference(repoName, reference string) (string, error) {
	if strings.HasPrefix(reference, "sha256:") {
		// 已经是摘要值了。
		return reference, nil
	}

	// 把标签解析为为摘要值。
	tagPath := d.tagPath(repoName, reference)
	digestBytes, err := os.ReadFile(tagPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 标签未找到，返回标准的清单未知错误。
			return "", types.NewError(types.ErrorCodeManifestUnknown, "manifest unknown", map[string]string{"reference": reference})
		}
		return "", err // 其他类型的错误（如权限问题）。
	}
	return string(digestBytes), nil
}

// calculateDigest 计算数据的 sha256 摘要。
func calculateDigest(content []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(content))
}

// --- Manifest API ---

func (d *fileSystemDriver) GetManifest(params types.GetManifestParams) (*types.Manifest, error) {
	// 1. 将 tag 处理为 digest
	digest, err := d.resolveReference(params.RepositoryName, params.Reference)
	if err != nil {
		return nil, err // 错误处理在 resolveReference 内
	}

	// 2. 通过 reference 获取 manifest
	manifestPath := d.manifestPath(params.RepositoryName, digest)
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.NewError(types.ErrorCodeManifestUnknown, "manifest unknown", map[string]string{"digest": digest})
		}
		return nil, err
	}

	// 3. 反序列化
	var manifest types.Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, types.NewError(types.ErrorCodeManifestInvalid, "failed to parse manifest", err.Error())
	}

	return &manifest, nil
}

func (d *fileSystemDriver) PutManifest(params types.PutManifestParams) (*types.ManifestData, error) {
	// 1. 计算上传内容的 digest。
	digest := calculateDigest(params.Content)

	// 2. 检测 digest 是否匹配
	if strings.HasPrefix(params.Reference, "sha256:") && params.Reference != digest {
		return nil, types.NewError(types.ErrorCodeDigestInvalid, "digest mismatch", nil)
	}

	// 3. 校验 manifest 引用的所有 blob 是否都存在。
	var manifest types.Manifest
	if err := json.Unmarshal(params.Content, &manifest); err != nil {
		return nil, types.NewError(types.ErrorCodeManifestInvalid, "failed to parse manifest", err.Error())
	}

	// 检查 config blob
	blobParams := types.GetBlobParams{RepositoryName: params.RepositoryName, Digest: manifest.Config.Digest}
	if _, err := d.BlobExists(blobParams); err != nil {
		return nil, types.NewError(types.ErrorCodeBlobUnknown, "config blob unknown", map[string]string{"digest": manifest.Config.Digest})
	}
	// 检查所有 layer blobs
	for _, layer := range manifest.Layers {
		blobParams.Digest = layer.Digest
		if _, err := d.BlobExists(blobParams); err != nil {
			return nil, types.NewError(types.ErrorCodeBlobUnknown, "layer blob unknown", map[string]string{"digest": layer.Digest})
		}
	}

	// 4. 存储 manifest 内容文件。
	manifestPath := d.manifestPath(params.RepositoryName, digest)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(manifestPath, params.Content, 0644); err != nil {
		return nil, err
	}

	// 5. 如果 reference 是 tag，则创建 tag 链接文件。
	if !strings.HasPrefix(params.Reference, "sha256:") {
		tagPath := d.tagPath(params.RepositoryName, params.Reference)
		if err := os.MkdirAll(filepath.Dir(tagPath), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(tagPath, []byte(digest), 0644); err != nil {
			return nil, err
		}
	}

	// 6. 返回成功结果。
	result := &types.ManifestData{
		Digest:   digest,
		Location: fmt.Sprintf("/v2/%s/manifests/%s", params.RepositoryName, digest),
	}
	return result, nil
}

func (d *fileSystemDriver) ManifestExists(params types.GetManifestParams) (*types.ManifestData, error) {
	// 1. 解析引用，获取 digest
	digest, err := d.resolveReference(params.RepositoryName, params.Reference)
	if err != nil {
		return nil, err // 错误已在 resolveReference 中创建
	}

	// 2. 获取 manifest 内容文件的路径
	manifestPath := d.manifestPath(params.RepositoryName, digest) // <-- 使用 digest，而不是 reference

	// 3. 使用 os.Stat 获取文件元信息，而不是读取整个文件
	info, err := os.Stat(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，返回标准的 manifest unknown 错误
			return nil, types.NewError(types.ErrorCodeManifestUnknown, "manifest unknown", map[string]string{"digest": digest})
		}
		return nil, err // 其他文件系统错误
	}

	// 4. 构建并返回 ManifestData
	manifestData := &types.ManifestData{
		Digest: digest,
		// Location 通常由 handler 构建，但在这里返回也很方便
		Location:      fmt.Sprintf("/v2/%s/manifests/%s", params.RepositoryName, digest),
		ContentLength: int(info.Size()), // 从文件信息中获取大小
	}

	return manifestData, nil
}

func (d *fileSystemDriver) DeleteManifest(params types.GetManifestParams) {
	panic("not implemented")
}

// --- Blob API ---

func (d *fileSystemDriver) InitiateBlobUpload(params types.InitiateBlobUploadParams) (*types.InitiateBlobUploadResponse, error) {
	panic("not implemented")
}

func (d *fileSystemDriver) BlobExists(params types.GetBlobParams) (*types.BlobStatus, error) {
	panic("not implemented")
}

func (d *fileSystemDriver) RetrieveBlob(params types.GetBlobParams) (*types.BlobStatus, error) {
	panic("not implemented")
}

func (d *fileSystemDriver) GetBlobUploadStatus(params types.GetBlobParams) (*types.BlobUploadStatus, error) {
	panic("not implemented")
}

func (d *fileSystemDriver) CompleteBlobUpload(params types.GetBlobParams) (*types.CompleteBlobUploadResponse, error) {
	panic("not implemented")
}

func (d *fileSystemDriver) UploadBlobChunk(params types.GetBlobParams) (*types.UploadBlobChunkResponse, error) {
	panic("not implemented")
}

func (d *fileSystemDriver) CancelBlobUpload(params types.GetBlobParams) (int, error) {
	panic("not implemented")
}
