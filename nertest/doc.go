// Package nertest scores a NER extractor against golden cases stored as
// JSON lines. Gold spans are given by surface text (n-th occurrence) and
// type; the report has strict (exact bytes) and partial (overlap) precision
// and recall per type plus the list of strict failures. A case may carry
// host Context (key/value data) that nertest ignores unless a WithTags
// option turns it into document tags.
package nertest
