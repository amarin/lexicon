# CLI: lexicon

> English version: [docs/en/cli.md](../en/cli.md).

`cmd/lexicon` — небольшая утилита, чтобы попробовать анализ и извлечение
сущностей на фразе, оценить эталонный набор и управлять каталогом
словарей. Хосты используют библиотеку; CLI — для
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
lexicon extract [--dicts DIR] [--ortho modern|prereform] [--gazetteer FILE]... [--rules FILE]... [--nest OUTER>INNER]... [--tags T,...] [--types T,...] [--explain] [--format table|jsonl] TEXT...|-
lexicon golden --cases FILE.jsonl [--dicts DIR] [--ortho modern|prereform] [--gazetteer FILE]... [--rules FILE]... [--nest OUTER>INNER]... [--min-precision N] [--min-recall N]
```

`extract` и `golden` — *(0.2.0)*.

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

## `extract` — спаны сущностей в тексте

*(0.2.0)* Запускает конвейер NER
([сценарии 15–20](scenarios.md#15-найти-сущности-по-словарям)) по каждому
аргументу TEXT или по stdin при `-` (документ на строку, ровно в том виде,
как прочитан, так что смещения отсчитываются от начала строки; пустые
строки пропускаются).

| Флаг | По умолчанию | Значение |
|---|---|---|
| `--dicts` | см. выше | словари морфологии, как у `analyze` |
| `--ortho` | `modern` | правила орфографии: `modern`, `prereform` |
| `--gazetteer FILE` | — | TSV газетира (`type<TAB>ref<TAB>canonical<TAB>alias[<TAB>flags[<TAB>k=v;…]]`); можно повторять; источник называется по базовому имени файла без расширения |
| `--rules FILE` | — | файл правил, YAML или JSON; можно повторять |
| `--nest OUTER>INNER` | — | разрешить спаны типа INNER внутри OUTER; можно повторять |
| `--tags T,...` | — | теги документа, выбирающие наборы правил |
| `--types T,...` | — | печатать только эти типы спанов (после разрешения) |
| `--explain` | выкл. | печатать обоснование каждого спана |
| `--format` | `table` | `table` или `jsonl` |

Плохие строки газетира — предупреждения в stderr; газетир, упавший
целиком, неверный файл правил или плохой `--nest` останавливают команду.

Строка таблицы на спан: номер документа (с 1), байтовые смещения, тип,
текст, нормальные формы и ссылки (через `|`), флаги, оценка. С
`--explain` после спана идут строки обоснования, затем строки `alt` для
типов, проигравших на том же диапазоне.

С базовым словарём OpenCorpora, таким `place.tsv`

```
division	d1	Лягушкино	Лягушкино		level=village
division	d2	Боровский	Боровский		level=uezd
```

и таким `place.yaml` ([сценарий 16](scenarios.md#16-слова-контекста-триггеры-и-теги-документа)):

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

«села Покровское» — кандидат от триггера: записи о нём нет, его
нормальные формы — леммы «Покровское», на настоящей базе их две, отсюда
`ambiguous` и штраф за неоднозначность.

`--format jsonl` печатает по объекту JSON на спан: `doc`, `start`, `end`,
`rune_start`, `rune_end`, `type`, `surface`, `normal`, `refs`, `attrs`,
`flags`, `score`, `evidence`, `alternatives` (пустые поля опускаются).

```bash
printf 'Лягушкино\nиз Боровского уезда\n' | lexicon extract --gazetteer place.tsv --rules place.yaml --format jsonl -
```
```
{"doc":1,"start":0,"end":18,"rune_start":0,"rune_end":9,"type":"division","surface":"Лягушкино","normal":["Лягушкино"],"refs":["d1"],"attrs":{"level":"village"},"score":3}
{"doc":2,"start":5,"end":36,"rune_start":3,"rune_end":19,"type":"division","surface":"Боровского уезда","normal":["Боровский"],"refs":["d2"],"attrs":{"level":"uezd"},"score":6.5}
```

**Поведение в 0.2:** один профиль анализатора, `text` (все включённые
виды словарей), обслуживает и документы, и псевдонимы; два файла
`--gazetteer` с одинаковым базовым именем без расширения (`a/x.tsv`,
`b/x.txt`) падают как дубликат имени источника. С `--explain` колонки
таблицы после блока обоснования могут съезжать.

## `golden` — оценить эталонный набор

*(0.2.0)* Прогоняет конвейер по эталонным примерам
([сценарий 19](scenarios.md#19-измерить-качество-на-эталонном-наборе)) и
печатает строгие (`P`, `R`, `F1`: точные байты и тип) и частичные (`P~`,
`R~`, `F1~`: пересечение) оценки по типам, строгие счётчики и каждый
пропущенный или лишний спан.

| Флаг | По умолчанию | Значение |
|---|---|---|
| `--cases FILE` | — (обязателен) | эталонные примеры, JSON Lines |
| `--min-precision N` | 0 | ошибка, если строгая точность типа ниже N |
| `--min-recall N` | 0 | ошибка, если строгая полнота типа ниже N |
| `--dicts`, `--ortho`, `--gazetteer`, `--rules`, `--nest` | | как у `extract` |

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

Последние две строки идут в stderr, код выхода — 1: годится как проверка
в CI.

**Поведение в 0.2:** `context` примера игнорируется (в CLI нет способа
превратить его в теги; из Go используйте `nertest.WithTags`), а пример с
`profile`, отличным от `text`, проваливает прогон.

## История

- 0.1.0 — `analyze`, `dicts list`, `dicts fetch`.
- 0.2.0 — `extract`, `golden`.
