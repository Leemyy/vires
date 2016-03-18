package mapgen

import "testing"

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
