package structure

type BiMap[K comparable, V comparable] struct {
	forward  map[K]V
	backward map[V]K
}

func NewBiMap[K comparable, V comparable]() *BiMap[K, V] {
	return &BiMap[K, V]{
		forward:  make(map[K]V),
		backward: make(map[V]K),
	}
}

func (b *BiMap[K, V]) Put(key K, value V) {
	if oldVal, ok := b.forward[key]; ok {
		delete(b.backward, oldVal)
	}
	if oldKey, ok := b.backward[value]; ok {
		delete(b.forward, oldKey)
	}
	b.forward[key] = value
	b.backward[value] = key
}

func (b *BiMap[K, V]) GetValue(key K) (V, bool) {
	val, ok := b.forward[key]
	return val, ok
}

func (b *BiMap[K, V]) GetKey(value V) (K, bool) {
	key, ok := b.backward[value]
	return key, ok
}

func (b *BiMap[K, V]) Delete(key K) {
	if val, ok := b.forward[key]; ok {
		delete(b.forward, key)
		delete(b.backward, val)
	}
}

func (b *BiMap[K, V]) Len() int {
	return len(b.forward)
}
