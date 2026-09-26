# Installation

> Russian version: [docs/ru/installation.md](../ru/installation.md).

## Library

```bash
go get github.com/amarin/lexicon
```

Requirements:
- Go 1.25 or later (the `go` directive follows "current Go minus two minor
  versions").
- [gomorphy](https://github.com/amarin/gomorphy) v1.2.0 or later — pulled
  in by `go get`.

What each package brings into your binary:

| Package | Imports besides the standard library |
|---|---|
| `github.com/amarin/lexicon` (root), `textnorm` | `gomorphy/pkg/morphology`, `golang.org/x/text` |
| `basefetch` (optional) | the above plus `gomorphy/pkg/pymorphy` and `github.com/amarin/logging` (zap) |
| `gazetteer` *(0.2.0)* | the root package and `textnorm` only |
| `rules`, `ner`, `nertest` *(0.2.0)* | the above plus `go.yaml.in/yaml/v3` (rule files) |
| `cmd/lexicon` | all of the above (it uses `basefetch` and `ner`) |

A host that does not import `basefetch` never compiles the pymorphy loader
or zap.

## CLI

```bash
go install github.com/amarin/lexicon/cmd/lexicon@latest
```

See [cli.md](cli.md).

## Dictionary data

lexicon ships no dictionary data. For general Russian morphology fetch the
OpenCorpora base (network access, ~15 MB, CC BY-SA 4.0):

```bash
lexicon dicts fetch                      # into ~/.local/share/lexicon/dicts
lexicon dicts fetch --dicts /var/lib/app/dicts
```

or call `basefetch.Fetch(dst)` from Go, or embed the compiled `.dat` in your
binary through `Options.Base`
([scenario 11](scenarios.md#11-the-base-dictionary-and-its-attribution)).
Your own dictionaries are TSV or `.dat` files in the same directory, or
built-ins ([scenario 10](scenarios.md#10-your-own-dictionaries-files-and-built-ins)).

Gazetteers and rule files for entity extraction are host data too: TSV and
YAML files or in-memory sources
([scenario 15](scenarios.md#15-find-entities-with-dictionaries)).

Everything except the base can be tried with no download:
`go run ./examples/index`, `go run ./examples/ner` and the other
[examples](../../examples/README.md).

## Platforms

`.dat` files in `Options.Dir` are memory-mapped by gomorphy
(`morphology.Open`), which works on Unix (Linux, macOS, BSD). On Windows
pass `.dat` bytes through `Options.Base` or a `FormatDat` built-in: they are
opened from memory (`morphology.OpenBytes`). TSV dictionaries are always
built in memory.

## Attribution

If you distribute the base dictionary, attribute OpenCorpora (CC BY-SA):
its source, version, URL and license are in `Entry.Manifest`
(`lexicon dicts list`).
