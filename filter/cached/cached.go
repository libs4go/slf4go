package cached

import (
	"sync"

	"github.com/libs4go/scf4go"
	"github.com/libs4go/slf4go"
)

type cachedBackend struct {
	sync.Mutex
	sync.Once
	syncCond *sync.Cond
	backend  slf4go.Backend
	cached   chan *slf4go.EventEntry
	closed   bool
	exitLoop bool
	filter   *cachedFilter
	initOnce sync.Once
}

func (cached *cachedFilter) newCachedBackend(backend slf4go.Backend) slf4go.Backend {

	wrapper := &cachedBackend{
		backend: backend,
		filter:  cached,
	}

	wrapper.syncCond = sync.NewCond(wrapper)

	return wrapper
}

func (cached *cachedBackend) notifyExit() {
	cached.Lock()
	defer cached.Unlock()

	cached.exitLoop = true
	cached.syncCond.Broadcast()

}

func (cached *cachedBackend) sendLoop() {
	for {
		if evt, ok := <-cached.cached; ok {
			cached.backend.Send(evt)
			continue
		}

		cached.notifyExit()
		return
	}
}

func (cached *cachedBackend) Config(config scf4go.Config) error {
	return cached.backend.Config(config)
}

func (cached *cachedBackend) createChan() {
	cached.initOnce.Do(func() {
		cached.Lock()
		defer cached.Unlock()
		cached.cached = make(chan *slf4go.EventEntry, cached.filter.cachedSize)
		go cached.sendLoop()
	})
}

func (cached *cachedBackend) Send(entry *slf4go.EventEntry) {

	if cached.closed {
		return
	}

	cached.createChan()

	defer func() {
		if err := recover(); err != nil {
			cached.closed = true
		}
	}()

	cached.cached <- entry
}

func (cached *cachedBackend) Sync() {

	cached.Lock()
	defer cached.Unlock()

	if cached.cached == nil {
		return
	}

	cached.Do(func() {
		close(cached.cached)
		cached.closed = true
	})

	if cached.exitLoop {
		return
	}

	cached.syncCond.Wait()
}

type cachedFilter struct {
	cachedSize int
}

func (cached *cachedFilter) Name() string {
	return "cached"
}

func (cached *cachedFilter) Config(config scf4go.Config) {
	cached.cachedSize = config.Get("size").Int(1000)
}

func (cached *cachedFilter) MakeChain(backend slf4go.Backend) slf4go.Backend {
	return cached.newCachedBackend(backend)
}

func init() {
	slf4go.RegisterFilter(&cachedFilter{
		cachedSize: 1000,
	})
}
