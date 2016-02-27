// Package timed provides a
// preemptive shortest time remaining
// scheduler for timers.
package timed

import (
	"container/heap"
	"time"
)

// timer represents a scheduled timer within timed.
type timer struct {
	at     time.Time
	action func()
}

// timed represents a min-heap of scheduled timers
// where the timer with the shortest remaining time
// is always the first element provided by Pop().
type timed []*timer

func (h timed) Len() int           { return len(h) }
func (h timed) Less(i, j int) bool { return h[i].at.Before(h[j].at) }
func (h timed) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *timed) Push(x interface{}) {
	*h = append(*h, x.(*timer))
}

func (h *timed) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Timed represents a preemptive
// shortest time remaining scheduler
// for timers.
type Timed struct {
	timers timed
	// use actions instead of
	// seperate start/stop channels
	// to avoid timers being stopped
	// before they are started
	// (which would result in stop being ignored)
	actions chan func()
	quit    chan struct{}
}

// New creates a new preemptive shortest time remaining scheduler
// for timers.
func New() *Timed {
	t := &Timed{[]*timer{}, make(chan func(), 8192), make(chan struct{})}
	go t.scheduler()
	return t
}

// scheduler runs timers and takes care of interrupting
// timers.
func (t *Timed) scheduler() {
	actions := t.actions
	quit := t.quit
	for {
		// execute actions until there is a timer
		for len(t.timers) == 0 {
			select {
			case a := <-actions:
				a()
			case <-quit:
				return
			}
		}
		first := t.timers[0]
		timer := time.NewTimer(first.at.Sub(time.Now()))
		select {
		case a := <-actions:
			a()
			// if the first timer isn't the currently
			// running one anymore, stop the timer
			if len(t.timers) == 0 || t.timers[0] != first {
				timer.Stop()
			}
		case <-timer.C:
			heap.Pop(&t.timers)
			first.action()
		case <-quit:
			return
		}
	}
}

func (t *Timed) sendAction(a func()) {
	select {
	case t.actions <- a:
	default:
		// if the channel is blocked, launch
		// in new goroutine to avoid
		// deadlock when calling
		// Start from within Start
		// (actions send would block
		// on the scheduler goroutine,
		// which reads actions).
		// this essentially trades a panic
		// for a possible slowdown.
		go func() { t.actions <- a }()
	}
}

// stop is a function indirectly returned to the user
// that stops the respective timer and removes it from
// scheduling.
func (t *Timed) stop(tim *timer) {
	t.sendAction(func() {
		for i, v := range t.timers {
			if v == tim {
				heap.Remove(&t.timers, i)
				break
			}
		}
	})
}

// Start schedules a timer at the specified time, executing the specified
// action when done.
//
// Start also returns a function with which the timer can be stopped
// and removed from scheduling.
func (t *Timed) Start(at time.Time, action func()) (stop func()) {
	tim := &timer{at, action}
	t.sendAction(func() { heap.Push(&t.timers, tim) })
	return func() { t.stop(tim) }
}

// Close closes the scheduler and stops scheduling all pending operations.
//
// Blocks until the scheduler has really been closed and the currently running
// action has been executed.
func (t *Timed) Close() {
	// send instead of close to block until closed
	t.quit <- struct{}{}
}

// alternative scheduler implementation using the go scheduler
/*
type Entry struct {
	Key   interface{}
	Timer *time.Timer
}

type Timed struct {
	sync.Mutex
	Entries []Entry
	actions chan func()
	quit    chan struct{}
}

func (t Timed) runActions() {
	for {
		select {
		case a := <-t.actions:
			a()
		case <-t.quit:
			return
		}
	}
}

func New() Timed {
	t := Timed{
		Entries: []Entry{},
		actions: make(chan func(), 1024),
		quit:    make(chan struct{}, 1),
	}
	go t.runActions()
	return t
}

func (t *Timed) add(key interface{}, timer *time.Timer) {
	t.Entries = append(t.Entries, Entry{key, timer})
}

func (t *Timed) Index(key interface{}) (i int, ok bool) {
	for i, e := range t.Entries {
		if e.Key == key {
			return i, true
		}
	}
	return -1, false
}

func (t *Timed) Start(key interface{}, after time.Duration, action func()) {
	t.add(key, time.AfterFunc(after, func() {
		t.actions <- func() {
			action()
			t.Lock()
			defer t.Unlock()
			i, _ := t.Index(key)
			t.Remove(i)
		}
	}))
}

func (t *Timed) Remove(i int) {
	es := t.Entries
	n := len(es)
	removed := es[i]
	// remove without preserving order
	es[i] = es[n-1]
	t.Entries = es[:n-1]
	removed.Timer.Stop()
}

func (t *Timed) Close() {
	t.Lock()
	defer t.Unlock()
	for _, e := range t.Entries {
		e.Timer.Stop()
	}
	close(t.quit)
}
*/
