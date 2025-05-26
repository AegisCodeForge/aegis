package auxfuncs

import (
	"cmp"
	"slices"
)

func SortedKeys[K cmp.Ordered, V any](m map[K]V) ([]K) {
	keys := make([]K, 0)
	for k, _ := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

