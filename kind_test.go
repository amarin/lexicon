package lexicon

import "testing"

func TestKindValid(t *testing.T) {
	for _, k := range []Kind{"base", "abbrev", "surname", "given_name", "t2"} {
		if !k.Valid() {
			t.Errorf("Kind(%q).Valid() = false", k)
		}
	}

	for _, k := range []Kind{"", "Base", "2x", "_x", "a-b", "a.b", "кот"} {
		if k.Valid() {
			t.Errorf("Kind(%q).Valid() = true", k)
		}
	}
}

// TestKindOf: the kind is the name prefix before the first dot (D17).
func TestKindOf(t *testing.T) {
	ok := map[string]Kind{"base.opencorpora": KindBase, "abbrev.scribes": KindAbbrev, "surname.volost": "surname", "custom.x.y": "custom"}
	for name, want := range ok {
		if got, err := kindOf(name); err != nil || got != want {
			t.Errorf("kindOf(%q) = %q, %v; want %q", name, got, err, want)
		}
	}

	for _, name := range []string{"mine", "Mine.x", "surname.", ".x"} {
		if _, err := kindOf(name); err == nil {
			t.Errorf("kindOf(%q): no error", name)
		}
	}
}

// TestKindSet: declared kinds (owner decision 7, D17). nil/empty accepts every
// valid kind; a non-empty set accepts base, abbrev and the declared kinds only.
func TestKindSet(t *testing.T) {
	all, err := newKindSet(nil)
	if err != nil || !all.accepts("surnme") || all.check("surnme") != nil {
		t.Fatalf("nil set: %v, accepts(surnme) = %v", err, all.accepts("surnme"))
	}

	s, err := newKindSet([]Kind{"surname", "given"})
	if err != nil {
		t.Fatal(err)
	}

	for _, k := range []Kind{KindBase, KindAbbrev, "surname", "given"} {
		if !s.accepts(k) || s.check(k) != nil {
			t.Errorf("accepts(%q) = false", k)
		}
	}

	err = s.check("surnme")
	if s.accepts("surnme") || err == nil || err.Error() != `unknown kind "surnme"` {
		t.Fatalf("check(surnme) = %v", err)
	}

	if _, err := newKindSet([]Kind{"surname", "Bad"}); err == nil {
		t.Fatal("newKindSet accepted an invalid kind")
	}
}

func TestFormatString(t *testing.T) {
	if FormatDat.String() != "dat" || FormatTSV.String() != "tsv" {
		t.Fatalf("formats %q %q", FormatDat, FormatTSV)
	}
}
