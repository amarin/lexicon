package lexicon

import "testing"

func BenchmarkRegistryParse(b *testing.B) {
	dir := b.TempDir()
	writeFile(b, dir, "surname.test.dat", datBytes(b, surnameForms))

	r := openTest(b, dir, nil)
	b.ReportAllocs()

	for b.Loop() {
		r.Parse("кота", nil)
		r.Parse("кузнецов", nil)
	}
}
