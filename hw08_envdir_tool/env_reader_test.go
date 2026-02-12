package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	t.Run("reads first lines and applies transformations", func(t *testing.T) {
		dir := t.TempDir()

		// First line only
		writeFile(t, filepath.Join(dir, "FOO"), "123\n456\n789")

		// Trim trailing spaces & tabs
		writeFile(t, filepath.Join(dir, "BAR"), "value\t \nnext")

		// Replace 0x00 with '\n'
		writeFile(t, filepath.Join(dir, "ZEROS"), "a\x00b\x00c\nignored")

		// Empty file => NeedRemove=true
		writeFileRaw(t, filepath.Join(dir, "EMPTY"), []byte{})

		// First line empty but file not empty => NeedRemove=false, Value=""
		writeFile(t, filepath.Join(dir, "FIRST_EMPTY_LINE"), "\nsecond")

		env, err := ReadDir(dir)
		if err != nil {
			t.Fatalf("ReadDir error: %v", err)
		}

		assertEnv(t, env, "FOO", "123", false)
		assertEnv(t, env, "BAR", "value", false)
		assertEnv(t, env, "ZEROS", "a\nb\nc", false)
		assertEnv(t, env, "EMPTY", "", true)
		assertEnv(t, env, "FIRST_EMPTY_LINE", "", false)
	})

	t.Run("ignores subdirectories", func(t *testing.T) {
		dir := t.TempDir()

		if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}

		env, err := ReadDir(dir)
		if err != nil {
			t.Fatalf("ReadDir error: %v", err)
		}

		if _, ok := env["subdir"]; ok {
			t.Fatalf("subdir must be ignored, but present in env")
		}
	})

	t.Run("invalid env var name with '=' returns ErrInvalidEnvName", func(t *testing.T) {
		dir := t.TempDir()

		writeFile(t, filepath.Join(dir, "BAD=NAME"), "123")

		_, err := ReadDir(dir)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidEnvName) {
			t.Fatalf("expected ErrInvalidEnvName, got: %v", err)
		}
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func writeFileRaw(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func assertEnv(t *testing.T, env Environment, key, wantValue string, wantRemove bool) {
	t.Helper()

	got, ok := env[key]
	if !ok {
		t.Fatalf("expected key %q to exist", key)
	}
	if got.Value != wantValue || got.NeedRemove != wantRemove {
		t.Fatalf("%s: got %+v, want Value=%q NeedRemove=%v", key, got, wantValue, wantRemove)
	}
}
