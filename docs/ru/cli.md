# CLI: lexicon

> English version: [docs/en/cli.md](../en/cli.md).

`cmd/lexicon` — небольшая утилита, чтобы попробовать анализ на фразе и
управлять каталогом словарей. Хосты используют библиотеку; CLI — для
людей.

```bash
go install github.com/amarin/lexicon/cmd/lexicon@latest
# или из корня репозитория:
go run ./cmd/lexicon ...
```

```
lexicon analyze [--dicts DIR] [--ortho modern|prereform] [--profile NAME[:KIND[G1|G2],KIND...]] [--mode index|full] TEXT...
lexicon dicts list [--dicts DIR]
lexicon dicts fetch [--dicts DIR] [--force]
```

Флаги идут перед текстом. Коды выхода: 0 — успех, 1 — ошибка, 2 — ошибка
использования.

## Каталог словарей

`--dicts DIR` по умолчанию берётся по порядку из `$LEXICON_DICTS`,
`$XDG_DATA_HOME/lexicon/dicts`, `~/.local/share/lexicon/dicts`. В каталоге
лежат файлы `<kind>.<name>.dat|tsv` с необязательными файлами `.meta`
([сценарии, 10](scenarios.md#10-свои-словари-файлы-и-встроенные)). CLI
принимает любой допустимый вид: своего словаря видов у него нет.

## `dicts fetch` — скачать базовый словарь

Скачивает pymorphy2-dicts-ru (данные OpenCorpora, ~15 МБ, CC BY-SA 4.0),
компилирует и атомарно записывает в каталог `base.opencorpora.dat` и его
`.meta`. Существующий файл сохраняется, если не указан `--force`.

```bash
lexicon dicts fetch
```
```
base dictionary pymorphy2-dicts-ru 2.4.417127.4579844 -> /home/me/.local/share/lexicon/dicts/base.opencorpora.dat
```

## `dicts list` — что загружено

По строке на словарь: имя, вид, формат, включён ли, источник (путь к файлу
или `builtin`), первые 12 шестнадцатеричных цифр хэша содержимого,
происхождение из `.meta`, ошибка загрузки. Затем строка-сводка и
`Registry.Version()`.

```bash
lexicon dicts list
```
```
NAME              KIND  FORMAT  ENABLED  ORIGIN                                  HASH          SOURCE                                                                                                                      ERROR
base.opencorpora  base  dat     true     /home/me/.local/share/lexicon/dicts/…  7aefee101d12  OpenCorpora via pymorphy2-dicts-ru; 2.4.417127.4579844; https://pypi.org/project/pymorphy2-dicts-ru/; license CC BY-SA 4.0
dictionaries: 1 of 1 enabled; base: yes; broken: 0
version: 896617ae5463de52c3d7119f443dffe1455f80755e83dd3a04e18f7c96fc9cef
```

Испорченный файл или файл недопустимого вида выводится со своей ошибкой и
считается сломанным; команду он не останавливает.

## `analyze` — термы текста

Печатает по строке на терм: `raw<TAB>form<TAB>lemmas`, каждая лемма как
`text[flags]` (без флагов — без скобок). Сводка по словарям и
`Analyzer.Version()` идут в stderr.

| Флаг | По умолчанию | Значение |
|---|---|---|
| `--ortho` | `modern` | правила орфографии: `modern`, `prereform` ([сценарий 1](scenarios.md#1-сравнивать-слова-независимо-от-орфографии)) |
| `--profile` | `text` | `NAME[:KIND[G1\|G2],KIND...]`: имя профиля, виды словарей, граммемы для вида; без видов — все словари ([сценарий 6](scenarios.md#6-профиль-на-поле-против-омонимии)) |
| `--mode` | `index` | `index` — термы поискового индекса, `full` — терм на каждый токен ([сценарии 4](scenarios.md#4-термы-для-поискового-индекса), [7](scenarios.md#7-разметить-каждый-токен-вход-для-ner-и-проверки)) |

Флаги лемм: `ambiguous`, `predicted`, `unknown`, `abbrev`, `stop`, `reform`
— см. [сценарий 4](scenarios.md#4-термы-для-поискового-индекса).

Только с базовым словарём OpenCorpora:

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

Без словаря сокращений «кр-нин» угадывается и делится на части, а «с» —
предлог (стоп-слово, отброшено); добавьте файл `abbrev.*.tsv`
([сценарий 9](scenarios.md#9-сокращения-в-записях)), чтобы раскрыть их.

```bash
lexicon analyze --mode full --profile 'name:base[Name|Surn|Patr]' "У Ивана сын Петр"
```
```
У	у	у[stop]
Ивана	ивана	иван
сын	сын	сын[unknown]
Петр	петр	петра[ambiguous] петр[ambiguous]
```

В профиле `name` у «сын» нет разбора-имени, и оно `unknown`, а не
угадано ([сценарий 4](scenarios.md#4-термы-для-поискового-индекса),
«Поведение в 0.1.0»).

## История

- 0.1.0 — `analyze`, `dicts list`, `dicts fetch`.
