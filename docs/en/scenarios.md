# Usage scenarios

> Russian version: [docs/ru/scenarios.md](../ru/scenarios.md).

What lexicon is for, task by task: what each scenario solves, which calls
do it, and a runnable example. [library.md](library.md) and
[cli.md](cli.md) are the reference for every call named here; this page is
the "why" and "which one".

Each scenario states the version it is available since; "X
(unreleased)" means the feature is on `main` but not yet in a tagged
release (install it with `go get github.com/amarin/lexicon@main`). Three more marks
keep a per-feature changelog next to the feature:

- **Behaviour in 0.1.0** (**Behaviour in 0.2**, …) — a rule that was a
  deliberate decision and may change later. When it changes, the line
  moves to **History**.
- **History** — what changed in later versions, a per-feature excerpt of
  [CHANGELOG.md](../../CHANGELOG.md), which stays the source of truth.
- **⚠ reindex** — the change alters produced forms or terms: it bumps
  `textnorm.Rules.Version` or the analyzer version, `Analyzer.Version()`
  changes, and hosts must rebuild derived data (see
  [scenario 13](#13-know-when-to-reindex)). Such a change also changes spans.
- **⚠ re-extract** — the change alters extracted spans but not index terms:
  it bumps the extractor version, `ner.Result.Version` changes, and hosts
  recompute stored spans or suggestions, not the search index.

Examples: `go run ./examples/<name>` from the repository root
([examples/](../../examples/README.md)); `ExampleXxx` functions are in
`example_test.go` and `textnorm/example_test.go` and render on pkg.go.dev.
Every example except `base` needs no download: its dictionaries,
gazetteers and rules are tiny TSVs and YAML strings in code.

| # | Scenario | Since | Example |
|---|---|---|---|
| 1 | [Compare words across orthographies](#1-compare-words-across-orthographies) | 0.1.0 | [orthography](../../examples/orthography/main.go), `ExampleNormalizeWord` |
| 2 | [Repair Latin letters inside Cyrillic words](#2-repair-latin-letters-inside-cyrillic-words) | 0.1.0 | [orthography](../../examples/orthography/main.go) |
| 3 | [Tokens with offsets into the original text](#3-tokens-with-offsets-into-the-original-text) | 0.1.0 | [tokenize](../../examples/tokenize/main.go), `ExampleTokenize`, `ExampleUTF16Offsets` |
| 4 | [Terms for a search index](#4-terms-for-a-search-index) | 0.1.0 | [index](../../examples/index/main.go), `ExampleAnalyzer_Analyze` |
| 5 | [Parse a search query as it is typed](#5-parse-a-search-query-as-it-is-typed) | 0.1.0 | [query](../../examples/query/main.go), `ExampleAnalyzer_ParseQuery` |
| 6 | [A profile per field against homonymy](#6-a-profile-per-field-against-homonymy) | 0.1.0 | [profiles](../../examples/profiles/main.go), `ExampleProfile` |
| 7 | [Mark up every token (input for NER and review)](#7-mark-up-every-token-input-for-ner-and-review) | 0.1.0 | [full](../../examples/full/main.go), `ExampleAnalyzer_Analyze_full` |
| 8 | [Pre-reform texts](#8-pre-reform-texts) | 0.1.0 | [full](../../examples/full/main.go), `ExampleReformVariants` |
| 9 | [Abbreviations of records](#9-abbreviations-of-records) | 0.1.0 | [abbrev](../../examples/abbrev/main.go) |
| 10 | [Your own dictionaries: files and built-ins](#10-your-own-dictionaries-files-and-built-ins) | 0.1.0 | [registry](../../examples/registry/main.go), [embed](../../examples/embed/main.go), `ExampleOpen` |
| 11 | [The base dictionary and its attribution](#11-the-base-dictionary-and-its-attribution) | 0.1.0 | [base](../../examples/base/main.go), `ExampleParseManifest` |
| 12 | [Switch dictionaries on and off, reload without restart](#12-switch-dictionaries-on-and-off-reload-without-restart) | 0.1.0 | [registry](../../examples/registry/main.go) |
| 13 | [Know when to reindex](#13-know-when-to-reindex) | 0.1.0 | [registry](../../examples/registry/main.go), `ExampleAnalyzer_Version` |
| 14 | [Inspect by hand](#14-inspect-by-hand) | 0.1.0 | CLI `analyze`, `dicts list` |
| 15 | [Find entities with dictionaries](#15-find-entities-with-dictionaries) | 0.2.0 | [ner](../../examples/ner/main.go) |
| 16 | [Context words, triggers and document tags](#16-context-words-triggers-and-document-tags) | 0.2.0 | [ner](../../examples/ner/main.go) |
| 17 | [Overlapping matches, nesting and explanations](#17-overlapping-matches-nesting-and-explanations) | 0.2.0 | [ner](../../examples/ner/main.go) |
| 18 | [Keep gazetteers current without a restart](#18-keep-gazetteers-current-without-a-restart) | 0.2.0 | [gazetteer](../../examples/gazetteer/main.go) |
| 19 | [Measure quality on a golden set](#19-measure-quality-on-a-golden-set) | 0.2.0 | [golden](../../examples/golden/main.go) |
| 20 | [Try NER by hand](#20-try-ner-by-hand) | 0.2.0 | CLI `extract`, `golden` |
| — | [Planned: patterns](#planned-patterns) | 0.3 | — |

## 1. Compare words across orthographies

**Task.** Make «Вѣра», «Вера» and «ВЕРА», «Ѳеодоръ» and «Федор», «ёлка»
and «елка» compare equal — in a search index, in deduplication, in
matching a query against old documents.

**How.** Pick a rule set: `textnorm.Modern` (lower case, NFC, ё→е,
combining marks dropped except in й) or `textnorm.PreReform` (Modern plus
ѣ→е, і→и, ѳ→ф, ѵ→и and other old letters, and the final ъ dropped from
every hyphen part). `textnorm.NormalizeWord(r, w)` is the form a word is
indexed and compared by; `textnorm.Orthography(r, s)` maps a whole string
(the final ъ is a word rule and stays). The results are for comparison
only; offsets never refer to them ([scenario 3](#3-tokens-with-offsets-into-the-original-text)).

```go
textnorm.NormalizeWord(textnorm.PreReform, "Санктъ-Петербургъ") // санкт-петербург
textnorm.NormalizeWord(textnorm.Modern, "Ёлка")                 // елка
```

A custom rule set is a new `textnorm.Rules` value with its own `Name` and
`Version`; never modify `Modern`/`PreReform`.

**Example:** [orthography](../../examples/orthography/main.go),
`ExampleNormalizeWord`, `ExampleOrthography`; CLI `analyze --ortho
prereform`.

**Available since:** 0.1.0 (`Modern` and `PreReform` at `Version` "3").

## 2. Repair Latin letters inside Cyrillic words

**Task.** OCR output and hand-typed text mix look-alike Latin letters into
Cyrillic words: «Kот», «Iоаннъ». Such a word would never match its
dictionary form.

**How.** Nothing to call: `Rules.Homoglyphs` does it in the tokenizer and
in `Orthography`. A word part (hyphens separate parts) that has a Cyrillic
letter and no letters other than known homoglyphs is Cyrillic: the
tokenizer keeps it one `ScriptCyrillic` word and its homoglyphs are
replaced. Parts with other Latin letters («Kотw») or without Cyrillic
letters («XIX») are left alone. `PreReform` also treats Latin i/I as
pre-reform і/І.

**Example:** [orthography](../../examples/orthography/main.go) («Kот»,
«Iоаннъ», «XIX»).

**Available since:** 0.1.0. Letters carrying combining marks (a Latin
letter with an accent inside a Cyrillic word) fold too.

## 3. Tokens with offsets into the original text

**Task.** Highlight a found word, underline an entity, or cut a quote from
the text exactly as stored — even though matching ran on a normalized
form, and even when the client is JavaScript (UTF-16 strings).

**How.** `textnorm.Tokenize(r, input)` returns every token — words of any
script, numbers, punctuation, symbols — with byte offsets
(`Start`/`End`, `input[Start:End] == Raw` always) and code-point offsets
(`RuneStart`/`RuneEnd`); boundaries fall on grapheme clusters.
`textnorm.UTF16Offsets(input, offs...)` converts byte offsets for
JavaScript. `textnorm.SplitHyphen(tok)` gives the parts of a hyphenated
word, each with its own offsets and `Form`. `Token.Dotted` marks a word
directly followed by «.» (an abbreviation candidate); `SentenceEnd` marks
«.», «!», «?», «;», «…» and the last token before a blank line.

```go
for _, t := range textnorm.Tokenize(textnorm.PreReform, input) {
    fmt.Println(input[t.Start:t.End], t.Form, t.RuneStart, t.RuneEnd)
}
```

**Example:** [tokenize](../../examples/tokenize/main.go),
`ExampleTokenize`, `ExampleSplitHyphen`, `ExampleUTF16Offsets`.

**Available since:** 0.1.0.

**Behaviour in 0.1.0:** offsets always refer to the input exactly as
passed to `Tokenize`, never to a normalized string. This is a binding
contract, not a default: it will not change.

## 4. Terms for a search index

**Task.** Turn a document field into index terms: «Кота», «котом» → «кот»;
service words dropped; numbers kept; an ambiguous form («стали» → «сталь»,
«стать») indexed under every lemma.

**How.** Open a `Registry` ([scenario 10](#10-your-own-dictionaries-files-and-built-ins)),
create one `Analyzer` with `lexicon.NewAnalyzer(reg, rules,
AnalyzerOptions{})` and call `Analyze(text, profile, lexicon.ModeIndex)`.
Each `Term` has the source `Token`, the `Form` looked up and `Lemmas`
(`Text`, `Flags`, and the `Tag`, `Kind`, `Dict` of the reading that gave
the lemma). Index the lemmas; keep `Form` too if you search by prefix
([scenario 5](#5-parse-a-search-query-as-it-is-typed)). A hyphenated word
is indexed whole and by parts. Only Cyrillic words and numbers are
indexed.

The flags say how much to trust a lemma: none — a dictionary word;
`ambiguous` — one of several lemmas; `predicted` — the word is in no
dictionary, the base dictionary guessed the lemma from the ending;
`unknown` — no reading at all, the lemma is the form itself; `abbrev`,
`reform` — see scenarios [9](#9-abbreviations-of-records) and
[8](#8-pre-reform-texts).

```go
a := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
for _, t := range a.Analyze("Кота и пса стали кормить", lexicon.Profile{Name: "text"}, lexicon.ModeIndex) {
    for _, l := range t.Lemmas {
        index.Add(docID, l.Text)
    }
}
```

An `Analyzer` is safe for concurrent use and caches lemmas per profile name
(`AnalyzerOptions.CacheSize`, default 50 000 entries; the cache is dropped
when the dictionary set changes). Use one `Analyzer` for indexing and
query parsing so the two cannot diverge.

**Example:** [index](../../examples/index/main.go),
`ExampleAnalyzer_Analyze`; with the real base: [base](../../examples/base/main.go).

**Available since:** 0.1.0.

**Behaviour in 0.1.0:**
- A stop word is a word whose exact readings are all service parts of
  speech (PREP, CONJ, PRCL, INTJ); readings tagged `Abbr` do not count (the
  OpenCorpora letter-name readings of «в», «с», «и»). There is no stop
  list: «что», «как», «из» have non-service readings and stay indexed.
- Predictions are used only when neither the form nor any pre-reform
  variant has an exact reading, and only from dictionaries of kind `base`.
  When the profile filters out every exact reading, the lemma is `unknown`
  — a guess never replaces a filtered-out dictionary word.

## 5. Parse a search query as it is typed

**Task.** A search box: «Кузнецова Ив» should find documents with «Кузнецов»
in any form and any word starting with «ив».

**How.** `Analyzer.ParseQuery(q, profile)` analyses the query exactly like
`ModeIndex`. The last word, when nothing follows it (not even a space),
becomes `Query.Partial`: search it by prefix over indexed forms and by its
lemmas (union). A trailing service word has no lemmas and is prefix-only.
`Query.Terms` are the complete words. A number or a dotted word at the end
is complete.

```go
q := a.ParseQuery("кота Кузн", name)
// q.Terms: кота → кот; q.Partial: кузн — prefix «кузн*» or its lemmas
```

**Example:** [query](../../examples/query/main.go),
`ExampleAnalyzer_ParseQuery`.

**Available since:** 0.1.0.

## 6. A profile per field against homonymy

**Task.** In a "name" field «Вера» is a given name, not "faith", and
«Мороз» a surname, not "frost"; in a "place" field «Покровском» is a
village, not an adjective. A single dictionary set answers all of them.

**How.** A `Profile` is host configuration: `Name` (the cache key, unique
per `Analyzer`), `Kinds` (dictionary kinds consulted; empty = all) and
`Grammemes` (per kind, a reading is kept only if its tag has one of the
listed grammemes). Host kinds (`surname`, `given`, `toponym`, …) are
declared in `Options.Kinds`. Readings come in registry order, and a lemma
keeps the tag of the first reading that gave it.

```go
name := lexicon.Profile{
    Name:      "name",
    Kinds:     []lexicon.Kind{lexicon.KindBase, "surname", "given", "patronymic"},
    Grammemes: map[lexicon.Kind][]string{lexicon.KindBase: {"Name", "Surn", "Patr"}},
}
place := lexicon.Profile{
    Name:      "place",
    Kinds:     []lexicon.Kind{lexicon.KindBase, "toponym", lexicon.KindAbbrev},
    Grammemes: map[lexicon.Kind][]string{lexicon.KindBase: {"Geox"}},
}
```

CLI: `--profile 'name:base[Name|Surn|Patr],surname'`.

**Example:** [profiles](../../examples/profiles/main.go), `ExampleProfile`.

**Available since:** 0.1.0.

**Behaviour in 0.1.0:** profiles are not part of `Analyzer.Version()`. A
host that changes a profile definition versions it itself and reindexes
the fields that use it.

## 7. Mark up every token (input for NER and review)

**Task.** Show a text with every word annotated — for pre-annotation before
human review, a highlighting UI, or your own entity rules — without losing
punctuation or service words, which matter for context.

**How.** `Analyze(text, profile, lexicon.ModeFull)` returns exactly one
`Term` per token, punctuation included (punctuation and symbols have no
lemmas). Service words keep their lemmas with the `stop` flag instead of
being dropped. Numbers and non-Cyrillic words get an `unknown` lemma
without a dictionary lookup. Offsets come from the token
([scenario 3](#3-tokens-with-offsets-into-the-original-text)).

**Example:** [full](../../examples/full/main.go),
`ExampleAnalyzer_Analyze_full`; CLI `analyze --mode full`.

**Available since:** 0.1.0.

**History:** 0.2.0 — dictionary NER runs on top of this mode:
gazetteers, rules and resolved spans, scenarios
[15](#15-find-entities-with-dictionaries)–[20](#20-try-ner-by-hand).

## 8. Pre-reform texts

**Task.** Analyse documents written before 1918: «Иванъ Кузнецовъ изъ
села Покровскаго» — old letters, final ъ, and adjective endings -аго,
-яго, -ыя, -ія that no modern dictionary knows.

**How.** Build the `Analyzer` with `textnorm.PreReform` (letters and ъ,
[scenario 1](#1-compare-words-across-orthographies)). For endings nothing
is needed: when a form has no exact reading, the analyzer tries its modern
variants from `textnorm.ReformVariants` in order and takes the first the
dictionaries know; the lemma is flagged `reform`, and the indexed `Form`
stays as written («покровскаго»).

```go
textnorm.ReformVariants("покровскаго") // [покровского]
textnorm.ReformVariants("хорошаго")    // [хорошого хорошего]: the ending depends on stress
```

**Example:** [full](../../examples/full/main.go) («Шуйскаго» →
«шуйский[reform]»), `ExampleReformVariants`; CLI `analyze --ortho
prereform`.

**Available since:** 0.1.0.

## 9. Abbreviations of records

**Task.** Parish registers and censuses are full of abbreviations: «кр-нин»
(крестьянин), «с.» (село or сын), «у.» (уезд), «губ.» (губерния). Index
and mark them up by their full words.

**How.** Supply a dictionary of kind `abbrev` — a file
`abbrev.<name>.tsv` or a built-in — whose wordform column is the
abbreviation with its dot or hyphen and whose lemma column is the full
word. A dotted word is first looked up with its dot, a hyphenated word
first as a whole; the lemma is flagged `abbrev`, and `Term.Form` includes
the dot («с.»). lexicon ships no abbreviation list: it is host data.

```
крестьянин	кр-нин	NOUN
село	с.	NOUN
сын	с.	NOUN
```

**Example:** [abbrev](../../examples/abbrev/main.go).

**Available since:** 0.1.0.

**Behaviour in 0.1.0:**
- A one-letter dotted abbreviation («с.», «ц.») is always `ambiguous`; an
  undotted one-letter word («с» — a preposition) never takes an
  abbreviation reading.
- Abbreviation lookups ignore the profile's `Kinds`: a profile without
  `abbrev` still expands «с.». Strict per-profile filtering is an open
  question ([todo](../todo.md)).
- Without an abbreviation dictionary the base dictionary decides: in the
  OpenCorpora base «г.» gives «г», «кр-нин» is guessed.

## 10. Your own dictionaries: files and built-ins

**Task.** Add host vocabulary — surnames, given names, toponyms, estates —
as dictionaries users can edit, and ship some of it inside the binary.

**How.** `lexicon.Open(ctx, lexicon.Options{...})`:
- `Dir` — a directory of `<kind>.<name>.dat` (gomorphy binary, memory-
  mapped) and `<kind>.<name>.tsv` files (`lemma<TAB>wordform[<TAB>tags]`,
  built at load with gomorphy `ImportTSV`). The dictionary name is the file
  name without the extension. Symbolic links are followed.
- `<file>.meta` — an optional provenance sidecar of `key: value` lines
  (`source`, `license`, `version`, `url`, `generated_from`), exposed as
  `Entry.Manifest`.
- `Builtin` — host dictionaries in memory (`//go:embed`), `FormatTSV` or
  `FormatDat`. A file with the same name in `Dir` replaces a built-in, so
  users can override shipped data without a rebuild.
- `Kinds` — the host's own kinds. A file of an undeclared kind (a typo like
  `surnme.x.tsv`) is listed with an error and not used; nil accepts every
  valid kind (the CLI does that).

Registry order: the base, then built-ins in order, then the remaining files
by name. A broken file never fails `Open`: `Registry.List()` shows it with
`Entry.Error`, `Registry.Summary()` counts it.

**Example:** [registry](../../examples/registry/main.go),
[embed](../../examples/embed/main.go), `ExampleOpen`; CLI `dicts list`.

**Available since:** 0.1.0.

**Behaviour in 0.1.0:** only dictionaries of kind `base` contribute
predictions. gomorphy predicts for every built dictionary, and a small
surname list would otherwise "recognise" almost any word.

## 11. The base dictionary and its attribution

**Task.** Get general Russian morphology (OpenCorpora, ~15 MB) without
committing it, and attribute it correctly (CC BY-SA 4.0) if you
distribute it.

**How.** Three ways:
- CLI: `lexicon dicts fetch [--dicts DIR]` downloads pymorphy2-dicts-ru,
  compiles it and writes `base.opencorpora.dat` plus its `.meta` sidecar.
- Go: `basefetch.Fetch(dst)`. Only this package pulls gomorphy's
  pymorphy loader and a logger (zap); hosts that do not import it never
  compile them.
- Embed: put the `.dat` bytes into `Options.Base` and the `.meta` that
  `basefetch` wrote, parsed with `lexicon.ParseManifest`, into
  `Options.BaseManifest`. It is registered as `base.builtin`; any `base.*`
  file in `Dir` replaces it.

`Entry.Manifest` of the base carries source, version, URL and license:
show it wherever you attribute data. lexicon itself ships no dictionary
data.

**Example:** [base](../../examples/base/main.go),
`ExampleParseManifest`; CLI `dicts fetch`, `dicts list`.

**Available since:** 0.1.0.

## 12. Switch dictionaries on and off, reload without restart

**Task.** Let an administrator disable a noisy dictionary, or drop in a new
surname list, while the service keeps answering queries.

**How.** `Registry.SetEnabled(ctx, name, on)` stores the state through
`Options.State` (a `StateStore` the host implements over its database;
nil = in memory, everything enabled) and swaps in a new snapshot.
`Registry.Reload(ctx)` rescans `Dir`, reopens every dictionary and swaps.
Readers never lock: `Parse` pins the current snapshot, and old dictionaries
are closed once in-flight calls finish.

Replace dictionary files atomically — write a temporary file, then rename
it over the old one. A `.dat` file is memory-mapped: overwriting it in
place can crash in-flight parses.

**Example:** [registry](../../examples/registry/main.go).

**Available since:** 0.1.0.

**Behaviour in 0.1.0:** `Reload` reopens every dictionary even when its
file did not change (reuse by hash is an open question,
[todo](../todo.md)).

## 13. Know when to reindex

**Task.** A search index stores lemmas computed by some version of the
rules and some set of dictionaries. Know when they are stale.

**How.** Store `Analyzer.Version()` with the derived data and rebuild when
it changes. It combines three parts:

| Part | Changes when | Example value |
|---|---|---|
| analyzer version | analysis rules of the library change terms | `analyzer-2` |
| `Rules.Name`-`Rules.Version` | an orthography table or rule changes forms | `prereform-3` |
| `Registry.Version()` | a dictionary is added, removed, enabled, disabled or its content changes | a content hash |

Extracted spans have their own version, `ner.Result.Version`
([scenario 15](#15-find-entities-with-dictionaries)). It changes with
`Analyzer.Version()` and also with:

| Part | Changes when | Example value |
|---|---|---|
| extractor version | NER rules of the library change spans (⚠ re-extract) | `ner-1` |
| gazetteer snapshot | a source is recompiled from new content or with a new analyzer | a hash |
| rule book | rule files change | a hash |
| pipeline configuration | profiles, weights, nesting, `MinLemmaMatchRunes` change | a hash |

Profiles are host definitions and are not part of it
([scenario 6](#6-a-profile-per-field-against-homonymy)). `Registry.Version()`
alone is enough to invalidate caches of dictionary lookups.

**Example:** [registry](../../examples/registry/main.go),
`ExampleAnalyzer_Version`; CLI `analyze` prints the version to stderr.

**Available since:** 0.1.0 (analyzer version "2", `Modern` and `PreReform`
"3").

**History:** every later entry marked ⚠ reindex or ⚠ re-extract on this
page names the part it bumps.

## 14. Inspect by hand

**Task.** Check what lexicon makes of a phrase, which dictionaries are
loaded, and where they came from — without writing code.

**How.** `lexicon analyze [--dicts DIR] [--ortho modern|prereform]
[--profile SPEC] [--mode index|full] TEXT...` prints one line per term:
raw text, form, lemmas with flags; the dictionary summary and
`Analyzer.Version()` go to stderr. `lexicon dicts list` prints every
dictionary with its kind, format, state, origin, hash, provenance and
error. See [cli.md](cli.md).

**Available since:** 0.1.0.

## 15. Find entities with dictionaries

**Task.** Mark up people, places and estates in a record: «Иван Петров из
деревни Лягушкино Боровского уезда» → a given name, a surname and two
places, each pointing back to the host's own record — matching inflected
forms («Боровского» → «Боровский») and spelling variants («Иоанн» →
«Иван»).

**How.** Three parts.
- **Gazetteer.** Aliases of host records come from `gazetteer.Source`s: a
  TSV file (`gazetteer.NewTSVSource(name, path)`), TSV bytes the host
  embeds (`NewTSVSourceData`), entries built in Go (`NewSliceSource`), or
  the host's own implementation over its database. A TSV line is
  `type<TAB>ref<TAB>canonical<TAB>alias[<TAB>flags[<TAB>k=v;k=v]]`;
  `# key: value` lines before the first entry are the provenance manifest.
  Aliases that share a `Ref` are variants of one record. `gazetteer.New`
  analyses every alias with your `Analyzer` (per type through
  `TypeProfiles`) and compiles its surface key and lemma keys (one per
  combination of the words' lemmas, at most `gazetteer.MaxLemmaKeys` = 8;
  none for `SurfaceOnly`).
- **Rules** (optional, [scenario 16](#16-context-words-triggers-and-document-tags)).
- **Pipeline.** `ner.New(ner.Config{Analyzer, Gazetteer, Rules, Profiles,
  DefaultProfile})`, then `Extract(ctx, ner.Doc{Text, Profile, Tags,
  Types})`.

```go
gz, _ := gazetteer.New(ctx, gazetteer.Config{
    Analyzer: an, DefaultProfile: text,
    Sources:  []gazetteer.Source{gazetteer.NewTSVSource("places", "place.tsv")},
})
p, _ := ner.New(ner.Config{Analyzer: an, Gazetteer: gz, Rules: book,
    Profiles: map[string]lexicon.Profile{"text": text}, DefaultProfile: "text"})
res, _ := p.Extract(ctx, ner.Doc{Text: "из деревни Лягушкино Боровского уезда"})
for _, s := range res.Spans {
    fmt.Println(s.Surface, s.Type, s.Refs, s.Normal) // Боровского уезда division [d2] [Боровский]
}
```

A `Span` has byte and code-point offsets into `Doc.Text`
(`Doc.Text[Start:End] == Surface`), the `Type`, the host `Refs` (opaque
keys; empty for a trigger candidate), `Normal` (the matched entries'
`Canonical` forms; for a trigger candidate the lemmas of its words), the
entries' `Attrs`, a `Score` and `Flags`: `Ambiguous` (several refs or
normal forms, a covered ambiguous abbreviation such as «с.» that no rule
absorbed, or a tie between types — not homonymy of an ordinary word),
`Predicted`, `Abbrev`, `Candidate` (proposed by a trigger, not a record)
and `Nested`. Entry flags tune matching: `SurfaceOnly` (no lemma key, for
abbreviations like «СПб»), `CaseSensitive`, `RequiresContext` (kept only
when a rule supports it: «Мороз» the surname, not "frost") and `Blocked`
(these words are not of this type). `Result.Version` covers the extractor,
analyzer, compiled gazetteer, rules and configuration: store it with
suggestions and recompute when it changes.

lexicon only marks up: linking a span to a specific person or place is the
host's job.

**Example:** [ner](../../examples/ner/main.go); CLI `extract`
([scenario 20](#20-try-ner-by-hand)).

**Available since:** 0.2.0.

**Behaviour in 0.2:**
- A one-word match found only by lemma is dropped when the alias is shorter
  than `Config.MinLemmaMatchRunes` (default 3 runes), so «с» never matches a
  one-letter alias by lemma.
- `Blocked` vetoes whatever the entry's other flags say: a blocked
  `CaseSensitive` entry vetoes in any letter case, and a blocked one-word
  lemma match vetoes even below `MinLemmaMatchRunes`.
- An alias matches only consecutive words with no punctuation between
  them («Большой, Лес» does not match «Большой Лес»); an abbreviation's own
  dot is skipped.
- A span ending in a dotted abbreviation excludes the dot («Калужской губ»,
  not «Калужской губ.»); extending the span over the dot is planned for 0.3.
- Variant groups (aliases sharing a `Ref`) are not used by extraction: a
  span's `Normal` comes from the `Canonical` fields of its matched entries
  ([scenario 18](#18-keep-gazetteers-current-without-a-restart)).
- The interner behind compiled aliases is process-wide and never shrinks.

## 16. Context words, triggers and document tags

**Task.** Use the words around a name: «деревни», «уезда», «ул.» say a
place is near; «село Покровское» is a place even when no record knows it;
in pre-1917 peasant records the word after «крестьянин» is probably a
surname, in other documents not.

**How.** A rule file is YAML (or JSON) with an optional top-level `meta:`
provenance mapping and `sets:` of rules; load it with `rules.LoadFile`,
`LoadNamed` or `Load` and compile one or more files with `rules.Compile`
into a `Book` for `ner.Config.Rules`.
- A **hint** is a keyword (`lemma: деревня|село`, `dotted: true` for «ул.»)
  that adds its `weight` to spans of its `type` whose edge is within
  `window` content words in `dir` (`right`, `left`, `both`; `window: 1` —
  right next to the keyword), satisfies their `RequiresContext`, and with
  `absorb: true` extends an adjacent span over the keyword. Hints have no
  `shape`.
- A **trigger** proposes a `Candidate` span of its `type` over the words
  after (or before) its keyword — `window: 1..2` words that fit `shape`
  (`case`, `script`), stopping at `stop_at` (`punct` always, `stop`,
  `number`, `latin`) — when no gazetteer span of that type overlaps them;
  otherwise it boosts the overlapping spans. Its `weight` is added either
  way: to the candidate it proposes or to the spans it boosts, and
  `absorb: true` extends both over an adjacent keyword. A `negative: true`
  trigger subtracts its weight from overlapping gazetteer spans of its type
  and never proposes or absorbs.
- A **rule set** with `when: [tags]` is active only for documents whose
  `Doc.Tags` include ALL of them; a set without `when` is always active.

```yaml
meta: {source: my rules, license: CC0-1.0}
sets:
  - name: places
    hints:
      - {lemma: уезд, type: division, dir: left, window: 1, weight: 2, absorb: true}
    triggers:
      - {lemma: деревня|село, type: division, window: 1..2, shape: {case: title, script: cyrillic}, absorb: true}
  - name: pre1917
    when: ["period:pre1917"]
    hints:
      - {lemma: крестьянин, type: surname, window: 2}
```

Keywords are lemmas or forms, compared after pre-reform normalization, so
one rule file serves modern and pre-reform text. Windows count content
words from the keyword to the span and never cross punctuation or a
sentence end (an abbreviation's dot does not break them); `rules.MaxWindow`
is 8. Errors name the place: `file.yaml:12: places/hint 0: …`.

**Example:** [ner](../../examples/ner/main.go) (a hint, a trigger
candidate, a set switched on by `period:pre1917`); CLI `extract --rules
--tags`.

**Available since:** 0.2.0.

**Behaviour in 0.2:**
- A negative trigger affects only gazetteer spans, never a trigger
  candidate. A span whose score drops to 0 or below is removed, so a
  negative trigger (or negative weights) can delete a match entirely.
- Hints run before triggers and never boost a trigger candidate.
- A `Blocked` entry also stops triggers: no candidate of the blocked type
  is proposed over a range that overlaps the blocked match.
- A hint measures its window from the span's current edge; after one hint
  absorbed its keyword, the next one measures from the new edge, so results
  can depend on the order of hints.
- A trigger's window can take in a following name of another type before
  that name's own candidate is considered; keep trigger windows short.
- Rules are compiled once: to change them, build a new `Pipeline`
  (a hot-swap is an open question, [todo](../todo.md)).

## 17. Overlapping matches, nesting and explanations

**Task.** «Боровского уезда» matches a place alias, a trigger and maybe a
surname at once; a street name may lie inside a city span. Pick one answer
per stretch of text, keep what lost, and show a reviewer why a span was
found.

**How.** `Extract` resolves candidates by weighted interval scheduling:
no two output spans cross, and a span may lie inside another only for an
outer/inner type pair listed in `Config.Nesting` (the inner one gets the
`Nested` flag). A candidate's score is a weighted sum
(`ner.Weights`, `DefaultWeights()` when zero): per word the origin weight
(surface 3, lemma 2, trigger 1) plus the type weight, a length bonus per
word beyond the first (0.5), the hint and trigger evidence, minus the
ambiguity penalty (0.25) times (the larger of the numbers of distinct refs
and normal forms − 1). Candidates scoring 0 or below are dropped before
resolution. Candidates of other
types on the winner's exact range become `Span.Alternatives`, best first;
an exact tie is never silent — the span is `Ambiguous`, and ties go to the
type name. `Doc.Types` hides unwanted types after resolution, so it never
changes which span wins. `ner.Explain()` fills `Span.Evidence`: «lemma
match «Боровский» (places)», «hint «уезда» → division +2».

```go
p, _ := ner.New(ner.Config{..., Nesting: map[string][]string{"city": {"street"}}})
res, _ := p.Extract(ctx, doc, ner.Explain())
```

**Example:** [ner](../../examples/ner/main.go) (evidence of every span);
CLI `extract --nest 'city>street' --explain` (quote `>` from the shell).

**Available since:** 0.2.0.

## 18. Keep gazetteers current without a restart

**Task.** A user adds a village or a surname variant in the host
application; extraction should see it within seconds, while other requests
keep running.

**How.** `Source.Version(ctx)` must be cheap and change with the content
(`TSVSource` hashes its bytes; a host source can return a revision
counter). `Gazetteer.Refresh(ctx)` recompiles only the sources whose
version (or the analyzer version) changed, `RefreshSource(ctx, name)` one
source regardless of its version; each returns `SourceReport`s (entries,
aliases, keys, capped expansions, blocked entries, duration, bad lines).
The new `Snapshot` is swapped in atomically: an `Extract` in flight pins
the snapshot it started with. A bad TSV line is listed in the report,
never fatal; a source that fails as a whole keeps its previous data.

`Gazetteer.Expand(lemma)` returns the lemma keys of the variant groups
(aliases sharing a `Ref`) that contain it — a search box can expand
«иван» to «иоанн»; `Canonical(key)` maps a key back to canonical forms.

**Example:** [gazetteer](../../examples/gazetteer/main.go).

**Available since:** 0.2.0.

**Behaviour in 0.2:**
- Only the gazetteer snapshot is pinned per `Extract`: the analyzer and its
  dictionary registry are live, and after a `Registry.Reload` the snapshot
  keeps aliases compiled with the old analyzer until the host calls
  `Refresh`.
- A `Refresh` cancelled through its context publishes nothing and discards
  the sources it had rebuilt; the next call rebuilds them. A source whose
  `Version` fails keeps its data, with the error in its report.
- `TSVSource` manifest keys are `[a-z_]` only; any other `# Key: value`
  line is a plain comment.

## 19. Measure quality on a golden set

**Task.** Know whether a new rule or dictionary made extraction better or
worse, and fail a CI run when a type regresses.

**How.** Keep cases as JSON lines: `{"id", "text", "profile", "tags",
"spans": [{"text", "type", "occurrence"}], "context"}` — a gold span is
given by its surface text (the n-th occurrence, default 1) and type.
`nertest.LoadCases`/`LoadCasesFile` read them (every gold span must be
found in its text); `nertest.Run(ctx, pipeline, cases)` returns a
`Report` with strict (exact bytes and type) and partial (overlap)
precision, recall and F1 per type, and the list of missed and spurious
spans. `Report.Write` prints it; `Report.Check(minPrecision, minRecall)`
returns one message per failed threshold («surname: strict recall 0.500 <
0.900») — an empty list passes. `context` is host metadata of the
document that lexicon ignores; `nertest.WithTags(func(Case) []string)` adds
the tags your function derives from it (for example `estate:peasant` from
`{"estate": "peasant"}` — the mapping is yours, not built in).

```go
rep, _ := nertest.Run(ctx, p, cases)
if v := rep.Check(0.9, 0.9); len(v) > 0 {
    t.Fatalf("golden set regressed: %q", v)
}
```

**Example:** [golden](../../examples/golden/main.go); CLI `golden`.

**Available since:** 0.2.0.

## 20. Try NER by hand

**Task.** Check what a gazetteer and a rule file find in a phrase, or score
a golden set, without writing Go.

**How.** `lexicon extract [--gazetteer FILE]... [--rules FILE]...
[--nest OUTER>INNER]... [--tags T,...] [--types T,...] [--explain]
[--format table|jsonl] TEXT...` prints one row (or JSON line) per span;
`-` reads documents from stdin, one per line, each exactly as read
(offsets refer to the line). `lexicon golden --cases FILE.jsonl
[--gazetteer FILE]... [--rules FILE]... [--nest OUTER>INNER]...
[--min-precision N] [--min-recall N]` prints the report and exits with 1
below a threshold. Morphology comes from `--dicts` and `--ortho` as for
`analyze`. See [cli.md](cli.md).

**Available since:** 0.2.0.

**Behaviour in 0.2:** the CLI configures one profile, `text` (every enabled
dictionary kind), for documents and aliases alike, so a golden case with
another `profile` fails the run; a gazetteer source is named after its
file's basename without the extension, so `a/x.tsv` and `b/x.txt` fail as
duplicates.

## Planned: patterns

Not available yet; the design is in the
[spec](../specs/2026-09-24-lexicon-design.md), the steps in the
[plan](../plans/2026-09-24-v0.3-patterns.md).

- **0.3** — sequence patterns over spans, lemmas and grammemes with
  actions (relabel, boost, emit a fact) in the rule sets of
  [scenario 16](#16-context-words-triggers-and-document-tags); a span
  ending in an abbreviation extended over its dot.

When they ship, they get scenarios here with their own "Available since".
