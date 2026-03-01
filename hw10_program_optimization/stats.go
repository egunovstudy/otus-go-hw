package hw10programoptimization

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return DomainStat{}, nil
	}
	suffix := "." + domain

	res := make(DomainStat)

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for sc.Scan() {
		line := sc.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		email, ok := extractEmail(line)
		if !ok {
			return nil, fmt.Errorf("invalid json line")
		}

		at := strings.LastIndexByte(email, '@')
		if at < 0 || at+1 >= len(email) {
			continue
		}

		domainPart := email[at+1:]
		if !hasSuffixFold(domainPart, suffix) {
			continue
		}

		res[toLowerASCII(domainPart)]++
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

var emailKey = []byte(`"Email":"`)

func extractEmail(line []byte) (string, bool) {
	i := bytes.Index(line, emailKey)
	if i < 0 {
		return "", false
	}
	start := i + len(emailKey)
	end := start
	for end < len(line) {
		switch line[end] {
		case '\\':
			end += 2
			continue
		case '"':
			return string(line[start:end]), true
		default:
			end++
		}
	}
	return "", false
}

func hasSuffixFold(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	s = s[len(s)-len(suffix):]
	for i := 0; i < len(suffix); i++ {
		cb := s[i]
		sb := suffix[i]
		if cb == sb {
			continue
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if cb != sb {
			return false
		}
	}
	return true
}

func toLowerASCII(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
