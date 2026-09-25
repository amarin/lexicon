# Changelog

Notable changes of the project are recorded in this file (format inspired by
[Keep a Changelog](https://keepachangelog.com/); versions follow semver).

## [Unreleased]

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
  and `ParseQuery`; per-analyzer lemma cache.
- `basefetch`: download and compile the OpenCorpora base dictionary.
- CLI `lexicon analyze`, `lexicon dicts list|fetch`.
