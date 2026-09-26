# Changelog

Notable changes of the project are recorded in this file (format inspired by
[Keep a Changelog](https://keepachangelog.com/); versions follow semver).

## [Unreleased]

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
