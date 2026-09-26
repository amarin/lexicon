# Changelog

Notable changes of the project are recorded in this file (format inspired by
[Keep a Changelog](https://keepachangelog.com/); versions follow semver).

## [Unreleased]

### Added

- `gazetteer`: `Entry`/`EntryFlag` (`RequiresContext`, `SurfaceOnly`,
  `CaseSensitive`, `Blocked`); `TSVSource` (files or in-memory data via
  `NewTSVSourceData`, `# key: value` manifest, bad lines listed not fatal)
  and `SliceSource`; a process-wide interner; per-source token tries over
  lemma and surface keys (lemma keys expand combinations, capped at 8);
  zero-allocation matching from every token, branching over lemma
  alternatives, bounded by the longest alias and sentence ends, overlaps
  allowed; variant groups (`Canonical`, `Expand`) by `(source, Ref)`;
  `TypeProfiles`; versioned `Snapshot` swapped atomically on `Refresh`/
  `RefreshSource`; a build report (entries, aliases, keys, capped, blocked,
  duration, errors).
- `rules`: `RuleSet`/`When`, `Hint`, `Trigger` compiled into a `Book`;
  YAML (`go.yaml.in/yaml/v3`) and JSON rule files with a top-level `meta:`
  manifest and tag-gated rule sets; `Load`/`LoadNamed`/`LoadFile` errors
  carry `<file>:<line>` (`Compile` errors additionally `<set>/<rule>`).
- `ner`: the extraction pipeline — `Doc`, `Span`, `SpanFlag`, `Result`,
  `Config` (weights, `Nesting`), `New`/`Extract`, `Explain`; filters
  (`Blocked`, `RequiresContext`, `CaseSensitive`, `MinLemmaMatchRunes`);
  hints and triggers with context windows, absorption and boosts;
  negative triggers subtracting weight from overlapping gazetteer
  candidates; a weighted-sum score (origin and type weight per word,
  length bonus, evidence, ambiguity penalty) with exact ties broken by
  type name; weighted interval scheduling over overlapping spans with
  configurable nesting; same-range losers of other types kept as
  `Span.Alternatives`; a `Types` filter applied after resolution; one
  pinned gazetteer snapshot per `Extract` call (the analyzer and its
  registry stay live), safe for concurrent use; `Result.Version` covering
  the extractor version, analyzer, snapshot (including the gazetteer's
  profiles), rules and pipeline configuration; linear in document size.
- `nertest`: a golden-JSONL test harness (`Case`, `Run`) reporting strict
  and partial precision/recall/F1 per entity type; `Case.Context` for
  host document metadata plus `WithTags` to derive rule-set tags from it.
- CLI: `lexicon extract` (text args or `-` for stdin, one document per
  line; table or JSONL output, `--gazetteer`, `--rules`, `--nest`,
  `--tags`, `--types`, `--explain`) and `lexicon golden` (`--cases`,
  `--min-precision`, `--min-recall`).
- New dependency: `go.yaml.in/yaml/v3` (v3.0.5), for `rules` YAML files.

### Documentation

- Usage scenarios — what each feature is for, since which version, with
  per-feature history and reindex marks — plus installation, CLI and
  library reference pages, in English (`docs/en/`) and Russian
  (`docs/ru/`).
- Runnable examples in `examples/` (all but `base` need no download; their
  output is checked by `go test ./examples/`) and godoc `Example*`
  functions for `textnorm` and `lexicon`.
- `docs/todo.md`: roadmap, open questions carried over from v0.1,
  follow-ups.
- README marks every feature with the version it appears in and separates
  released from planned.

## [0.1.0] - 2026-09-25

### Added

- `textnorm`: `Modern` and `PreReform` orthography rule sets with versions;
  tokenizer emitting every token (words of any script, numbers, punctuation,
  symbols) with byte and code-point offsets into the input as passed,
  grapheme-cluster boundaries, script, case, dotted and sentence-end marks;
  `SplitHyphen`, `UTF16Offsets`, pre-reform ending variants.
- `lexicon`: dictionary `Registry` (host built-ins, `<kind>.<name>.dat|tsv`
  files, provenance sidecars, broken files listed not fatal, content-hash
  version, enable/disable through a `StateStore`, hot `Reload` with
  reference-counted retirement); host-defined `Profile`s; `Analyzer` with
  `ModeIndex` (genodex search P1 behaviour), `ModeFull` (one term per token)
  and `ParseQuery`; per-analyzer lemma cache. Stop words are words whose
  exact readings are all service parts of speech (PREP, CONJ, PRCL, INTJ),
  readings tagged `Abbr` ignored (OpenCorpora letter-name readings of «в»,
  «с», «и»). Base predictions are used only when neither the form nor any
  pre-reform variant has exact readings: exact readings all filtered out by
  the profile give the unknown lemma. Symlinked dictionary files are
  followed.
- `basefetch`: download and compile the OpenCorpora base dictionary.
- CLI `lexicon analyze`, `lexicon dicts list|fetch`.
