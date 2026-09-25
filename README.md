# lexicon

**lexicon** is a Go library for analyzing Russian text and extracting named
entities with dictionaries, rules and morphology. It needs no ML models, no
LLMs and no external services.

It gives you a single analysis pipeline that works the same way whether you
are building a search index, parsing a user query or marking up entities in
a document:

- **Normalization.** Modern and pre-reform orthography (ѣ, і, ѳ, final ъ),
  Latin/Cyrillic homoglyph repair for OCR and typed input, and tokenization
  whose byte, rune and UTF-16 offsets always point back into the original
  text.
- **Morphology.** Lemmatization through
  [gomorphy](https://github.com/amarin/gomorphy), with a separate dictionary
  profile for each field (names, places, general vocabulary).
- **Dictionary NER.** Multi-word aliases matched by lemmas or surface forms,
  variant groups, abbreviation hints, trigger words, and pattern rules
  switched on by document tags. Overlapping matches are resolved, and every
  span can explain why it was produced.
- **Dictionaries as data.** Gazetteers are TSV files and rules are YAML
  files, each with a provenance manifest. You can rebuild one source in
  milliseconds and swap it in atomically without locking readers.

## Use cases

- Search engines that need matching normalization at index time and at
  query time
- Entity extraction from archival and historical documents (parish
  registers, censuses, letters)
- Genealogy and digital-humanities tools
- Pre-annotation of text before human review, or a lightweight provider
  inside a larger NER pipeline
- Domain-specific NER where you control the entity types and dictionaries

lexicon only marks up text. It does not link the spans it finds to your
data: resolving a mention to a specific person or place, storing the
annotations and exposing an HTTP or MCP API are left to the host application.

## Status

v0.1 implemented: text analysis — `textnorm` orthography and tokenization,
the dictionary `Registry`, and the `Analyzer` (`ModeIndex`/`ModeFull`,
`ParseQuery`), plus `basefetch` and the `lexicon` CLI. Dictionary NER
(gazetteers, rules, `cmd/lexicon extract`) is v0.2.

## Installation

```bash
go get github.com/amarin/lexicon
```

lexicon requires [gomorphy](https://github.com/amarin/gomorphy) v1.2.0 or
later. It ships no dictionary data: fetch the OpenCorpora base dictionary
with the CLI (network access, ~15 MB, CC BY-SA 4.0 — see Attribution) —

```bash
go run github.com/amarin/lexicon/cmd/lexicon dicts fetch
```

— or supply your own base bytes through `Options.Base`.

## Usage

```go
reg, err := lexicon.Open(ctx, lexicon.Options{
	Dir:          "/var/lib/app/dicts",           // base.opencorpora.dat, surname.x.tsv, ...
	Builtin:      []lexicon.BuiltinDict{abbrevs},  // host data, e.g. //go:embed
	Kinds:        []lexicon.Kind{"surname", "given", "patronymic"}, // nil = any kind
	State:        myStateStore,                    // nil: in memory, everything enabled
	Base:         embeddedBase,                    // optional: embed the base instead of Dir
	BaseManifest: embeddedBaseManifest,             // provenance sidecar for Base
})
if err != nil { ... }
defer reg.Close()

a := lexicon.NewAnalyzer(reg, textnorm.PreReform, lexicon.AnalyzerOptions{})
name := lexicon.Profile{
	Name:      "name",
	Kinds:     []lexicon.Kind{lexicon.KindBase, "surname", "given", "patronymic"},
	Grammemes: map[lexicon.Kind][]string{lexicon.KindBase: {"Name", "Surn", "Patr"}},
}
terms := a.Analyze("Кузнецова Ивана", name, lexicon.ModeIndex)
query := a.ParseQuery("Кузнецов", name) // OR of lemma variants, for a search box
version := a.Version() // store with derived data; recompute when it changes (profiles: version them yourself)
```

Replace dictionary files atomically (write a temporary file, rename), then call
`reg.Reload(ctx)`: `.dat` files are memory-mapped.

## CLI

```bash
go run ./cmd/lexicon dicts fetch                      # base dictionary into ~/.local/share/lexicon/dicts
go run ./cmd/lexicon dicts list
go run ./cmd/lexicon analyze --ortho prereform "Кр-нин с. Покровскаго, 1834 г."
go run ./cmd/lexicon analyze --mode full --profile 'name:base[Name|Surn|Patr],surname' "У Ивана сын Петр"
```

## Attribution

lexicon ships no dictionary data. The base dictionary fetched by `basefetch` is
built from [OpenCorpora](http://opencorpora.org) data (via pymorphy2-dicts-ru),
distributed under CC BY-SA; hosts that distribute it must attribute it — the
provenance is in `Entry.Manifest` (`lexicon dicts list`).

## Documentation

- [Design spec](docs/specs/2026-09-24-lexicon-design.md)
- Implementation plans:
  - [v0.1 — analysis](docs/plans/2026-09-24-v0.1-analysis.md),
    [lexicon](docs/plans/2026-09-24-v0.1-analysis-lexicon.md)
  - [v0.2 — NER](docs/plans/2026-09-24-v0.2-ner.md)
  - [v0.3 — patterns](docs/plans/2026-09-24-v0.3-patterns.md)
- [v0.1 implementation write-up](docs/implementation/v0.1-analysis.md)

## Development

See [AGENTS.md](AGENTS.md).
