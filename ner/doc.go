// Package ner extracts entity spans from text with dictionaries and rules.
//
// Pipeline.Extract runs: Analyzer.Analyze(ModeFull) → gazetteer matches →
// filters (blocked, case-sensitive, short lemma matches) → hints →
// triggers → patterns → RequiresContext filter → scoring → weighted
// interval scheduling (no crossing spans, nesting only for Config.Nesting
// pairs) → Doc.Types filter → facts. Spans carry opaque gazetteer Refs; linking them to host
// data is the host's job. A Pipeline is safe for concurrent use; every call
// pins one gazetteer snapshot. Only the snapshot is pinned: the analyzer
// (and the dictionary registry behind it) is live, and the snapshot may have
// been compiled with an older analyzer until the host calls
// Gazetteer.Refresh. Spans a pattern creates with label can hold their
// parts (Config.Nesting); the facts patterns emit are in Result.Facts.
//
// What each feature is for, with runnable examples:
// https://github.com/amarin/lexicon/blob/main/docs/en/scenarios.md#15-find-entities-with-dictionaries
// (Russian: docs/ru/scenarios.md).
package ner
