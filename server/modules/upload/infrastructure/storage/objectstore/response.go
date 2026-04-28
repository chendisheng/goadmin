package objectstore

import (
	"encoding/json"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
)

type storedObjectMeta struct {
	Key                string            `json:"key"`
	StorageName        string            `json:"storage_name"`
	Size               int64             `json:"size"`
	ContentType        string            `json:"content_type"`
	ETag               string            `json:"etag"`
	ChecksumSHA256     string            `json:"checksum_sha256"`
	Metadata           map[string]string `json:"metadata"`
	Visibility         string            `json:"visibility"`
	Filename           string            `json:"filename"`
	ContentDisposition string            `json:"content_disposition"`
	PublicURL          string            `json:"public_url"`
	StoredAt           time.Time         `json:"stored_at"`
}

func objectInfoFromMeta(key string, stat os.FileInfo, meta *storedObjectMeta, publicURL string) *storagecontract.ObjectInfo {
	info := &storagecontract.ObjectInfo{
		Key:       key,
		Size:      stat.Size(),
		ModTime:   stat.ModTime(),
		Metadata:  map[string]string{},
		PublicURL: publicURL,
	}
	if meta != nil {
		if stringsTrimSpace(meta.ContentType) != "" {
			info.ContentType = meta.ContentType
		}
		if stringsTrimSpace(meta.ETag) != "" {
			info.ETag = meta.ETag
		}
		if stringsTrimSpace(meta.PublicURL) != "" {
			info.PublicURL = meta.PublicURL
		}
		info.Metadata = cloneStringMap(meta.Metadata)
	}
	if info.ContentType == "" {
		info.ContentType = mime.TypeByExtension(filepath.Ext(key))
	}
	if info.ContentType == "" {
		info.ContentType = "application/octet-stream"
	}
	if info.Metadata == nil {
		info.Metadata = map[string]string{}
	}
	return info
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".meta-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func readMeta(path string) (*storedObjectMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var meta storedObjectMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	if meta.Metadata == nil {
		meta.Metadata = map[string]string{}
	}
	return &meta, nil
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func stringsTrimSpace(value string) string {
	return strings.TrimSpace(value)
}
