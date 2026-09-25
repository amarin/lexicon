package lexicon

import "testing"

// TestMemState: the zero value is empty and ready; Enabled returns a copy.
func TestMemState(t *testing.T) {
	var s MemState

	ctx := t.Context()

	got, err := s.Enabled(ctx)
	if err != nil || len(got) != 0 {
		t.Fatalf("Enabled = %v, %v", got, err)
	}

	if err := s.SetEnabled(ctx, "surname.x", false); err != nil {
		t.Fatal(err)
	}

	got, _ = s.Enabled(ctx)
	got["surname.x"] = true

	again, _ := s.Enabled(ctx)
	if on, ok := again["surname.x"]; !ok || on {
		t.Fatalf("state %v: want surname.x=false", again)
	}
}
