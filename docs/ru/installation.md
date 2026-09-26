# Установка

> English version: [docs/en/installation.md](../en/installation.md).

## Библиотека

```bash
go get github.com/amarin/lexicon
```

Требования:
- Go 1.25 или новее (директива `go` следует правилу «текущий Go минус две
  минорные версии»).
- [gomorphy](https://github.com/amarin/gomorphy) v1.2.0 или новее —
  подтягивается `go get`.

Что каждый пакет добавляет в ваш бинарник:

| Пакет | Импорты помимо стандартной библиотеки |
|---|---|
| `github.com/amarin/lexicon` (корень), `textnorm` | `gomorphy/pkg/morphology`, `golang.org/x/text` |
| `basefetch` (необязательный) | то же плюс `gomorphy/pkg/pymorphy` и `github.com/amarin/logging` (zap) |
| `gazetteer` *(0.2.0)* | только корневой пакет и `textnorm` |
| `rules`, `ner`, `nertest` *(0.2.0)* | то же плюс `go.yaml.in/yaml/v3` (файлы правил) |
| `cmd/lexicon` | всё перечисленное (использует `basefetch` и `ner`) |

Хост, не импортирующий `basefetch`, не компилирует ни загрузчик pymorphy,
ни zap.

## CLI

```bash
go install github.com/amarin/lexicon/cmd/lexicon@latest
```

См. [cli.md](cli.md).

## Словарные данные

lexicon не поставляет словарных данных. Для общей морфологии русского
языка скачайте базовый словарь OpenCorpora (нужна сеть, ~15 МБ, CC BY-SA
4.0):

```bash
lexicon dicts fetch                      # в ~/.local/share/lexicon/dicts
lexicon dicts fetch --dicts /var/lib/app/dicts
```

или вызовите `basefetch.Fetch(dst)` из Go, или встройте скомпилированный
`.dat` в бинарник через `Options.Base`
([сценарий 11](scenarios.md#11-базовый-словарь-и-его-атрибуция)). Свои
словари — это файлы TSV или `.dat` в том же каталоге или встроенные
словари ([сценарий 10](scenarios.md#10-свои-словари-файлы-и-встроенные)).

Газетиры и файлы правил для извлечения сущностей — тоже данные хоста:
файлы TSV и YAML или источники в памяти
([сценарий 15](scenarios.md#15-найти-сущности-по-словарям)).

Всё, кроме базового словаря, можно попробовать без скачивания:
`go run ./examples/index`, `go run ./examples/ner` и другие
[примеры](../../examples/README.md).

## Платформы

Файлы `.dat` из `Options.Dir` gomorphy отображает в память
(`morphology.Open`), что работает на Unix (Linux, macOS, BSD). На Windows
передавайте байты `.dat` через `Options.Base` или встроенный словарь
`FormatDat`: они открываются из памяти (`morphology.OpenBytes`). Словари
TSV всегда собираются в памяти.

## Атрибуция

Если вы распространяете базовый словарь, укажите атрибуцию OpenCorpora
(CC BY-SA): источник, версия, URL и лицензия есть в `Entry.Manifest`
(`lexicon dicts list`).
