// Package ner extracts entity spans from text with dictionaries and rules.
//
// Pipeline.Extract runs: Analyzer.Analyze(ModeFull) → gazetteer matches →
// filters (blocked, case-sensitive, short lemma matches) → hints →
// triggers → RequiresContext filter → scoring → weighted interval
// scheduling (no crossing spans, nesting only for Config.Nesting pairs) →
// Doc.Types filter. Spans carry opaque gazetteer Refs; linking them to host
// data is the host's job. A Pipeline is safe for concurrent use; every call
// pins one gazetteer snapshot.
package ner
