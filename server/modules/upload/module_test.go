package upload

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	coretransport "goadmin/core/transport"
	uploadhttp "goadmin/modules/upload/transport/http"
)

type routeCall struct {
	method string
	path   string
}

type recordingRegistrar struct {
	prefix string
	calls  *[]routeCall
}

func newRecordingRegistrar() *recordingRegistrar {
	calls := make([]routeCall, 0, 16)
	return &recordingRegistrar{calls: &calls}
}

func (r *recordingRegistrar) Group(path string, _ ...coretransport.Middleware) coretransport.RouteRegistrar {
	return &recordingRegistrar{prefix: joinRoute(r.prefix, path), calls: r.calls}
}

func (r *recordingRegistrar) GET(path string, _ coretransport.HandlerFunc, _ ...coretransport.Middleware) {
	r.record("GET", path)
}

func (r *recordingRegistrar) POST(path string, _ coretransport.HandlerFunc, _ ...coretransport.Middleware) {
	r.record("POST", path)
}

func (r *recordingRegistrar) PUT(path string, _ coretransport.HandlerFunc, _ ...coretransport.Middleware) {
	r.record("PUT", path)
}

func (r *recordingRegistrar) PATCH(path string, _ coretransport.HandlerFunc, _ ...coretransport.Middleware) {
	r.record("PATCH", path)
}

func (r *recordingRegistrar) DELETE(path string, _ coretransport.HandlerFunc, _ ...coretransport.Middleware) {
	r.record("DELETE", path)
}

func (r *recordingRegistrar) Any(path string, _ coretransport.HandlerFunc, _ ...coretransport.Middleware) {
	r.record("ANY", path)
}

func (r *recordingRegistrar) record(method, path string) {
	if r == nil || r.calls == nil {
		return
	}
	*r.calls = append(*r.calls, routeCall{method: method, path: joinRoute(r.prefix, path)})
}

func joinRoute(prefix, suffix string) string {
	if prefix == "" {
		if suffix == "" {
			return "/"
		}
		if strings.HasPrefix(suffix, "/") {
			return suffix
		}
		return "/" + suffix
	}
	if suffix == "" {
		return prefix
	}
	return strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(suffix, "/")
}

func TestUploadModuleMetadataAndManifest(t *testing.T) {
	t.Parallel()

	mod := NewModule()
	if mod.Name != Name {
		t.Fatalf("NewModule().Name = %q, want %q", mod.Name, Name)
	}
	if mod.ManifestPath != ManifestPath {
		t.Fatalf("NewModule().ManifestPath = %q, want %q", mod.ManifestPath, ManifestPath)
	}

	manifestPath := filepath.Clean("manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest %s: %v", manifestPath, err)
	}
	manifest := string(data)
	for _, marker := range []string{
		"name: upload",
		"kind: core-module",
		"path: /api/v1/uploads/files",
		"path: /api/v1/uploads/files/:id/download",
		"path: /api/v1/uploads/files/:id/preview",
		"path: /system/upload",
		"permission: upload:file:list",
		"object: upload-file",
		"capabilities:",
	} {
		if !strings.Contains(manifest, marker) {
			t.Fatalf("manifest missing marker %q", marker)
		}
	}
}

func TestUploadRouterRegistersExpectedRoutes(t *testing.T) {
	t.Parallel()

	recorder := newRecordingRegistrar()
	uploadhttp.Register(recorder, uploadhttp.Dependencies{})

	want := map[routeCall]struct{}{
		{method: "GET", path: "/uploads/files/storage/default"}: {},
		{method: "PUT", path: "/uploads/files/storage/default"}: {},
		{method: "GET", path: "/uploads/files"}:                 {},
		{method: "GET", path: "/uploads/files/:id"}:             {},
		{method: "POST", path: "/uploads/files"}:                {},
		{method: "DELETE", path: "/uploads/files/:id"}:          {},
		{method: "GET", path: "/uploads/files/:id/download"}:    {},
		{method: "GET", path: "/uploads/files/:id/preview"}:     {},
		{method: "POST", path: "/uploads/files/:id/bind"}:       {},
		{method: "DELETE", path: "/uploads/files/:id/bind"}:     {},
	}
	got := make(map[routeCall]struct{}, len(*recorder.calls))
	for _, call := range *recorder.calls {
		got[call] = struct{}{}
	}
	if len(got) != len(want) {
		t.Fatalf("registered route count = %d, want %d; routes=%v", len(got), len(want), *recorder.calls)
	}
	for call := range want {
		if _, ok := got[call]; !ok {
			t.Fatalf("missing registered route %+v; got=%v", call, *recorder.calls)
		}
	}
}
