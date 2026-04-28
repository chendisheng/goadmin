package service

import (
	"bytes"
	"context"
	"io"
	"testing"

	"goadmin/core/config"
	apperrors "goadmin/core/errors"
	"goadmin/modules/upload/domain/model"
	uploadrepo "goadmin/modules/upload/domain/repository"
	"goadmin/modules/upload/infrastructure/storage/local"
)

type memoryRepo struct {
	items                map[string]*model.FileAsset
	defaultStorageDriver string
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{items: map[string]*model.FileAsset{}}
}

func (r *memoryRepo) List(ctx context.Context, filter uploadrepo.ListFilter) ([]model.FileAsset, int64, error) {
	items := make([]model.FileAsset, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item.Clone())
	}
	return items, int64(len(items)), nil
}

func (r *memoryRepo) Get(ctx context.Context, id string) (*model.FileAsset, error) {
	if item, ok := r.items[id]; ok {
		clone := item.Clone()
		return &clone, nil
	}
	return nil, uploadrepo.ErrNotFound
}

func (r *memoryRepo) Create(ctx context.Context, item *model.FileAsset) (*model.FileAsset, error) {
	clone := item.Clone()
	if clone.Id == "" {
		clone.Id = "asset-1"
	}
	r.items[clone.Id] = &clone
	return &clone, nil
}

func (r *memoryRepo) Update(ctx context.Context, item *model.FileAsset) (*model.FileAsset, error) {
	clone := item.Clone()
	r.items[clone.Id] = &clone
	return &clone, nil
}

func (r *memoryRepo) Delete(ctx context.Context, id string) error {
	delete(r.items, id)
	return nil
}

func (r *memoryRepo) Bind(ctx context.Context, id string, binding model.FileBinding) (*model.FileAsset, error) {
	item, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	item.BizModule = binding.BizModule
	item.BizType = binding.BizType
	item.BizId = binding.BizId
	item.BizField = binding.BizField
	return r.Update(ctx, item)
}

func (r *memoryRepo) Unbind(ctx context.Context, id string) (*model.FileAsset, error) {
	item, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	item.BizModule = ""
	item.BizType = ""
	item.BizId = ""
	item.BizField = ""
	return r.Update(ctx, item)
}

func (r *memoryRepo) DefaultStorageDriver(ctx context.Context, fallback string) (string, error) {
	if r.defaultStorageDriver == "" {
		return fallback, nil
	}
	return r.defaultStorageDriver, nil
}

func (r *memoryRepo) SetDefaultStorageDriver(ctx context.Context, driver string) error {
	r.defaultStorageDriver = driver
	return nil
}

func TestUploadAndDeleteFlow(t *testing.T) {
	t.Parallel()

	driver, err := local.NewDriver(config.LocalStorageConfig{BaseDir: t.TempDir(), PublicBaseURL: "/uploads/files", UseProxyDownload: true})
	if err != nil {
		t.Fatalf("new local driver: %v", err)
	}
	repo := newMemoryRepo()
	svc, err := New(repo, driver, config.StoragePolicyConfig{
		MaxUploadSize:     "1mb",
		AllowedExtensions: []string{".txt"},
		AllowedMIMETypes:  []string{"text/plain"},
		VisibilityDefault: "private",
		PathPrefix:        "uploads",
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	created, err := svc.Upload(context.Background(), UploadRequest{
		File:        bytes.NewBufferString("hello goadmin"),
		Filename:    "note.txt",
		ContentType: "text/plain",
		Size:        int64(len("hello goadmin")),
		UploadedBy:  "admin",
	})
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if created.Id == "" {
		t.Fatal("expected uploaded asset id")
	}
	if created.StorageKey == "" {
		t.Fatalf("unexpected created asset: %+v", created)
	}
	if created.PublicURL != "" {
		t.Fatalf("private file should not expose public_url: %+v", created)
	}
	if _, _, err := driver.Get(context.Background(), created.StorageKey); err != nil {
		t.Fatalf("stored file missing: %v", err)
	}

	reader, info, opened, err := svc.Open(context.Background(), created.Id)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		_ = reader.Close()
		t.Fatalf("read opened file: %v", err)
	}
	_ = reader.Close()
	if string(data) != "hello goadmin" {
		t.Fatalf("unexpected opened file content: %q", string(data))
	}
	if info == nil || info.ContentType == "" {
		t.Fatalf("expected object info content type, got %+v", info)
	}
	if opened == nil || opened.StorageKey != created.StorageKey {
		t.Fatalf("unexpected opened asset: %+v", opened)
	}

	if err := svc.Delete(context.Background(), created.Id); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if exists, err := driver.Exists(context.Background(), created.StorageKey); err != nil || exists {
		t.Fatalf("file should be removed, exists=%v err=%v", exists, err)
	}
}

func TestUploadPublicFileExposesPublicURL(t *testing.T) {
	t.Parallel()

	driver, err := local.NewDriver(config.LocalStorageConfig{BaseDir: t.TempDir(), PublicBaseURL: "/uploads/files", UseProxyDownload: true})
	if err != nil {
		t.Fatalf("new local driver: %v", err)
	}
	svc, err := New(newMemoryRepo(), driver, config.StoragePolicyConfig{
		AllowedExtensions: []string{".txt"},
		AllowedMIMETypes:  []string{"text/plain"},
		VisibilityDefault: "private",
		PathPrefix:        "uploads",
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	created, err := svc.Upload(context.Background(), UploadRequest{
		File:        bytes.NewBufferString("hello public"),
		Filename:    "note.txt",
		ContentType: "text/plain",
		Size:        int64(len("hello public")),
		UploadedBy:  "admin",
		Visibility:  string(model.FileVisibilityPublic),
	})
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if created.PublicURL == "" {
		t.Fatalf("public file should expose public_url: %+v", created)
	}
	if created.Visibility != model.FileVisibilityPublic {
		t.Fatalf("unexpected visibility: %+v", created)
	}
}

func TestUploadRejectsDisallowedExtension(t *testing.T) {
	t.Parallel()

	driver, err := local.NewDriver(config.LocalStorageConfig{BaseDir: t.TempDir(), PublicBaseURL: "/uploads/files", UseProxyDownload: true})
	if err != nil {
		t.Fatalf("new local driver: %v", err)
	}
	svc, err := New(newMemoryRepo(), driver, config.StoragePolicyConfig{
		AllowedExtensions: []string{".txt"},
		AllowedMIMETypes:  []string{"text/plain"},
		VisibilityDefault: "private",
		PathPrefix:        "uploads",
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = svc.Upload(context.Background(), UploadRequest{
		File:        bytes.NewBufferString("hello"),
		Filename:    "bad.png",
		ContentType: "image/png",
		Size:        5,
	})
	if err == nil {
		t.Fatal("expected upload rejection for extension")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Key != "upload.extension_not_allowed" {
		t.Fatalf("unexpected error key: %q", appErr.Key)
	}
}
