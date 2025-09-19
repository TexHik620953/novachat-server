package safemap

import "sync"

// Thread safe map
type Safemap[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, val V)
	Exists(key K) bool
	Foreach(it func(K, V))
	Count() int
	Remove(key K)
}

type safemapImpl[K comparable, V any] struct {
	data  map[K]V
	mutex sync.RWMutex
}

// Thread safe map
func New[K comparable, V any]() Safemap[K, V] {
	return &safemapImpl[K, V]{
		data:  make(map[K]V),
		mutex: sync.RWMutex{},
	}
}

func (h *safemapImpl[K, V]) Get(key K) (V, bool) {
	h.mutex.RLock()
	v, ex := h.data[key]
	h.mutex.RUnlock()
	return v, ex
}

func (h *safemapImpl[K, V]) Set(key K, val V) {
	h.mutex.Lock()
	h.data[key] = val
	h.mutex.Unlock()
}

func (h *safemapImpl[K, V]) Remove(key K) {
	h.mutex.Lock()
	delete(h.data, key)
	h.mutex.Unlock()
}

func (h *safemapImpl[K, V]) Exists(key K) bool {
	h.mutex.Lock()
	_, ex := h.data[key]
	h.mutex.Unlock()
	return ex
}

func (h *safemapImpl[K, V]) Foreach(it func(K, V)) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	for k, v := range h.data {
		it(k, v)
	}
}

func (h *safemapImpl[K, V]) Count() int {
	return len(h.data)
}
