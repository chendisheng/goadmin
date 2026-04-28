package objectstore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	apperrors "goadmin/core/errors"
)

func normalizeKey(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_key_required", "upload storage key is required")
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = pathClean(trimmed)
	if trimmed == "." || trimmed == "" {
		return "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_key_invalid", "upload storage key is invalid")
	}
	if strings.HasPrefix(trimmed, "../") || trimmed == ".." || strings.Contains(trimmed, "/../") {
		return "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_key_traversal", "upload storage key contains path traversal")
	}
	return strings.TrimPrefix(trimmed, "/"), nil
}

func pathClean(value string) string {
	cleaned := filepath.Clean(strings.ReplaceAll(value, "/", string(os.PathSeparator)))
	return strings.ReplaceAll(cleaned, string(os.PathSeparator), "/")
}

func sanitizeSegment(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "unknown"
	}
	return result
}

func stableConfigKey(cfg storageConfig) string {
	h := sha256.New()
	_, _ = io.WriteString(h, strings.Join([]string{
		strings.TrimSpace(cfg.Endpoint),
		strings.TrimSpace(cfg.Region),
		strings.TrimSpace(cfg.Bucket),
		strings.TrimSpace(cfg.AccessKeyID),
		strings.TrimSpace(cfg.AccessKeySecret),
		fmt.Sprintf("%t", cfg.UseSSL),
		fmt.Sprintf("%t", cfg.PathStyle),
		strings.TrimSpace(cfg.PublicBaseURL),
	}, "|"))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func defaultPublicBaseURL(name string, cfg storageConfig) string {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	bucket := strings.TrimSpace(cfg.Bucket)
	switch name {
	case "cos":
		if endpoint != "" && bucket != "" {
			return joinURL(endpoint, bucket)
		}
	case "oss", "s3-compatible", "minio":
		if endpoint != "" && bucket != "" {
			return joinURL(endpoint, bucket)
		}
	}
	return joinURL("/api/v1/uploads/files", bucket)
}

func validateConfig(name string, cfg storageConfig) error {
	needEndpoint := name == "s3-compatible" || name == "oss" || name == "minio"
	needRegion := name == "cos"
	needAccess := name != "cos"
	if needEndpoint && strings.TrimSpace(cfg.Endpoint) == "" {
		return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_endpoint_required", fmt.Sprintf("upload storage %s endpoint is required", name))
	}
	if needRegion && strings.TrimSpace(cfg.Region) == "" {
		return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_cos_region_required", "upload storage cos region is required")
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_bucket_required", fmt.Sprintf("upload storage %s bucket is required", name))
	}
	if needAccess && strings.TrimSpace(cfg.AccessKeyID) == "" {
		return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_access_key_id_required", fmt.Sprintf("upload storage %s access_key_id is required", name))
	}
	if needAccess && strings.TrimSpace(cfg.AccessKeySecret) == "" {
		return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_access_key_secret_required", fmt.Sprintf("upload storage %s access_key_secret is required", name))
	}
	if name == "cos" {
		if strings.TrimSpace(cfg.AccessKeyID) == "" {
			return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_cos_secret_id_required", "upload storage cos secret_id is required")
		}
		if strings.TrimSpace(cfg.AccessKeySecret) == "" {
			return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_cos_secret_key_required", "upload storage cos secret_key is required")
		}
	}
	return nil
}

func joinURL(base, suffix string) string {
	base = strings.TrimSpace(base)
	suffix = strings.TrimSpace(suffix)
	if base == "" {
		return strings.TrimPrefix(suffix, "/")
	}
	if suffix == "" {
		return strings.TrimRight(base, "/")
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(suffix, "/")
}
