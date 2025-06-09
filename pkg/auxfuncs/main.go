package auxfuncs

import (
	"cmp"
	"math/rand"
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

const passchdict = "abcdefghijklmnopqrstuvwxyz0123456789-_"
func GenSym(n int) string {
	res := make([]byte, 0)
	for _ = range n {
		res = append(res, passchdict[rand.Intn(len(passchdict))])
	}
	return string(res)
}
