package nertest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// LoadCases reads JSON lines; blank lines are skipped, a missing ID becomes
// "line<N>", and every gold span must be found in its text.
func LoadCases(r io.Reader) ([]Case, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var out []Case
	line := 0
	for sc.Scan() {
		line++
		b := bytes.TrimSpace(sc.Bytes())
		if len(b) == 0 {
			continue
		}
		var c Case
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&c); err != nil {
			return nil, fmt.Errorf("nertest: line %d: %w", line, err)
		}
		if c.ID == "" {
			c.ID = fmt.Sprintf("line%d", line)
		}
		if _, err := c.goldBounds(); err != nil {
			return nil, fmt.Errorf("nertest: line %d (%s): %w", line, c.ID, err)
		}
		out = append(out, c)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("nertest: %w", err)
	}
	return out, nil
}

// LoadCasesFile reads cases from path.
func LoadCasesFile(path string) ([]Case, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("nertest: %w", err)
	}
	defer f.Close()
	return LoadCases(f)
}
