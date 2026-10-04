package ner

import (
	"context"
	"strings"
	"testing"
)

func oneKB() string {
	const sentence = "Крестьянин деревни Лягушкиной Иван Петров, 25 лет, жил на ул. Мира. "
	return strings.Repeat(sentence, 1024/len(sentence)+1)
}

// BenchmarkExtract1KB: spec target — under 1 ms per 1 KB with a warm cache.
func BenchmarkExtract1KB(b *testing.B) {
	p, _ := newPipeline(b, testEntries(), placesRules)
	doc := Doc{Text: oneKB()}
	if _, err := p.Extract(context.Background(), doc); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(doc.Text)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := p.Extract(context.Background(), doc); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExtract64KBDense: an entity-dense 64 KB document; Extract must
// stay close to linear in the document size.
func BenchmarkExtract64KBDense(b *testing.B) {
	p, _ := newPipeline(b, testEntries(), placesRules)
	const chunk = "деревня Лягушкино Иван Петров "
	doc := Doc{Text: strings.Repeat(chunk, 64*1024/len(chunk)+1)}
	if _, err := p.Extract(context.Background(), doc); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(doc.Text)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := p.Extract(context.Background(), doc); err != nil {
			b.Fatal(err)
		}
	}
}

// personsSentence is name-dense text for the person rule set.
const personsSentence = "Крестьянин Иван Петров Сидоров и Анна Михаилова Кузнецова, ответчик Сидоров. "

// BenchmarkExtract1KBPatterns: 1 KB of name-dense text through the person
// rule set (fix-up and assembly patterns, nesting on).
func BenchmarkExtract1KBPatterns(b *testing.B) {
	benchPatterns(b, 1024)
}

// BenchmarkExtract64KBPatterns: the same text at 64 KB; the pattern stage
// must stay linear in the document size (MB/s close to the 1 KB run).
func BenchmarkExtract64KBPatterns(b *testing.B) {
	benchPatterns(b, 64*1024)
}

func benchPatterns(b *testing.B, size int) {
	p := personsPipeline(b, true)
	doc := Doc{Text: strings.Repeat(personsSentence, size/len(personsSentence)+1)}
	if _, err := p.Extract(context.Background(), doc); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(doc.Text)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := p.Extract(context.Background(), doc); err != nil {
			b.Fatal(err)
		}
	}
}
