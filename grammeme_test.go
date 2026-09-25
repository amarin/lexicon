package lexicon

import "testing"

func TestHasGrammeme(t *testing.T) {
	cases := []struct {
		tag, g string
		want   bool
	}{
		{"NOUN,anim,masc,Sgtm,Surn sing,ablt", "Surn", true},
		{"NOUN,anim,masc,Sgtm,Surn sing,ablt", "ablt", true},
		{"NOUN,anim,masc,Sgtm,Surn sing,ablt", "NOUN", true},
		{"N;GEN;SG", "GEN", true},
		{"NOUN", "NOU", false},
		{"", "NOUN", false},
	}
	for _, c := range cases {
		if got := hasGrammeme(c.tag, c.g); got != c.want {
			t.Errorf("hasGrammeme(%q, %q) = %v", c.tag, c.g, got)
		}
	}
}

func TestPOS(t *testing.T) {
	for tag, want := range map[string]string{"PREP": "PREP", "NOUN,anim sing": "NOUN", "N;GEN": "N", "": ""} {
		if got := pos(tag); got != want {
			t.Errorf("pos(%q) = %q, want %q", tag, got, want)
		}
	}
}

// TestAllStop: every reading must be a service part of speech.
func TestAllStop(t *testing.T) {
	if !allStop([]Reading{{Tag: "PREP"}, {Tag: "CONJ"}}) {
		t.Error("PREP+CONJ is not stop")
	}

	if allStop([]Reading{{Tag: "PREP"}, {Tag: "NOUN,inan"}}) || allStop(nil) {
		t.Error("NOUN or nothing counted as stop")
	}
}
