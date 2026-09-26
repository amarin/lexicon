package gazetteer

import (
	"sync"
	"testing"
)

func TestInterner(t *testing.T) {
	in := newInterner()
	a := in.intern("лягушкино")
	b := in.intern("деревня")
	if a == 0 || b == 0 || a == b {
		t.Fatalf("ids: %d %d", a, b)
	}
	if in.intern("лягушкино") != a || in.lookup("лягушкино") != a {
		t.Fatal("ids are not stable")
	}
	if in.lookup("нет такого") != 0 {
		t.Fatal("unknown text must map to 0")
	}
	if in.text(b) != "деревня" || in.text(0) != "" {
		t.Fatal("text() mismatch")
	}
}

func TestInternerConcurrent(t *testing.T) {
	in := newInterner()
	var wg sync.WaitGroup
	ids := make([]uint32, 8)
	for g := range ids {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				in.intern(synthKey(i))
			}
			ids[g] = in.lookup(synthKey(999))
		}(g)
	}
	wg.Wait()
	for _, id := range ids {
		if id == 0 || id != ids[0] {
			t.Fatalf("ids differ: %v", ids)
		}
	}
}

func synthKey(i int) string { return string(rune('а'+i%32)) + string(rune('а'+i/32%32)) }
