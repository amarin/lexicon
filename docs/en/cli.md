# CLI: lexicon

> Russian version: [docs/ru/cli.md](../ru/cli.md).

`cmd/lexicon` is a small tool for trying the analysis and entity
extraction on a phrase, scoring a golden set, and managing a dictionary
directory. Hosts use the library; the CLI is for
people.

```bash
go install github.com/amarin/lexicon/cmd/lexicon@latest
# or, from the repository root:
go run ./cmd/lexicon ...
```

```
lexicon analyze [--dicts DIR] [--ortho modern|prereform] [--profile NAME[:KIND[G1|G2],KIND...]] [--mode index|full] TEXT...
lexicon dicts list [--dicts DIR]
lexicon dicts fetch [--dicts DIR] [--force]
lexicon extract [--dicts DIR] [--ortho modern|prereform] [--gazetteer FILE]... [--rules FILE]... [--nest OUTER>INNER]... [--tags T,...] [--types T,...] [--explain] [--format table|jsonl] TEXT...|-
lexicon golden --cases FILE.jsonl [--dicts DIR] [--ortho modern|prereform] [--gazetteer FILE]... [--rules FILE]... [--nest OUTER>INNER]... [--min-precision N] [--min-recall N]
```

`extract` and `golden` are *(0.2.0)*.

Flags go before the text. Exit codes: 0 — success, 1 — failure, 2 — usage
error.

## Dictionary directory

`--dicts DIR` defaults to, in order: `$LEXICON_DICTS`,
`$XDG_DATA_HOME/lexicon/dicts`, `~/.local/share/lexicon/dicts`. The
directory holds `<kind>.<name>.dat|tsv` files with optional `.meta`
sidecars ([scenarios, 10](scenarios.md#10-your-own-dictionaries-files-and-built-ins)).
The CLI accepts every valid kind: it has no host vocabulary.

## `dicts fetch` — download the base dictionary

Downloads pymorphy2-dicts-ru (OpenCorpora data, ~15 MB, CC BY-SA 4.0),
compiles it and writes `base.opencorpora.dat` and its `.meta` sidecar into
the directory atomically. An existing file is kept unless `--force` is
given.

```bash
lexicon dicts fetch
```
```
base dictionary pymorphy2-dicts-ru 2.4.417127.4579844 -> /home/me/.local/share/lexicon/dicts/base.opencorpora.dat
```

## `dicts list` — what is loaded

One row per dictionary: name, kind, format, enabled, origin (file path or
`builtin`), the first 12 hex digits of the content hash, provenance from
the sidecar, load error. Then the summary line and `Registry.Version()`.

```bash
lexicon dicts list
```
```
NAME              KIND  FORMAT  ENABLED  ORIGIN                                  HASH          SOURCE                                                                                                                      ERROR
base.opencorpora  base  dat     true     /home/me/.local/share/lexicon/dicts/…  7aefee101d12  OpenCorpora via pymorphy2-dicts-ru; 2.4.417127.4579844; https://pypi.org/project/pymorphy2-dicts-ru/; license CC BY-SA 4.0
dictionaries: 1 of 1 enabled; base: yes; broken: 0
version: 896617ae5463de52c3d7119f443dffe1455f80755e83dd3a04e18f7c96fc9cef
```

A broken file or a file of an invalid kind is listed with its error and
counted as broken; it never stops the command.

## `analyze` — terms of a text

Prints one line per term: `raw<TAB>form<TAB>lemmas`, each lemma as
`text[flags]` (flags omitted when empty). The dictionary summary and
`Analyzer.Version()` go to stderr.

| Flag | Default | Meaning |
|---|---|---|
| `--ortho` | `modern` | orthography rules: `modern`, `prereform` ([scenario 1](scenarios.md#1-compare-words-across-orthographies)) |
| `--profile` | `text` | `NAME[:KIND[G1\|G2],KIND...]`: profile name, dictionary kinds, grammemes per kind; no kinds = all dictionaries ([scenario 6](scenarios.md#6-a-profile-per-field-against-homonymy)) |
| `--mode` | `index` | `index` — search-index terms, `full` — one term per token ([scenarios 4](scenarios.md#4-terms-for-a-search-index), [7](scenarios.md#7-mark-up-every-token-input-for-ner-and-review)) |

Flags: `ambiguous`, `predicted`, `unknown`, `abbrev`, `stop`, `reform` —
see [scenario 4](scenarios.md#4-terms-for-a-search-index).

With the OpenCorpora base only:

```bash
lexicon analyze "Стали кормить кота"
```
```
Стали	стали	стать[ambiguous] сталь[ambiguous]
кормить	кормить	кормить
кота	кота	кот
dictionaries: 1 of 1 enabled; base: yes; broken: 0; version analyzer-2/modern-3/896617ae…
```

```bash
lexicon analyze --ortho prereform "Кр-нин с. Покровскаго, 1834 г."
```
```
Кр-нин	кр-нин	кр-нин[predicted]
Кр	кр	крый[ambiguous,predicted] кр[ambiguous,predicted] кра[ambiguous,predicted]
нин	нин	нина
Покровскаго	покровскаго	покровский[ambiguous,reform] покровское[ambiguous,reform]
1834	1834	1834[unknown]
г	г	г
```

Without an abbreviation dictionary «кр-нин» is guessed and split, and «с»
is a preposition (a stop word, dropped); add an `abbrev.*.tsv` file
([scenario 9](scenarios.md#9-abbreviations-of-records)) to expand them.

```bash
lexicon analyze --mode full --profile 'name:base[Name|Surn|Patr]' "У Ивана сын Петр"
```
```
У	у	у[stop]
Ивана	ивана	иван
сын	сын	сын[unknown]
Петр	петр	петра[ambiguous] петр[ambiguous]
```

Under the `name` profile «сын» has no name reading and is `unknown`,
not guessed ([scenario 4](scenarios.md#4-terms-for-a-search-index),
"Behaviour in 0.1.0").

## `extract` — entity spans of a text

*(0.2.0)* Runs the NER pipeline
([scenarios 15–20](scenarios.md#15-find-entities-with-dictionaries)) over
each TEXT argument, or over stdin with `-` (one document per line, taken
exactly as read so offsets refer to the line; blank lines skipped).

| Flag | Default | Meaning |
|---|---|---|
| `--dicts` | see above | morphology dictionaries, as for `analyze` |
| `--ortho` | `modern` | orthography rules: `modern`, `prereform` |
| `--gazetteer FILE` | — | a gazetteer TSV (`type<TAB>ref<TAB>canonical<TAB>alias[<TAB>flags[<TAB>k=v;…]]`); repeatable; the source is named after the file's basename without the extension |
| `--rules FILE` | — | a rule file, YAML or JSON; repeatable |
| `--nest OUTER>INNER` | — | allow spans of type INNER inside OUTER; repeatable |
| `--tags T,...` | — | document tags selecting rule sets |
| `--types T,...` | — | print only these span types (applied after resolution) |
| `--explain` | off | print the evidence of every span |
| `--format` | `table` | `table` or `jsonl` |

Bad gazetteer lines are warnings on stderr; a gazetteer that fails as a
whole, an invalid rule file or a bad `--nest` stops the command.

A table row per span: document number (1-based), byte offsets, type,
surface, normal forms and refs (`|`-joined), flags, score. With
`--explain` the evidence lines follow the span, then `alt` lines for the
types that lost on the same range.

With the OpenCorpora base, this `place.tsv`

```
division	d1	Лягушкино	Лягушкино		level=village
division	d2	Боровский	Боровский		level=uezd
```

and this `place.yaml` ([scenario 16](scenarios.md#16-context-words-triggers-and-document-tags)):

```yaml
sets:
  - name: places
    hints:
      - {lemma: деревня, type: division, window: 1, weight: 2, absorb: true}
      - {lemma: уезд, type: division, dir: left, window: 1, weight: 2, absorb: true}
    triggers:
      - {lemma: село, type: division, shape: {case: title, script: cyrillic}, absorb: true}
```

```bash
lexicon extract --gazetteer place.tsv --rules place.yaml "из деревни Лягушкино Боровского уезда и села Покровское"
```
```
DOC  START  END  TYPE      SURFACE            NORMAL                 REFS  FLAGS                SCORE
1    5      38   division  деревни Лягушкино  Лягушкино              d1                         8.50
1    39     70   division  Боровского уезда   Боровский              d2                         6.50
1    74     103  division  села Покровское    покровский|покровское        ambiguous,candidate  3.25
```

«села Покровское» is a trigger candidate: no record knows it, its normal
forms are the lemmas of «Покровское» — two on the real base, hence
`ambiguous` and the ambiguity penalty.

`--format jsonl` prints one JSON object per span: `doc`, `start`, `end`,
`rune_start`, `rune_end`, `type`, `surface`, `normal`, `refs`, `attrs`,
`flags`, `score`, `evidence`, `alternatives` (empty fields omitted).

```bash
printf 'Лягушкино\nиз Боровского уезда\n' | lexicon extract --gazetteer place.tsv --rules place.yaml --format jsonl -
```
```
{"doc":1,"start":0,"end":18,"rune_start":0,"rune_end":9,"type":"division","surface":"Лягушкино","normal":["Лягушкино"],"refs":["d1"],"attrs":{"level":"village"},"score":3}
{"doc":2,"start":5,"end":36,"rune_start":3,"rune_end":19,"type":"division","surface":"Боровского уезда","normal":["Боровский"],"refs":["d2"],"attrs":{"level":"uezd"},"score":6.5}
```

**Behaviour in 0.2:** one analyzer profile, `text` (every enabled
dictionary kind), serves documents and aliases alike; two `--gazetteer`
files with the same basename without extension (`a/x.tsv`, `b/x.txt`) fail
as a duplicate source name. With `--explain` the table columns after an
evidence block may be misaligned.

## `golden` — score a golden set

*(0.2.0)* Runs the pipeline over golden cases
([scenario 19](scenarios.md#19-measure-quality-on-a-golden-set)) and
prints strict (`P`, `R`, `F1`: exact bytes and type) and partial (`P~`,
`R~`, `F1~`: overlap) scores per type, the strict counts, and every missed
or spurious span.

| Flag | Default | Meaning |
|---|---|---|
| `--cases FILE` | — (required) | golden cases, JSON lines |
| `--min-precision N` | 0 | fail when a type's strict precision is below N |
| `--min-recall N` | 0 | fail when a type's strict recall is below N |
| `--dicts`, `--ortho`, `--gazetteer`, `--rules`, `--nest` | | as for `extract` |

```bash
lexicon golden --gazetteer place.tsv --rules place.yaml --cases cases.jsonl --min-precision 0.9
```
```
TYPE      P      R      F1     P~     R~     F1~    TP  FP  FN
division  0.750  1.000  0.857  0.750  1.000  0.857  3   1   0
cases: 2, failures: 1
spurious	bare	division	«Боровский»
golden: division: strict precision 0.750 < 0.900
lexicon: golden: below threshold
```

The last two lines go to stderr, and the exit code is 1 — usable as a CI
gate.

**Behaviour in 0.2:** a case's `context` is ignored (the CLI has no way to
map it to tags; use `nertest.WithTags` from Go), and a case with a
`profile` other than `text` fails the run.

## History

- 0.1.0 — `analyze`, `dicts list`, `dicts fetch`.
- 0.2.0 — `extract`, `golden`.
