# CLI: lexicon

> Russian version: [docs/ru/cli.md](../ru/cli.md).

`cmd/lexicon` is a small tool for trying the analysis on a phrase and
managing a dictionary directory. Hosts use the library; the CLI is for
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
```

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

## History

- 0.1.0 — `analyze`, `dicts list`, `dicts fetch`.
