# AGENTS.md

Project instructions for AI agents (Codex, Claude, LGTM).

## About the project
- `github.com/amarin/lexicon` is a Go **library** for text analysis and (from v0.2)
  dictionary NER over gomorphy. No HTTP/MCP, no logger: packages return errors and
  expose state; hosts (first: genodex, `../genodex`) do the rest.
- Design: `docs/specs/2026-09-24-lexicon-design.md`. Plans: `docs/plans/`.
  Implementation write-ups: `docs/implementation/`.
- Packages: `textnorm` (orthography rule sets, tokenizer with byte + rune offsets,
  pre-reform endings; no dictionaries), root package `lexicon` (dictionary registry,
  profiles, `Analyzer`), `basefetch` (optional: download and compile the OpenCorpora
  base dictionary; imports gomorphy's pymorphy loader), `cmd/lexicon` (CLI).

## Conventions
- Code comments, error texts, docs, CLI output: English. Test fixtures: Russian text.
- One struct (or named type with methods) per file; helpers live in the file of their
  main function; snake_case file names.
- Dependencies between packages are interfaces declared in the consumer's `deps.go`.
  No mockgen: tests use small hand-written fakes in `*_test.go`.
- Library dependencies: gomorphy and `golang.org/x/text` only. Tests: stdlib
  `testing` only. A new dependency needs the owner's approval.
- Go version policy: `go` directive = current Go minus two minor versions
  (`go 1.25.0`, `toolchain go1.27.1` as of 2026-09); dependency updates must not
  raise it. Language/library features newer than 1.25 are not used.
- Offsets contract (binding): offsets refer to the input string exactly as passed;
  byte and code-point offsets on every token; boundaries on grapheme clusters;
  `input[Start:End] == Raw`. Never return offsets into a normalized string.
- Changing a table in `textnorm.Modern`/`textnorm.PreReform` requires bumping its
  `Version` (hosts reindex).

## Main commands
- `go build ./...`, `go vet ./...`, `gofmt -l .` (must be empty).
- `go test ./... -count=1`; `go test -race ./... -count=1`.
- `go test ./textnorm/ -run '^$' -fuzz FuzzTokenize -fuzztime 30s` — one fuzz target per run.
- `go test ./... -run '^$' -bench . -benchmem` — benchmarks.
- `LEXICON_BASE_DAT=/path/base.opencorpora.dat go test -tags integration ./... -count=1` —
  tests with the real base dictionary (`LEXICON_FETCH=1` also exercises the download).
- `go run ./cmd/lexicon analyze [--dicts DIR] [--ortho modern|prereform] [--profile SPEC] [--mode index|full] TEXT...`
- `go run ./cmd/lexicon dicts list|fetch [--dicts DIR]`

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
- Changing exported API of `textnorm`, `lexicon` or `basefetch` once a host depends on it.
- Changing orthography tables or analyzer behaviour that changes index terms
  (hosts must reindex).
- Adding a dependency.

## Restricted files
- `*.dat` — generated dictionaries, never committed.
