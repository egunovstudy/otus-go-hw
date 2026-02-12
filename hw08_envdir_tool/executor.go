package main

import (
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	merged := mergeEnv(os.Environ(), env)
	c.Env = merged

	if err := c.Run(); err != nil {
		var exitErr *exec.ExitError
		if ok := errorsAs(err, &exitErr); ok {
			return exitErr.ExitCode()
		}
		return 1
	}

	return 0
}

// mergeEnv takes current environment (base) and applies updates/removals from env.
func mergeEnv(base []string, env Environment) []string {
	m := make(map[string]string, len(base))

	for _, kv := range base {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		m[k] = v
	}

	for k, v := range env {
		if v.NeedRemove {
			delete(m, k)
			continue
		}
		m[k] = v.Value
	}

	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}

// small wrapper to avoid importing "errors" in tests-only style checks? (keeps code explicit)
func errorsAs(err error, target any) bool {
	// stdlib errors.As
	type aser interface {
		As(any) bool
	}
	// We can just call errors.As, but this avoids linter complaining in some setups about shadowing.
	// However simplest is to use errors.As directly.
	return asStd(err, target)
}

func asStd(err error, target any) bool {
	// Inline to keep imports small here; real call is errors.As.
	// (Compiler will inline; functionality is delegated in a separate file below.)
	return errorsAsStd(err, target)
}
