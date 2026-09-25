# lexicon — text analysis and dictionary NER over gomorphy: design

Status: **spec, agreed in discussion 2026-09-24, not implemented.**
Companion specs:

- gomorphy — `gomorphy/docs/en/superpowers/specs/2026-09-24-ner-support-design.md`
  (library changes lexicon depends on);
- genodex — `genodex/docs/plans/2026-09-24-ner-lexicon-integration-spec.md`
  (first host: search normalizer + NER over lexicon).

## Context

genodex (personal genealogy store, Go, single binary, SQLite, no external
LLMs) planned three internal packages for its search (P1): `textnorm`
(orthography, tokens), `dicts` (registry of gomorphy dictionaries) and
`normalizer` (text → terms with lemmas). The same normalizer is mandated for
NER (E9), duplicate detection (E10) and matching (F5) — genodex decision
"one normalizer" (`docs/search/index.md`, decision 5). NER itself was
designed as dictionary + rules + morphology, entirely in Go
(`docs/ux/extraction.md`).

The discussion that produced this spec concluded:

1. The text-analysis and NER machinery is reusable beyond genodex (other
   projects with their own entity types and dictionaries; a provider for the
   `nerman` orchestrator). It becomes a separate Go module — **lexicon** —
   with packages `textnorm`, `lexicon`, `gazetteer`, `rules`, `ner`.
2. The P1 packages move into lexicon **together** with NER, so that search and
   NER share one normalizer physically, not by convention. P1 is not
   implemented yet, so the move costs nothing but a plan rewrite.
3. lexicon is a **library, not a service**. HTTP/MCP exposure is the business
   of hosts (genodex) or of `nerman` (via a provider adapter living in nerman).
4. lexicon **marks up text; it does not link text to host data**. The output
   carries opaque references to dictionary records only; linking a span to a
   person/place instance by document context, privacy and storage of
   annotations are host concerns.

## Goals

- One analysis pipeline for index-building, query parsing and NER:
  orthography rules → tokens with offsets → lemmas via gomorphy with
  per-field dictionary profiles.
- Dictionary NER: multi-word aliases matched by lemma sequences and surface
  forms, variant groups, abbreviation hints, trigger-based candidates, rule
  sets switched by document tags, in-text conflict resolution, explainable
  output.
- Dictionaries and rules are **data**, portable between hosts: gomorphy
  `.dat`/TSV for morphology, TSV for gazetteers, YAML for rules, each with a
  provenance manifest.
- Cheap updates: rebuild one gazetteer source in milliseconds, swap
  atomically, no locks on the read path.
- Host-agnostic: no knowledge of genodex models, SQL, auth or privacy.

## Non-goals

- Linking spans to host entities by context (parish, district, document date),
  ranking candidate instances, privacy filtering — host.
- Storage of annotations/mentions, suggestion queues — host.
- Network services, MCP, HTTP — host / nerman.
- ML models, LLMs.
- Fuzzy (typo-tolerant) alias matching — deferred to a later version (see
  "Roadmap"); v1 relies on orthography normalization + morphology.
- Fact extraction beyond what patterns emit as plain data (`Fact`); building
  domain facts (relations, residences) is host logic.

## Module and dependencies

- Module path: `github.com/amarin/lexicon`, public GitHub repository (Q1 resolved).
- Go version (owner decision 2026-09-24): policy "current Go minus two minor
  versions", the same as gomorphy (its spec item F) — `go 1.25.0` with
  `toolchain go1.27.1` now; dependency updates must not raise the `go` line.
- Dependencies: `github.com/amarin/gomorphy` (≥ v1.2.0: `Reading.Predicted`,
  `OpenBytes`, `ContentHash`, consistent case handling), `golang.org/x/text`
  (`unicode/norm`), `go.yaml.in/yaml/v3` (rule files, package `rules` only).
  Nothing else in library packages; tests use stdlib `testing`.
- Optional subpackage `lexicon/basefetch` imports gomorphy's `pkg/pymorphy`
  loader (which pulls `amarin/logging`/zap); it is isolated so that hosts not
  fetching the base dictionary do not compile it.
- No logger dependency: packages return errors and expose state (`List`,
  `Summary`); hosts log.
- Language of code comments, errors, docs: English (library, like gomorphy).
  Test fixtures are Russian text.

## Package map

```
textnorm   orthography rule sets, tokenizer with byte+rune offsets,
           pre-reform ending variants, offset helpers        (no dictionaries)
lexicon    morphology dictionary registry (gomorphy), kinds, profiles,
           Analyzer: text → Terms (index mode / full mode), query parsing
gazetteer  alias entries → compiled matchers per source, variant groups,
           snapshot with atomic swap, TSV source
rules      hints (abbreviation → type), triggers (keyword → candidate),
           patterns (FSM over labelled tokens), rule sets gated by doc tags
ner        Pipeline: Analyzer + Gazetteer + rules → resolved Spans (+ Facts),
           explain, version
cmd/lexicon  CLI: analyze, extract, golden, dicts
nertest    golden-set harness (JSONL cases with optional host context,
           precision/recall per type)
basefetch  optional: download + compile the OpenCorpora base dictionary
```

Dependency direction: `textnorm` ← `lexicon` ← `gazetteer` ← `ner`;
`rules` depends on `textnorm` and `lexicon` types only; `ner` wires all.
Inter-package dependencies are interfaces declared in the consumer
(`deps.go`), matching genodex conventions.

## textnorm

Ported from genodex P1 tasks 1–2 with three changes: rule sets instead of one
hard-coded table, an offset map, and a tokenizer that emits **all** tokens.

### Rule sets

```go
type Rules struct {
    Name    string            // "modern", "prereform"
    Version string            // part of every derived version hash
    Letters map[rune]string   // after lower-casing and NFD: ѣ→е, і→и, ѳ→ф …
    KeepMarks []rune          // precomposed letters kept intact: й
    DropFinalHardSign bool    // «Ивановъ» → «иванов», per hyphen part
    Homoglyphs map[rune]rune  // Latin look-alikes inside Cyrillic words: «Kот» → «кот», «Iоаннъ» (PreReform)
}
var Modern, PreReform Rules
```

- `Modern`: lower-case, NFC, strip combining marks except `й`, `ё→е`, Latin
  homoglyphs inside Cyrillic words (A a B C c E e H K k M O o P p T X x y Y →
  their Cyrillic look-alikes; an OCR/typing issue, so in both rule sets).
- `PreReform`: `Modern` + genodex P1 table (ѣ і ѵ ѳ ѡ ꙋ ѹ ѧ ѯ ѱ, U+2010 → `-`,
  final `ъ`); its homoglyph table adds Latin `i/I` → `і/І` («Iоаннъ»).
- Homoglyph rule: a hyphen-separated part of a word that has at least one
  Cyrillic letter and whose other letters are all homoglyphs is one Cyrillic
  word; `Form` has the homoglyphs replaced, `Raw` and offsets stay as in the
  input. Parts with other Latin letters («Kотw») or without Cyrillic letters
  («XIX») are left alone. A word made only of Latin homoglyphs («MOCKBA»,
  «HOME») has no Cyrillic letter and stays Latin — intended (owner decision
  2026-09-24).
- `Rules.Version` feeds `lexicon.Analyzer.Version()`; changing a table
  requires a new version (hosts reindex).

`NormalizeWord(r Rules, w string) string` and `Orthography(r Rules, s string)
string` keep P1 semantics.

### Tokens and offsets

```go
type Token struct {
    Raw        string   // input[Start:End]
    Form       string   // NormalizeWord(Raw); "" for punctuation/space
    Start, End int      // byte offsets in the input string
    RuneStart, RuneEnd int // code-point offsets in the input string
    Kind       TokenKind  // Word, Number, Punct, Symbol
    Script     Script     // Cyrillic, Latin, Mixed, None
    Case       Case       // Lower, Title, Upper, Mixed
    Dotted     bool       // followed by '.' (abbreviation candidate)
    SentenceEnd bool      // '.', '!', '?', ';', '…' or a blank line ends a sentence here
}
func Tokenize(r Rules, input string) []Token
```

Contract (binding for all packages and hosts):

1. **Offsets refer to the input string exactly as passed.** All internal
   transformations (NFC, letter maps, mark stripping, final `ъ`) are hidden;
   no function returns offsets into a normalized string.
2. Every token carries both byte and code-point offsets, computed in one pass.
   UTF-16 is produced by `textnorm.UTF16Offsets(input, byteOffs...)` for
   JavaScript clients; hosts call it at their API boundary.
3. **Token and span boundaries fall on grapheme-cluster boundaries** of the
   input: combining marks (stress, titlo) belong to the preceding letter; a
   span never splits «и́».
4. `input[t.Start:t.End] == t.Raw` for every token; the same invariant holds
   for `ner.Span`.
5. Hyphenated words are one `Word` token; parts are available through
   `SplitHyphen(t Token) []Token` (offsets preserved). The P1 rule "check the
   abbreviation dictionary first, otherwise index whole + parts" stays in
   `lexicon`.
6. Whitespace is not a token; `Punct` tokens are kept (sentence and clause
   boundaries matter to NER and patterns).

`ReformVariants(form string) []string` — P1 pre-reform ending variants
(`-аго/-яго/-ыя/-ия`), unchanged.

## lexicon

Ported from genodex P1 `dicts` + `normalizer`, generalized: kinds and
profiles are host configuration, state storage is an interface, a full
analysis mode is added.

### Kinds and dictionary files

- `Kind` is a string. The library defines `KindBase` ("base") and
  `KindAbbrev` ("abbrev"); every other kind is host-defined (genodex:
  `surname`, `given`, `patronymic`, `toponym`, `estate`, `title`, `custom`)
  and must match `[a-z][a-z0-9_]*`.
- **Declared kinds** (owner decision 7, 2026-09-24): the host declares its
  kinds in `Options.Kinds`. A dictionary file whose kind prefix is neither
  `base`, `abbrev` nor in `Options.Kinds` is listed in `Registry.List()` with
  the error `unknown kind "<kind>"` and does not take part in parsing (a typo
  such as `surnme.x.tsv` is visible, not silently consulted by no profile).
  Empty/nil `Options.Kinds` accepts any syntactically valid kind (used by the
  lexicon CLI, which has no host vocabulary). A built-in dictionary
  (`Options.Builtin`) with an undeclared kind is a programming error: `Open`
  returns an error.
- File naming in a dictionary directory: `<kind>.<name>.dat` (gomorphy
  binary, opened with `morphology.Open`, mmap) or `<kind>.<name>.tsv`
  (`lemma<TAB>wordform[<TAB>tags]`, built at load with
  `morphology.ImportTSV`). Dictionary name = file name without extension.
- Built-in dictionaries are passed by the host as `Builtin []BuiltinDict{Name,
  Kind, Data []byte, Format}`; the base dictionary may be built in (host
  `//go:embed`, opened with `morphology.OpenBytes`) or a file. A file
  `base.<name>.*` in the directory replaces the built-in base.
- **Provenance manifest** (alienability): optional sidecar `<file>.meta`
  with `key: value` lines — `source`, `license`, `version`, `url`,
  `generated_from` (free text, e.g. "genodex surnames export"). For `.dat`,
  gomorphy `BuildInfo` is used when no sidecar exists. A built-in base
  (`Options.Base` bytes) has no sidecar: the host embeds the `.meta` that
  `basefetch` writes next to the `.dat`, parses it with `ParseManifest` and
  passes it as `Options.BaseManifest`; the `base.builtin` entry merges it over
  `BuildInfo` (non-empty `BaseManifest` fields win, empty ones are filled from
  `BuildInfo`; zero = `BuildInfo` only). `Registry.List()` exposes it; hosts
  use it for attribution (OpenCorpora CC BY-SA).
- A broken file never fails `Open`: it is listed with its error and does not
  take part in parsing (P1 rule "no silent dropping").

### Registry

```go
type StateStore interface { // host persistence of enabled/disabled
    Enabled(ctx context.Context) (map[string]bool, error)
    SetEnabled(ctx context.Context, name string, on bool) error
}
type Options struct {
    Dir          string
    Base         []byte        // built-in base dictionary (gomorphy format) or nil
    BaseManifest Manifest      // provenance of Base (embedded basefetch .meta); wins over BuildInfo
    Builtin      []BuiltinDict // e.g. host abbreviation TSVs
    Kinds        []Kind        // host-declared kinds besides base/abbrev; nil → any valid kind
    State        StateStore    // nil → MemState (everything enabled)
}
func Open(ctx context.Context, o Options) (*Registry, error)
func (r *Registry) List() []Entry          // name, kind, origin, hash, enabled, info, error
func (r *Registry) Version() string        // sha256 over (name, kind, content hash) of enabled
func (r *Registry) Parse(word string, kinds []Kind) []Reading
func (r *Registry) SetEnabled(ctx context.Context, name string, on bool) error
func (r *Registry) Reload(ctx context.Context) error // rescan Dir, rebuild snapshot, swap
func (r *Registry) Summary() string
func (r *Registry) Close() error
```

- `Reading{Normal, Tag string; Kind Kind; Dict string; Predicted bool}`.
- Snapshot of enabled dictionaries is immutable and swapped atomically;
  **hot reload is supported from v1** (P1 deferred it to restart). Retired
  snapshots are closed after in-flight `Parse` calls finish (reference
  counting inside the registry), because gomorphy `Close` unmaps memory.
- Content hash: gomorphy `Dictionary.ContentHash()` (excludes the `info`
  section, so rebuilding identical content keeps the version); TSV → sha256
  of the text.
- Parse rules (from P1): predicted readings only when no exact reading exists
  and only from `KindBase` dictionaries; abbreviation dictionaries never
  predict.

### Profiles

```go
type Profile struct {
    Name  string
    Kinds []Kind                  // dictionaries consulted, in order
    Grammemes map[Kind][]string   // readings of Kind kept only if tag has any of these
}
```

Hosts define their profiles (genodex: `text` = all kinds; `name` = base
filtered by `Name|Surn|Patr` + surname/given/patronymic; `place` = base
filtered by `Geox` + toponym + abbrev). This is the main defence against
homonymy («Вера», «Мороз»). Grammeme checks use gomorphy tag helpers
(`HasGrammeme`), falling back to token matching on the tag string until those
ship.

### Analyzer

```go
type Mode int
const (
    ModeIndex Mode = iota // P1 behaviour: Cyrillic words only, stop-words dropped
    ModeFull              // every token; stop-words flagged, Latin/numbers/punct kept
)
type Flag uint16 // Ambiguous, Predicted, Unknown, Abbrev, Stop, Reform
type Lemma struct { Text string; Flags Flag; Tag string; Kind Kind; Dict string }
type Term  struct { Token textnorm.Token; Form string; Lemmas []Lemma }

func NewAnalyzer(d Dictionaries, r textnorm.Rules, opts AnalyzerOptions) *Analyzer
func (a *Analyzer) Analyze(text string, p Profile, m Mode) []Term
func (a *Analyzer) ParseQuery(q string, p Profile) Query  // P1 task 7, unchanged semantics
func (a *Analyzer) Version() string // rules version + registry version
```

- `ModeFull` exists because NER patterns need service words («у X сын Y»,
  «в»), numbers («N лет»), Latin, punctuation and sentence ends, all of which
  `ModeIndex` drops.
- Abbreviations: dotted/hyphenated/plain forms per P1; one-letter
  abbreviations («с.» = село, сын, сестра, священник) are recognized only with
  a dot and flagged `Ambiguous`; disambiguation is NER's job (hints).
- A per-analyzer lemma cache (bounded LRU keyed by form + profile) is dropped
  on `Version()` change.
- `Dictionaries` is an interface (`Parse(word, kinds)`), so tests run on fakes
  without gomorphy data.

## gazetteer

### Entries and sources

```go
type Entry struct {
    Alias     string            // surface text of the alias, e.g. «дер. Лягушкино», «СПб»
    Type      string            // host entity type: "surname", "settlement", "estate" …
    Ref       string            // opaque host key of the dictionary record; may be ""
    Canonical string            // normal form of the record, e.g. «Лягушкино»
    Attrs     map[string]string // passthrough: since/until of a rename, level, …
    Flags     EntryFlag         // RequiresContext, SurfaceOnly, CaseSensitive, Blocked
}
type Source interface {
    Name() string
    Version(ctx context.Context) (string, error)   // cheap; unchanged → no rebuild
    Entries(ctx context.Context, yield func(Entry) error) error
}
```

- lexicon ships `TSVSource` (file format below) and `SliceSource`.
- Hosts implement `Source` over their storage (genodex: SQLite dictionaries).
- **`Ref` points only to dictionary records** chosen by the host (genodex:
  `Surname`, `GivenName`, `AdministrativeDivision`, `Estate`…), never to
  instance data. lexicon never interprets it.

TSV format (one alias per line, `#` comments, header lines `# key: value`
form the manifest):

```
type<TAB>ref<TAB>canonical<TAB>alias[<TAB>flags[<TAB>k=v;k=v]]
```

### Compilation and matching

- Each alias is analyzed with the `Analyzer` in `ModeFull` using the profile
  configured for its `Type` (`TypeProfiles map[string]lexicon.Profile`).
- Two keys per alias: **lemma key** (sequence of lemma IDs; tokens with several
  lemmas expand to all combinations, capped at 8 per alias, overflow logged in
  the build report) and **surface key** (sequence of normalized forms).
  `SurfaceOnly` entries get only the surface key.
- Lemma and form strings are interned into a process-wide append-only
  interner (`uint32` IDs); an unknown text lemma maps to ID 0, which never
  matches.
- Matcher per source: token trie over IDs (sorted child arrays); matching walks
  from every token, branching over the token's lemma alternatives, bounded by
  the longest alias and by sentence ends. Output: all matches, overlapping
  allowed — selection is `ner`'s job.
- `Snapshot` = immutable set of compiled sources + groups, swapped with
  `atomic.Pointer`. `Gazetteer.Refresh(ctx)` asks every source for `Version`
  and recompiles only the changed ones; `Gazetteer.RefreshSource(ctx, name)`
  forces one.
- Build report per source: entries, aliases, keys, capped expansions, blocked,
  duration, errors (bad lines do not fail the source).

### Variant groups

Four kinds of "variants" were mixed in genodex docs; the gazetteer covers the
ones that are not time- or context-dependent:

| Kind | Example | Where |
|---|---|---|
| Abbreviations | дер. → деревня | `lexicon` abbreviation dictionaries (morphology level) |
| Spellings of one record | Лягушкино / Лягушкина | aliases sharing `Ref` + `Canonical` |
| Name forms (church / folk) | Иоанн ~ Иван, Евдокия ~ Авдотья | aliases sharing a `Ref` (host record or a synthetic `grp:` ref from a data file) |
| Renames with periods | old/new district name | aliases with `Attrs{since, until}`; validity by date is host logic |

- **Group = set of aliases sharing a `(source, Ref)`**; many-to-many is natural
  (an alias may carry several refs).
- API: `Canonical(key) []string` (NER output normal forms) and
  `Expand(lemma string) []string` (search query expansion: all lemma keys of
  every group the lemma belongs to). One implementation for search and NER.
- gomorphy "synonyms" (Stage 20) is **not** used; groups are record relations,
  not morphology.

## rules

Rules are data (YAML via `go.yaml.in/yaml/v3`; since YAML is a JSON superset,
JSON rule files load too), grouped in rule sets:

```go
type RuleSet struct {
    Name     string
    When     []string   // document tags that must all be present; empty = always
    Hints    []Hint
    Triggers []Trigger
    Patterns []Pattern
}
```

- **Hint** — an abbreviation/keyword that sets or boosts the type of an
  adjacent span: `{Lemma: "ул", Dotted: true, Type: "street", Dir: Right,
  Window: 1, Weight: 2}`. Also resolves `RequiresContext` entries and
  disambiguates types («ул. Мира» → street, not the noun «мир»).
- **Trigger** — a keyword that creates a *candidate* span when no gazetteer
  match exists: `{Lemma: "деревня", Type: "settlement", Dir: Right, Window:
  1..3, Shape: {Case: Title, Script: Cyrillic}, StopAt: [Punct, Stop],
  Negative: false, Weight: 1}`. Candidates have no `Ref`; their normal form is
  the lemma sequence of the covered tokens.
- **Pattern** — a sequence over labelled tokens, compiled to an NFA:
  elements match a span type, a lemma, a grammeme, a token kind, or a nested
  group with `?`/`*`/`+`. Actions: `Relabel` (change a span type, e.g. a
  surname-typed «Петров» after a given name in a peasant context →
  patronymic; the previous type with its refs and normal forms stays as a
  `Span.Alternatives` entry with evidence "relabelled by pattern <name>", and
  the relabelled span takes the gazetteer refs of the new type on that range
  when such a record exists, else none), `Boost`, `Emit Fact{Kind, Args
  map[role]spanIndex}`.
  Example rule sets (genealogy) live in the host, not in lexicon.
- **Document tags** (`period:pre1917`, `estate:peasant`, `record:birth`) are
  host-chosen strings passed with the document; rule sets gated by `When` are
  how domain naming rules stay data while depending on document metadata.
- Rule files carry the same manifest as TSVs (top-level `meta:` mapping).

## ner

```go
type Doc struct {
    Text    string
    Profile string   // lexicon profile for free text
    Tags    []string // document tags for rule sets
    Types   []string // optional output filter, applied AFTER resolution
}
type Span struct {
    Start, End         int      // bytes in Doc.Text
    RuneStart, RuneEnd int      // code points in Doc.Text
    Surface            string   // Doc.Text[Start:End]
    Type               string
    Normal             []string // canonical forms; >1 when ambiguous
    Refs               []string // opaque gazetteer refs; empty for trigger candidates
    Attrs              map[string]string
    Flags              SpanFlag // Ambiguous, Predicted, Abbrev, Candidate, Nested
    Score              float32
    Evidence           []string      // only with Explain
    Alternatives       []Alternative // other readings of the same range, best first
}
type Alternative struct {
    Type     string
    Refs     []string
    Normal   []string
    Attrs    map[string]string
    Score    float32
    Evidence []string // only with Explain
}
type Result struct { Spans []Span; Facts []Fact; Version string }

func New(cfg Config) (*Pipeline, error)
func (p *Pipeline) Extract(ctx context.Context, d Doc, o ...Option) (Result, error)
```

Pipeline: `Analyzer.Analyze(ModeFull)` → gazetteer matches → hints →
triggers → patterns → resolution → output filter.

Resolution:

- Filters: blocked entries, `RequiresContext` without a hint, `CaseSensitive`
  mismatch, single-token lemma matches of words shorter than
  `MinLemmaMatchRunes` (default 3) unless surface-matched.
- Score = source weight (surface > lemma > trigger) + length + type weight +
  Σ hint/pattern evidence − ambiguity penalty. Weights are `Config` values.
- Selection: weighted interval scheduling. **Crossing spans are forbidden;
  nesting is allowed only for configured type pairs** (`Config.Nesting
  map[outer][]inner`, e.g. settlement ⊃ district for «дер. Лягушкиной Н-ского
  уезда» when the host models it) — matches genodex `annotations.md`.
- Ambiguity is never silently resolved: several `Refs`/`Normal` + `Ambiguous`.
- Alternatives (owner decision 2026-09-24): when resolution picks a winner
  among candidates on the **same range** with **different types** (a tie or a
  lower score), the losers become `Span.Alternatives` of the winner (type,
  refs, normal forms, attrs, score; evidence only with `Explain`), sorted by
  score desc, then type (deterministic). Only a tie sets `Ambiguous`. A
  pattern `Relabel` keeps the reading it replaced as an alternative too (see
  "rules"). The `Types` filter looks at the winner's type only.
- `Types` filter applies after resolution, so filtering never changes which
  span wins.
- `Explain` option fills `Evidence` ("lemma match «лягушкино» (toponym)",
  "hint «дер.» → settlement +2", "pattern peasant-patronymic relabel").

`Result.Version` = hash(rules version, registry version, gazetteer snapshot
version, rule sets version) — hosts store it with suggestions to know when to
recompute.

`Pipeline` is safe for concurrent use; each call pins one snapshot of every
component.

## CLI and testing

- `lexicon analyze [--dicts DIR] [--rules prereform] [--profile P] [--mode
  index|full] TEXT…` — terms with lemmas and flags.
- `lexicon extract --dicts DIR --gazetteer FILE… --rules FILE… [--tags …]
  [--explain] TEXT…|-` — spans as a table or JSONL.
- `lexicon golden --cases FILE.jsonl …` — runs `nertest`: strict and partial
  precision/recall per type, diff of failures; non-zero exit below thresholds.
  A case's `context` is ignored by the CLI (no `WithTags`).
- `lexicon dicts list|fetch` — registry listing; `fetch` via `basefetch`.
- Unit tests per package on fake dictionaries; integration tests behind a build
  tag with the real base dictionary; `go test -fuzz` for offsets (valid
  grapheme boundaries, `input[Start:End] == Raw`); benchmarks for `Analyze`,
  gazetteer build (100k aliases) and `Extract` (1 KB text).

Performance targets (to be confirmed by benchmarks, not promises): `Extract`
on a 1 KB text under 1 ms with a warm lemma cache; recompiling a 100k-alias
source under 1 s; zero allocations per token in the matcher walk.

## Versioning and releases

- Semver, `v0.x` until genodex search P2 and NER E9 run on it.
- Milestones:
  - **v0.1** — `textnorm`, `lexicon` (port of genodex P1 + full mode, profiles,
    hot reload), `basefetch`, CLI `analyze`/`dicts`. Unblocks genodex search P1.
  - **v0.2** — `gazetteer`, `rules` (hints, triggers), `ner` resolution, CLI
    `extract`, `nertest`/`golden`.
  - **v0.3** — `rules` patterns + facts, rule sets by document tags.
- Development against local gomorphy via a local, uncommitted `go.work`
  (`use . ../gomorphy` or the matching worktree path; `go.work` and
  `go.work.sum` in `.gitignore`), never `replace` in `go.mod` (owner decision
  2026-09-25). `go.mod` always lists real versions: tagged releases, or before
  a release a pseudo-version of a pushed commit (a placeholder only in branch
  commits when nothing is pushed yet). Tagged lexicon releases require released
  gomorphy versions; before merging to `main` (`git merge --ff-only`), build
  and test with `GOWORK=off`.

## Roadmap (not in v0.x)

- Fuzzy alias matching (gomorphy `FuzzyTop` over an alias-word dictionary,
  scored below exact matches).
- Morphology overlays generated from host data (inflected surnames etc.,
  genodex search P7).
- `nerman` provider adapter (lives in nerman).
- A compact line DSL for patterns, if YAML proves verbose.

## Reconciled with the plans (2026-09-24)

Planning (`docs/plans/2026-09-24-v0.1-analysis*.md`, `…-v0.2-ner.md`,
`…-v0.3-patterns.md`) fixed details this spec left open. Where they differ
from the sections above, **these win**:

- Package `lexicon` is the **module root** (`import "github.com/amarin/lexicon"`),
  not `lexicon/lexicon`.
- Constants are prefixed (the unprefixed `Mixed` would collide between
  `Script` and `Case`): `textnorm.TokenWord|TokenNumber|TokenPunct|TokenSymbol`,
  `ScriptNone|ScriptCyrillic|ScriptLatin|ScriptMixed`,
  `CaseNone|CaseLower|CaseTitle|CaseUpper|CaseMixed`;
  `lexicon.FlagAmbiguous|FlagPredicted|FlagUnknown|FlagAbbrev|FlagStop|FlagReform`.
- `lexicon.Dictionaries` is `{Parse(word, kinds) []Reading; Version() string}`
  (version drives the lemma cache and `Analyzer.Version`).
- Latin homoglyphs inside Cyrillic words (owner decision 2026-09-24): a
  hyphen part with a Cyrillic letter whose other letters are all
  `Rules.Homoglyphs` keys is one `ScriptCyrillic` word with the homoglyphs
  replaced in `Form` («Kот» → «кот»; `ModeIndex` deliberately differs from
  genodex P1 here, which indexed «от»). Any other part splits where Cyrillic
  meets non-Cyrillic letters («Kотw» → «K» + «от» + «w»), and parts without
  Cyrillic letters are untouched («XIX-го» → «XIX» + «го»); `SentenceEnd` is set on `. ! ? ; …` and on the last
  token before a blank line or U+2029; a single line break is a space (Q5).
- Dictionary files with an invalid kind prefix or a duplicate name are listed
  with an error (no fallback to `custom`); a missing `Dir` is treated as empty.
- Declared kinds (owner decision 7, 2026-09-24): `Options.Kinds []Kind`. With
  a non-empty list, a file whose kind is not `base`, `abbrev` or declared is
  listed with `unknown kind "<kind>"` and not parsed; a built-in dictionary
  with an undeclared kind makes `Open` fail. Nil/empty `Kinds` accepts any
  syntactically valid kind; the lexicon CLI passes nil. This resolves the
  "kind allow-list" open question of the v0.1 plan.
  TSV forms are converted ё→е at load, so lexicon does not need gomorphy's
  public `CharPolicy`.
- Built-in base provenance (owner decision 2026-09-25):
  `Options.BaseManifest Manifest` is the manifest of `Options.Base` — hosts
  embed the `.meta` written by `basefetch` next to the `.dat` and parse it with
  `ParseManifest`. The `base.builtin` entry uses it when non-zero, merged with
  gomorphy `BuildInfo`: `BaseManifest` fields win, empty ones are filled from
  `BuildInfo`. A `base.*` file that replaces the built-in base keeps its own
  sidecar/`BuildInfo`. The CLI passes no `Base` and is unaffected.
- Gazetteer: `gazetteer.New(ctx, Config{Analyzer, TypeProfiles,
  DefaultProfile, Sources}) (*Gazetteer, error)`; `Refresh(ctx)
  ([]SourceReport, error)`, `RefreshSource(ctx, name) (SourceReport, error)`;
  `NewTSVSource(name, path)` and `NewTSVSourceData(name, data)` (hosts embed
  TSVs). A dot after an abbreviation or an initial does not break a match
  («дер. Лягушкино», «И. Петров»).
- Rules (YAML, see Q2): `rules.Load(io.Reader) (File, error)`,
  `rules.LoadNamed(name string, r io.Reader) (File, error)`,
  `rules.LoadFile(path) (File, error)` → `rules.Compile(files...)
  (*Book, error)`; `Book.Active(tags)` selects rule sets. Source positions
  (owner decision 2026-09-24): loading records the file name and the line of
  every rule set, hint, trigger (v0.2), pattern and element (v0.3) in
  unexported fields (JSON input too); load errors read `<file>:<line>:
  <message>`, every `Compile`/validation error `<file>:<line>: <set>/<rule>:
  <message>` (`Load(r)` names the input `<input>`). In v0.3 `Element.Repeat`
  is a typed `Repeat` (`""`, `"?"`, `"*"`, `"+"`); any other value fails at
  load with its line. Hints and triggers get
  `Absorb` (the span includes the keyword, e.g. «дер.»). `Blocked` vetoes its
  type on that exact range. A trigger creates a candidate only when no
  overlapping candidate of the same type exists; otherwise it boosts.
- ner: `Config{Analyzer, Gazetteer, Rules *rules.Book, Profiles,
  DefaultProfile, Nesting, Weights, MinLemmaMatchRunes}`; score is per word plus
  a length bonus (a flat per-span score would prefer fragments); ties between
  types on one range → deterministic winner + `Ambiguous`; every loser of
  another type on the winner's exact range (tie or lower score) is kept in
  `Span.Alternatives` (owner decision 2026-09-24; v0.2 decision D14).
  `Result.Version` also hashes weights, nesting and min-runes (alternatives add
  no configuration, so they do not change it).
- v0.3 patterns run over the pre-resolution candidate lattice, per sentence,
  leftmost-first, non-overlapping; `Relabel` on a token capture creates a
  candidate span (so «N лет» can be a fact argument); `Fact` gains `Rule`.
  `Relabel` of a span (owner decision 2026-09-24): an existing candidate of the
  new type on the same range wins with its own refs; otherwise the span is
  retyped in place with empty `Refs` and lemma-sequence normal forms. Either
  way the replaced reading (old type, refs, normal forms, attrs) is kept as an
  alternative with evidence "relabelled by pattern <name>" and never sets
  `Ambiguous` («крестьянин … Иван Петров» → patronymic «Петров» with a surname
  alternative carrying the surname ref).
- CLI: orthography flag is `--ortho modern|prereform` in every command
  (`--rules` means rule files in `extract`).
- Golden-set context (owner decision 2026-09-25): `nertest.Case` gains
  `Context map[string]string` (JSON `context`, omitted when empty) — host data
  about the document, ignored by lexicon itself.
  `nertest.Run(ctx, ex, cases, opts ...Option)`; the option
  `nertest.WithTags(func(Case) []string)` appends its result to `Case.Tags`
  for the extraction (the case is not modified), so hosts derive document tags
  from context (`{"estate": "peasant"}` → `estate:peasant` activates rule sets
  gated by that tag). CLI `lexicon golden` ignores `Context`.

## Open questions

- Q5. RESOLVED 2026-09-24: only a blank line (or U+2029) ends a sentence; a
  single line break is a space (hard-wrapped archival text).
- Q6. RESOLVED 2026-09-24: lexicon code is MIT (LICENSE file in the bootstrap task). The OpenCorpora attribution string is "CC BY-SA 4.0" (gomorphy docs). Previously open: the exact
  OpenCorpora licence string for attribution.
- Q1. RESOLVED 2026-09-24: module path `github.com/amarin/lexicon`, public GitHub repository (like gomorphy).
- Q2. RESOLVED 2026-09-24: rule files are YAML (`go.yaml.in/yaml/v3`); JSON
  files still load (YAML superset).
- Q3. RESOLVED 2026-09-24: no data presets in lexicon; all data (abbreviations,
  name forms) lives in hosts. Revisit when a second host appears.
- Q4. Nesting defaults: empty (hosts configure) — or ship a sensible default
  for place types?
- Q7. RESOLVED 2026-09-25 (owner): provenance of an embedded base dictionary
  comes from `Options.BaseManifest` (the host embeds the `basefetch` `.meta`
  and parses it with `ParseManifest`), merged over gomorphy `BuildInfo`
  (v0.1 plan Q11).
- Q8. RESOLVED 2026-09-25 (owner): golden cases carry optional host `context`;
  `nertest.WithTags` maps it to document tags; the CLI ignores it (v0.2 plan
  Q-v02-12).
