package textnorm

import (
	"strings"
	"testing"
)

// benchText is about 1 KB of pre-reform text.
var benchText = strings.Repeat("Кр-нин с. Покровскаго Iоаннъ Ѳеодоровъ сынъ Кузнецо́въ, 1834 г. ", 10)

func BenchmarkTokenize(b *testing.B) {
	b.SetBytes(int64(len(benchText)))
	b.ReportAllocs()

	for b.Loop() {
		Tokenize(PreReform, benchText)
	}
}

func BenchmarkNormalizeWord(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		NormalizeWord(PreReform, "Покровскаго")
	}
}
