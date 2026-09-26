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
