package textnorm

import (
	"slices"
	"testing"
)

// TestUTF16Offsets: surrogate pairs count two units; order is preserved.
func TestUTF16Offsets(t *testing.T) {
	s := "a😀б" // a: 1 byte/1 unit, 😀: 4 bytes/2 units, б: 2 bytes/1 unit
	if got, want := UTF16Offsets(s, 7, 0, 5, 1), []int{4, 0, 3, 1}; !slices.Equal(got, want) {
		t.Fatalf("UTF16Offsets = %v, want %v", got, want)
	}
}

// TestUTF16OffsetsOutOfRange: offsets outside the input panic.
func TestUTF16OffsetsOutOfRange(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("no panic for offset 3 of a 2-byte string")
		}
	}()

	UTF16Offsets("ab", 3)
}
