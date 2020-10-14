package utils

/* This code is from github
Implementation of Range Map
*/

import
(
"sort"
)

type Range struct {
	L int
	U int
}

type RangeMap struct {
	Keys []Range
	Values []string
}

// works like Guava RangeMap com.google.common.collect.ImmutableRangeMap
// https://github.com/google/guava/blob/1efc5ce983c02cfaa2ce3338bdfaa1f583b66128/guava/src/com/google/common/collect/ImmutableRangeMap.java#L159
func (rm RangeMap) Get(key int) (string, bool) {
	i := sort.Search(len(rm.Keys), func(i int) bool {
		//fmt.Printf("search %v at index %d for %v is %v\n", rm.Keys[i], i, key, key < rm.Keys[i].L)
		return key < rm.Keys[i].L
	})

	i -= 1
	if i >= 0 && i < len(rm.Keys) && key <= rm.Keys[i].U {
		return rm.Values[i], true
	}
	return "", false
}