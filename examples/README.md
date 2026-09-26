# Examples

Minimal, runnable programs for the usage scenarios in
[docs/en/scenarios.md](../docs/en/scenarios.md)
([Russian](../docs/ru/scenarios.md)). The API reference is
[docs/en/library.md](../docs/en/library.md); shorter `ExampleXxx` functions
are in [example_test.go](../example_test.go) and
[textnorm/example_test.go](../textnorm/example_test.go) (rendered on
pkg.go.dev).

| Example | What it shows | Scenarios | Needs external data? |
|---|---|---|---|
| [tokenize](tokenize/main.go) | `textnorm.Tokenize`: byte, rune and UTF-16 offsets, `SplitHyphen` | 3 | no |
| [orthography](orthography/main.go) | `Modern` vs `PreReform`, homoglyphs, `ReformVariants` | 1, 2, 8 | no |
| [index](index/main.go) | `Analyze` in `ModeIndex`: search-index terms, stop words, flags | 4 | no — a tiny base TSV in code |
| [query](query/main.go) | `ParseQuery`: complete words and the `Partial` last word | 5 | no |
| [profiles](profiles/main.go) | `Profile` kinds and grammemes against homonymy | 6 | no |
| [full](full/main.go) | `ModeFull` markup of a pre-reform text, `stop`/`reform`/`predicted` | 7, 8 | no |
| [abbrev](abbrev/main.go) | an `abbrev` dictionary: «с.», «у.», «кр-нин» | 9 | no |
| [registry](registry/main.go) | a dictionary directory: `.meta`, broken files, `SetEnabled`, `Reload`, `Version` | 10, 12, 13 | no — files written to a temp dir |
| [embed](embed/main.go) | a built-in dictionary via `//go:embed`, overridden by a file | 10 | no — `abbrev.records.tsv` is committed |
| [ner](ner/main.go) | *(0.2.0)* gazetteer + rules → spans: hints, a trigger candidate, a rule set by document tag, `Explain` | 15, 16, 17 | no — morphology, gazetteer and rules in code |
| [gazetteer](gazetteer/main.go) | *(0.2.0)* sources, the build report, `Refresh` of a changed host source, raw matches, variant groups | 15, 18 | no |
| [golden](golden/main.go) | *(0.2.0)* `nertest`: strict/partial precision and recall, `Check`, `WithTags` from case context | 19 | no |
| [base](base/main.go) | the real OpenCorpora base: provenance, `text` vs `name` profile | 4, 6, 11 | yes — `lexicon dicts fetch` first |

Run any of them from the repository root:

```bash
go run ./examples/tokenize
go run ./examples/orthography
go run ./examples/index
go run ./examples/query
go run ./examples/profiles
go run ./examples/full
go run ./examples/abbrev
go run ./examples/registry
go run ./examples/embed
go run ./examples/ner
go run ./examples/gazetteer
go run ./examples/golden

go run ./cmd/lexicon dicts fetch
go run ./examples/base -dicts ~/.local/share/lexicon/dicts
```

The examples without external data stand in for the OpenCorpora base with
a few TSV rows, so words outside them come out `unknown` or `predicted`
where the real base would know them — compare with `examples/base`.

Each example is a short, self-contained `package main`. Examples that need
no external data end `main` with an `// Output:` comment; `go test
./examples/` runs each of them and checks its output against that
comment, so the outputs shown here never go stale.
