package gazetteer

import "testing"

func TestParseEntryFlags(t *testing.T) {
	f, err := ParseEntryFlags(" surface_only, requires_context ,")
	if err != nil {
		t.Fatal(err)
	}
	if !f.Has(SurfaceOnly) || !f.Has(RequiresContext) || f.Has(Blocked) || f.Has(CaseSensitive) {
		t.Fatalf("flags = %v", f)
	}
	if got := f.String(); got != "requires_context,surface_only" {
		t.Fatalf("String() = %q", got)
	}
	if f, err := ParseEntryFlags(""); err != nil || f != 0 {
		t.Fatalf("empty: %v %v", f, err)
	}
	if _, err := ParseEntryFlags("blocked,shiny"); err == nil {
		t.Fatal("unknown flag accepted")
	}
}
