package game

import "time"

type Timed map[interface{}]*time.Timer

func NewTimed() Timed {
	return make(map[interface{}]*time.Timer)
}

func (t Timed) Start(key interface{}, after time.Duration, action func()) {
	t[key] = time.AfterFunc(after, action)
}

func (t Timed) Remove(key interface{}) {
	timer, ok := t[key]
	if ok {
		delete(t, key)
		timer.Stop()
	}
}

func (t Timed) Halt() {
	for _, timer := range t {
		timer.Stop()
	}
}
