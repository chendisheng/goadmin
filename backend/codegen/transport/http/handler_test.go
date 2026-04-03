package http

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	downloadapp "goadmin/codegen/application/download"
	"goadmin/core/response"
)

type fakeContext struct {
	requestContext context.Context
	params         map[string]string
	payload        any
	status         int
	jsonBody       any
	attachmentPath string
	attachmentName string
	headers        map[string]string
	values         map[string]any
}

func (c *fakeContext) RequestContext() context.Context {
	if c.requestContext == nil {
		return context.Background()
	}
	return c.requestContext
}

func (c *fakeContext) SetRequestContext(ctx context.Context) { c.requestContext = ctx }
func (c *fakeContext) Method() string                        { return "POST" }
func (c *fakeContext) Path() string                          { return "/api/v1/codegen/dsl/generate-download" }
func (c *fakeContext) Header(string) string                  { return "" }
func (c *fakeContext) SetHeader(key, value string) {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers[key] = value
}
func (c *fakeContext) Param(key string) string {
	if c.params == nil {
		return ""
	}
	return c.params[key]
}
func (c *fakeContext) Query(string) string { return "" }
func (c *fakeContext) Set(key string, value any) {
	if c.values == nil {
		c.values = make(map[string]any)
	}
	c.values[key] = value
}
func (c *fakeContext) Get(key string) (any, bool) {
	if c.values == nil {
		return nil, false
	}
	value, ok := c.values[key]
	return value, ok
}
func (c *fakeContext) ShouldBindJSON(v any) error {
	data, err := json.Marshal(c.payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
func (c *fakeContext) ShouldBindQuery(any) error { return nil }
func (c *fakeContext) BindJSON(v any) error      { return c.ShouldBindJSON(v) }
func (c *fakeContext) JSON(status int, payload any) {
	c.status = status
	c.jsonBody = payload
}
func (c *fakeContext) FileAttachment(path, name string) {
	c.attachmentPath = path
	c.attachmentName = name
}
func (c *fakeContext) AbortWithStatusJSON(status int, payload any) {
	c.status = status
	c.jsonBody = payload
}

func TestHandlerGenerateDownloadAndArtifact(t *testing.T) {
	t.Parallel()

	handler := NewHandler(Dependencies{
		ProjectRoot:     t.TempDir(),
		ArtifactEnabled: true,
		ArtifactBaseDir: t.TempDir(),
		ArtifactTTL:     time.Hour,
	})
	generateCtx := &fakeContext{
		payload: GenerateDownloadRequest{
			DSL: strings.TrimSpace(`
module: inventory
kind: business-module
framework:
  backend: gin
  frontend: vue3
entity:
  name: item
  fields:
    - name: id
      type: string
      primary: true
    - name: name
      type: string
      required: true
pages:
  - list
permissions:
  - inventory:view
`),
			PackageName: "inventory-module",
		},
	}

	handler.GenerateDownload(generateCtx)
	if generateCtx.status != 200 {
		t.Fatalf("GenerateDownload status = %d, want 200, body=%#v", generateCtx.status, generateCtx.jsonBody)
	}
	envelope, ok := generateCtx.jsonBody.(response.Envelope)
	if !ok {
		t.Fatalf("GenerateDownload body type = %T, want response.Envelope", generateCtx.jsonBody)
	}
	artifact, ok := envelope.Data.(downloadapp.ArtifactInfo)
	if !ok {
		t.Fatalf("GenerateDownload data type = %T, want download.ArtifactInfo", envelope.Data)
	}
	if artifact.TaskID == "" {
		t.Fatal("expected task id")
	}
	if artifact.DownloadURL == "" {
		t.Fatal("expected download url")
	}

	downloadCtx := &fakeContext{params: map[string]string{"taskID": artifact.TaskID}}
	handler.DownloadArtifact(downloadCtx)
	if downloadCtx.attachmentPath == "" {
		t.Fatal("expected attachment path")
	}
	if downloadCtx.attachmentName == "" {
		t.Fatal("expected attachment name")
	}
	if got := downloadCtx.headers["Cache-Control"]; got != "private, max-age=300" {
		t.Fatalf("Cache-Control = %q, want private, max-age=300", got)
	}
}
