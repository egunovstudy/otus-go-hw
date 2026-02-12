package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

var ErrInvalidEnvName = errors.New("invalid env name")

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	env := make(Environment, len(entries))

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if strings.Contains(name, "=") {
			return nil, fmt.Errorf("%w: %s", ErrInvalidEnvName, name)
		}

		fullPath := filepath.Join(dir, name)

		info, err := e.Info()
		if err != nil {
			return nil, err
		}

		// empty file => remove variable
		if info.Size() == 0 {
			env[name] = EnvValue{Value: "", NeedRemove: true}
			continue
		}

		val, err := readFirstLine(fullPath)
		if err != nil {
			return nil, err
		}

		// trim trailing spaces and tabs
		val = strings.TrimRight(val, " \t")

		// replace terminal zeros with '\n' (actually: replace 0x00 in value)
		val = string(bytes.ReplaceAll([]byte(val), []byte{0x00}, []byte{'\n'}))

		env[name] = EnvValue{Value: val, NeedRemove: false}
	}

	return env, nil
}

func readFirstLine(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	r := bufio.NewReader(f)
	line, err := r.ReadBytes('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	// Remove trailing '\n' and optional '\r' (in case of CRLF).
	line = bytes.TrimSuffix(line, []byte{'\n'})
	line = bytes.TrimSuffix(line, []byte{'\r'})

	return string(line), nil
}
