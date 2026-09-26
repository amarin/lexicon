// Package textnorm turns text into comparable forms and tokens: orthography
// rule sets (modern and pre-reform Russian), a tokenizer that emits every token
// with byte and code-point offsets into the input exactly as passed, helpers
// for hyphenated words and UTF-16 offsets, and pre-reform adjective ending
// variants. It knows nothing about dictionaries.
//
// What each feature is for, with runnable examples:
// https://github.com/amarin/lexicon/blob/main/docs/en/scenarios.md
// (Russian: docs/ru/scenarios.md).
package textnorm
