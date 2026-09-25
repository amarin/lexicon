package lexicon

import (
	"testing"

	"github.com/amarin/lexicon/textnorm"
)

// TestLemmaCache: repeated forms hit the cache; a new dictionary version drops
// it; a negative size or an unnamed profile disables it.
func TestLemmaCache(t *testing.T) {
	d := &countingDicts{fakeDicts: fake, version: "v1"}
	a := NewAnalyzer(d, textnorm.PreReform, AnalyzerOptions{})

	a.Analyze("Кота Кота", profText, ModeIndex)
	a.Analyze("Кота", profText, ModeFull)

	if n := d.calls.Load(); n != 1 {
		t.Fatalf("Parse calls = %d, want 1", n)
	}

	d.setVersion("v2")
	a.Analyze("Кота", profText, ModeIndex)

	if n := d.calls.Load(); n != 2 {
		t.Fatalf("Parse calls after version change = %d, want 2", n)
	}

	a.Analyze("Кота", Profile{}, ModeIndex)

	if n := d.calls.Load(); n != 3 {
		t.Fatalf("unnamed profile hit the cache: calls = %d, want 3", n)
	}

	off := &countingDicts{fakeDicts: fake, version: "v1"}
	NewAnalyzer(off, textnorm.PreReform, AnalyzerOptions{CacheSize: -1}).Analyze("Кота Кота", profText, ModeIndex)

	if n := off.calls.Load(); n != 2 {
		t.Fatalf("disabled cache: calls = %d, want 2", n)
	}
}

// TestLemmaCacheCopies: modifying returned lemmas does not change the cache.
func TestLemmaCacheCopies(t *testing.T) {
	a := newTestAnalyzer()

	first := a.Analyze("с. Кота", profText, ModeIndex)
	first[0].Lemmas[0].Text = "x"
	first[1].Lemmas[0].Text = "y"

	if got := show(a.Analyze("с. Кота", profText, ModeIndex)); got != "с.=село[ambiguous,abbrev],сын[ambiguous,abbrev]; кота=кот[]" {
		t.Fatalf("cached lemmas modified: %q", got)
	}
}
