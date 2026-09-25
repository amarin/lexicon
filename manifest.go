package lexicon

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

// Manifest is the provenance of a dictionary, read from a sidecar file
// "<file>.meta" of "key: value" lines (source, license, version, url,
// generated_from; other keys go to Extra). For .dat files without a sidecar it
// is filled from gomorphy BuildInfo; the built-in base takes it from
// Options.BaseManifest merged over BuildInfo. Hosts use it for attribution
// (OpenCorpora data is CC BY-SA).
type Manifest struct {
	Source        string
	License       string
	Version       string
	URL           string
	GeneratedFrom string
	Extra         map[string]string
}

// ManifestPath is the sidecar path of a dictionary file.
func ManifestPath(dictPath string) string { return dictPath + ".meta" }

// ParseManifest reads "key: value" lines; blank lines, '#' comments and lines
// without ':' are skipped; keys are case-insensitive.
func ParseManifest(r io.Reader) (Manifest, error) {
	var m Manifest

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		key, value = strings.ToLower(strings.TrimSpace(key)), strings.TrimSpace(value)

		switch key {
		case "source":
			m.Source = value
		case "license":
			m.License = value
		case "version":
			m.Version = value
		case "url":
			m.URL = value
		case "generated_from":
			m.GeneratedFrom = value
		default:
			if m.Extra == nil {
				m.Extra = map[string]string{}
			}

			m.Extra[key] = value
		}
	}

	if err := sc.Err(); err != nil {
		return Manifest{}, fmt.Errorf("lexicon: manifest: %w", err)
	}

	return m, nil
}

// WriteTo writes the manifest in sidecar format: known keys in fixed order,
// then Extra sorted by key; empty values are omitted.
func (m Manifest) WriteTo(w io.Writer) (int64, error) {
	var b strings.Builder

	for _, kv := range [][2]string{
		{"source", m.Source}, {"license", m.License}, {"version", m.Version}, {"url", m.URL}, {"generated_from", m.GeneratedFrom},
	} {
		if kv[1] != "" {
			fmt.Fprintf(&b, "%s: %s\n", kv[0], kv[1])
		}
	}

	for _, k := range slices.Sorted(maps.Keys(m.Extra)) {
		fmt.Fprintf(&b, "%s: %s\n", k, m.Extra[k])
	}

	n, err := io.WriteString(w, b.String())

	return int64(n), err
}

// String is a one-line attribution: "source; version; url; license L".
func (m Manifest) String() string {
	var parts []string

	for _, s := range []string{m.Source, m.Version, m.URL} {
		if s != "" {
			parts = append(parts, s)
		}
	}

	if m.License != "" {
		parts = append(parts, "license "+m.License)
	}

	return strings.Join(parts, "; ")
}

// IsZero reports whether the manifest carries no information.
func (m Manifest) IsZero() bool {
	return m.Source == "" && m.License == "" && m.Version == "" && m.URL == "" && m.GeneratedFrom == "" && len(m.Extra) == 0
}

// mergeManifest returns primary with its empty fields and missing Extra keys
// filled from fallback; primary is not modified (Options.BaseManifest over
// gomorphy BuildInfo, owner decision 2026-09-25).
func mergeManifest(primary, fallback Manifest) Manifest {
	m := primary
	m.Source = cmp.Or(m.Source, fallback.Source)
	m.License = cmp.Or(m.License, fallback.License)
	m.Version = cmp.Or(m.Version, fallback.Version)
	m.URL = cmp.Or(m.URL, fallback.URL)
	m.GeneratedFrom = cmp.Or(m.GeneratedFrom, fallback.GeneratedFrom)

	if len(fallback.Extra) > 0 {
		m.Extra = maps.Clone(fallback.Extra)
		maps.Copy(m.Extra, primary.Extra)
	}

	return m
}
