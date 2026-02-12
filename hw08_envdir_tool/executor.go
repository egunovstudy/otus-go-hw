package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	// #nosec G204 -- This is a CLI tool: executing user-provided command is expected behavior.
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	c.Env = mergeEnv(os.Environ(), env)

	if err := c.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
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
