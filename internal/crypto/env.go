package cryptutils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sort"
	"strings"
)

func ParseEnv(data []byte) (map[string]string, error) {
	envs := make(map[string]string)

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid env line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		envs[key] = val
	}

	return envs, nil
}

func NormalizeEnv(env map[string]string) []byte {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(env[k])
		b.WriteString("\n")
	}

	return []byte(b.String())
}

func CompressEnv(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DecompressEnv(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	return io.ReadAll(gr)
}

func PrepareEnvForStorage(raw []byte) ([]byte, error) {
	parsed, err := ParseEnv(raw)
	if err != nil {
		return nil, err
	}

	normalized := NormalizeEnv(parsed)

	compressed, err := CompressEnv(normalized)
	if err != nil {
		return nil, err
	}

	return compressed, nil
}

func ReadEnvFromStorage(data []byte) (map[string]string, error) {
	decompressed, err := DecompressEnv(data)
	if err != nil {
		return nil, err
	}

	return ParseEnv(decompressed)
}

type DiffingResult struct {
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
	Modified []string `json:"modified"`
}

func DiffEnvVersions(oldVersion, newVersion map[string]string) DiffingResult {

	var Added, Removed, Modified []string

	for key, val := range newVersion {
		if _, exists := oldVersion[key]; !exists {
			Added = append(Added, key)
		} else {
			if val != oldVersion[key] {
				Modified = append(Modified, key)
			}
		}
	}
	for key := range oldVersion {
		if _, exists := newVersion[key]; !exists {
			Removed = append(Removed, key)
		}
	}

	return DiffingResult{Added: Added, Removed: Removed, Modified: Modified}
}
