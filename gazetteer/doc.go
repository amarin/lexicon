// Package gazetteer compiles dictionary aliases into token tries and matches
// them against analyzed text.
//
// Each Source is compiled into a lemma trie (sequences of lemma IDs, with
// ambiguous alias tokens expanded to at most MaxLemmaKeys combinations) and a
// surface trie (sequences of normalized forms). Strings are interned into a
// process-wide append-only table of uint32 IDs. Compiled sources form an
// immutable Snapshot that Gazetteer swaps atomically; readers never lock.
// Matching reports every alias match, overlapping ones included: choosing
// between them is the job of package ner.
//
// What each feature is for, with runnable examples:
// https://github.com/amarin/lexicon/blob/main/docs/en/scenarios.md#15-find-entities-with-dictionaries
// (Russian: docs/ru/scenarios.md).
package gazetteer
