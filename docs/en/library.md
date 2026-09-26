# Library reference

> Russian version: [docs/ru/library.md](../ru/library.md).

The reference for the public API, package by package. What each part is
for, with runnable examples: [scenarios.md](scenarios.md). The godoc of every
symbol is the final word; this page groups them and states the contracts.

- [textnorm](#textnorm) — orthography and tokens, no dictionaries
- [lexicon](#lexicon) — dictionaries (`Registry`), `Profile`, `Analyzer`
- [basefetch](#basefetch) — download the base dictionary
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

## Contracts

- **Offsets.** Offsets refer to the input exactly as passed: byte and
  code-point offsets on every token, boundaries on grapheme clusters,
  `input[Start:End] == Raw`. Offsets never refer to a normalized string.
- **Versions.** `Analyzer.Version()` = `analyzer-<N>/<rules name>-<rules
  version>/<Registry.Version()>`. Library changes that alter forms or terms
  bump `Rules.Version` or the analyzer version; store the value with
  derived data and rebuild when it changes
  ([scenario 13](scenarios.md#13-know-when-to-reindex)).
- **Concurrency.** `Registry` and `Analyzer` are safe for concurrent use.
  `Parse` never locks; `SetEnabled`, `Reload` and `Close` swap snapshots
  atomically and close old dictionaries after in-flight calls.
- **Files.** Replace dictionary files atomically (temporary file + rename):
  `.dat` files are memory-mapped.
- **Errors, no logging.** Packages return errors and expose state
  (`List`, `Summary`); the host logs.
- **Marking up only.** lexicon never links a span to host data, stores
  annotations or serves HTTP/MCP.
