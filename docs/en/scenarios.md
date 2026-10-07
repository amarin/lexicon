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
| 21 | [Share dictionaries with the gomorphy CLI](#21-share-dictionaries-with-the-gomorphy-cli) | 0.1.0 | CLI `dicts list`, `gomorphy lookup` |
| 22 | [Word forms and number agreement](#22-word-forms-and-number-agreement) | 0.3.0 | [inflect](../../examples/inflect/main.go), `ExampleRegistry_Inflect`, `ExampleNumeralGrammemes` |
| 23 | [Sequence patterns](#23-sequence-patterns) | 0.4.0 | [persons](../../examples/persons/main.go) |
| 24 | [A person as one span](#24-a-person-as-one-span) | 0.4.0 | [persons](../../examples/persons/main.go) |
| 25 | [Facts from patterns](#25-facts-from-patterns) | 0.4.0 | [persons](../../examples/persons/main.go) |
| — | [Planned](#planned) | — | — |

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
listed grammemes as a whole token — gomorphy `HasGrammeme`, so `Surn` does
not match `Surname`). Host kinds (`surname`, `given`, `toponym`, …) are
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
first as a whole; the lemma of an abbreviated form is flagged `abbrev`,
and `Term.Form` includes the dot («с.»). lexicon ships no abbreviation list: it is host data.

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

**History:** 0.2.1 — `abbrev` marks only an abbreviated form: a full word
that an abbreviation dictionary also knows («сын», «деревня», «город») is no
longer flagged `abbrev`, so a NER span over it is no longer `Abbrev` (nor
`Ambiguous` because of it), and a sentence dot after it ends the sentence
(⚠ re-extract: `ner-2`). Lemmas and index terms are unchanged.

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

The `gomorphy` CLI is a fourth way (`gomorphy update pymorphy -o
DIR/base.opencorpora.dat`), but it writes no `.meta`: see
[scenario 21](#21-share-dictionaries-with-the-gomorphy-cli).

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
| extractor version | NER rules of the library change spans (⚠ re-extract) | `ner-3` |
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
- 0.2.1 — extractor version `ner-2`: the `abbrev` flag fix
  ([scenario 9](#9-abbreviations-of-records)) changes span flags and
  sentence ends (⚠ re-extract).
- 0.4.0 — extractor version `ner-3`: a span ending in a dotted
  abbreviation includes the dot ([scenario 15](#15-find-entities-with-dictionaries);
  ⚠ re-extract).

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
- Variant groups (aliases sharing a `Ref`) are not used by extraction: a
  span's `Normal` comes from the `Canonical` fields of its matched entries
  ([scenario 18](#18-keep-gazetteers-current-without-a-restart)).
- The interner behind compiled aliases is process-wide and never shrinks.

**History:**
- 0.4: a span ending in a dotted abbreviation includes the dot («Калужской губ.»). Before 0.4 it excluded it. ⚠ re-extract.

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
`Nested` flag). A span created by a pattern `label` may hold an allowed span
on its own whole range (a one-word person and its surname); two dictionary
spans on one range still compete. *(0.4.0)* A candidate's score is a weighted sum
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

## 21. Share dictionaries with the gomorphy CLI

**Task.** Use the `gomorphy` command-line tool next to `lexicon`: build or
refresh the base dictionary with `gomorphy build|update|import|merge`,
check words with `gomorphy lookup` over the lexicon dictionary directory,
and know when files from one tool work in the other.

**How.**
- **One format.** Both tools read and write gomorphy `.dat` (GMOR).
  `lexicon dicts fetch` (`basefetch.Fetch`) compiles pymorphy2-dicts-ru the
  same way `gomorphy build pymorphy` does, so either file can be the
  lexicon base. The reader accepts only its own format version; it has been
  1 in every gomorphy 1.x release, and an incompatible file is listed by
  `lexicon dicts list` with a `format: unsupported version` error.
- **gomorphy → lexicon.** Write straight into the lexicon directory under a
  `<kind>.<name>.dat` name; gomorphy's default output
  (`.data/<type>/<type>.dat`) and a bare `pymorphy.dat` are not usable
  there (an invalid name is listed as broken):

  ```bash
  gomorphy update pymorphy -o ~/.local/share/lexicon/dicts/base.opencorpora.dat
  gomorphy import tsv surnames.tsv -o ~/.local/share/lexicon/dicts/surname.parish.dat --source parish
  ```

  gomorphy writes no `.meta` sidecar. Without one, `Entry.Manifest` comes
  from the file's BuildInfo — source, version and URL, **no license**; add
  the `.meta` yourself if you attribute from it. Delete the `.meta` that an
  earlier `lexicon dicts fetch` left next to the file you replace, or it
  keeps describing the old data.
- **lexicon → gomorphy.** `gomorphy -d DIR` (or `GOMORPHY_DICTIONARY=DIR`)
  loads every `*.dat` directly in `DIR` and ignores `.tsv` and `.meta`.
  Unlike lexicon ([scenario 10](#10-your-own-dictionaries-files-and-built-ins)),
  gomorphy predicts from every dictionary, so `gomorphy lookup` over a
  directory with user `.dat` files may show `(predicted)` readings that
  lexicon never returns, and it does not see TSV dictionaries at all. Pass
  only the base (`-d DIR/base.opencorpora.dat`) to compare like with like.
- **Rebuilds change the content hash.** Different gomorphy versions (the
  `gomorphy version` of the CLI vs the gomorphy version the lexicon binary
  was built with) or a newer pymorphy2-dicts-ru can produce a different
  file from the same command — for example, gomorphy 1.3.0 builds
  prediction for `opencorpora`/`unimorph` and groups Builder/TSV lemmas by
  part of speech. A different file means a different `Entry.Hash` and
  `Registry.Version()`: reindex as in [scenario 13](#13-know-when-to-reindex).
  The same source built by the same gomorphy version gives the same hash
  whichever tool built it (the hash ignores build time and other BuildInfo).
- **Replace files atomically.** gomorphy's `SaveTo` writes a temporary file
  and renames it, like `dicts fetch`, so `gomorphy build -o` over a file a
  running lexicon host has memory-mapped is safe; then call
  `Registry.Reload` ([scenario 12](#12-switch-dictionaries-on-and-off-reload-without-restart)).
  Do not copy over the file in place.
- **Tag vocabulary.** Profiles, abbreviations and rule books name
  OpenCorpora grammemes (`Surn`, `Geox`, `Abbr`, …). A base built with
  `gomorphy build unimorph` carries UniMorph tags, so those grammeme
  filters stop matching; keep an OpenCorpora-derived base
  (`pymorphy`/`opencorpora`).
- A dictionary merged with `gomorphy merge` is a regular `.dat`: named
  `base.*` it is the base and predicts; named with another kind it only
  gives exact readings.

**Example:** CLI `dicts list` after replacing a file (hash, provenance,
load error); `gomorphy lookup -d DIR/base.opencorpora.dat WORD`.

**Available since:** 0.1.0.

## 22. Word forms and number agreement

**Task.** Show a dictionary word in the form the interface needs: a plural
heading («Уезд» → «Уезды»), a label that agrees with a count («1 уезд»,
«2 уезда», «5 уездов»), a case form for a generated phrase.

**How.**
- `Registry.Inflect(word, kinds, from, want)` returns the forms of `word`
  that have every grammeme of `want` (`plur nomn`, `gent sing`, …).
  `kinds` selects dictionaries as in `Parse` (empty = all).
- **Pick the reading with `from`.** A written word is often several words:
  «село» is a noun and a past form of «сесть», «округ» is also a genitive
  plural of «округа». `from` lists grammemes the source reading must have —
  `NOUN nomn sing` for a dictionary headword; empty `from` inflects every
  reading and returns all their forms.
- **Take the first form.** The result holds distinct forms: dictionaries in
  registry order, and within a reading the form closest to the source one
  first («корпусы» before the informal «корпуса»). Empty `want` returns
  every form of the lexeme, the source form first.
- **Counts.** `NumeralGrammemes(n)` gives `want` for a noun after a number in
  a nominative phrase: 1, 21, 101 → `nomn sing`; 2–4, 22 → `gent sing`;
  0, 5–20, 11–14, 111 → `gent plur`. Other cases («о пяти уездах») you state
  yourself.
- **Letter ё.** «поселок» and «посёлок» find the same word. Forms come back
  as the dictionary stores them: with ё from the base («посёлки», «сёла»),
  with е from TSV dictionaries; fold ё yourself if your interface writes е.
  When the base and a TSV dictionary both know the word you get both
  spellings — pass `kinds` to choose.
- Grammemes are those of the answering dictionary — OpenCorpora for the
  base (`nomn gent datv accs ablt loct`, `sing plur`, `NOUN ADJF`, …).

**Limits.**
- Only words a dictionary knows are inflected. A word that `Parse` only
  predicts gives nothing, and so does every word when no base dictionary is
  loaded: keep a fallback («5 × стан», or the word unchanged).
- A dictionary does not separate senses: «корпус» gives «корпусы», a
  military corps needs «корпуса»; «год» after a count gives «годов» before
  «лет». Keep an override table for such terms, or put them in a `custom`
  dictionary and ask for it by `kinds`.
- Multi-word terms («отдельный батальон») are inflected word by word, each
  with its own `from` and `want`: after 2–4 the adjective takes `gent plur`
  («2 отдельных батальона»).
- Forms are lower-case; restore capitals yourself.

**Example:** [inflect](../../examples/inflect/main.go);
`ExampleRegistry_Inflect`, `ExampleNumeralGrammemes`.

**Available since:** 0.3.0.

## 23. Sequence patterns

**Task.** Say what a word is by what stands around it: after a given name
and a patronymic a village name is a surname; in a peasant record the word
after the given name is a patronymic; an unknown capitalised word between
name words is part of the name.

**How.** A rule set has `patterns:` next to `hints:` and `triggers:`. A
pattern is a sequence of `elements` and a list of `actions`.
- An element with `type` consumes one candidate span of that type (several
  types: `type: surname|patronymic`). Without `type` it consumes one term:
  `lemma: сын|дочь`, `grammeme: Name`, `token: word|number|punct|symbol`,
  `shape: {case: title, script: cyrillic}`, or `any: true`. Conditions on
  one element all have to hold; next to `type` they are checked on every
  word of the span.
- `not:` is a condition that must not match at the same position:
  `not: {type: given_name}` — no given-name candidate starts here.
- `group:` nests a sequence; `repeat: "?"`, `"*"`, `"+"` (quoted, greedy)
  repeats an element or a group; `role:` names what the element consumed.
- Actions: `relabel: {role, type}` retypes a span — the previous reading
  stays in `Span.Alternatives`, and a dictionary record of the new type on
  the same words wins with its refs; `boost: {role, weight}` adds evidence;
  `label: {role, type}` creates a span over the captured words (flag
  `Candidate` unless it covers spans the pattern consumed); `emit` records
  a fact (scenario 25).

```yaml
sets:
  - name: persons
    patterns:
      - name: place-as-surname
        elements:
          - {type: given_name}
          - {type: patronymic}
          - {type: division, shape: {case: title}, role: s}
        actions:
          - relabel: {role: s, type: surname}
      - name: unknown-patronymic
        elements:
          - {type: given_name}
          - {token: word, shape: {case: title}, not: {type: given_name|patronymic|surname}, role: p}
          - {type: surname}
        actions:
          - label: {role: p, type: patronymic}
```

Patterns run after triggers, before overlaps are resolved, so they see
every candidate reading of a word. They apply in file order, each on what
the previous one left: put patterns that fix parts before patterns that
use them. A pattern never crosses a sentence end; its matches do not
overlap; the leftmost match wins and repetitions are greedy. Errors name
the place: `file.yaml:12: persons/pattern "person": …`.

**Example:** [persons](../../examples/persons/main.go); CLI `extract
--rules`.

**Available since:** 0.4.0.

**Behaviour in 0.4:**
- A repeated `role` keeps its last repetition; there are no lazy
  repetitions ([todo](../todo.md)).
- A slot for an unknown word belongs after a name element, never first in
  a pattern: the first word of a sentence is capitalised too.
- `type` next to `shape` skips a span that absorbed a lower-case keyword
  («дер. Головина»).
- Punctuation is a term. The dot of an abbreviation between two spans
  («крест. Иван») needs `{token: punct, repeat: "?"}` between the two
  `type` elements.
- `Extract` stays linear in the document size for patterns without an
  unbounded repeat. An unbounded repeat (`{any: true, repeat: "*"}`) costs
  time quadratic in the length of a sentence: take the estate or period
  context from document tags (`when:`), not from an `any*` prefix.

## 24. A person as one span

**Task.** Get «Иван Петров Сидоров» as one mention with its parts — given
name, patronymic, surname — each with its own dictionary reference, also
when two persons stand back to back and when a person is one surname after
«ответчик».

**How.** Name parts are spans of your own types. A pattern groups them and
labels the group; the host allows the parts inside the whole:

```yaml
      - name: person
        elements:
          - group:
              - {type: given_name}
              - {type: patronymic|surname}
              - {type: surname, repeat: "?"}
            role: who
        actions:
          - label: {role: who, type: person}
      - name: person-by-defining-word
        elements:
          - {lemma: крестьянин|мещанин|ответчик|истец}
          - group:
              - {type: surname}
            role: who
        actions:
          - label: {role: who, type: person}
```

```go
ner.Config{Nesting: map[string][]string{"person": {"given_name", "patronymic", "surname"}}}
```

The person span has no `Refs`; its `Normal` joins the normal forms of its
parts. The parts follow it in `Result.Spans` with the flag `Nested`; the
type of a part is its role. A one-word person («ответчик Сидоров») covers
the name alone, with the surname nested on the same range. Without the
`Nesting` pairs you get the parts and no person. Each person is one match,
so «Иван Петров Сидоров Анна Михайлова Кузнецова» gives two persons.

A complete rule set with golden cases:
[persons.rules.yaml](../../nertest/testdata/persons.rules.yaml),
[persons.jsonl](../../nertest/testdata/persons.jsonl).

**Example:** [persons](../../examples/persons/main.go); CLI `extract --ortho prereform
--rules persons.yaml --nest 'person>given_name' --nest 'person>surname'`.

**Available since:** 0.4.0.

**Behaviour in 0.4:**
- Whether one name word is a person is your rule set's choice: write a
  pattern with a defining word. A lone given name stays a `given_name`.
- A person takes the word flags `Predicted` and `Ambiguous` (an ambiguous
  abbreviation) only from words outside its parts: a part settled its own
  words and carries their flags itself. `Abbrev` is set for an abbreviation
  anywhere in the span. A span a pattern created without parts inside it
  is `Predicted` when one of its words is known only by predicted lemmas.
- A part that a pattern created from an unknown word is a `Candidate`
  without `Refs`; a part that was relabelled keeps its previous reading in
  `Alternatives` (a surname that is also a village).
- The person's `Normal` joins the canonical forms of its dictionary parts
  with the lower-case lemmas of the parts a pattern labelled or relabelled
  and of the words outside parts («Анна Петров головин»). Agreed forms
  are an open question ([todo](../todo.md), Q-v03-10).
- A part that requires context (`requires_context`: a surname that is also
  a common word, «Мороз») is kept when a pattern assembles a person over
  it — the match is its context: «Иван Мороз» is a person with the surname
  part, «ударил мороз» gives nothing.
- A one-word person can stand over an unknown word («ответчик Лаптев»):
  one pattern labels the word a `surname`, a later one labels the person
  over it (`unknown-surname-by-defining-word` in the rule set above). The
  surname is nested on the same range and is a `Candidate`.
- Other readings of a one-word person's range (the surname is also a
  village) are `Alternatives` of the nested part, not of the person, and a
  tie marks the part `Ambiguous`.

## 25. Facts from patterns

**Task.** Get «дочь X» and «N лет» as plain data pointing at spans, and
build your own relations from it.

**How.** `emit: {kind, args}` records a fact; `args` maps your fact roles
to pattern roles. A role must be a span: bound to a `type` element, or
labelled or relabelled earlier in the same pattern.

```yaml
      - name: age
        elements:
          - {type: person|given_name, role: person}
          - {token: punct, repeat: "?"}
          - {token: number, role: age}
          - lemma: год
        actions:
          - label: {role: age, type: age}
          - emit: {kind: age, args: {person: person, age: age}}
```

`Result.Facts` holds `Fact{Kind, Args, Rule}`; `Args` are indexes into
`Result.Spans`, `Rule` is the pattern name. A fact is dropped when one of
its spans did not make it to the output (lost an overlap, filtered by
`Doc.Types`); equal facts are reported once.

**Example:** [persons](../../examples/persons/main.go); CLI `extract`
prints fact rows.

**Available since:** 0.4.0.

**Behaviour in 0.4:**
- A list after one keyword («восприемники: …») yields one fact per match,
  not one fact with a list ([todo](../todo.md)).
- `nertest` golden sets score spans only, not facts.

## 26. Dates from record columns

Parish and civil records write dates in pre-reform Russian: «21 генваря»,
«1 іюня 1883 г.», «3 сент.», or just «21.01» when the year sits in a
heading. `dates.Parse(text, ctxYear)` reads such a date; `dates.Month(word)`
reads a lone month word. Spelling goes through `textnorm.PreReform`, so
ѣ, і, ѳ and letter case are not your problem.

```go
d, ok := dates.Parse("21 генваря 1883 года", 1884) // 1883-01-21, from text
d, ok = dates.Parse("3 сент.", 1885)                // 1885-09-03, YearFromContext
_, ok = dates.Parse("3 сент.", 0)                   // false: no year anywhere
```

The year from the text always wins over `ctxYear`; `YearFromContext` is set
only when the year was substituted. Impossible dates (31 февраля, month 13,
day 0, 29 февраля in a non-leap year), unknown month words and any other
text give `ok=false`. A two-digit year («21.01.84») is rejected as
ambiguous, and numeric dates are always read as day.month[.year]. A «г.» suffix needs a year. Dates are taken as written: no Julian/Gregorian conversion.

**Example:** `ExampleParse` and `ExampleMonth` in `dates/example_test.go`.

**Available since:** 0.5.0.

## Planned

Not available yet.

- **Under consideration** (not scheduled) — agreed normal forms for spans
  without dictionary hits: «Калужская губерния» instead of the lemma
  sequence «калужский губерния», through the inflection of
  [scenario 22](#22-word-forms-and-number-agreement).

When they ship, they get scenarios here with their own "Available since".
