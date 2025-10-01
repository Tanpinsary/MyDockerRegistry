package types

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode represents a Docker Registry API error code.
type ErrorCode string

// Standard Docker Registry error codes for 400 & 404 responses.
const (
	// 404 Not Found
	ErrorCodeBlobUnknown       ErrorCode = "BLOB_UNKNOWN"
	ErrorCodeManifestUnknown   ErrorCode = "MANIFEST_UNKNOWN"
	ErrorCodeNameUnknown       ErrorCode = "NAME_UNKNOWN"
	ErrorCodeBlobUploadUnknown ErrorCode = "BLOB_UPLOAD_UNKNOWN"

	// 400 Bad Request
	ErrorCodeDigestInvalid     ErrorCode = "DIGEST_INVALID"
	ErrorCodeSizeInvalid       ErrorCode = "SIZE_INVALID"
	ErrorCodeManifestInvalid   ErrorCode = "MANIFEST_INVALID"
	ErrorCodeBlobUploadInvalid ErrorCode = "BLOB_UPLOAD_INVALID"
	ErrorCodeNameInvalid       ErrorCode = "NAME_INVALID"
	ErrorCodeUnsupported       ErrorCode = "UNSUPPORTED"
	ErrorCodeRangeInvalid      ErrorCode = "RANGE_INVALID"
)

// RegistryError defines the structure for a single error.
type RegistryError struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail,omitempty"`
}

// RegistryErrorResponse defines the structure for the error response body.
type RegistryErrorResponse struct {
	Errors []RegistryError `json:"errors"`
}

// Error implements the error interface for RegistryError.
func (e RegistryError) Error() string {
	return fmt.Sprintf("registry error: %s - %s", e.Code, e.Message)
}

// WriteErrorResponse is a helper function to write a standard error response.
func WriteErrorResponse(w http.ResponseWriter, statusCode int, errs ...RegistryError) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	response := RegistryErrorResponse{Errors: errs}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If encoding fails, fall back to a plain text error.
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// NewError is a generic constructor for a RegistryError.
func NewError(code ErrorCode, message string, detail interface{}) RegistryError {
	return RegistryError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

// Standard error constructors based on Docker Registry HTTP API V2 specification

// NewBlobUnknownError creates a 404 error for unknown blob
func NewBlobUnknownError(digest string) RegistryError {
	return RegistryError{
		Code:    ErrorCodeBlobUnknown,
		Message: "blob unknown to registry",
		Detail:  map[string]string{"digest": digest},
	}
}

// NewManifestUnknownError creates a 404 error for unknown manifest
func NewManifestUnknownError(name, reference string) RegistryError {
	return RegistryError{
		Code:    ErrorCodeManifestUnknown,
		Message: "manifest unknown",
		Detail:  map[string]string{"name": name, "tag": reference},
	}
}

// NewNameUnknownError creates a 404 error for unknown repository
func NewNameUnknownError(name string) RegistryError {
	return RegistryError{
		Code:    ErrorCodeNameUnknown,
		Message: "repository name not known to registry",
		Detail:  map[string]string{"name": name},
	}
}

// NewBlobUploadUnknownError creates a 404 error for unknown upload session
func NewBlobUploadUnknownError(uuid string) RegistryError {
	return RegistryError{
		Code:    ErrorCodeBlobUploadUnknown,
		Message: "blob upload unknown to registry",
		Detail:  map[string]string{"uuid": uuid},
	}
}

// NewNameInvalidError creates a 400 error for invalid name or reference
func NewNameInvalidError(name string) RegistryError {
	return RegistryError{
		Code:    ErrorCodeNameInvalid,
		Message: "invalid repository name",
		Detail:  map[string]string{"name": name},
	}
}

// NewManifestInvalidError creates a 400 error for invalid manifest
func NewManifestInvalidError(reason string) RegistryError {
	return RegistryError{
		Code:    ErrorCodeManifestInvalid,
		Message: "manifest invalid",
		Detail:  map[string]string{"reason": reason},
	}
}

// NewDigestInvalidError creates a 400 error for invalid digest
func NewDigestInvalidError(digest string) RegistryError {
	return RegistryError{
		Code:    ErrorCodeDigestInvalid,
		Message: "provided digest did not match uploaded content",
		Detail:  map[string]string{"digest": digest},
	}
}

// NewRangeInvalidError creates a 400 error for invalid range
func NewRangeInvalidError(rangeHeader string) RegistryError {
	return RegistryError{
		Code:    ErrorCodeRangeInvalid,
		Message: "invalid content range",
		Detail:  map[string]string{"range": rangeHeader},
	}
}

// NewBlobUploadInvalidError creates a 400 error for invalid upload parameters
func NewBlobUploadInvalidError(reason string) RegistryError {
	return RegistryError{
		Code:    ErrorCodeBlobUploadInvalid,
		Message: "blob upload invalid",
		Detail:  map[string]string{"reason": reason},
	}
}
