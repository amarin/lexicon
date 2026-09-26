# AGENTS.md

Project instructions for AI agents (Codex, Claude, LGTM).

## About the project
- `github.com/amarin/lexicon` is a Go **library** for text analysis and (from v0.2)
  dictionary NER over gomorphy. No HTTP/MCP, no logger: packages return errors and
  expose state; hosts (first: genodex, `../genodex`) do the rest.
- Design: `docs/specs/2026-09-24-lexicon-design.md`. Plans: `docs/plans/`.
  Implementation write-ups: `docs/implementation/`. Roadmap and open questions:
  `docs/todo.md`.
- User documentation: `docs/en/` (index, scenarios, installation, cli, library) and
  the same pages in Russian in `docs/ru/`. Runnable examples: `examples/<name>/main.go`
  (checked by `examples/examples_test.go`); godoc examples: `example_test.go`,
  `textnorm/example_test.go`.
- Packages: `textnorm` (orthography rule sets, tokenizer with byte + rune offsets,
  pre-reform endings; no dictionaries), root package `lexicon` (dictionary registry,
  profiles, `Analyzer`), `basefetch` (optional: download and compile the OpenCorpora
  base dictionary; imports gomorphy's pymorphy loader), `cmd/lexicon` (CLI). From
  v0.2: `gazetteer` (TSV/in-memory alias sources compiled into versioned trie
  snapshots, zero-allocation matching), `rules` (YAML/JSON rule books: hints,
  triggers, tag-gated rule sets), `ner` (the extraction pipeline — `Doc`, `Span`,
  `Result`, filters, scoring, overlap resolution — built on `gazetteer` and
  `rules`), `nertest` (golden-JSONL test harness with strict/partial
  precision/recall), and `internal/fakedict` (the `lexicon.Dictionaries`/
  `Analyzer` test fixture shared by `gazetteer`, `ner` and `nertest` tests, not
  part of the public API).

## Conventions
- Code comments, error texts, docs, CLI output: English. Test fixtures: Russian text.
  Exception: `docs/ru/` holds Russian translations of the `docs/en/` user pages;
  project documents (spec, plans, write-ups, todo) stay English only.
- One struct (or named type with methods) per file; helpers live in the file of their
  main function; snake_case file names.
- Dependencies between packages are interfaces declared in the consumer's `deps.go`.
  No mockgen: tests use small hand-written fakes in `*_test.go`.
- Library dependencies: the root package `lexicon` (and `textnorm`) imports only
  `github.com/amarin/gomorphy/pkg/morphology` and `golang.org/x/text` (check with
  `go list -deps .`). `basefetch` — and through it `cmd/lexicon` — additionally
  imports `gomorphy/pkg/pymorphy` and `github.com/amarin/logging` (which pull zap);
  hosts that do not import `basefetch` never compile them. `gazetteer` imports
  only the root `lexicon` package and `textnorm` — no yaml, no new dependency.
  `rules` — and through it `ner`, `nertest` and `cmd/lexicon` — additionally
  imports `go.yaml.in/yaml/v3` (v3.0.5, added in v0.2 for rule files; check with
  `go list -f '{{join .Imports "\n"}}' ./<pkg>`). Tests: stdlib `testing` only.
  A new dependency needs the owner's approval.
- Go version policy: `go` directive = current Go minus two minor versions
  (`go 1.25.0`, `toolchain go1.27.1` as of 2026-09); dependency updates must not
  raise it. Language/library features newer than 1.25 are not used.
- Offsets contract (binding): offsets refer to the input string exactly as passed;
  byte and code-point offsets on every token; boundaries on grapheme clusters;
  `input[Start:End] == Raw`. Never return offsets into a normalized string.
- Changing a table in `textnorm.Modern`/`textnorm.PreReform`, or any rule that changes
  their output, requires bumping its `Version`; changing analyzer rules that change
  produced terms requires bumping `analyzerVersion` (hosts reindex). Host profile
  definitions are not part of `Analyzer.Version`: hosts version their profiles
  themselves.
- Changing NER behaviour that changes extracted spans (candidate generation, filters,
  rule application, scoring, resolution, output flags) requires bumping
  `ner.extractorVersion` (part of `Result.Version`; hosts recompute stored spans).
  Before the first release of a milestone (while it is "unreleased") the version is
  not bumped: no host stores spans from it.
- Documentation follows behaviour, in both `docs/en/` and `docs/ru/`:
  - a new user-visible feature gets a scenario in `scenarios.md` ("Available since"),
    an entry in `library.md`/`cli.md`, and a runnable example (an `// Output:` block
    when it needs no downloaded data) or an `ExampleXxx` function;
  - a behaviour change moves or adds a line under the scenario's **History**
    (a "Behaviour in X" line becomes history when it changes) and marks
    ⚠ reindex when it bumps `Rules.Version` or `analyzerVersion`, ⚠ re-extract
    when it bumps `ner.extractorVersion` only (RU: «⚠ переиндексация»,
    «⚠ переизвлечение»); CHANGELOG stays the source of truth;
  - README marks features with the version they appear in (`*(0.1.0)*`;
    implemented but not yet released: `*(0.3, unreleased)*`; `*(planned: 0.3)*`).

## Main commands
- `go build ./...`, `go vet ./...`, `gofmt -l .` (must be empty).
- `go test ./... -count=1`; `go test -race ./... -count=1` (includes `go test ./examples/`,
  which builds and runs every example with an `// Output:` block; `-short` skips it).
- `go test ./textnorm/ -run '^$' -fuzz FuzzTokenize -fuzztime 30s`,
  `go test ./ner/ -run '^$' -fuzz FuzzExtractOffsets -fuzztime 30s -fuzzminimizetime 5s` —
  one fuzz target per run (for `ner`, the default 60 s input minimization would stall a short run).
- `go test ./... -run '^$' -bench . -benchmem` — benchmarks.
- `LEXICON_BASE_DAT=/path/base.opencorpora.dat go test -tags integration ./... -count=1` —
  tests with the real base dictionary (`LEXICON_FETCH=1` also exercises the download).
- `go run ./cmd/lexicon analyze [--dicts DIR] [--ortho modern|prereform] [--profile SPEC] [--mode index|full] TEXT...`
- `go run ./cmd/lexicon dicts list|fetch [--dicts DIR]`
- `go run ./cmd/lexicon extract [--dicts DIR] [--ortho modern|prereform] [--gazetteer FILE...] [--rules FILE...] [--nest outer>inner...] [--tags T,...] [--types T,...] [--explain] [--format table|jsonl] TEXT...|-` —
  NER over CLI args or stdin (`-`, one document per line).
- `go run ./cmd/lexicon golden [--dicts DIR] [--ortho modern|prereform] [--gazetteer FILE...] [--rules FILE...] --cases CASES.jsonl [--min-precision N] [--min-recall N]` —
  run a golden-JSONL case set and report strict/partial precision/recall per type.
- `go run ./examples/<name>` — runnable examples (`examples/README.md`).

## Sibling modules during development
- gomorphy is used as a released module (currently v1.2.0): `go.mod` requires it
  directly, no `replace` directive, ever.
- A local, uncommitted `go.work` (`go work init . ../gomorphy` or the matching
  worktree path) is only for trying unreleased gomorphy changes; never commit it —
  `/go.work` and `/go.work.sum` go in `.gitignore` the moment one is created.
- `main` gets tagged releases only; merges are `git merge --ff-only`. Final check
  before merging: `go build ./... && go test ./... -count=1` against the tagged
  gomorphy version.

## One working copy
- Work happens in the main checkout unless a session says otherwise (e.g. parallel
  work needing isolation: `git worktree add ../lexicon-<topic> -b <topic>`).
- Stage only files you changed yourself (`git add <paths>`, not `git add -A`).

## When to ask the owner
- Changing exported API of `textnorm`, `lexicon`, `basefetch`, `gazetteer`, `rules`
  or `ner` once a host depends on it.
- Changing orthography tables or analyzer behaviour that changes index terms
  (hosts must reindex).
- Adding a dependency.

## Restricted files
- `*.dat` — generated dictionaries, never committed.
