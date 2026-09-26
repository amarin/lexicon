# Справочник по библиотеке

> English version: [docs/en/library.md](../en/library.md).

Справочник по публичному API, по пакетам. Зачем нужна каждая часть и
работающие примеры — в [scenarios.md](scenarios.md). Последнее слово за
godoc каждого символа; эта страница группирует их и фиксирует контракты.

- [textnorm](#textnorm) — орфография и токены, без словарей
- [lexicon](#lexicon) — словари (`Registry`), `Profile`, `Analyzer`
- [basefetch](#basefetch) — скачивание базового словаря
- [gazetteer](#gazetteer) — записи хоста, скомпилированные для сопоставления *(0.2, не выпущено)*
- [rules](#rules) — подсказки, триггеры, наборы правил по тегам документа *(0.2, не выпущено)*
- [ner](#ner) — конвейер извлечения *(0.2, не выпущено)*
- [nertest](#nertest) — оценка на эталонном наборе *(0.2, не выпущено)*
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

## gazetteer

`import "github.com/amarin/lexicon/gazetteer"` — *(0.2, не выпущено)*

Компилирует псевдонимы записей хоста в префиксные деревья по токенам и
сопоставляет их с разобранным текстом. Сообщает о каждом совпадении,
включая пересекающиеся; выбирать между ними — работа `ner`. Импортирует
только корневой пакет и `textnorm`.

### Записи и источники

```go
type Entry struct {
    Alias     string            // текст псевдонима: «дер. Лягушкино», «СПб»
    Type      string            // тип сущности хоста: "surname", "division", …
    Ref       string            // непрозрачный ключ хоста; псевдонимы с общим Ref — варианты одной записи
    Canonical string            // нормальная форма записи; пусто — Alias
    Attrs     map[string]string // передаются в спаны как есть
    Flags     EntryFlag         // RequiresContext, SurfaceOnly, CaseSensitive, Blocked
}
type Source interface {
    Name() string                                                // уникально в пределах Gazetteer
    Version(ctx context.Context) (string, error)                 // дешёвый; не изменилась — без перекомпиляции
    Entries(ctx context.Context, yield func(Entry) error) error // останавливается на первой ошибке yield
}
func NewTSVSource(name, path string) *TSVSource        // читает path при каждом Version/Entries
func NewTSVSourceData(name string, data []byte) *TSVSource // из памяти, например //go:embed
func NewSliceSource(name, version string, entries []Entry) *SliceSource
```

TSV: `type<TAB>ref<TAB>canonical<TAB>alias[<TAB>flags[<TAB>k=v;k=v]]`,
флаги через запятую (`requires_context,surface_only,case_sensitive,blocked`,
`ParseEntryFlags`); строки на `#` — комментарии, строки `# key: value`
(ключи `[a-z_]`) до первой записи образуют манифест
(`TSVSource.Manifest()`). В строке от 4 до 6 полей, каждое обрезается по
краям. `Version` — sha256 содержимого: `NewTSVSource` читает и хэширует
весь файл при каждом вызове. Плохая строка пропускается и возвращается как
`*ParseErrors` после всех хороших записей — сборщик кладёт её в отчёт;
строка длиннее 1 МиБ роняет весь источник (`SourceReport.Err`).
`Manifest()` пуст до первого вызова `Entries`.

| Флаг | Действие |
|---|---|
| `RequiresContext` | остаётся, только если совпадение поддержала подсказка или триггер |
| `SurfaceOnly` | без ключа по леммам: сопоставляется только по нормализованной форме |
| `CaseSensitive` | остаётся, только если у каждого слова регистр как у псевдонима |
| `Blocked` | запрещает любое совпадение типа записи на этом диапазоне, независимо от остальных флагов |

### Gazetteer

```go
func New(ctx context.Context, cfg Config) (*Gazetteer, error)
type Config struct {
    Analyzer       Analyzer                   // *lexicon.Analyzer; его Version входит в каждый скомпилированный источник
    TypeProfiles   map[string]lexicon.Profile // профиль на тип записи
    DefaultProfile lexicon.Profile            // для остальных типов
    Sources        []Source                   // компилируются по порядку; имена уникальны и непусты
}
```

| Метод | |
|---|---|
| `Snapshot() *Snapshot` | текущее скомпилированное состояние, без блокировок; держите его на время одной операции |
| `Refresh(ctx) ([]SourceReport, error)` | перекомпилировать источники, у которых изменилась `Version` (или версия анализатора); атомарная подмена |
| `RefreshSource(ctx, name) (SourceReport, error)` | перекомпилировать один источник независимо от версии; `ErrUnknownSource` |
| `Canonical(key)`, `Expand(lemma)` | делегируют текущему снимку |

`New` падает на неверной конфигурации или отменённом `ctx` (он выполняет
первый `Refresh`); сбой источника виден в `Snapshot().Reports()`.
`Refresh` и `RefreshSource` выполняются по очереди, поэтому один
`Gazetteer` никогда не вызывает `Source` конкурентно. `Refresh` также
повторяет каждый источник, который ещё ни разу не собрался; он возвращает
отчёты пересобранных источников и тех, у которых не удался `Version`.
Каждый псевдоним разбирается (`ModeFull`) с профилем своего типа в ключ по
поверхностной форме и ключи по леммам —
неоднозначное слово раскрывается в комбинации, не больше `MaxLemmaKeys`
(8) на псевдоним. `SourceReport` считает записи, псевдонимы, ключи по
леммам и по поверхностной форме, урезанные и заблокированные псевдонимы,
пропущенные пустые, длительность и нефатальные `Errors`; `Err` — фатальный
сбой, после которого источник сохраняет прежние данные.

### Snapshot и сопоставление

| Метод | |
|---|---|
| `Match(tx *Text, out []Match) []Match` | дописать в `out` каждое совпадение псевдонима в `tx`; память выделяется только на рост `out` |
| `Version() string` | меняется, когда источник пересобран из новой `Version` или новым анализатором (принудительный `RefreshSource` без изменений её сохраняет) |
| `Reports() []SourceReport` | последний отчёт каждого источника, в порядке конфигурации |
| `Canonical(key) []string` | канонические формы псевдонимов, чей ключ по леммам или по поверхностной форме (через пробел) равен `key` |
| `Expand(lemma) []string` | ключи по леммам всех групп вариантов, содержащих `lemma`, — для расширения запроса |

Группы вариантов — псевдонимы одного источника с общим `Ref`; псевдонимы
с пустым `Ref` или без ключей по леммам (`SurfaceOnly`) ни в одну группу
не входят. Псевдонимы с `Blocked` входят, а неоднозначный псевдоним даёт
все свои леммы (`Expand("сталь")` → `[сталь стать]`). Ключи — нормализованные
строки в нижнем регистре. `ner` группы не использует: `Normal` спана
берётся из `Entry.Canonical`.

`Prepare(terms)` строит `Text` из термов `Analyzer.Analyze(ModeFull)`.
`Match` — это значимые позиции `[Start, End)` (слова и числа; обратно в
термы — через `Text.TermIndex`), способ `Kind` (`ByLemma`, `BySurface`) и
псевдонимы `Aliases`, чей ключ закончился здесь. Совпадение покрывает
только подряд идущие слова: любая пунктуация или символ между словами его
прерывает (собственная точка сокращения — нет). Псевдонимы и снимки общие и только для чтения.

## rules

`import "github.com/amarin/lexicon/rules"` — *(0.2, не выпущено)*

Правила как данные, проверенные и собранные в неизменяемую `Book`.
Импортирует `go.yaml.in/yaml/v3`.

```go
func Load(r io.Reader) (File, error)              // в ошибках — "<input>"
func LoadNamed(name string, r io.Reader) (File, error)
func LoadFile(path string) (File, error)          // .yaml, .yml или .json
func Compile(files ...File) (*Book, error)        // имена наборов уникальны во всех файлах

type File struct {
    Meta map[string]string // `meta:` — source, license, version, url, generated_from
    Sets []RuleSet         // `sets:`
}
type RuleSet struct {
    Name     string    // `name:`
    When     []string  // `when:` — активен, когда у документа есть ВСЕ эти теги; пусто — всегда
    Hints    []Hint
    Triggers []Trigger
}
```

Один документ YAML на файл (JSON — тоже YAML); неизвестные поля —
ошибка. Ошибки правил выглядят как `<file>:<line>: <message>`, ошибки
`Compile` — `<file>:<line>: <set>/<rule>: <message>` (`hint 0`,
`trigger 1`); у файла, который не открылся (`rules: open …`) или пуст
(`<file>: empty rule file`), номера строки нет.

| Поле подсказки | YAML | По умолчанию | |
|---|---|---|---|
| `Lemma` | `lemma` | — | лемма или форма ключевого слова, варианты через `\|` |
| `Dotted` | `dotted` | false | за ключевым словом должна идти «.» |
| `Type` | `type` | — | тип спанов, которые она поддерживает |
| `Dir` | `dir` | `right` | `right`, `left`, `both` |
| `Window` | `window` | 1 | расстояние в значимых словах от ключевого слова до границы спана, ≤ `MaxWindow` (8); 1 — вплотную |
| `Weight` | `weight` | 1 | прибавляется к оценке спана |
| `Absorb` | `absorb` | false | расширить соседний спан на ключевое слово |

У подсказок нет `shape`, и на кандидатов от триггеров они не действуют
(выполняются раньше).

У `Trigger` те же `lemma`, `dotted`, `type`, `weight`, `absorb`, а
также `dir` (`right` или `left`), `window` как диапазон (`N` = 1..N,
`"min..max"`), `shape` (`case`: `lower`/`title`/`upper`, `script`:
`cyrillic`/`latin`) и `stop_at` (`punct` подразумевается, `stop`,
`number`, `latin`), ограничивающие покрываемые слова, и `negative`.
Триггер предлагает `Candidate` на покрытых словах, если их не перекрывает
ни один спан газетира его типа, а иначе усиливает эти спаны; его `weight`
прибавляется в обоих случаях, а `absorb` расширяет и тех, и других на
соседнее ключевое слово. Триггер с `negative` только вычитает вес из
перекрывающих спанов газетира. Кандидат не предлагается на диапазоне,
пересекающем совпадение с `Blocked` того же типа.

| Метод `Book` | |
|---|---|
| `Active(tags) Active` | подсказки и триггеры наборов, все теги `When` которых есть в `tags`, в порядке файлов и наборов; общие, не изменяйте |
| `Sets() []string` | имена наборов по порядку |
| `Version() string` | меняется при изменении любого правила или значения meta |

Ключевые слова сравниваются после `textnorm.NormalizeWord(textnorm.PreReform, …)`
и без завершающей точки, какую бы орфографию ни использовал документ.

## ner

`import "github.com/amarin/lexicon/ner"` — *(0.2, не выпущено)*

```go
func New(cfg Config) (*Pipeline, error)
type Config struct {
    Analyzer           Analyzer                   // *lexicon.Analyzer
    Gazetteer          Gazetteer                  // *gazetteer.Gazetteer
    Rules              *rules.Book                // nil — без правил
    Profiles           map[string]lexicon.Profile // Doc.Profile → профиль анализатора
    DefaultProfile     string
    Nesting            map[string][]string        // внешний тип → внутренние типы, допустимые внутри него
    Weights            Weights                    // DefaultWeights(), если Surface, Lemma и Trigger все 0
    MinLemmaMatchRunes int                        // отбрасывать однословные совпадения только по лемме у более коротких псевдонимов; 0 → 3
}
func (p *Pipeline) Extract(ctx context.Context, d Doc, opts ...Option) (Result, error)
func Explain() Option // заполнить Span.Evidence

type Doc struct {
    Text    string
    Profile string   // пусто — Config.DefaultProfile; неизвестный — ErrUnknownProfile
    Tags    []string // выбирают наборы правил
    Types   []string // фильтр результата, после разрешения
}
type Result struct {
    Spans   []Span // по Start, затем длинные первыми, затем по Type
    Version string // экстрактор, анализатор, снимок газетира, правила, конфигурация
}
type Span struct {
    Start, End         int // байты в Doc.Text; Doc.Text[Start:End] == Surface
    RuneStart, RuneEnd int // кодовые точки
    Surface, Type      string
    Normal             []string          // канонические формы; >1 при неоднозначности
    Refs               []string          // ссылки газетира; пусто у кандидата
    Attrs              map[string]string // атрибуты записей, конфликтующие ключи отброшены
    Flags              SpanFlag          // Ambiguous, Predicted, Abbrev, Candidate, Nested
    Score              float32
    Evidence           []string          // с Explain
    Alternatives       []Alternative     // проигравшие типы на том же диапазоне, лучшие первыми
}
```

`Extract` выполняет: разбор (`ModeFull`) → совпадения газетира → фильтры
(`Blocked`, `CaseSensitive`, короткие совпадения по лемме) → подсказки →
триггеры → фильтр `RequiresContext` → оценка → взвешенное планирование
интервалов (без пересечений; вложенность только для пар из
`Config.Nesting`) → фильтр `Doc.Types`. Оценка = (вес происхождения +
`Types[type]`) × слова + `LengthBonus` × (слова − 1) + вклад подсказок и
триггеров − `AmbiguityPenalty` × (max(различные ссылки, различные
нормальные формы, 1) − 1), с округлением до 1e-6; кандидаты с оценкой 0
или ниже отбрасываются до разрешения. `DefaultWeights()`: поверхность 3,
лемма 2, триггер 1, бонус за длину 0,5, штраф за неоднозначность 0,25; он
заменяет `Config.Weights`, только если `Surface`, `Lemma` и `Trigger` все
нулевые (`Types` сохраняются) — задав один вес происхождения, задавайте
все. Отрицательный `MinLemmaMatchRunes` отключает фильтр коротких лемм.
Точная ничья между типами помечает спан `Ambiguous` и разрешается по
имени типа. `Span.Flags`: `Ambiguous` — несколько ссылок или нормальных
форм, ничья типов или покрытое неоднозначное сокращение, не поглощённое
правилом (омонимия обычного слова не считается); `Predicted` — покрытое
слово известно только по предсказанным леммам (не ставится у совпадений
по поверхностной форме); `Abbrev`, `Candidate`, `Nested`. `Normal`
кандидата от триггера — последовательности лемм его слов в нижнем
регистре (без поглощённого ключевого слова, не больше
`gazetteer.MaxLemmaKeys`).

`New` возвращает ошибку при nil `Analyzer` или `Gazetteer` и при
`DefaultProfile`, которого нет в `Profiles`; карты хоста он копирует.
`Extract` линеен по размеру документа и проверяет `ctx` между этапами.

## nertest

`import "github.com/amarin/lexicon/nertest"` — *(0.2, не выпущено)*

```go
type Case struct {
    ID      string            `json:"id,omitempty"`      // нет — "line<N>"
    Text    string            `json:"text"`
    Profile string            `json:"profile,omitempty"`
    Tags    []string          `json:"tags,omitempty"`
    Spans   []Gold            `json:"spans"`
    Context map[string]string `json:"context,omitempty"` // данные хоста, игнорируются без WithTags
}
type Gold struct {
    Text       string `json:"text"`
    Type       string `json:"type"`
    Occurrence int    `json:"occurrence,omitempty"` // с 1, по умолчанию 1
}
func LoadCases(r io.Reader) ([]Case, error) // JSON Lines; пустые строки пропускаются; неизвестные поля — ошибка
func LoadCasesFile(path string) ([]Case, error)
func Run(ctx context.Context, ex Extractor, cases []Case, opts ...Option) (*Report, error)
func WithTags(f func(Case) []string) Option // дополнительные теги только для извлечения
```

`Extractor` — всё, у чего есть `Extract` как у `ner.Pipeline`. В `Report`
есть `Cases`, `Types map[string]*TypeScore` (`Strict` — точные байты и
тип, `Partial` — пересечение и тип; каждый — `Counts{TP, FP, FN}` с
`Precision()`, `Recall()`, `F1()`) и `Failures` (строгие пропущенные
(`missed`) и лишние (`spurious`) спаны). `Report.Write(w)` печатает таблицу
и расхождения (`Failures`); `Report.Check(minPrecision, minRecall)`
возвращает по сообщению на каждый нарушенный порог, например `surname:
strict recall 0.500 < 0.900`.

`Run` оценивает все спаны результата, включая `Nested` и `Candidate`;
задать `Doc.Types` он не может и останавливается на первой ошибке
`Extract`. `LoadCases` также отвергает эталонные спаны с пустым текстом или
типом и строки длиннее 4 МиБ.

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
  `ner.Result.Version` делает то же для извлечённых спанов: покрывает
  версию экстрактора, `Analyzer.Version()`, `Snapshot.Version()` газетира,
  `Book.Version()` и конфигурацию конвейера.
- **Конкурентность.** `Registry` и `Analyzer` безопасны для конкурентного
  использования. `Parse` не блокируется; `SetEnabled`, `Reload` и `Close`
  атомарно подменяют снимки и закрывают старые словари после начатых
  вызовов.
  `Gazetteer`, `Snapshot`, `Book` и `Pipeline` тоже безопасны для
  конкурентного использования: `Refresh` атомарно подменяет снимки
  газетира, а каждый `Extract` фиксирует один снимок (только снимок —
  анализатор и его реестр остаются живыми).
- **Файлы.** Заменяйте файлы словарей атомарно (временный файл +
  переименование): `.dat` отображается в память.
- **Ошибки, без логирования.** Пакеты возвращают ошибки и показывают
  состояние (`List`, `Summary`); логирует хост.
- **Только разметка.** lexicon не связывает спаны с данными хоста, не
  хранит разметку и не поднимает HTTP/MCP.
