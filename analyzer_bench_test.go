package lexicon

import (
	"strings"
	"testing"

	"github.com/amarin/lexicon/textnorm"
)

// analyzerBenchText is about 1 KB of mixed text.
var analyzerBenchText = strings.Repeat("Кр-нин с. Покровскаго Iоаннъ сынъ Кузнецова, 1834 г. кот в селе. ", 12)

func benchAnalyze(b *testing.B, opts AnalyzerOptions, m Mode) {
	a := NewAnalyzer(fake, textnorm.PreReform, opts)
	a.Analyze(analyzerBenchText, profText, m) // warm the cache

	b.SetBytes(int64(len(analyzerBenchText)))
	b.ReportAllocs()

	for b.Loop() {
		a.Analyze(analyzerBenchText, profText, m)
	}
}

func BenchmarkAnalyzeIndex(b *testing.B) { benchAnalyze(b, AnalyzerOptions{}, ModeIndex) }

func BenchmarkAnalyzeFull(b *testing.B) { benchAnalyze(b, AnalyzerOptions{}, ModeFull) }

func BenchmarkAnalyzeFullNoCache(b *testing.B) {
	benchAnalyze(b, AnalyzerOptions{CacheSize: -1}, ModeFull)
}

func BenchmarkParseQuery(b *testing.B) {
	a := NewAnalyzer(fake, textnorm.PreReform, AnalyzerOptions{})
	b.ReportAllocs()

	for b.Loop() {
		a.ParseQuery("Кузнецова Покро", profText)
	}
}
