package download

import "time"

type GenerateRequest struct {
	DSL           string
	Force         bool
	PackageName   string
	IncludeReadme bool
	IncludeReport bool
	IncludeDSL    bool
}

type ArtifactInfo struct {
	TaskID      string    `json:"task_id"`
	DownloadURL string    `json:"download_url"`
	Filename    string    `json:"filename"`
	SizeBytes   int64     `json:"size_bytes"`
	FileCount   int       `json:"file_count"`
	ExpiresAt   time.Time `json:"expire_at"`
}

type ResolvedArtifact struct {
	Filename    string
	PackagePath string
}
