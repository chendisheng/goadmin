package model

import "testing"

func TestFileAssetTableNameAndClone(t *testing.T) {
	t.Parallel()

	asset := FileAsset{
		Id:            "1",
		OriginalName:  "demo.png",
		StorageDriver: "local",
		Visibility:    FileVisibilityPrivate,
		Status:        FileStatusActive,
	}

	if got := asset.TableName(); got != "upload_file" {
		t.Fatalf("TableName() = %q, want %q", got, "upload_file")
	}

	clone := asset.Clone()
	if clone.Id != asset.Id || clone.OriginalName != asset.OriginalName || clone.StorageDriver != asset.StorageDriver {
		t.Fatalf("Clone() = %+v, want %+v", clone, asset)
	}
}

func TestFileVisibilityAndStatusConstants(t *testing.T) {
	t.Parallel()

	if FileVisibilityPrivate != "private" {
		t.Fatalf("FileVisibilityPrivate = %q", FileVisibilityPrivate)
	}
	if FileVisibilityPublic != "public" {
		t.Fatalf("FileVisibilityPublic = %q", FileVisibilityPublic)
	}
	if FileStatusActive != "active" || FileStatusArchived != "archived" || FileStatusDeleted != "deleted" {
		t.Fatalf("unexpected file status constants: %q %q %q", FileStatusActive, FileStatusArchived, FileStatusDeleted)
	}
}
