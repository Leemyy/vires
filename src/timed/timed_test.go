package timed

import (
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestTimed(t *testing.T) {
	const n = 1500
	const maxms = 1000
	const chance = 0.3
	timed := New()
	results := make(chan int, n)
	expected := make([]int, 0, n)
	var wg sync.WaitGroup
	now := time.Now()
	for i := 0; i < n; i++ {
		d := rand.Intn(maxms)
		at := now.Add(time.Duration(d) * time.Nanosecond)
		wg.Add(1)
		stop := timed.Start(at, func() {
			results <- d
			wg.Done()
		})
		if rand.Float64() < chance {
			stop()
			wg.Add(-1)
		} else {
			expected = append(expected, d)
		}
	}
	wg.Wait()
	close(results)
	got := make([]int, 0, n)
	for r := range results {
		got = append(got, r)
	}
	sort.Ints(expected)
	fatal := func() { t.Fatalf("Timed: expected %v\ngot %v", expected, got) }
	if len(expected) != len(got) {
		t.Errorf("Timed: expected len %d, got %d", len(expected), len(got))
		fatal()
	}
	for i, v := range got {
		if v != expected[i] {
			t.Errorf("Timed: expected elem %d, got %d", expected[i], v)
			fatal()
		}
	}
}
