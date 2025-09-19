package linq

func ToList[K comparable, V any, T any](data map[K]V, selector func(K, V) T) []T {
	r := make([]T, len(data))
	c := 0
	for k, v := range data {
		r[c] = selector(k, v)
		c++
	}
	return r
}
