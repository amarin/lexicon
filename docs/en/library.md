# Library reference

> Russian version: [docs/ru/library.md](../ru/library.md).

The reference for the public API, package by package. What each part is
for, with runnable examples: [scenarios.md](scenarios.md). The godoc of every
symbol is the final word; this page groups them and states the contracts.

- [textnorm](#textnorm) — orthography and tokens, no dictionaries
- [lexicon](#lexicon) — dictionaries (`Registry`), `Profile`, `Analyzer`
- [basefetch](#basefetch) — download the base dictionary
- [gazetteer](#gazetteer) — host records compiled for matching *(0.2.0)*
- [rules](#rules) — hints, triggers, rule sets by document tags *(0.2.0)*
- [ner](#ner) — the extraction pipeline *(0.2.0)*
- [nertest](#nertest) — golden-set scoring *(0.2.0)*
- [Contracts](#contracts) — offsets, versions, concurrency, errors

## textnorm

`import "github.com/amarin/lexicon/textnorm"`

### Rule sets

```go
type Rules struct {
    Name, Version     string
    Letters           map[rune]string // after lower-casing and decomposition: ѣ→е, ѯ→кс
    KeepMarks         []rune          // precomposed letters kept intact: й
    DropFinalHardSign bool            // NormalizeWord drops a final ъ of every hyphen part
    Homoglyphs        map[rune]rune   // Latin look-alikes inside Cyrillic word parts
}
var Modern, PreReform Rules
```

`Modern` and `PreReform` must not be modified; build a new `Rules` value
with its own `Name` and `Version` to customise. Any table change needs a
new `Version` (it is part of `Analyzer.Version()`).

| Function | Result |
|---|---|
| `Orthography(r, s) string` | `s` for comparison: NFC, homoglyphs replaced, lower case, marks dropped (except `KeepMarks`), `Letters` applied; the final ъ stays |
| `NormalizeWord(r, w) string` | `Orthography` plus the final ъ rule: the form a word is indexed by |
| `ReformVariants(form) []string` | modern spellings of a pre-reform adjective ending in the order to try; nil when the ending is not pre-reform |

### Tokens

```go
func Tokenize(r Rules, input string) []Token
type Token struct {
    Raw         string    // input[Start:End]
    Form        string    // words: NormalizeWord(Raw); numbers: the digits; "" otherwise
    Start, End  int       // byte offsets in the input
    RuneStart   int       // code-point offsets
    RuneEnd     int
    Kind        TokenKind // TokenWord, TokenNumber, TokenPunct, TokenSymbol
    Script      Script    // words only: ScriptCyrillic, ScriptLatin, …
    Case        Case      // words only: CaseLower, CaseTitle, CaseUpper, …
    Dotted      bool      // word directly followed by '.'
    SentenceEnd bool      // '.', '!', '?', ';', '…', or last token before a blank line / U+2029
}
func SplitHyphen(t Token) []Token             // parts of a hyphenated word; nil otherwise
func UTF16Offsets(input string, byteOffs ...int) []int
```

Every token of the input is returned — words of any script, numbers,
punctuation, symbols. Token boundaries fall on grapheme clusters.
`UTF16Offsets` panics on an offset outside `[0, len(input)]`.

## lexicon

`import "github.com/amarin/lexicon"`

### Kinds and dictionary files

`Kind` is a string matching `[a-z][a-z0-9_]*`. The library defines
`KindBase` ("base": general morphology, the only kind whose predictions
are used) and `KindAbbrev` ("abbrev": abbreviation → full word); every
other kind is host-defined and declared in `Options.Kinds`.

A dictionary is named `<kind>.<name>`. In a directory it is a file
`<kind>.<name>.dat` (gomorphy binary, memory-mapped) or
`<kind>.<name>.tsv` (`lemma<TAB>wordform[<TAB>tags]`, built at load);
`ManifestPath(path)` is its optional provenance sidecar `<file>.meta`.

### Opening a registry

```go
func Open(ctx context.Context, o Options) (*Registry, error)

type Options struct {
    Dir          string        // <kind>.<name>.dat|tsv files; "" or missing = none; never created
    Base         []byte        // built-in base in gomorphy binary format, registered as BaseBuiltinName
    BaseManifest Manifest      // provenance of Base (the .meta basefetch writes, via ParseManifest)
    Builtin      []BuiltinDict // host dictionaries, registered after the base, in order
    Kinds        []Kind        // host kinds besides base and abbrev; nil = any valid kind
    State        StateStore    // enabled/disabled persistence; nil = in memory, all enabled
}

type BuiltinDict struct {
    Name     string // "<kind>.<name>"
    Kind     Kind   // "" = from the name
    Format   Format // FormatDat or FormatTSV
    Data     []byte // FormatDat data is used in place: do not modify it while the registry lives
    Manifest Manifest
}
```

Registry order: `base.*` files (or the built-in base), built-ins in order
(a file with the same name replaces a built-in in place), remaining files
by name. `Open` fails only on a state-store error, an unreadable directory,
an invalid `Options.Kinds` entry, or a built-in of an undeclared kind (a
programming error). A broken file, an undeclared kind, a duplicate name or
a dangling symlink is listed with `Entry.Error` and does not take part in
parsing.

### Registry

| Method | |
|---|---|
| `Parse(word, kinds) []Reading` | readings from enabled dictionaries of `kinds` (empty = all), in registry order: exact ones, or — when there are none — predictions of `base` dictionaries |
| `List() []Entry` | copies of all entries, broken ones included |
| `Summary() string` | one line for host logs: `dictionaries: N of M enabled; base: yes; broken: K` |
| `Version() string` | identifies the enabled set and its contents |
| `SetEnabled(ctx, name, on) error` | store the state, swap in a new snapshot; `ErrUnknownDictionary` for an unknown name |
| `Reload(ctx) error` | rescan `Dir`, reopen everything, re-read the state, swap |
| `Close() error` | release dictionaries once in-flight calls finish; idempotent. Afterwards `Parse` returns nil, `List` is empty, `SetEnabled` and `Reload` fail with `ErrClosed` |

```go
type Entry struct {
    Name     string
    Kind     Kind
    Format   Format
    Origin   string   // OriginBuiltin or the file path
    Hash     string   // content hash, hex
    Enabled  bool
    Manifest Manifest // Source, License, Version, URL, GeneratedFrom, Extra
    Error    string   // load error; "" when loaded
}
type Reading struct {
    Normal, Tag string // lemma as stored (may contain ё); gomorphy tag, part of speech first
    Kind        Kind
    Dict        string
    Predicted   bool
}
type StateStore interface {
    Enabled(ctx context.Context) (map[string]bool, error) // no record = enabled
    SetEnabled(ctx context.Context, name string, on bool) error
}
```

`MemState` is the in-memory `StateStore`. `ParseManifest(r)` reads a
sidecar (`key: value` lines, `#` comments, case-insensitive keys);
`Manifest.WriteTo` writes one; `Manifest.String()` is a one-line summary.

### Profiles

```go
type Profile struct {
    Name      string            // cache key; unique per Analyzer; "" disables caching
    Kinds     []Kind            // dictionaries consulted; empty = all
    Grammemes map[Kind][]string // per kind: keep a reading only if its tag has one of these
}
```

Profiles are host configuration and are not part of `Analyzer.Version()`.

### Analyzer

```go
func NewAnalyzer(d Dictionaries, r textnorm.Rules, opts AnalyzerOptions) *Analyzer
type AnalyzerOptions struct{ CacheSize int } // 0 = DefaultCacheSize (50 000), negative = no cache

func (a *Analyzer) Analyze(text string, p Profile, m Mode) []Term // ModeIndex or ModeFull
func (a *Analyzer) ParseQuery(q string, p Profile) Query
func (a *Analyzer) Version() string

type Dictionaries interface { // *Registry implements it; tests can pass a fake
    Parse(word string, kinds []Kind) []Reading
    Version() string
}
type Term struct {
    Token  textnorm.Token
    Form   string  // the form looked up; «с.» with its dot for a dotted abbreviation
    Lemmas []Lemma // none for punctuation and symbols
}
type Lemma struct {
    Text  string // normalized like words (textnorm.NormalizeWord)
    Flags Flag   // FlagAmbiguous, FlagPredicted, FlagUnknown, FlagAbbrev, FlagStop, FlagReform
    Tag   string // tag of the first reading that gave the lemma
    Kind  Kind   // "" for unknown lemmas
    Dict  string
}
type Query struct {
    Terms   []Term // complete words, as ModeIndex
    Partial *Term  // the last word when nothing follows it; nil otherwise
}
```

Lookup of one word: exact readings (then pre-reform ending variants, flag
`reform`); if all exact readings are service parts of speech (PREP, CONJ,
PRCL, INTJ; `Abbr` readings ignored) — a stop word; otherwise the profile
filters the readings; with no exact readings at all — base predictions
(`predicted`); nothing — the form itself (`unknown`). Several distinct
lemmas are all `ambiguous`. Dotted words are first looked up with their dot
in `abbrev` dictionaries, hyphenated words first whole; abbreviation
lookups ignore the profile's kinds.

`ModeIndex`: Cyrillic words and numbers only, stop words dropped, a
hyphenated word whole and by parts. `ModeFull`: exactly one term per token,
stop words kept with `FlagStop`, numbers and non-Cyrillic words `unknown`
without lookup, punctuation without lemmas. `Flag.String()` gives
`ambiguous,predicted,unknown,abbrev,stop,reform` names.

## basefetch

`import "github.com/amarin/lexicon/basefetch"`

```go
const DefaultName = "base.opencorpora.dat"
func Fetch(dst string) (version string, err error)
```

Downloads pymorphy2-dicts-ru, compiles it, and writes `dst` and its sidecar
`lexicon.ManifestPath(dst)` atomically (temporary file + rename). Needs
network access. This package pulls gomorphy's pymorphy loader and zap;
import it only where you fetch.

## gazetteer

`import "github.com/amarin/lexicon/gazetteer"` — *(0.2.0)*

Compiles aliases of host records into token tries and matches them against
analyzed text. It reports every match, overlapping ones included; choosing
between them is `ner`'s job. Imports only the root package and `textnorm`.

### Entries and sources

```go
type Entry struct {
    Alias     string            // surface text: «дер. Лягушкино», «СПб»
    Type      string            // host entity type: "surname", "division", …
    Ref       string            // opaque host key; aliases sharing a Ref are variants of one record
    Canonical string            // normal form of the record; Alias when empty
    Attrs     map[string]string // passthrough to spans
    Flags     EntryFlag         // RequiresContext, SurfaceOnly, CaseSensitive, Blocked
}
type Source interface {
    Name() string                                                // unique within a Gazetteer
    Version(ctx context.Context) (string, error)                 // cheap; unchanged = no recompilation
    Entries(ctx context.Context, yield func(Entry) error) error // stops at the first yield error
}
func NewTSVSource(name, path string) *TSVSource        // reads path on every Version/Entries call
func NewTSVSourceData(name string, data []byte) *TSVSource // in-memory, e.g. //go:embed
func NewSliceSource(name, version string, entries []Entry) *SliceSource
```

TSV: `type<TAB>ref<TAB>canonical<TAB>alias[<TAB>flags[<TAB>k=v;k=v]]`,
flags comma-separated (`requires_context,surface_only,case_sensitive,blocked`,
`ParseEntryFlags`); `#` lines are comments, `# key: value` lines (keys
`[a-z_]`) before the first entry form the manifest (`TSVSource.Manifest()`).
A line has 4 to 6 fields, each trimmed. Its `Version` is the sha256 of
the content — `NewTSVSource` reads and hashes the whole file on every call.
A bad line is skipped and returned as `*ParseErrors` after every good
entry — the builder puts it in the report; a line longer than 1 MiB fails
the whole source (`SourceReport.Err`). `Manifest()` is empty until the
first `Entries` call.

| Flag | Effect |
|---|---|
| `RequiresContext` | kept only when a hint or trigger supports the match |
| `SurfaceOnly` | no lemma key: matched by normalized form only |
| `CaseSensitive` | kept only when every word has the alias's letter case |
| `Blocked` | vetoes every match of the entry's type on the matched range, whatever its other flags |

### Gazetteer

```go
func New(ctx context.Context, cfg Config) (*Gazetteer, error)
type Config struct {
    Analyzer       Analyzer                   // *lexicon.Analyzer; its Version is part of every compiled source
    TypeProfiles   map[string]lexicon.Profile // profile per entry type
    DefaultProfile lexicon.Profile            // for other types
    Sources        []Source                   // compiled in order; unique non-empty names
}
```

| Method | |
|---|---|
| `Snapshot() *Snapshot` | the current compiled state, lock-free; hold it for one operation |
| `Refresh(ctx) ([]SourceReport, error)` | recompile the sources whose `Version` (or the analyzer version) changed; swap atomically |
| `RefreshSource(ctx, name) (SourceReport, error)` | recompile one source regardless of its version; `ErrUnknownSource` |
| `Canonical(key)`, `Expand(lemma)` | delegate to the current snapshot |

`New` fails on an invalid configuration or a cancelled `ctx` (it runs the
first `Refresh`); a failing source is in `Snapshot().Reports()`.
`Refresh` and `RefreshSource` are serialized, so a `Source` is never called
concurrently by one `Gazetteer`. `Refresh` also retries every source that
has never built; it returns reports of the sources it rebuilt or whose
`Version` failed. Each alias is analysed (`ModeFull`) with its
type's profile into a surface key and lemma keys — an ambiguous word
expands into combinations, at most `MaxLemmaKeys` (8) per alias.
`SourceReport` counts entries, aliases, lemma and surface keys, capped and
blocked aliases, skipped empty ones, the duration and non-fatal `Errors`;
`Err` is a fatal failure, after which the source keeps its previous data.

### Snapshot and matching

| Method | |
|---|---|
| `Match(tx *Text, out []Match) []Match` | append every alias match in `tx`; allocates only to grow `out` |
| `Version() string` | changes when a source is rebuilt from a new `Version` or with a new analyzer version (a forced `RefreshSource` of unchanged content keeps it) |
| `Reports() []SourceReport` | the last report of every source, in configuration order |
| `Canonical(key) []string` | canonical forms of the aliases whose lemma or surface key (space-joined) is `key` |
| `Expand(lemma) []string` | lemma keys of every variant group containing `lemma` — for query expansion |

Variant groups are aliases of one source sharing a `Ref`; aliases with an
empty `Ref` or without lemma keys (`SurfaceOnly`) are not in any group.
`Blocked` aliases are, and an ambiguous alias contributes every lemma
(`Expand("сталь")` → `[сталь стать]`). Keys are lower-case normalized
strings. `ner` does not use the groups: span `Normal` comes from
`Entry.Canonical`.

`Prepare(terms)` builds a `Text` from `Analyzer.Analyze(ModeFull)` terms.
A `Match` is content positions `[Start, End)` (words and numbers; map back
with `Text.TermIndex`), its `Kind` (`ByLemma`, `BySurface`) and the
`Aliases` whose key ended there. A match covers consecutive words only:
any punctuation or symbol between words breaks it (an abbreviation's own
dot does not).
Aliases and snapshots are shared and read-only.

## rules

`import "github.com/amarin/lexicon/rules"` — *(0.2.0)*

Rules as data, validated into an immutable `Book`. Imports
`go.yaml.in/yaml/v3`.

```go
func Load(r io.Reader) (File, error)              // errors say "<input>"
func LoadNamed(name string, r io.Reader) (File, error)
func LoadFile(path string) (File, error)          // .yaml, .yml or .json
func Compile(files ...File) (*Book, error)        // set names unique across files

type File struct {
    Meta map[string]string // `meta:` — source, license, version, url, generated_from
    Sets []RuleSet         // `sets:`
}
type RuleSet struct {
    Name     string    // `name:`
    When     []string  // `when:` — active when the document has ALL these tags; empty = always
    Hints    []Hint
    Triggers []Trigger
}
```

One YAML document per file (JSON is valid YAML); unknown fields are
errors. Rule errors read `<file>:<line>: <message>`, `Compile` errors
`<file>:<line>: <set>/<rule>: <message>` (`hint 0`, `trigger 1`); a file
that cannot be opened (`rules: open …`) or is empty (`<file>: empty rule
file`) has no line.

| Hint field | YAML | Default | |
|---|---|---|---|
| `Lemma` | `lemma` | — | keyword lemma or form, alternatives with `\|` |
| `Dotted` | `dotted` | false | keyword must be followed by «.» |
| `Type` | `type` | — | span type it supports |
| `Dir` | `dir` | `right` | `right`, `left`, `both` |
| `Window` | `window` | 1 | distance in content words from the keyword to the span's edge, ≤ `MaxWindow` (8); 1 = right next to it |
| `Weight` | `weight` | 1 | added to the span's score |
| `Absorb` | `absorb` | false | extend an adjacent span over the keyword |

Hints have no `shape` and never affect trigger candidates (they run
first).

A `Trigger` has the same `lemma`, `dotted`, `type`, `weight`, `absorb`,
plus `dir` (`right` or `left`), `window` as a range (`N` = 1..N,
`"min..max"`), `shape` (`case`: `lower`/`title`/`upper`, `script`:
`cyrillic`/`latin`) and `stop_at` (`punct` implied, `stop`, `number`,
`latin`) restricting the words it covers, and `negative`. A trigger
proposes a `Candidate` over the covered words when no gazetteer span of
its type overlaps them, and otherwise boosts those spans; its `weight` is
added in both cases and `absorb` extends both over an adjacent keyword. A
`negative` trigger only subtracts its weight from overlapping gazetteer
spans. No candidate is proposed over a range overlapping a `Blocked` match
of the type.

| `Book` method | |
|---|---|
| `Active(tags) Active` | hints and triggers of the sets whose `When` tags are all in `tags`, in file and set order; shared, do not modify |
| `Sets() []string` | set names in order |
| `Version() string` | changes when any rule or meta value changes |

Keywords are compared after `textnorm.NormalizeWord(textnorm.PreReform, …)`
and without a trailing dot, whatever orthography the document uses.

## ner

`import "github.com/amarin/lexicon/ner"` — *(0.2.0)*

```go
func New(cfg Config) (*Pipeline, error)
type Config struct {
    Analyzer           Analyzer                   // *lexicon.Analyzer
    Gazetteer          Gazetteer                  // *gazetteer.Gazetteer
    Rules              *rules.Book                // nil: no rules
    Profiles           map[string]lexicon.Profile // Doc.Profile → analyzer profile
    DefaultProfile     string
    Nesting            map[string][]string        // outer type → inner types allowed inside it
    Weights            Weights                    // DefaultWeights() when Surface, Lemma, Trigger are all 0
    MinLemmaMatchRunes int                        // drop one-word lemma-only matches of shorter aliases; 0 → 3
}
func (p *Pipeline) Extract(ctx context.Context, d Doc, opts ...Option) (Result, error)
func Explain() Option // fill Span.Evidence

type Doc struct {
    Text    string
    Profile string   // Config.DefaultProfile when empty; unknown → ErrUnknownProfile
    Tags    []string // select rule sets
    Types   []string // output filter, applied after resolution
}
type Result struct {
    Spans   []Span // by Start, then longer first, then Type
    Version string // extractor, analyzer, gazetteer snapshot, rules, configuration
}
type Span struct {
    Start, End         int // bytes in Doc.Text; Doc.Text[Start:End] == Surface
    RuneStart, RuneEnd int // code points
    Surface, Type      string
    Normal             []string          // canonical forms; >1 when ambiguous
    Refs               []string          // gazetteer refs; empty for a candidate
    Attrs              map[string]string // entry attributes, conflicting keys dropped
    Flags              SpanFlag          // Ambiguous, Predicted, Abbrev, Candidate, Nested
    Score              float32
    Evidence           []string          // with Explain
    Alternatives       []Alternative     // losing types on the same range, best first
}
```

`Extract` runs: analysis (`ModeFull`) → gazetteer matches → filters
(`Blocked`, `CaseSensitive`, short lemma matches) → hints → triggers →
`RequiresContext` filter → scoring → weighted interval scheduling (no
crossing spans; nesting only for `Config.Nesting` pairs) → the
`Doc.Types` filter. Score = (origin weight + `Types[type]`) × words +
`LengthBonus` × (words − 1) + hint/trigger evidence − `AmbiguityPenalty` ×
(max(distinct refs, distinct normal forms, 1) − 1), rounded to 1e-6;
candidates scoring 0 or below are dropped before resolution.
`DefaultWeights()` is surface 3, lemma 2, trigger 1, length bonus 0.5,
ambiguity penalty 0.25; it replaces `Config.Weights` only when `Surface`,
`Lemma` and `Trigger` are all 0 (keeping `Types`) — set one origin weight
and you set them all. A negative `MinLemmaMatchRunes` disables the short
lemma filter. An exact tie between types marks the span `Ambiguous` and
goes to the type name. `Span.Flags`: `Ambiguous` — several refs or normal
forms, a type tie, or a covered ambiguous abbreviation no rule absorbed
(homonymy of an ordinary word does not count); `Predicted` — a covered
word known only by predicted lemmas (not set on surface matches); `Abbrev`,
`Candidate`, `Nested`. A trigger candidate's `Normal` is the lower-case
lemma sequences of its words (the absorbed keyword excluded, at most
`gazetteer.MaxLemmaKeys`).

`New` returns an error for a nil `Analyzer` or `Gazetteer` and for a
`DefaultProfile` missing from `Profiles`; it copies the host's maps. `Extract` is linear in the document size and
checks `ctx` between stages.

## nertest

`import "github.com/amarin/lexicon/nertest"` — *(0.2.0)*

```go
type Case struct {
    ID      string            `json:"id,omitempty"`      // "line<N>" when missing
    Text    string            `json:"text"`
    Profile string            `json:"profile,omitempty"`
    Tags    []string          `json:"tags,omitempty"`
    Spans   []Gold            `json:"spans"`
    Context map[string]string `json:"context,omitempty"` // host data, ignored unless WithTags
}
type Gold struct {
    Text       string `json:"text"`
    Type       string `json:"type"`
    Occurrence int    `json:"occurrence,omitempty"` // 1-based, default 1
}
func LoadCases(r io.Reader) ([]Case, error) // JSON lines; blank lines skipped; unknown fields are errors
func LoadCasesFile(path string) ([]Case, error)
func Run(ctx context.Context, ex Extractor, cases []Case, opts ...Option) (*Report, error)
func WithTags(f func(Case) []string) Option // extra tags for the extraction only
```

`Extractor` is anything with `ner.Pipeline`'s `Extract`. `Report` has
`Cases`, `Types map[string]*TypeScore` (`Strict` — exact bytes and type,
`Partial` — overlap and type; each `Counts{TP, FP, FN}` with
`Precision()`, `Recall()`, `F1()`) and `Failures` (strict `missed` and
`spurious` spans). `Report.Write(w)` prints the table and failures;
`Report.Check(minPrecision, minRecall)` returns one message per failed
threshold, e.g. `surname: strict recall 0.500 < 0.900`.

`Run` scores every output span, `Nested` and `Candidate` ones included;
it cannot set `Doc.Types` and stops at the first `Extract` error.
`LoadCases` also rejects gold spans with empty text or type, and lines
over 4 MiB.

## Contracts

- **Offsets.** Offsets refer to the input exactly as passed: byte and
  code-point offsets on every token, boundaries on grapheme clusters,
  `input[Start:End] == Raw`. Offsets never refer to a normalized string.
- **Versions.** `Analyzer.Version()` = `analyzer-<N>/<rules name>-<rules
  version>/<Registry.Version()>`. Library changes that alter forms or terms
  bump `Rules.Version` or the analyzer version; store the value with
  derived data and rebuild when it changes
  ([scenario 13](scenarios.md#13-know-when-to-reindex)).
  `ner.Result.Version` does the same for extracted spans: it covers the
  extractor version, `Analyzer.Version()`, the gazetteer
  `Snapshot.Version()`, `Book.Version()` and the pipeline configuration.
- **Concurrency.** `Registry` and `Analyzer` are safe for concurrent use.
  `Parse` never locks; `SetEnabled`, `Reload` and `Close` swap snapshots
  atomically and close old dictionaries after in-flight calls.
  `Gazetteer`, `Snapshot`, `Book` and `Pipeline` are safe for concurrent
  use too: `Refresh` swaps gazetteer snapshots atomically, and each
  `Extract` pins one snapshot (only the snapshot — the analyzer and its
  registry stay live).
- **Files.** Replace dictionary files atomically (temporary file + rename):
  `.dat` files are memory-mapped.
- **Errors, no logging.** Packages return errors and expose state
  (`List`, `Summary`); the host logs.
- **Marking up only.** lexicon never links a span to host data, stores
  annotations or serves HTTP/MCP.
