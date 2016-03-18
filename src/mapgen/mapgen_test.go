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

func benchmarkMapgenN(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		Generate(n)
	}
}

func BenchmarkMapgen2(b *testing.B) {
	benchmarkMapgenN(b, 2)
}

func BenchmarkMapgen5(b *testing.B) {
	benchmarkMapgenN(b, 5)
}

func BenchmarkMapgen10(b *testing.B) {
	benchmarkMapgenN(b, 10)
}

func BenchmarkMapgen15(b *testing.B) {
	benchmarkMapgenN(b, 15)
}

func BenchmarkMapgen20(b *testing.B) {
	benchmarkMapgenN(b, 20)
}
