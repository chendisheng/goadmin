package command

type CreateCodegenConsole struct {
	Name    string `json:"name,omitempty"`
	Enabled int64  `json:"enabled,omitempty"`
}

type UpdateCodegenConsole struct {
	Name    string `json:"name,omitempty"`
	Enabled int64  `json:"enabled,omitempty"`
}
