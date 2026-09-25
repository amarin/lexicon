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

Design stage, not implemented yet.

## Documentation

- [Design spec](docs/specs/2026-09-24-lexicon-design.md)
- Implementation plans:
  - [v0.1 — analysis](docs/plans/2026-09-24-v0.1-analysis.md),
    [lexicon](docs/plans/2026-09-24-v0.1-analysis-lexicon.md)
  - [v0.2 — NER](docs/plans/2026-09-24-v0.2-ner.md)
  - [v0.3 — patterns](docs/plans/2026-09-24-v0.3-patterns.md)
