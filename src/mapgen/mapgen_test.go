package mapgen

import "testing"

func TestMapgen(t *testing.T) {
	f := Generate(20)
	if len(f.StartCells) != 20 {
		t.Fatalf("len(Generate(20).StartCells) = %d", len(f.StartCells))
	}
	for c1 := range f.Cells {
		for c2 := range f.Cells {
			if c1 == c2 {
				continue
			}
			if tooClose(c1, c2) {
				t.Fatalf("Generate(20).Cells: too close: %+v, %+v", c1, c2)
			}
		}
	}
}

func BenchmarkMapgen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Generate(20)
	}
}
