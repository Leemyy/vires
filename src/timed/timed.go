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
	action func(actual time.Time)
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

func (h *timed) Remove(t *timer) (int, bool) {
	for i, v := range *h {
		if v == t {
			heap.Remove(h, i)
			return i, true
		}
	}
	return -1, false
}

// Timed represents a preemptive
// shortest time remaining scheduler
// for timers.
type Timed struct {
	timers     chan timed
	newFirst   chan *timer
	removeLast chan struct{}
	quit       chan struct{}
}

// New creates a new preemptive shortest time remaining scheduler
// for timers.
func New() *Timed {
	timed := make(chan timed, 1)
	timed <- []*timer{}
	// small buffer size for both transmission channels
	// to leave a bit of room for Start() with the
	// earliest timer to remove being spammed faster
	// than schedule() can handle it (so the caller of Start
	// doesn't block unless 10 Start()s are waiting to replace
	// the first element) and to leave room for Start() - stop() spam
	// with only a single element so the caller of stop()
	// doesn't block unless 10 stop()s are waiting to
	// tell schedule() that the last element was removed.
	// ... an alternative would be to launch all
	// calls in a seperate goroutine, but this kind of optimization
	// is up to the caller (downside: request spam in some context might
	// kill the GC through too many goroutines being spawned
	// while a limit of 10 would just block the spam if it's too much)
	t := &Timed{timed, make(chan *timer, 10), make(chan struct{}, 10), make(chan struct{})}
	go t.scheduler()
	return t
}

// takeFirst takes the first available timer
// and eventually blocks until a timer is available
func (t *Timed) takeFirst() *timer {
	timers := <-t.timers
	if len(timers) == 0 {
		t.timers <- timers
		// wait for first element to arrive
		return <-t.newFirst
	}
	// first element is present, take it
	first := timers[0]
	t.timers <- timers
	return first
}

// scheduler runs timers and takes care of interrupting
// timers.
func (t *Timed) scheduler() {
	first := t.takeFirst()
	for {
		timer := time.NewTimer(first.at.Sub(time.Now()))
		select {
		case first = <-t.newFirst:
			// a new timer has been set as first, replace first preemptively
			timer.Stop()
		case <-t.removeLast:
			// the last timer has been removed from externally, take the next first
			timer.Stop()
			first = t.takeFirst()
		case actual := <-timer.C:
			// the first timer has expired, execute its action
			first.action(actual)
			// acquire t.timers after action to allow for recursive calls to Timed.Start
			// (which would otherwise deadlock)
			timers := <-t.timers
			// remove instead of pop - pop might remove a newly added first timer
			// that was added after the timer expired and before t.timers was locked
			// (moving the lock before first.action doesn't solve the issue either
			// because the goroutine might get unscheduled right after <-timer.C
			// and also increases lock time to the time taken by first.action,
			// which might be very long)

			// ignore remove return because if first wasn't found then it was already
			// removed from externally while action was executing
			timers.Remove(first)
			t.timers <- timers
			first = t.takeFirst()
		case <-t.quit:
			timer.Stop()
			return
		}
	}
}

// stop is a function indirectly returned to the user
// that stops the respective timer, removes it from
// scheduling and returns whether the timer
// was already stopped.
func (t *Timed) stop(tim *timer) bool {
	timers := <-t.timers
	defer func() { t.timers <- timers }()
	i, ok := timers.Remove(tim)
	switch {
	case !ok:
		return false
	// we removed the first and the first was the last one remaining
	case i == 0 && len(timers) == 0:
		t.removeLast <- struct{}{}
		return true
	// we removed the first
	case i == 0:
		first := timers[0]
		t.newFirst <- first
		return true
	}
	return true
}

// Start schedules a timer at the specified time, executing the specified
// action when done.
//
// actions parameter actual contains the actual time
// action was run at, as opposed to when it was scheduled.
//
// Start also returns a function with which the timer can be stopped
// and removed from scheduling. The return value of that functions
// determines whether that timer was already scheduled.
func (t *Timed) Start(at time.Time, action func(actual time.Time)) (stop func() bool) {
	tim := &timer{at, action}
	timers := <-t.timers
	if len(timers) == 0 || at.Before(timers[0].at) {
		// temporarily release t.timers.
		// if t.newFirst blocks because we are Start()ing
		// more early timers than schedule() can handle
		// then t.timers would be locked by this goroutine and
		// to unblock t.newFirst we might have to acquire
		// t.timers after executing the action of a timer.
		// this goroutine would wait for t.newFirst to
		// unblock and the scheduling goroutine would wait
		// for t.timers to be released - a deadlock occurs.
		t.timers <- timers
		t.newFirst <- tim
		timers = <-t.timers
	}
	heap.Push(&timers, tim)
	t.timers <- timers
	return func() bool { return t.stop(tim) }
}

// Close closes the scheduler and stops scheduling all pending operations.
//
// Blocks until the scheduler has really been closed and the currently running
// action has been executed.
func (t *Timed) Close() {
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
