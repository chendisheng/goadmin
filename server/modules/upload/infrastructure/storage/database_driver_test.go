package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDatabaseDriverLifecycle(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(testDatabaseStorageSQLiteDSN(t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	driver := NewDatabaseDriver()
	if err := driver.SetDB(db); err != nil {
		t.Fatalf("set db: %v", err)
	}

	payload := []byte("hello database storage")
	result, err := driver.Put(context.Background(), storagecontract.PutObjectRequest{
		Key:         "uploads/demo.txt",
		Reader:      bytes.NewReader(payload),
		Size:        int64(len(payload)),
		ContentType: "text/plain",
		Filename:    "demo.txt",
		Metadata: map[string]string{
			"source":     "test",
			"public_url": "https://cdn.example.com/uploads/demo.txt",
		},
		Visibility: "public",
	})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if result.Key != "uploads/demo.txt" {
		t.Fatalf("Put key = %q, want %q", result.Key, "uploads/demo.txt")
	}
	if result.ChecksumSHA256 == "" || result.Size != int64(len(payload)) {
		t.Fatalf("unexpected Put result: %+v", result)
	}
	if result.URL != "https://cdn.example.com/uploads/demo.txt" {
		t.Fatalf("Put URL = %q, want %q", result.URL, "https://cdn.example.com/uploads/demo.txt")
	}

	exists, err := driver.Exists(context.Background(), result.Key)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Fatal("expected stored object to exist")
	}

	reader, info, err := driver.Get(context.Background(), result.Key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read data: %v", err)
	}
	if string(data) != string(payload) {
		t.Fatalf("Get payload = %q, want %q", string(data), string(payload))
	}
	if info == nil || info.ContentType != "text/plain" {
		t.Fatalf("unexpected object info: %+v", info)
	}
	if got := info.Metadata["source"]; got != "test" {
		t.Fatalf("metadata[source] = %q, want %q", got, "test")
	}
	if got := info.PublicURL; got != "https://cdn.example.com/uploads/demo.txt" {
		t.Fatalf("info.PublicURL = %q, want %q", got, "https://cdn.example.com/uploads/demo.txt")
	}

	publicURL, err := driver.PublicURL(context.Background(), result.Key)
	if err != nil {
		t.Fatalf("PublicURL: %v", err)
	}
	if publicURL != "https://cdn.example.com/uploads/demo.txt" {
		t.Fatalf("PublicURL = %q, want %q", publicURL, "https://cdn.example.com/uploads/demo.txt")
	}

	if err := driver.Delete(context.Background(), result.Key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	exists, err = driver.Exists(context.Background(), result.Key)
	if err != nil {
		t.Fatalf("Exists after delete: %v", err)
	}
	if exists {
		t.Fatal("expected stored object to be removed")
	}
}

func TestDatabaseDriverRejectsMissingDB(t *testing.T) {
	t.Parallel()

	driver := NewDatabaseDriver()
	if err := driver.SetDB(nil); err == nil {
		t.Fatal("expected SetDB(nil) to fail")
	}
}

func testDatabaseStorageSQLiteDSN(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "default"
	}
	name = strings.NewReplacer(" ", "_", "/", "_", "\\", "_").Replace(name)
	return fmt.Sprintf("file:storage-%s?mode=memory&cache=shared", name)
}
