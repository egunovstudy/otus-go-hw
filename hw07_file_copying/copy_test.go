package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy_WholeFile(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src.txt")
	dstPath := filepath.Join(dir, "dst.txt")

	want := []byte("hello world\nthis is a test\n")
	if err := os.WriteFile(srcPath, want, 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := Copy(srcPath, dstPath, 0, 0); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("unexpected content:\n got: %q\nwant: %q", string(got), string(want))
	}
}

func TestCopy_WithOffset(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src.bin")
	dstPath := filepath.Join(dir, "dst.bin")

	data := []byte("0123456789")
	if err := os.WriteFile(srcPath, data, 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := Copy(srcPath, dstPath, 3, 0); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	want := []byte("3456789")
	if string(got) != string(want) {
		t.Fatalf("unexpected content: got %q want %q", string(got), string(want))
	}
}

func TestCopy_WithLimit(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src.bin")
	dstPath := filepath.Join(dir, "dst.bin")

	data := []byte("0123456789")
	if err := os.WriteFile(srcPath, data, 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := Copy(srcPath, dstPath, 2, 4); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	want := []byte("2345")
	if string(got) != string(want) {
		t.Fatalf("unexpected content: got %q want %q", string(got), string(want))
	}
}

func TestCopy_LimitExceedsFileSize(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src.bin")
	dstPath := filepath.Join(dir, "dst.bin")

	data := []byte("abc")
	if err := os.WriteFile(srcPath, data, 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := Copy(srcPath, dstPath, 0, 1000); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("unexpected content: got %q want %q", string(got), string(data))
	}
}

func TestCopy_OffsetEqualsFileSize(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src.bin")
	dstPath := filepath.Join(dir, "dst.bin")

	data := []byte("abc")
	if err := os.WriteFile(srcPath, data, 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := Copy(srcPath, dstPath, int64(len(data)), 0); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty dst, got %q", string(got))
	}
}

func TestCopy_OffsetExceedsFileSize(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src.bin")
	dstPath := filepath.Join(dir, "dst.bin")

	data := []byte("abc")
	if err := os.WriteFile(srcPath, data, 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	err := Copy(srcPath, dstPath, int64(len(data)+1), 0)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrOffsetExceedsFileSize) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrOffsetExceedsFileSize)
	}
}

func TestCopy_UnsupportedFile(t *testing.T) {
	dir := t.TempDir()

	dstPath := filepath.Join(dir, "dst.bin")

	err := Copy(dir, dstPath, 0, 0)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUnsupportedFile) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrUnsupportedFile)
	}
}
