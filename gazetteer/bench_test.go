package gazetteer

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/internal/fakedict"
)

// synthWord returns a distinct Title-case Cyrillic word unknown to the fake
// dictionary, so aliases fall back to their normalized forms.
func synthWord(i int) string {
	letters := []rune("абвгдежзиклмнопрстуфхцчшщэюя")
	var b []rune
	for {
		b = append(b, letters[i%len(letters)])
		i /= len(letters)
		if i == 0 {
			break
		}
	}
	return "К" + string(b)
}

func synthEntries(n int) []Entry {
	out := make([]Entry, n)
	for i := range out {
		words := []string{synthWord(i)}
		for j := 0; j < i%3; j++ {
			words = append(words, synthWord(i*7+j+1))
		}
		out[i] = Entry{Alias: strings.Join(words, " "), Type: "division", Ref: fmt.Sprintf("division:%d", i), Canonical: words[0]}
	}
	return out
}

func benchGazetteer(b *testing.B, n int) (*Gazetteer, *lexicon.Analyzer) {
	an := fakedict.NewAnalyzer()
	g, err := New(context.Background(), Config{
		Analyzer:       an,
		DefaultProfile: fakedict.Profiles()["text"],
		Sources:        []Source{NewSliceSource("synthetic", "1", synthEntries(n))},
	})
	if err != nil {
		b.Fatal(err)
	}
	return g, an
}

// BenchmarkBuild100k: spec target — recompiling a 100k-alias source under 1 s.
func BenchmarkBuild100k(b *testing.B) {
	g, _ := benchGazetteer(b, 100_000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := g.RefreshSource(context.Background(), "synthetic"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMatch: spec target — zero allocations per token in the walk.
func BenchmarkMatch(b *testing.B) {
	g, an := benchGazetteer(b, 100_000)
	var words []string
	for i := 0; i < 150; i++ {
		if i%5 == 0 {
			words = append(words, synthWord(i*13))
		} else {
			words = append(words, []string{"крестьянин", "деревни", "Лягушкиной", "Иван"}[i%4])
		}
	}
	tx := Prepare(an.Analyze(strings.Join(words, " "), fakedict.Profiles()["text"], lexicon.ModeFull))
	snap := g.Snapshot()
	buf := make([]Match, 0, 256)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = snap.Match(tx, buf[:0])
	}
}
