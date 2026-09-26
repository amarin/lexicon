package gazetteer

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"os"
	"strings"
	"sync"
)

// TSVSource reads entries from UTF-8 TSV text — a file or in-memory data —
// with one alias per line:
//
//	type<TAB>ref<TAB>canonical<TAB>alias[<TAB>flags[<TAB>k=v;k=v]]
//
// Lines starting with '#' are comments; "# key: value" lines before the
// first entry form the provenance manifest (source, license, version, url,
// generated_from). Bad lines are skipped and returned as *ParseErrors after
// all good entries were yielded. Version is the sha256 of the content.
type TSVSource struct {
	name, path string
	data       []byte // in-memory content when inline
	inline     bool
	mu         sync.Mutex
	manifest   map[string]string
}

// NewTSVSource returns a source reading path on every Version and Entries call.
func NewTSVSource(name, path string) *TSVSource {
	return &TSVSource{name: name, path: path}
}

// NewTSVSourceData returns a source over in-memory TSV content, e.g. a file
// the host embeds with //go:embed. The slice is not copied and must not change.
func NewTSVSourceData(name string, data []byte) *TSVSource {
	return &TSVSource{name: name, data: data, inline: true}
}

// Name implements Source.
func (s *TSVSource) Name() string { return s.name }

// content returns the in-memory data or the current file content.
func (s *TSVSource) content() ([]byte, error) {
	if s.inline {
		return s.data, nil
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("gazetteer: source %s: %w", s.name, err)
	}
	return data, nil
}

// Version implements Source.
func (s *TSVSource) Version(context.Context) (string, error) {
	data, err := s.content()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Manifest returns the header of the last Entries run (a copy).
func (s *TSVSource) Manifest() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return maps.Clone(s.manifest)
}

// Entries implements Source.
func (s *TSVSource) Entries(ctx context.Context, yield func(Entry) error) error {
	data, err := s.content()
	if err != nil {
		return err
	}

	manifest := map[string]string{}
	var bad []LineError
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	line, seenEntry := 0, false
	for sc.Scan() {
		line++
		if line%1024 == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		text := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(text) == "" {
			continue
		}
		if strings.HasPrefix(text, "#") {
			if !seenEntry {
				if k, v, ok := manifestLine(text); ok {
					manifest[k] = v
				}
			}
			continue
		}
		seenEntry = true
		e, err := parseTSVLine(text)
		if err != nil {
			bad = append(bad, LineError{Line: line, Msg: err.Error()})
			continue
		}
		if err := yield(e); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("gazetteer: source %s: %w", s.name, err)
	}
	s.mu.Lock()
	s.manifest = manifest
	s.mu.Unlock()
	if len(bad) > 0 {
		return &ParseErrors{Source: s.name, Lines: bad}
	}
	return nil
}

func manifestLine(text string) (string, string, bool) {
	body := strings.TrimSpace(strings.TrimPrefix(text, "#"))
	k, v, ok := strings.Cut(body, ":")
	if !ok {
		return "", "", false
	}
	k = strings.TrimSpace(k)
	if k == "" || strings.ContainsFunc(k, func(r rune) bool { return !(r >= 'a' && r <= 'z' || r == '_') }) {
		return "", "", false
	}
	return k, strings.TrimSpace(v), true
}

func parseTSVLine(text string) (Entry, error) {
	f := strings.Split(text, "\t")
	if len(f) < 4 || len(f) > 6 {
		return Entry{}, fmt.Errorf("want 4 to 6 tab-separated fields, got %d", len(f))
	}
	e := Entry{
		Type:      strings.TrimSpace(f[0]),
		Ref:       strings.TrimSpace(f[1]),
		Canonical: strings.TrimSpace(f[2]),
		Alias:     strings.TrimSpace(f[3]),
	}
	if e.Type == "" {
		return Entry{}, errors.New("empty type")
	}
	if e.Alias == "" {
		return Entry{}, errors.New("empty alias")
	}
	if len(f) >= 5 {
		flags, err := ParseEntryFlags(f[4])
		if err != nil {
			return Entry{}, err
		}
		e.Flags = flags
	}
	if len(f) == 6 && strings.TrimSpace(f[5]) != "" {
		e.Attrs = map[string]string{}
		for _, kv := range strings.Split(f[5], ";") {
			kv = strings.TrimSpace(kv)
			if kv == "" {
				continue
			}
			k, v, ok := strings.Cut(kv, "=")
			if !ok || strings.TrimSpace(k) == "" {
				return Entry{}, fmt.Errorf("bad attribute %q", kv)
			}
			e.Attrs[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return e, nil
}
