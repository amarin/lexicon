# lexicon

**lexicon** is a Go library for analyzing Russian text and extracting named
entities with dictionaries, rules and morphology. It needs no ML models, no
LLMs and no external services.

It gives you a single analysis pipeline that works the same way whether you
are building a search index, parsing a user query or marking up entities in
a document:

- **Normalization** *(0.1.0)*. Modern and pre-reform orthography (ѣ, і,
  ѳ, final ъ), Latin/Cyrillic homoglyph repair for OCR and typed input, and
  tokenization whose byte, rune and UTF-16 offsets always point back into
  the original text.
- **Morphology** *(0.1.0)*. Lemmatization through
  [gomorphy](https://github.com/amarin/gomorphy), with a separate dictionary
  profile for each field (names, places, general vocabulary); pre-reform
  adjective endings and abbreviations of records; search-index terms,
  search-query parsing and per-token markup from one analyzer.
- **Dictionary NER** *(planned: 0.2, patterns 0.3)*. Multi-word aliases
  matched by lemmas or surface forms, variant groups, abbreviation hints,
  trigger words, and pattern rules switched on by document tags.
  Overlapping matches are resolved, and every span can explain why it was
  produced.
- **Dictionaries as data.** Morphology dictionaries are gomorphy `.dat` or
  TSV files with provenance manifests, switched on and off and hot-reloaded
  without locking readers *(0.1.0)*. Gazetteers as TSV and rules as YAML,
  rebuilt one source in milliseconds and swapped in atomically *(planned:
  0.2)*.

What each feature is for, how to use it and since which version:
[usage scenarios](docs/en/scenarios.md) ([по-русски](docs/ru/scenarios.md)).

## Use cases

- Search engines that need matching normalization at index time and at
  query time *(0.1.0)*
- Entity extraction from archival and historical documents (parish
  registers, censuses, letters) — the text analysis *(0.1.0)*, dictionary
  NER *(planned: 0.2)*
- Genealogy and digital-humanities tools
- Pre-annotation of text before human review, or a lightweight provider
  inside a larger NER pipeline — per-token markup *(0.1.0)*, entity spans
  *(planned: 0.2)*
- Domain-specific NER where you control the entity types and dictionaries
  *(planned: 0.2)*

lexicon only marks up text. It does not link the spans it finds to your
data: resolving a mention to a specific person or place, storing the
annotations and exposing an HTTP or MCP API are left to the host application.

## Status

0.1.0 is released: text analysis — `textnorm` orthography and
tokenization, the dictionary `Registry`, and the `Analyzer` (`ModeIndex`,
`ModeFull`, `ParseQuery`), plus `basefetch` and the `lexicon` CLI.
Dictionary NER (gazetteers, rules, `lexicon extract`) is planned for 0.2,
patterns for 0.3 — see the [roadmap](docs/todo.md).

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

— or supply your own base bytes through `Options.Base`. Requirements, what
each package pulls into your binary, platforms:
[installation](docs/en/installation.md).

## Usage

```go
reg, err := lexicon.Open(ctx, lexicon.Options{
	Dir:     "/var/lib/app/dicts",                              // base.opencorpora.dat, surname.parish.tsv, ...
	Builtin: []lexicon.BuiltinDict{abbrevs},                    // host data, e.g. //go:embed
	Kinds:   []lexicon.Kind{"surname", "given", "patronymic"}, // the host's own dictionary kinds
})
if err != nil { ... }
defer reg.Close()

a := lexicon.NewAnalyzer(reg, textnorm.PreReform, lexicon.AnalyzerOptions{})
name := lexicon.Profile{
	Name:      "name",
	Kinds:     []lexicon.Kind{lexicon.KindBase, "surname", "given", "patronymic"},
	Grammemes: map[lexicon.Kind][]string{lexicon.KindBase: {"Name", "Surn", "Patr"}},
}
terms := a.Analyze("Кузнецова Ивана", name, lexicon.ModeIndex) // index terms with lemmas
query := a.ParseQuery("Кузнецов", name)                         // for a search box
version := a.Version() // store with derived data; rebuild when it changes
```

Runnable programs for every scenario, most with no download:
[examples/](examples/README.md) (`go run ./examples/index`).

Replace dictionary files atomically (write a temporary file, rename), then call
`reg.Reload(ctx)`: `.dat` files are memory-mapped.

## CLI

```bash
go run ./cmd/lexicon dicts fetch                      # base dictionary into ~/.local/share/lexicon/dicts
go run ./cmd/lexicon dicts list
go run ./cmd/lexicon analyze --ortho prereform "Кр-нин с. Покровскаго, 1834 г."
go run ./cmd/lexicon analyze --mode full --profile 'name:base[Name|Surn|Patr],surname' "У Ивана сын Петр"
```

Flags, output format and sample output: [CLI](docs/en/cli.md).

## Attribution

lexicon ships no dictionary data. The base dictionary fetched by `basefetch` is
built from [OpenCorpora](http://opencorpora.org) data (via pymorphy2-dicts-ru),
distributed under CC BY-SA; hosts that distribute it must attribute it — the
provenance is in `Entry.Manifest` (`lexicon dicts list`).

## Documentation

User documentation is in English and Russian; project documents are in
English only. Index: [docs/en/index.md](docs/en/index.md),
[docs/ru/index.md](docs/ru/index.md).

| Document | English | Русский |
|---|---|---|
| Usage scenarios: what each feature is for, since which version | [docs/en/scenarios.md](docs/en/scenarios.md) | [docs/ru/scenarios.md](docs/ru/scenarios.md) |
| Installation | [docs/en/installation.md](docs/en/installation.md) | [docs/ru/installation.md](docs/ru/installation.md) |
| CLI: `analyze`, `dicts list`, `dicts fetch` | [docs/en/cli.md](docs/en/cli.md) | [docs/ru/cli.md](docs/ru/cli.md) |
| Library reference: packages, types, contracts | [docs/en/library.md](docs/en/library.md) | [docs/ru/library.md](docs/ru/library.md) |
| Runnable examples | [examples/](examples/README.md) | — |
| Roadmap and open questions | [docs/todo.md](docs/todo.md) | — |
| Design spec, implementation plans and write-ups | [docs/en/index.md](docs/en/index.md#project-english-only) | — |
| Release history | [CHANGELOG.md](CHANGELOG.md) | — |

## Development

See [AGENTS.md](AGENTS.md).
