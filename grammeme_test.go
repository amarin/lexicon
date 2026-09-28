package lexicon

import "testing"

// TestAllStop: every reading must be a service part of speech.
func TestAllStop(t *testing.T) {
	if !allStop([]Reading{{Tag: "PREP"}, {Tag: "CONJ"}}) {
		t.Error("PREP+CONJ is not stop")
	}

	if allStop([]Reading{{Tag: "PREP"}, {Tag: "NOUN,inan"}}) || allStop(nil) {
		t.Error("NOUN or nothing counted as stop")
	}
}
