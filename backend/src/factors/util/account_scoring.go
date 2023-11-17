package util

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func GetDateOnlyFromTimestamp(ts int64) string {
	year, month, date := time.Unix(ts, 0).UTC().Date()
	data := fmt.Sprintf("%d%02d%02d", year, month, date)
	return data
}

func GenDateStringsForLastNdays(currDate int64, salewindow int64) []string {

	dateStrings := make([]string, 0)

	currts := time.Unix(currDate, 0)
	dstring := GetDateOnlyFromTimestamp(currDate)
	dateStrings = append(dateStrings, dstring)

	for idx := salewindow; idx > 0; idx-- {
		ts := currts.AddDate(0, 0, -1*int(idx))
		dstring := GetDateOnlyFromTimestamp(ts.Unix())
		dateStrings = append(dateStrings, dstring)
	}
	return dateStrings
}

func ComputeDayDifference(ts1 int64, ts2 int64) int {

	t1 := time.Unix(ts1, 0)
	t2 := time.Unix(ts2, 0)
	daydiff := t2.Sub(t1).Seconds() / float64(24*60*60)
	return int(math.Abs(daydiff))

}

func GetDateFromString(ts string) int64 {
	year := ts[0:4]
	month := ts[4:6]
	date := ts[6:8]
	t, _ := strconv.ParseInt(year, 10, 64)
	t1, _ := strconv.ParseInt(month, 10, 64)
	t2, _ := strconv.ParseInt(date, 10, 64)
	t3 := time.Date(int(t), time.Month(t1), int(t2), 0, 0, 0, 0, time.UTC).Unix()
	return t3
}

type keyVal struct {
	Key   string
	Value float64
}

func TakeTopKOnSortedmap(wordFreq map[string]float64, topk int) map[string]float64 {
	resultMap := make(map[string]float64)
	resultString := SortOnFreq(wordFreq, true)
	log.Debugf("result string --- %v", resultString)

	for idx := 0; idx < topk; idx++ {
		wrd := resultString[idx]
		log.Debugf("wrd --- %s", wrd)
		resultMap[wrd] = wordFreq[wrd]
	}

	return resultMap
}

func SortOnFreq(pq map[string]float64, desending bool) []string {
	ll := make([]string, len(pq))
	idx := 0
	for k, _ := range pq {
		ll[idx] = k
		idx += 1
	}
	log.Debugf("top ll :%v", ll)
	return SortOnFreqList(ll, pq, desending)
}

func SortOnFreqList(ll []string, pq map[string]float64, desending bool) []string {
	// sort the string-ints in ascending order
	//https://play.golang.org/p/fpv2lVpzVCO
	res := make([]string, len(ll))

	var ss []keyVal
	ss = make([]keyVal, 0)
	for _, k := range ll {
		v := pq[k]
		ss = append(ss, keyVal{k, v})
	}

	sort.SliceStable(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	if !desending {
		for i, kv := range ss {
			res[i] = kv.Key
		}
	} else {
		for i := len(ss) - 1; i >= 0; i-- {
			res[i] = ss[i].Key
		}
	}

	return res
}
