# Справочник по библиотеке

> English version: [docs/en/library.md](../en/library.md).

Справочник по публичному API, по пакетам. Зачем нужна каждая часть и
работающие примеры — в [scenarios.md](scenarios.md). Последнее слово за
godoc каждого символа; эта страница группирует их и фиксирует контракты.

- [textnorm](#textnorm) — орфография и токены, без словарей
- [lexicon](#lexicon) — словари (`Registry`), `Profile`, `Analyzer`
- [basefetch](#basefetch) — скачивание базового словаря
- [Контракты](#контракты) — смещения, версии, конкурентность, ошибки

## textnorm

`import "github.com/amarin/lexicon/textnorm"`

### Наборы правил

```go
type Rules struct {
    Name, Version     string
    Letters           map[rune]string // после нижнего регистра и декомпозиции: ѣ→е, ѯ→кс
    KeepMarks         []rune          // составные буквы, которые не разбираются: й
    DropFinalHardSign bool            // NormalizeWord удаляет конечный ъ каждой части через дефис
    Homoglyphs        map[rune]rune   // похожие латинские буквы в кириллических частях слова
}
var Modern, PreReform Rules
```

`Modern` и `PreReform` менять нельзя; для своих правил создайте новое
значение `Rules` со своими `Name` и `Version`. Любое изменение таблицы
требует новой `Version` (она входит в `Analyzer.Version()`).

| Функция | Результат |
|---|---|
| `Orthography(r, s) string` | `s` для сравнения: NFC, гомоглифы заменены, нижний регистр, диакритика удалена (кроме `KeepMarks`), применены `Letters`; конечный ъ остаётся |
| `NormalizeWord(r, w) string` | `Orthography` плюс правило конечного ъ: форма, по которой слово индексируется |
| `ReformVariants(form) []string` | современные написания дореформенного окончания прилагательного в порядке перебора; nil, если окончание не дореформенное |

### Токены

```go
func Tokenize(r Rules, input string) []Token
type Token struct {
    Raw         string    // input[Start:End]
    Form        string    // слова: NormalizeWord(Raw); числа: цифры; иначе ""
    Start, End  int       // байтовые смещения во входе
    RuneStart   int       // смещения в кодовых точках
    RuneEnd     int
    Kind        TokenKind // TokenWord, TokenNumber, TokenPunct, TokenSymbol
    Script      Script    // только слова: ScriptCyrillic, ScriptLatin, …
    Case        Case      // только слова: CaseLower, CaseTitle, CaseUpper, …
    Dotted      bool      // сразу за словом идёт '.'
    SentenceEnd bool      // '.', '!', '?', ';', '…' или последний токен перед пустой строкой / U+2029
}
func SplitHyphen(t Token) []Token             // части слова через дефис; иначе nil
func UTF16Offsets(input string, byteOffs ...int) []int
```

Возвращаются все токены входа — слова любой письменности, числа,
пунктуация, символы. Границы токенов проходят по графемным кластерам.
`UTF16Offsets` паникует на смещении вне `[0, len(input)]`.

## lexicon

`import "github.com/amarin/lexicon"`

### Виды и файлы словарей

`Kind` — строка вида `[a-z][a-z0-9_]*`. Библиотека определяет `KindBase`
("base": общая морфология, единственный вид, чьи предсказания
используются) и `KindAbbrev` ("abbrev": сокращение → полное слово); все
остальные виды задаёт хост и объявляет в `Options.Kinds`.

Словарь называется `<kind>.<name>`. В каталоге это файл `<kind>.<name>.dat`
(бинарный формат gomorphy, отображается в память) или `<kind>.<name>.tsv`
(`lemma<TAB>wordform[<TAB>tags]`, собирается при загрузке);
`ManifestPath(path)` — путь его необязательного файла происхождения
`<file>.meta`.

### Открытие реестра

```go
func Open(ctx context.Context, o Options) (*Registry, error)

type Options struct {
    Dir          string        // файлы <kind>.<name>.dat|tsv; "" или нет каталога — нет файлов; не создаётся
    Base         []byte        // встроенная база в бинарном формате gomorphy, регистрируется как BaseBuiltinName
    BaseManifest Manifest      // происхождение Base (.meta от basefetch, через ParseManifest)
    Builtin      []BuiltinDict // словари хоста, регистрируются после базы по порядку
    Kinds        []Kind        // виды хоста помимо base и abbrev; nil — любой допустимый вид
    State        StateStore    // хранение вкл/выкл; nil — в памяти, всё включено
}

type BuiltinDict struct {
    Name     string // "<kind>.<name>"
    Kind     Kind   // "" — взять из имени
    Format   Format // FormatDat или FormatTSV
    Data     []byte // данные FormatDat используются на месте: не меняйте их, пока жив реестр
    Manifest Manifest
}
```

Порядок реестра: файлы `base.*` (или встроенная база), встроенные словари
по порядку (файл с тем же именем заменяет встроенный на его месте),
остальные файлы по имени. `Open` падает только при ошибке хранилища
состояния, нечитаемом каталоге, недопустимом элементе `Options.Kinds` или
встроенном словаре необъявленного вида (ошибка программиста). Испорченный
файл, необъявленный вид, повтор имени или битая символическая ссылка
выводятся с `Entry.Error` и в разборе не участвуют.

### Registry

| Метод | |
|---|---|
| `Parse(word, kinds) []Reading` | разборы из включённых словарей видов `kinds` (пусто — все) в порядке реестра: точные, а если их нет — предсказания словарей `base` |
| `List() []Entry` | копии всех записей, включая сломанные |
| `Summary() string` | строка для логов хоста: `dictionaries: N of M enabled; base: yes; broken: K` |
| `Version() string` | идентифицирует набор включённых словарей и их содержимое |
| `SetEnabled(ctx, name, on) error` | сохранить состояние, подменить снимок; `ErrUnknownDictionary` для неизвестного имени |
| `Reload(ctx) error` | пересканировать `Dir`, переоткрыть всё, перечитать состояние, подменить снимок |
| `Close() error` | освободить словари после завершения начатых вызовов; идемпотентен. После него `Parse` возвращает nil, `List` пуст, `SetEnabled` и `Reload` возвращают `ErrClosed` |

```go
type Entry struct {
    Name     string
    Kind     Kind
    Format   Format
    Origin   string   // OriginBuiltin или путь к файлу
    Hash     string   // хэш содержимого, hex
    Enabled  bool
    Manifest Manifest // Source, License, Version, URL, GeneratedFrom, Extra
    Error    string   // ошибка загрузки; "" если загружен
}
type Reading struct {
    Normal, Tag string // лемма как в словаре (может содержать ё); тег gomorphy, первой идёт часть речи
    Kind        Kind
    Dict        string
    Predicted   bool
}
type StateStore interface {
    Enabled(ctx context.Context) (map[string]bool, error) // нет записи — включён
    SetEnabled(ctx context.Context, name string, on bool) error
}
```

`MemState` — `StateStore` в памяти. `ParseManifest(r)` читает файл
происхождения (строки `key: value`, комментарии `#`, ключи без учёта
регистра); `Manifest.WriteTo` его пишет; `Manifest.String()` — сводка в
одну строку.

### Профили

```go
type Profile struct {
    Name      string            // ключ кэша; уникален в пределах Analyzer; "" отключает кэш
    Kinds     []Kind            // словари, по которым ищем; пусто — все
    Grammemes map[Kind][]string // для вида: разбор остаётся, только если в теге есть одна из граммем
}
```

Профили — конфигурация хоста и в `Analyzer.Version()` не входят.

### Analyzer

```go
func NewAnalyzer(d Dictionaries, r textnorm.Rules, opts AnalyzerOptions) *Analyzer
type AnalyzerOptions struct{ CacheSize int } // 0 — DefaultCacheSize (50 000), отрицательное — без кэша

func (a *Analyzer) Analyze(text string, p Profile, m Mode) []Term // ModeIndex или ModeFull
func (a *Analyzer) ParseQuery(q string, p Profile) Query
func (a *Analyzer) Version() string

type Dictionaries interface { // реализует *Registry; в тестах можно передать фейк
    Parse(word string, kinds []Kind) []Reading
    Version() string
}
type Term struct {
    Token  textnorm.Token
    Form   string  // форма поиска; «с.» с точкой для сокращения с точкой
    Lemmas []Lemma // нет у пунктуации и символов
}
type Lemma struct {
    Text  string // нормализована как слова (textnorm.NormalizeWord)
    Flags Flag   // FlagAmbiguous, FlagPredicted, FlagUnknown, FlagAbbrev, FlagStop, FlagReform
    Tag   string // тег первого разбора, давшего лемму
    Kind  Kind   // "" у лемм unknown
    Dict  string
}
type Query struct {
    Terms   []Term // законченные слова, как в ModeIndex
    Partial *Term  // последнее слово, если после него ничего нет; иначе nil
}
```

Поиск одного слова: точные разборы (затем дореформенные варианты
окончаний, флаг `reform`); если все точные разборы — служебные части речи
(PREP, CONJ, PRCL, INTJ; разборы `Abbr` не учитываются) — стоп-слово;
иначе профиль фильтрует разборы; если точных разборов нет вовсе —
предсказания базы (`predicted`); ничего — сама форма (`unknown`). Если
различных лемм несколько, все они `ambiguous`. Слова с точкой сначала
ищутся с точкой в словарях `abbrev`, слова через дефис — сначала целиком;
поиск сокращений не учитывает виды профиля.

`ModeIndex`: только кириллические слова и числа, стоп-слова отброшены,
слово через дефис — целиком и по частям. `ModeFull`: ровно один терм на
токен, стоп-слова сохраняются с `FlagStop`, числа и некириллические слова
`unknown` без поиска, пунктуация без лемм. `Flag.String()` даёт имена
`ambiguous,predicted,unknown,abbrev,stop,reform`.

## basefetch

`import "github.com/amarin/lexicon/basefetch"`

```go
const DefaultName = "base.opencorpora.dat"
func Fetch(dst string) (version string, err error)
```

Скачивает pymorphy2-dicts-ru, компилирует его и атомарно (временный файл
+ переименование) записывает `dst` и файл происхождения
`lexicon.ManifestPath(dst)`. Нужна сеть. Пакет тянет загрузчик pymorphy
из gomorphy и zap; импортируйте его только там, где скачиваете.

## Контракты

- **Смещения.** Смещения указывают на вход ровно в том виде, как он
  передан: байтовые смещения и смещения в кодовых точках у каждого токена,
  границы по графемным кластерам, `input[Start:End] == Raw`. Смещения
  никогда не указывают на нормализованную строку.
- **Версии.** `Analyzer.Version()` = `analyzer-<N>/<имя правил>-<версия
  правил>/<Registry.Version()>`. Изменения библиотеки, меняющие формы или
  термы, поднимают `Rules.Version` или версию анализатора; храните значение
  с производными данными и пересобирайте их при изменении
  ([сценарий 13](scenarios.md#13-понять-когда-переиндексировать)).
- **Конкурентность.** `Registry` и `Analyzer` безопасны для конкурентного
  использования. `Parse` не блокируется; `SetEnabled`, `Reload` и `Close`
  атомарно подменяют снимки и закрывают старые словари после начатых
  вызовов.
- **Файлы.** Заменяйте файлы словарей атомарно (временный файл +
  переименование): `.dat` отображается в память.
- **Ошибки, без логирования.** Пакеты возвращают ошибки и показывают
  состояние (`List`, `Summary`); логирует хост.
- **Только разметка.** lexicon не связывает спаны с данными хоста, не
  хранит разметку и не поднимает HTTP/MCP.
