package service

import (
	"time"
)

type configWatcher struct {
	revision     func() (uint64, error)
	reload       func()
	stopC        chan struct{}
	stoppedC     chan struct{}
	lastRevision uint64
}

func newConfigWatcher(revision func() (uint64, error), reload func()) *configWatcher {
	rev, _ := revision()
	return &configWatcher{
		revision:     revision,
		reload:       reload,
		stopC:        make(chan struct{}, 1),
		stoppedC:     make(chan struct{}, 1),
		lastRevision: rev,
	}
}

func (w *configWatcher) start(interval uint) {
	go w.loop(interval)
}

func (w *configWatcher) stop() <-chan struct{} {
	w.stopC <- struct{}{}
	return w.stoppedC
}

func (w *configWatcher) loop(interval uint) {
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
Loop:
	for {
		select {
		case <-ticker.C:
			revision, err := w.revision()
			if err == nil && revision != w.lastRevision {
				w.lastRevision = revision
				w.reload()
			}
		case <-w.stopC:
			ticker.Stop()
			break Loop
		}
	}
	w.stoppedC <- struct{}{}
}
