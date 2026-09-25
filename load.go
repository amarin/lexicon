package lexicon

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// BaseBuiltinName is the registry name of Options.Base.
const BaseBuiltinName = "base.builtin"

// load opens every dictionary of o in registry order — base (files of kind
// base, else the built-in), built-ins in order (a same-name file replaces a
// built-in in place), remaining files by name — and applies the stored state.
func load(ctx context.Context, o Options) ([]*dict, error) {
	states, err := o.State.Enabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("lexicon: dictionary state: %w", err)
	}

	kinds, err := newKindSet(o.Kinds)
	if err != nil {
		return nil, err
	}

	files, err := scanDir(o.Dir, kinds)
	if err != nil {
		return nil, fmt.Errorf("lexicon: dictionary directory: %w", err)
	}

	var all, rest []*dict

	for _, f := range files {
		if f.entry.Kind == KindBase {
			all = append(all, f)
		} else {
			rest = append(rest, f)
		}
	}

	if len(all) == 0 && o.Base != nil {
		base := openBuiltin(BuiltinDict{Name: BaseBuiltinName, Kind: KindBase, Format: FormatDat, Data: o.Base})
		// Options.BaseManifest wins field by field over BuildInfo (owner decision 2026-09-25).
		base.entry.Manifest = mergeManifest(o.BaseManifest, base.entry.Manifest)
		all = append(all, base)
	}

	for _, b := range o.Builtin {
		if i := slices.IndexFunc(rest, func(f *dict) bool { return f.entry.Name == b.Name }); i >= 0 {
			all = append(all, rest[i])
			rest = slices.Delete(rest, i, i+1)

			continue
		}

		all = append(all, openBuiltin(b))
	}

	all = append(all, rest...)

	for _, x := range all {
		if on, ok := states[x.entry.Name]; ok {
			x.entry.Enabled = on
		}
	}

	return all, nil
}

// scanDir opens the .dat and .tsv files of dir sorted by name; a missing dir
// is empty. Invalid names, undeclared kinds and duplicates are listed with an
// error, unopened.
func scanDir(dir string, kinds kindSet) ([]*dict, error) {
	if dir == "" {
		return nil, nil
	}

	items, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var out []*dict

	seen := map[string]bool{}

	for _, it := range items {
		ext := filepath.Ext(it.Name())
		if !it.Type().IsRegular() || (ext != ".dat" && ext != ".tsv") {
			continue
		}

		path := filepath.Join(dir, it.Name())
		name := strings.TrimSuffix(it.Name(), ext)

		if seen[name] {
			x := newDict(Entry{Name: name, Origin: path, Enabled: true})
			x.entry.Error = fmt.Sprintf("lexicon: duplicate dictionary name %q", name)
			out = append(out, x)

			continue
		}

		seen[name] = true
		out = append(out, loadFile(path, name, ext, kinds))
	}

	slices.SortStableFunc(out, func(a, b *dict) int { return strings.Compare(a.entry.Name, b.entry.Name) })

	return out, nil
}

// loadFile opens one directory file; its sidecar manifest, if any, wins over
// BuildInfo. A file of a kind not accepted by kinds is listed (with its kind)
// and an "unknown kind" error, unopened (owner decision 7).
func loadFile(path, name, ext string, kinds kindSet) *dict {
	format := FormatDat
	if ext == ".tsv" {
		format = FormatTSV
	}

	x := newDict(Entry{Name: name, Format: format, Origin: path, Enabled: true})

	kind, err := kindOf(name)
	if err != nil {
		x.entry.Error = err.Error()

		return x
	}

	x.entry.Kind = kind

	if err := kinds.check(kind); err != nil {
		x.entry.Error = fmt.Sprintf("lexicon: dictionary %q: %v", name, err)

		return x
	}

	if format == FormatTSV {
		data, err := os.ReadFile(path)
		if err != nil {
			x.entry.Error = err.Error()

			return x
		}

		x.openTSV(data)
	} else {
		x.openDat(path)
	}

	if m, ok := readManifest(ManifestPath(path)); ok {
		x.entry.Manifest = m
	}

	return x
}

// openBuiltin opens a host built-in; its Kind must match the name prefix.
func openBuiltin(b BuiltinDict) *dict {
	x := newDict(Entry{Name: b.Name, Kind: b.Kind, Format: b.Format, Origin: OriginBuiltin, Enabled: true})

	kind, err := kindOf(b.Name)

	switch {
	case err != nil:
		x.entry.Error = err.Error()

		return x
	case b.Kind != "" && b.Kind != kind:
		x.entry.Error = fmt.Sprintf("lexicon: built-in %q: kind %q does not match the name", b.Name, b.Kind)

		return x
	}

	x.entry.Kind = kind

	switch b.Format {
	case FormatTSV:
		x.openTSV(b.Data)
	case FormatDat:
		x.openBytes(b.Data)
	default:
		x.entry.Error = fmt.Sprintf("lexicon: built-in %q: unknown format %d", b.Name, b.Format)

		return x
	}

	if !b.Manifest.IsZero() {
		x.entry.Manifest = b.Manifest
	}

	return x
}

// readManifest reads a sidecar; false when it is missing or unreadable.
func readManifest(path string) (Manifest, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Manifest{}, false
	}
	defer f.Close()

	m, err := ParseManifest(f)

	return m, err == nil
}
