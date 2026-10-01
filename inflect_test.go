package lexicon

import (
	"slices"
	"strings"
	"sync"
	"testing"
)

// inflectForms is the base of the Inflect tests, stored with ё.
var inflectForms = []wordForm{
	{"уезд", "уезд", "NOUN,inan,masc sing,nomn"},
	{"уезда", "уезд", "NOUN,inan,masc sing,gent"},
	{"уезд", "уезд", "NOUN,inan,masc sing,accs"},
	{"уезды", "уезд", "NOUN,inan,masc plur,nomn"},
	{"уездов", "уезд", "NOUN,inan,masc plur,gent"},
	{"посёлок", "посёлок", "NOUN,inan,masc sing,nomn"},
	{"посёлки", "посёлок", "NOUN,inan,masc plur,nomn"},
	{"село", "село", "NOUN,inan,neut sing,nomn"},
	{"сёла", "село", "NOUN,inan,neut plur,nomn"},
	{"сёл", "село", "NOUN,inan,neut plur,gent"},
	{"сесть", "сесть", "INFN,perf,intr"},
	{"село", "сесть", "VERB,perf,intr neut,sing,past,indc"},
	{"сели", "сесть", "VERB,perf,intr plur,past,indc"},
	{"корпус", "корпус", "NOUN,inan,masc sing,nomn"},
	{"корпусы", "корпус", "NOUN,inan,masc plur,nomn"},
}

// inflectCustomTSV is a host dictionary (lemma, form, tag): another plural
// of «корпус» and «посёлок», which a TSV dictionary stores with е (D28).
const inflectCustomTSV = "корпус\tкорпус\tNOUN,inan,masc sing,nomn\n" +
	"корпус\tкорпуса\tNOUN,inan,masc plur,nomn\n" +
	"посёлок\tпосёлок\tNOUN,inan,masc sing,nomn\n" +
	"посёлок\tпосёлки\tNOUN,inan,masc plur,nomn\n"

// openInflect opens base (inflectForms) + custom.terms (inflectCustomTSV).
func openInflect(t testing.TB) *Registry {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, dir, "custom.terms.tsv", []byte(inflectCustomTSV))

	r, err := Open(t.Context(), Options{Dir: dir, Base: datBytes(t, inflectForms)})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(func() { _ = r.Close() })

	return r
}

// TestInflect: forms by grammemes — the source reading filter, homonyms,
// duplicates, ё, dictionary order, kinds, empty want, misses.
func TestInflect(t *testing.T) {
	r := openInflect(t)

	var (
		noun   = []string{"NOUN", "nomn", "sing"}
		custom = []Kind{"custom"}
		base   = []Kind{KindBase}
	)

	cases := []struct {
		name       string
		word       string
		kinds      []Kind
		from, want []string
		forms      string
	}{
		{"plural", "уезд", nil, noun, []string{"plur", "nomn"}, "уезды"},
		{"genitive singular", "уезд", nil, noun, []string{"gent", "sing"}, "уезда"},
		{"genitive plural", "уезд", nil, noun, []string{"gent", "plur"}, "уездов"},
		{"the source form itself", "уезд", nil, noun, []string{"nomn", "sing"}, "уезд"},
		{"from an oblique form", "уездов", nil, []string{"gent"}, []string{"nomn", "sing"}, "уезд"},
		{"upper case input", "Уезд", nil, noun, []string{"plur", "nomn"}, "уезды"},
		{"readings agree: one form", "уезд", nil, nil, []string{"plur", "nomn"}, "уезды"},
		{"homonyms without from", "село", nil, nil, []string{"plur"}, "сёла|сёл|сели"},
		{"from picks the noun", "село", nil, []string{"NOUN"}, []string{"plur"}, "сёла|сёл"},
		{"from picks the verb", "село", nil, []string{"VERB"}, []string{"plur"}, "сели"},
		{"empty want: every form, source first", "уезда", nil, []string{"gent"}, nil, "уезда|уезд|уездов|уезды"},
		{"е finds ё in the base", "поселок", base, noun, []string{"plur", "nomn"}, "посёлки"},
		{"ё in the base", "посёлок", base, noun, []string{"plur", "nomn"}, "посёлки"},
		{"е in a TSV dictionary", "поселок", custom, noun, []string{"plur", "nomn"}, "поселки"},
		{"ё folded for a TSV dictionary", "посёлок", custom, noun, []string{"plur", "nomn"}, "поселки"},
		{"ё: both dictionaries answer", "посёлок", nil, noun, []string{"plur", "nomn"}, "посёлки|поселки"},
		{"registry order", "корпус", nil, noun, []string{"plur", "nomn"}, "корпусы|корпуса"},
		{"kinds select the dictionary", "корпус", custom, noun, []string{"plur", "nomn"}, "корпуса"},
		{"kind without the word", "уезд", custom, noun, []string{"plur", "nomn"}, ""},
		{"no form with want", "уезд", nil, noun, []string{"datv"}, ""},
		{"no reading with from", "уезд", nil, []string{"VERB"}, []string{"plur"}, ""},
		{"unknown word", "уездик", nil, nil, []string{"plur", "nomn"}, ""},
		{"empty word", "", nil, nil, nil, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := r.Inflect(c.word, c.kinds, c.from, c.want)
			if s := strings.Join(got, "|"); s != c.forms {
				t.Fatalf("Inflect(%q, %v, %v, %v) = %q, want %q", c.word, c.kinds, c.from, c.want, s, c.forms)
			}

			if c.forms == "" && got != nil {
				t.Fatalf("a miss must be nil, got %#v", got)
			}
		})
	}
}

// TestInflectAfterClose: like Parse, nil on a closed registry.
func TestInflectAfterClose(t *testing.T) {
	r := openInflect(t)

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	if got := r.Inflect("уезд", nil, nil, []string{"plur", "nomn"}); got != nil {
		t.Fatalf("Inflect after Close = %v", got)
	}
}

// TestInflectDuringReload (run with -race): Inflect pins its snapshot, so
// reloads never close a dictionary under it.
func TestInflectDuringReload(t *testing.T) {
	r := openInflect(t)

	var wg sync.WaitGroup

	stop := make(chan struct{})

	for range 4 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-stop:
					return
				default:
				}

				got := r.Inflect("уезд", nil, []string{"NOUN", "nomn", "sing"}, []string{"plur", "nomn"})
				if !slices.Equal(got, []string{"уезды"}) {
					t.Errorf("Inflect during Reload = %v", got)

					return
				}
			}
		}()
	}

	for range 20 {
		if err := r.Reload(t.Context()); err != nil {
			t.Errorf("Reload: %v", err)
		}
	}

	close(stop)
	wg.Wait()
}
