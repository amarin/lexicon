// Package lexicon analyzes text for search indexing, query parsing and
// dictionary NER over gomorphy: a Registry of morphology dictionaries
// (host built-ins and <kind>.<name>.dat|tsv files of a directory,
// enabled/disabled state, hot reload), host-defined field Profiles, and an
// Analyzer that turns text into Terms with lemmas in an index mode and a full
// mode. Orthography and tokens come from package textnorm.
//
// What each feature is for, with runnable examples:
// https://github.com/amarin/lexicon/blob/main/docs/en/scenarios.md
// (Russian: docs/ru/scenarios.md).
package lexicon
