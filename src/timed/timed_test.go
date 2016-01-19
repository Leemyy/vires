package timed

import (
	"fmt"
	"testing"
	"time"
)

func TestTimed(t *testing.T) {
	// not an automatable test, just a very rough one
	out := make(chan string, 10)
	cases := []struct {
		key    interface{}
		in     time.Duration
		action func()
	}{
		{"1", 0, func() { out <- "0" }},
		{"2", 0, func() { out <- "1" }},
		{"3", 10, func() { out <- "2" }},
		{"4", 20, func() { out <- "3" }},
		{"5", 30, func() { out <- "4" }},
		{"6", 100, func() { out <- "5"; close(out) }},
		{"7", 40, func() { out <- "before 5 - 6" }},
	}
	tim := New()
	for _, c := range cases {
		tim.Lock()
		tim.Start(c.key, c.in, c.action)
		tim.Unlock()
	}
	for v := range out {
		fmt.Println(v)
	}
	tim.Close()
}
