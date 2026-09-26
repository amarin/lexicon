// Package rules holds NER rules as data: hints (a keyword that boosts and
// gives context to an adjacent span), triggers (a keyword that proposes a
// candidate span when the gazetteer has none) and, from v0.3, patterns.
// Rules are grouped in rule sets activated by document tags (When). Files
// are YAML (JSON documents load too) with a provenance header in a
// top-level meta mapping; Compile validates them into a Book.
package rules
