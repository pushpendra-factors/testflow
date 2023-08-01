package util

import (
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

func RemoveByindex(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func RemoveString(s []string, c string) []string {

	if len(s) > 0 {

		for i, p := range s {
			if strings.Compare(p, c) == 0 {
				return RemoveByindex(s, i)
			}
		}
	}
	return s
}

func In(strList []string, str string) bool {
	s := make(map[string]bool)

	for _, v := range strList {
		s[v] = false
	}
	_, ok := s[str]
	return ok

}

type kv struct {
	Key   string
	Value int
}

func SortOnPriorityTable(pq map[string]int, desending bool) []string {
	ll := make([]string, 0)
	for k, _ := range pq {
		ll = append(ll, k)
	}
	return SortOnPriority(ll, pq, desending)
}

func SortOnPriority(ll []string, pq map[string]int, desending bool) []string {
	// sort the string-ints in ascending order
	//https://play.golang.org/p/fpv2lVpzVCO
	res := make([]string, 0)

	var ss []kv
	for _, k := range ll {
		v := pq[k]
		ss = append(ss, kv{k, v})
	}

	sort.SliceStable(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	if !desending {
		for _, kv := range ss {
			res = append(res, kv.Key)
		}
	} else {
		for i := len(ss) - 1; i >= 0; i-- {
			res = append(res, ss[i].Key)
		}
	}

	return res
}

func RankByWordFreq(wordFrequencies map[string]int) ([]string, map[string]int) {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))

	wl := make([]string, len(pl))
	wlMap := make(map[string]int)
	for _, v := range pl {
		wl = append(wl, v.Key)
		wlMap[v.Key] = v.Value
	}
	return wl, wlMap
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func CheckUniqueTrans(trns []string) bool {
	trnsMap := make(map[string]bool, 0)

	for _, tr := range trns {

		if _, ok := trnsMap[tr]; ok {
			return false
		} else {
			trnsMap[tr] = true
		}
	}
	return true
}

func MakeUniqueTrans(trns []string) []string {
	trnsMap := make(map[string]bool, 0)
	trnsSet := make([]string, 0)
	for _, tr := range trns {
		trnsMap[tr] = true
	}
	for k, _ := range trnsMap {
		trnsSet = append(trnsSet, k)
	}
	return trnsSet
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	c := string(b) + strconv.FormatInt(time.Now().UnixNano(), 32)
	return c
}

func PrintStats(mem runtime.MemStats) {
	runtime.ReadMemStats(&mem)
	fmt.Println("mem.Alloc:", mem.Alloc)
	fmt.Println("mem.TotalAlloc:", mem.TotalAlloc)
	fmt.Println("mem.HeapAlloc:", mem.HeapAlloc)
	fmt.Println("mem.NumGC:", mem.NumGC)
	fmt.Println("-----")
}
