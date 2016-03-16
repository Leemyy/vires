package mapgen

import "testing"

func BenchmarkMapgen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Generate(20)
	}
}
