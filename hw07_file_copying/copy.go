package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

// Copy copies data from fromPath to toPath starting at offset and copying up to limit bytes.
// If limit == 0, it copies until EOF.
// If limit exceeds remaining bytes, it copies until EOF.
// It prints copy progress to stdout in percents.
func Copy(fromPath, toPath string, offset, limit int64) error {
	if offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	if limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}

	src, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	size := info.Size()
	if offset > size {
		return ErrOffsetExceedsFileSize
	}

	if _, err := src.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	remaining := size - offset
	toCopy := remaining
	if limit > 0 && limit < toCopy {
		toCopy = limit
	}

	dst, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	if toCopy == 0 {
		return nil
	}

	const bufSize = 32 * 1024
	buf := make([]byte, bufSize)

	var copied int64
	lastPercent := int64(-1)

	printPercent := func(p int64) {
		fmt.Printf("\r%d%%", p)
	}

	for copied < toCopy {
		want := int64(len(buf))
		left := toCopy - copied
		if left < want {
			want = left
		}

		n, rerr := src.Read(buf[:want])
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
			copied += int64(n)

			percent := copied * 100 / toCopy
			if percent != lastPercent {
				lastPercent = percent
				printPercent(percent)
			}
		}

		if rerr != nil {
			if errors.Is(rerr, io.EOF) {
				break
			}
			return rerr
		}
	}

	if lastPercent != 100 {
		printPercent(100)
	}
	fmt.Print("\n")
	return nil
}
