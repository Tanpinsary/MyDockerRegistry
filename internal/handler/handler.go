package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"my_docker_registry/internal/storage"
	"my_docker_registry/internal/types"

	"github.com/gorilla/mux"
)

// RegistryHandler 包含所有 API 端点的处理逻辑
type RegistryHandler struct {
	storage storage.StorageDriver
}

// NewRegistryHandler 创建一个新的 RegistryHandler 实例
func NewRegistryHandler(storageDriver storage.StorageDriver) *RegistryHandler {
	return &RegistryHandler{
		storage: storageDriver,
	}
}

// writeErrorResponse 写入标准的错误响应
func (h *RegistryHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	// 检查是否是 RegistryError
	if regErr, ok := err.(types.RegistryError); ok {
		types.WriteErrorResponse(w, statusCode, regErr)
		return
	}

	// 否则创建一个通用错误
	genericError := types.NewError(types.ErrorCodeUnsupported, err.Error(), nil)
	types.WriteErrorResponse(w, statusCode, genericError)
}

// parseContentRange 解析 Content-Range 头部
func parseContentRange(rangeHeader string) (int64, int64, error) {
	if rangeHeader == "" {
		return 0, 0, fmt.Errorf("missing Content-Range header")
	}

	// Content-Range: bytes 0-65535
	if !strings.HasPrefix(rangeHeader, "bytes ") {
		return 0, 0, fmt.Errorf("invalid Content-Range format")
	}

	rangeStr := strings.TrimPrefix(rangeHeader, "bytes ")
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range start: %v", err)
	}

	end, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range end: %v", err)
	}

	return start, end, nil
}

// API version check handler
func (h *RegistryHandler) APIVersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	w.WriteHeader(http.StatusOK)
	log.Println("Responded to /v2/ API version check")
}

// === Manifest Handlers ===

// ManifestHandler 处理 manifest 相关的请求 (GET, PUT, HEAD, DELETE)
func (h *RegistryHandler) ManifestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	reference := vars["reference"]

	switch r.Method {
	case http.MethodGet:
		h.getManifest(w, r, name, reference)
	case http.MethodPut:
		h.putManifest(w, r, name, reference)
	case http.MethodHead:
		h.headManifest(w, r, name, reference)
	case http.MethodDelete:
		h.deleteManifest(w, r, name, reference)
	default:
		h.writeErrorResponse(w, http.StatusMethodNotAllowed,
			types.NewError(types.ErrorCodeUnsupported, "Method not allowed", nil))
	}
}

// getManifest 处理 GET /v2/{name}/manifests/{reference}
func (h *RegistryHandler) getManifest(w http.ResponseWriter, r *http.Request, name, reference string) {
	params := types.GetManifestParams{
		RepositoryName: name,
		Reference:      reference,
	}

	manifest, err := h.storage.GetManifest(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeManifestUnknown:
				// 404 Manifest not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewManifestUnknownError(name, reference))
			case types.ErrorCodeNameUnknown:
				// 404 Repository not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewNameUnknownError(name))
			case types.ErrorCodeNameInvalid:
				// 400 Invalid name or reference
				types.WriteErrorResponse(w, http.StatusBadRequest, types.NewNameInvalidError(name))
			default:
				types.WriteErrorResponse(w, http.StatusBadRequest, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
	w.Header().Set("Docker-Content-Digest", manifest.Config.Digest)
	w.WriteHeader(http.StatusOK)

	// 返回 manifest JSON
	if err := json.NewEncoder(w).Encode(manifest); err != nil {
		log.Printf("Failed to encode manifest: %v", err)
	}
}

// putManifest 处理 PUT /v2/{name}/manifests/{reference}
func (h *RegistryHandler) putManifest(w http.ResponseWriter, r *http.Request, name, reference string) {
	// 读取请求体
	content, err := io.ReadAll(r.Body)
	if err != nil {
		types.WriteErrorResponse(w, http.StatusBadRequest,
			types.NewManifestInvalidError("Failed to read manifest"))
		return
	}
	defer r.Body.Close()

	params := types.PutManifestParams{
		RepositoryName: name,
		Reference:      reference,
		Content:        content,
	}

	result, err := h.storage.PutManifest(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeNameUnknown:
				// 404 Repository not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewNameUnknownError(name))
			case types.ErrorCodeNameInvalid, types.ErrorCodeManifestInvalid:
				// 400 Invalid name, reference, or manifest
				types.WriteErrorResponse(w, http.StatusBadRequest, regErr)
			default:
				types.WriteErrorResponse(w, http.StatusBadRequest, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 设置响应头
	w.Header().Set("Location", result.Location)
	w.Header().Set("Docker-Content-Digest", result.Digest)
	w.WriteHeader(http.StatusCreated)
}

// headManifest 处理 HEAD /v2/{name}/manifests/{reference}
func (h *RegistryHandler) headManifest(w http.ResponseWriter, r *http.Request, name, reference string) {
	params := types.GetManifestParams{
		RepositoryName: name,
		Reference:      reference,
	}

	manifestData, err := h.storage.ManifestExists(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeManifestUnknown:
				// 404 Manifest not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewManifestUnknownError(name, reference))
			default:
				types.WriteErrorResponse(w, http.StatusNotFound, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
	w.Header().Set("Docker-Content-Digest", manifestData.Digest)
	w.Header().Set("Content-Length", strconv.Itoa(manifestData.ContentLength))
	w.WriteHeader(http.StatusOK)
}

// deleteManifest 处理 DELETE /v2/{name}/manifests/{reference}
func (h *RegistryHandler) deleteManifest(w http.ResponseWriter, r *http.Request, name, reference string) {
	params := types.GetManifestParams{
		RepositoryName: name,
		Reference:      reference,
	}

	err := h.storage.DeleteManifest(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeManifestUnknown, types.ErrorCodeNameUnknown:
				// 404 Manifest or repository not found
				types.WriteErrorResponse(w, http.StatusNotFound, regErr)
			default:
				types.WriteErrorResponse(w, http.StatusNotFound, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 成功删除，返回 202 Accepted
	w.WriteHeader(http.StatusAccepted)
}

// === Blob Handlers ===

// BlobHandler 处理 blob 相关的请求 (HEAD, GET)
func (h *RegistryHandler) BlobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	digest := vars["digest"]

	switch r.Method {
	case http.MethodHead:
		h.headBlob(w, r, name, digest)
	case http.MethodGet:
		h.getBlob(w, r, name, digest)
	default:
		h.writeErrorResponse(w, http.StatusMethodNotAllowed,
			types.NewError(types.ErrorCodeUnsupported, "Method not allowed", nil))
	}
}

// headBlob 处理 HEAD /v2/{name}/blobs/{digest}
func (h *RegistryHandler) headBlob(w http.ResponseWriter, r *http.Request, name, digest string) {
	params := types.GetBlobParams{
		RepositoryName: name,
		Digest:         digest,
	}

	status, err := h.storage.BlobExists(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeBlobUnknown:
				// 404 Blob not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewBlobUnknownError(digest))
			default:
				types.WriteErrorResponse(w, http.StatusNotFound, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 设置响应头
	w.Header().Set("Content-Length", strconv.Itoa(status.ContentLength))
	w.Header().Set("Docker-Content-Digest", status.Digest)
	w.WriteHeader(http.StatusOK)
}

// getBlob 处理 GET /v2/{name}/blobs/{digest}
func (h *RegistryHandler) getBlob(w http.ResponseWriter, r *http.Request, name, digest string) {
	params := types.GetBlobParams{
		RepositoryName: name,
		Digest:         digest,
	}

	status, err := h.storage.RetrieveBlob(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeBlobUnknown:
				// 404 Blob not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewBlobUnknownError(digest))
			default:
				types.WriteErrorResponse(w, http.StatusNotFound, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 设置响应头
	w.Header().Set("Content-Length", strconv.Itoa(status.ContentLength))
	w.Header().Set("Content-Type", status.ContentType)
	w.Header().Set("Docker-Content-Digest", status.Digest)
	w.WriteHeader(http.StatusOK)

	// 注意：这里我们只返回了头部信息，实际的文件内容需要从存储层读取并写入响应体
	// 在真实实现中，需要添加文件流式传输逻辑
}

// InitiateBlobUploadHandler 处理 POST /v2/{name}/blobs/uploads/
func (h *RegistryHandler) InitiateBlobUploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// 获取查询参数
	mount := r.URL.Query().Get("mount")
	from := r.URL.Query().Get("from")

	params := types.InitiateBlobUploadParams{
		RepositoryName: name,
		Mount:          mount,
		From:           from,
	}

	response, err := h.storage.InitiateBlobUpload(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeNameUnknown:
				// 404 Repository not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewNameUnknownError(name))
			default:
				h.writeErrorResponse(w, http.StatusInternalServerError, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 根据响应的状态设置不同的响应头和状态码
	if response.MountedStatus != nil {
		// 201 Created - 挂载成功
		w.Header().Set("Location", response.MountedStatus.Location)
		w.Header().Set("Docker-Content-Digest", response.MountedStatus.Digest)
		w.WriteHeader(http.StatusCreated)
	} else if response.InitiatedStatus != nil {
		// 202 Accepted - 上传会话创建
		w.Header().Set("Location", response.InitiatedStatus.Location)
		w.Header().Set("Range", response.InitiatedStatus.Range)
		w.Header().Set("Docker-Upload-UUID", response.InitiatedStatus.UUID)
		w.WriteHeader(http.StatusAccepted)
	} else {
		h.writeErrorResponse(w, http.StatusInternalServerError,
			fmt.Errorf("invalid response from storage layer"))
	}
}

// BlobUploadHandler 处理 blob upload 相关的请求 (GET, PATCH, PUT, DELETE)
func (h *RegistryHandler) BlobUploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	uuid := vars["uuid"]

	switch r.Method {
	case http.MethodGet:
		h.getBlobUploadStatus(w, r, name, uuid)
	case http.MethodPatch:
		h.uploadBlobChunk(w, r, name, uuid)
	case http.MethodPut:
		h.completeBlobUpload(w, r, name, uuid)
	case http.MethodDelete:
		h.cancelBlobUpload(w, r, name, uuid)
	default:
		h.writeErrorResponse(w, http.StatusMethodNotAllowed,
			types.NewError(types.ErrorCodeUnsupported, "Method not allowed", nil))
	}
}

// getBlobUploadStatus 处理 GET /v2/{name}/blobs/uploads/{uuid}
func (h *RegistryHandler) getBlobUploadStatus(w http.ResponseWriter, r *http.Request, name, uuid string) {
	params := types.GetBlobParams{
		RepositoryName: name,
		UUID:           uuid,
	}

	status, err := h.storage.GetBlobUploadStatus(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeBlobUploadUnknown:
				// 404 Upload session not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewBlobUploadUnknownError(uuid))
			default:
				types.WriteErrorResponse(w, http.StatusNotFound, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 设置响应头
	w.Header().Set("Location", status.Location)
	w.Header().Set("Range", status.Range)
	w.Header().Set("Docker-Upload-UUID", status.UUID)
	w.WriteHeader(http.StatusNoContent)
}

// uploadBlobChunk 处理 PATCH /v2/{name}/blobs/uploads/{uuid}
func (h *RegistryHandler) uploadBlobChunk(w http.ResponseWriter, r *http.Request, name, uuid string) {
	// 读取 Content-Range 头
	contentRange := r.Header.Get("Content-Range")
	if contentRange == "" {
		types.WriteErrorResponse(w, http.StatusBadRequest,
			types.NewRangeInvalidError("Content-Range header required"))
		return
	}

	rangeFrom, rangeTo, err := parseContentRange(contentRange)
	if err != nil {
		types.WriteErrorResponse(w, http.StatusBadRequest,
			types.NewRangeInvalidError(contentRange))
		return
	}

	// 读取请求体
	content, err := io.ReadAll(r.Body)
	if err != nil {
		types.WriteErrorResponse(w, http.StatusBadRequest,
			types.NewBlobUploadInvalidError("Failed to read chunk data"))
		return
	}
	defer r.Body.Close()

	params := types.UploadBlobChunkParams{
		RepositoryName: name,
		UUID:           uuid,
		Content:        content,
		RangeFrom:      rangeFrom,
		RangeTo:        rangeTo,
	}

	response, err := h.storage.UploadBlobChunk(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeBlobUploadUnknown:
				// 404 Upload session not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewBlobUploadUnknownError(uuid))
			case types.ErrorCodeRangeInvalid:
				// 400 Malformed content or range
				types.WriteErrorResponse(w, http.StatusBadRequest, regErr)
			default:
				types.WriteErrorResponse(w, http.StatusBadRequest, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 设置响应头
	w.Header().Set("Location", response.Location)
	w.Header().Set("Range", response.Range)
	w.Header().Set("Docker-Upload-UUID", response.UUID)
	w.WriteHeader(http.StatusAccepted)
}

// completeBlobUpload 处理 PUT /v2/{name}/blobs/uploads/{uuid}
func (h *RegistryHandler) completeBlobUpload(w http.ResponseWriter, r *http.Request, name, uuid string) {
	// 从查询参数获取 digest
	digest := r.URL.Query().Get("digest")
	if digest == "" {
		types.WriteErrorResponse(w, http.StatusBadRequest,
			types.NewDigestInvalidError("digest parameter required"))
		return
	}

	params := types.GetBlobParams{
		RepositoryName: name,
		UUID:           uuid,
		Digest:         digest,
	}

	response, err := h.storage.CompleteBlobUpload(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeBlobUploadUnknown:
				// 404 Upload session not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewBlobUploadUnknownError(uuid))
			case types.ErrorCodeDigestInvalid:
				// 400 Invalid digest or missing parameters
				types.WriteErrorResponse(w, http.StatusBadRequest, types.NewDigestInvalidError(digest))
			default:
				types.WriteErrorResponse(w, http.StatusBadRequest, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 设置响应头
	w.Header().Set("Location", response.Location)
	w.Header().Set("Docker-Content-Digest", response.Digest)
	w.WriteHeader(http.StatusCreated)
}

// cancelBlobUpload 处理 DELETE /v2/{name}/blobs/uploads/{uuid}
func (h *RegistryHandler) cancelBlobUpload(w http.ResponseWriter, r *http.Request, name, uuid string) {
	params := types.GetBlobParams{
		RepositoryName: name,
		UUID:           uuid,
	}

	statusCode, err := h.storage.CancelBlobUpload(params)
	if err != nil {
		if regErr, ok := err.(types.RegistryError); ok {
			switch regErr.Code {
			case types.ErrorCodeBlobUploadUnknown:
				// 404 Upload session not found
				types.WriteErrorResponse(w, http.StatusNotFound, types.NewBlobUploadUnknownError(uuid))
			default:
				types.WriteErrorResponse(w, http.StatusNotFound, regErr)
			}
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// 使用存储层返回的状态码
	w.WriteHeader(statusCode)
}
