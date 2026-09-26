package gazetteer

import "testing"

func TestTrie(t *testing.T) {
	a := &Alias{Entry: Entry{Alias: "a"}}
	b := &Alias{Entry: Entry{Alias: "b"}}
	c := &Alias{Entry: Entry{Alias: "c"}}
	bld := newTrieBuilder()
	bld.insert([]uint32{7, 3}, a)
	bld.insert([]uint32{7, 3}, a) // duplicate key of the same alias is ignored
	bld.insert([]uint32{7, 3}, b)
	bld.insert([]uint32{7, 1, 9}, c)
	bld.insert([]uint32{5}, c)
	tr := bld.freeze()

	if tr.depth != 3 {
		t.Fatalf("depth = %d", tr.depth)
	}
	n7, ok := tr.child(0, 7)
	if !ok || len(tr.terminal(n7)) != 0 {
		t.Fatal("7 must be an inner node")
	}
	n73, ok := tr.child(n7, 3)
	if got := tr.terminal(n73); !ok || len(got) != 2 || got[0] != a || got[1] != b {
		t.Fatalf("7-3 terminal = %v", got)
	}
	if _, ok := tr.child(n7, 2); ok {
		t.Fatal("7-2 must not exist")
	}
	n71, _ := tr.child(n7, 1)
	n719, ok := tr.child(n71, 9)
	if !ok || tr.terminal(n719)[0] != c {
		t.Fatal("7-1-9 missing")
	}
	n5, ok := tr.child(0, 5)
	if !ok || tr.terminal(n5)[0] != c {
		t.Fatal("5 missing")
	}
	if _, ok := tr.child(0, 0); ok {
		t.Fatal("ID 0 must never match")
	}
	empty := newTrieBuilder().freeze()
	if empty.depth != 0 {
		t.Fatal("empty trie depth")
	}
	if _, ok := empty.child(0, 7); ok {
		t.Fatal("empty trie has children")
	}
}
