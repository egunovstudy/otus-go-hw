package main

import (
	"os"
	"testing"
)

func TestRunCmd(t *testing.T) {
	t.Run("returns same exit code as program", func(t *testing.T) {
		env := Environment{
			"GO_WANT_HELPER_PROCESS": {Value: "1", NeedRemove: false},
		}

		code := RunCmd(helperCmd(t, "exit", "7"), env)
		if code != 7 {
			t.Fatalf("expected 7, got %d", code)
		}
	})

	t.Run("applies env vars and removes vars with NeedRemove", func(t *testing.T) {
		t.Setenv("REMOVE_ME", "YES")

		env := Environment{
			"GO_WANT_HELPER_PROCESS": {Value: "1", NeedRemove: false},
			"FOO":                    {Value: "bar", NeedRemove: false},
			"EMPTYVAL":               {Value: "", NeedRemove: false},
			"REMOVE_ME":              {Value: "nope", NeedRemove: true},
		}

		code := RunCmd(helperCmd(t, "checkenv"), env)
		if code != 0 {
			t.Fatalf("expected 0, got %d", code)
		}
	})
}

func helperCmd(t *testing.T, action string, args ...string) []string {
	t.Helper()

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}

	cmd := []string{
		exe,
		"-test.run=TestHelperProcess",
		"--",
		action,
	}
	cmd = append(cmd, args...)
	return cmd
}

func TestHelperProcess(t *testing.T) {
	// IMPORTANT: when IDE runs tests including TestHelperProcess, it must not do anything.
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		t.Skip("helper process")
	}

	sep := -1
	for i, a := range os.Args {
		if a == "--" {
			sep = i
			break
		}
	}
	if sep == -1 || sep+1 >= len(os.Args) {
		os.Exit(2)
	}

	action := os.Args[sep+1]
	rest := os.Args[sep+2:]

	switch action {
	case "exit":
		if len(rest) != 1 {
			os.Exit(2)
		}
		if rest[0] == "7" {
			os.Exit(7)
		}
		os.Exit(1)

	case "checkenv":
		if os.Getenv("FOO") != "bar" {
			os.Exit(10)
		}
		if os.Getenv("EMPTYVAL") != "" {
			os.Exit(11)
		}
		if _, ok := os.LookupEnv("REMOVE_ME"); ok {
			os.Exit(12)
		}
		os.Exit(0)

	default:
		os.Exit(2)
	}
}
