package errors

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type frame struct {
	caller  uintptr
	file    string
	line    int
	callers []uintptr
}

type iFrame struct {
	i uint64
	f frame
}

type locationFrames struct {
	m  map[uint64]frame
	ch chan iFrame
	mu sync.RWMutex
}

func (lf *locationFrames) put(i uint64, f frame) {
	go func() {
		lf.flush()
	}()
	lf.ch <- iFrame{i: i, f: f}
}

func (lf *locationFrames) flush() {
	for {
		select {
		case f := <-lf.ch:
			lf.mu.Lock()
			lf.m[f.i] = f.f
			lf.mu.Unlock()
		default:
			return
		}
	}
}

func (lf *locationFrames) get(i uint64) frame {
	lf.flush()
	lf.mu.RLock()
	defer lf.mu.RUnlock()
	return lf.m[i]
}

func (lf *locationFrames) del(i uint64) {
	lf.flush()
	lf.mu.Lock()
	defer lf.mu.Unlock()
	delete(lf.m, i)
}

var lFrames = locationFrames{
	m:  make(map[uint64]frame),
	ch: make(chan iFrame, 32),
}

var lFramesCounter uint64

type location uint64

func (l *location) Location() (string, int) {
	return lFrames.get(uint64(*l)).file, lFrames.get(uint64(*l)).line
}

func (l *location) SetLocation(skip int) {
	i := atomic.AddUint64(&lFramesCounter, 1)
	var f frame
	f.caller, f.file, f.line, _ = runtime.Caller(skip + 1)

	pc := make([]uintptr, 32)
	n := runtime.Callers(skip+2, pc)
	f.callers = pc[0:n]

	*l = location(i)
	lFrames.put(i, f)

	runtime.SetFinalizer(l, func(l *location) {
		lFrames.del(uint64(*l))
	})
}

func (l *location) Caller() uintptr { return lFrames.get(uint64(*l)).caller }

func (l *location) Callers() []uintptr { return lFrames.get(uint64(*l)).callers }
