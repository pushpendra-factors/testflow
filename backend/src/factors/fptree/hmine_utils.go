package fptree

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

type kv struct {
	Key   string
	Value int
}

func filterSetPriority(fc Fmap, support int) error {
	for k, v := range fc.Count {
		if v < support {
			delete(fc.Count, k)
		}
	}

	prItmList := SortOnPriorityTable(fc.Count, false)
	for idx, p := range prItmList {
		fc.Pri[p] = idx
		fc.InvPri[idx] = p
	}

	if len(fc.Count) != len(fc.InvPri) || len(fc.InvPri) != len(fc.Pri) {
		log.Fatal("fmaps are not equal")
		return fmt.Errorf("fmaps are not equal")
	}
	return nil
}

func filterSetPrioritySecondary(fc Fmap, fmain Fmap, support int) {

	if &fc == nil {
		return
	}

	itmList := make([]string, 0)

	for k, v := range fc.Count {
		if v < support {
			delete(fc.Count, k)
		}
		itmList = append(itmList, k)

	}

	if len(fc.Count) > 0 {
		prItmList := SortOnPriority(itmList, fc.Pri, true)
		for idx, p := range prItmList {
			fc.Pri[p] = idx
			fc.InvPri[idx] = p
			// log.Infof("counts in hmine help :%s,%d,%d", p, fmain.Pri[p], idx)
		}

		for k, _ := range fc.Pri {
			if _, ok := fc.Count[k]; !ok {
				delete(fc.Pri, k)
			}
		}

		for k, v := range fc.InvPri {
			if _, ok := fc.Count[v]; !ok {
				delete(fc.InvPri, k)
			}
		}
	}
	if len(fc.Count) != len(fc.InvPri) || len(fc.InvPri) != len(fc.Pri) {
		log.Fatal("fmaps are not equal")
	}
	// log.Debugf("secondary fmap fcount after priority :%v", fc)

}

func BuildHTransaction(Trans []string) PrjTrans {
	var pjt PrjTrans
	pjt.Trans = make([]*PrjTransItm, len(Trans))
	for idx, itm := range Trans {

		var pti PrjTransItm

		pti.ItemId = itm
		pti.Count = 1
		pti.HLink = nil
		pjt.Trans[idx] = &pti
	}
	// log.Debugf("Inserted transaction : %v", Trans)
	return pjt
}

func BuildHDB(allTrans [][]string) PrjDB {
	// input : all transaction items
	// output: projected db where transactions are sorted based on counts
	// create a projected DB where each transactions is sorted based on item couts
	var pjDB PrjDB
	pjDB.DB = make([]*PrjTrans, 0)

	for _, trn := range allTrans {
		pjTrn := BuildHTransaction(trn)
		pjDB.DB = append(pjDB.DB, &pjTrn)
	}
	return pjDB

}

func PrintTrans(plink *PrjTrans) []string {
	// stdout a transaction with all items
	if plink == nil {
		return []string{}
	}
	itms := make([]string, 0)
	for _, a := range plink.Trans {
		itms = append(itms, a.ItemId)
	}
	return itms
}

func CountHlink(itm string, lk *PrjTrans) (int, error) {
	// log.Debugf("counting link: %s", itm)
	infLoop := make(map[*PrjTrans]int)
	plink := lk
	lcount := 0
	for plink != nil {

		if _, ok := infLoop[plink]; ok {
			// log.Fatalf("visiting trans again ,%s-> %v", itm, PrintTrans(plink))
			// return -1, fmt.Errorf("visiting trans again")

			for _, v := range plink.Trans {
				if strings.Compare(v.ItemId, itm) == 0 {
					v.HLink = nil
				}
			}
		} else {
			infLoop[plink] = 1
		}

		for _, itmCurr := range plink.Trans {
			if strings.Compare(itmCurr.ItemId, itm) == 0 {
				// log.Debugf("looping :%v", PrintTrans(plink))
				plink = itmCurr.HLink
				lcount += 1
			}
		}
	}
	return lcount, nil
}

func SetNextLink(pItem string, pTrans []*PrjTransItm) *PrjTrans {

	var r *PrjTrans
	for _, pi := range pTrans {
		if strings.Compare(pi.ItemId, pItem) == 0 {
			r = pi.HLink
		}
	}
	return r
}

func GetNextItemInPriority(CurrItm string, fmp Fmap) (string, int) {

	CurrPri := fmp.Pri[CurrItm]
	if CurrPri+1 > len(fmp.Count) {
		return "", -1
	} else {
		return fmp.InvPri[CurrPri+1], CurrPri + 1
	}

}

func verifyloop(htb HeadTable, s string) {
	chckItm := "$salesforce_account_ownerid::0051V000007lF8wQAE"
	// log.Debugf("start of check :%s", s)
	if h, ok := htb.Hmap[chckItm]; ok {
		_, err := CountHlink(h.ItemId, h.HyperLink)
		// log.Debugf("count of in %s checking is %s ->%d", s, chckItm, n)
		if err != nil {
			// log.Debugf("in loop %s", s)
			log.Fatal(err)
		}

	} else {
	}
}

// ConvertHmineFptreeContainer read patterns from hmine container adapt it for fptree struct
func ConvertHmineFptreeContainer(hm FpContainer) []ConditionalPattern {
	dst := make([]ConditionalPattern, 0)
	for _, h := range hm.Fpts {
		var ds ConditionalPattern
		ds.Items = h.FpItm
		ds.Count = h.FpCounts
		dst = append(dst, ds)
	}
	return dst
}

// EncodeData read transaction and create a pseudo name for counting
func EncodeData(db [][]string) ([][]string, map[string]string) {
	// items starting with similar counts and same name will have issues while assigning priority.
	// hence rewriting the encoding the item name.names of items can also stretch to multiple paragraph.
	// hence gnerate a dummy_name with no priority clashes
	map_a := make(map[string]int)
	encoder := make(map[string]string)
	decoder := make(map[string]string)

	dbEncoded := make([][]string, 0)
	for _, tr := range db {
		for _, it := range tr {
			map_a[it] += 1
		}
	}
	cnt := 0
	for k, _ := range map_a {
		label := fmt.Sprintf("a_%d", cnt)
		encoder[k] = label
		decoder[label] = k
		cnt += 1
	}

	for _, tr := range db {
		t := make([]string, 0)
		for _, it := range tr {
			itmEnc := encoder[it]
			t = append(t, itmEnc)
		}
		dbEncoded = append(dbEncoded, t)
	}

	return dbEncoded, decoder
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

		if ss[i].Value != ss[j].Value {
			return ss[i].Value > ss[j].Value
		} else {
			tmp := strings.Compare(ss[i].Key, ss[j].Key)
			if tmp == 1 {
				return true
			} else {
				return false
			}
		}

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

func ResortTrans(trns [][]string, support int) [][]string {
	countMap := make(map[string]int)
	res := make([][]string, 0)
	for _, tr := range trns {
		for _, itm := range tr {
			countMap[itm] += 1
		}
	}

	for _, tr := range trns {
		var y []string
		r := SortOnPriority(tr, countMap, false)
		for _, sortedItem := range r {
			if val, ok := countMap[sortedItem]; ok {
				if val > support {
					y = append(y, sortedItem)
				}
			}
		}
		if len(y) > 0 {
			res = append(res, y)
			// log.Infof("re-sorted trns :%v", r)
		}
	}

	lenSorted := sortOnLength(res)
	return lenSorted
}

func WritePropertiesToFile(fname string, props [][]string) error {

	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)

	for _, lineWrite := range props {
		var dt FpHProperties
		dt.Prop = lineWrite

		eventDetailsBytes, _ := json.Marshal(dt)
		lineWrite := string(eventDetailsBytes)

		if _, err := w.WriteString(fmt.Sprintf("%s\n", lineWrite)); err != nil {
			log.WithFields(log.Fields{"line": lineWrite, "err": err}).Error("Unable to write to file.")
			return err
		}

	}

	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func WriteSinglePropertyToFile(fname string, prop []string) error {
	file, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)

	var dt FpHProperties
	dt.Prop = prop

	eventDetailsBytes, _ := json.Marshal(dt)
	lineWrite := string(eventDetailsBytes)

	if _, err := w.WriteString(fmt.Sprintf("%s\n", lineWrite)); err != nil {
		log.WithFields(log.Fields{"line": lineWrite, "err": err}).Error("Unable to write to file.")
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func WriteHmineResultsToFile(fname string, res FpContainer) error {
	log.Infof("writting results to file :%d %s", len(res.Fpts), fname)
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)

	for _, pts := range res.Fpts {
		patternBytes, _ := json.Marshal(pts)
		lineWrite := string(patternBytes)

		if _, err := w.WriteString(fmt.Sprintf("%s\n", lineWrite)); err != nil {
			log.WithFields(log.Fields{"line": lineWrite, "err": err}).Error("Unable to write to file.")
			return err
		}

	}

	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func computeSupport(support_percentage float32, numTrans int) int {

	totSupport := int(float32(numTrans) * support_percentage)
	log.Debugf("num of trans:%d support_percent:%f totalSupport :%d", numTrans, support_percentage, totSupport)
	return totSupport
}

func ProcessHmineResults(decoder map[string]string, rawresults FpContainer, count_all int, topK []int) (FpContainer, error) {
	// rank results based on counts
	// topk_patterns : list containing the number of items to be counted for each pattern length
	// count_all : -1 for no ranking
	var processedResults FpContainer
	processedResults.Fpts = make([]FPatternCount, 0)
	processedResults.FptMap = make(map[int][]FPatternCount)

	for _, pts := range rawresults.FptMap {

		for _, v := range pts {
			if len(v.FpItm) < MAX_LENGTH && v.FpCounts > 0 {

				var f FPatternCount
				l := len(v.FpItm)
				f.FpItm = make([]string, l)

				for idx, itm := range v.FpItm {

					if dec, ok := decoder[itm]; ok {
						f.FpItm[idx] = dec

					} else {
						err := fmt.Errorf("while writing results data not available in decoder  :%s", itm)
						return FpContainer{}, err
					}
				}
				f.FpCounts = v.FpCounts
				if _, ok := processedResults.FptMap[l]; ok {
					processedResults.FptMap[l] = append(processedResults.FptMap[l], f)
				} else {
					processedResults.FptMap[l] = make([]FPatternCount, 0)
					processedResults.FptMap[l] = append(processedResults.FptMap[l], f)
				}
			}
		}

	}

	if count_all == 1 {
		for kv, mv := range processedResults.FptMap {
			pr := getTopHminePatterns(mv, topK[kv])
			processedResults.Fpts = append(processedResults.Fpts, pr...)
			log.Infof("inside process hmine results add :%d,%d", kv, len(pr))

		}
	} else {

		for _, mv := range processedResults.FptMap {
			processedResults.Fpts = append(processedResults.Fpts, mv...)
		}
	}

	log.Infof("inside process hmine results computes :%d", len(processedResults.Fpts))
	return processedResults, nil

}

func CreateScannerFromReader(r io.Reader) *bufio.Scanner {
	const MAX_PATTERN_BYTES = 20 * 1024 * 1024
	scanner := bufio.NewScanner(r)
	// Adjust scanner buffer capacity to MAX_PATTERN_BYTES per line.
	buf := make([]byte, MAX_PATTERN_BYTES)
	scanner.Buffer(buf, MAX_PATTERN_BYTES)
	return scanner
}
func ReadFileSliceStrings(fname string) ([][]string, error) {

	fileName := fname

	f, err := os.Open(fileName)
	if err != nil {
		log.Errorf("error opening file:%s", fileName)
		return nil, err
	}
	defer f.Close()
	unItem := make(map[string]int)

	scanner := CreateScannerFromReader(f)
	tmp := make([][]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		var eventDetails FpHProperties
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Read failed")
			return nil, err
		}
		tmp = append(tmp, eventDetails.Prop)

		for _, e := range eventDetails.Prop {
			unItem[e] += 1
		}
	}
	return tmp, nil

}

func getTopHminePatterns(fpts []FPatternCount, topKPatterns int) []FPatternCount {

	if topKPatterns == 0 {
		return []FPatternCount{}
	}

	if len(fpts) > 0 && topKPatterns > 1 {
		sort.Slice(fpts, func(i, j int) bool { return fpts[i].FpCounts > fpts[j].FpCounts })
		if len(fpts) > topKPatterns {
			return fpts[0:topKPatterns]
		}
		return fpts
	}
	return fpts

}

func ReadHminePatterns(fname string) (FpContainer, error) {

	var rawPatterns FpContainer
	rawPatterns.Fpts = make([]FPatternCount, 0)
	fileName := fname

	if _, err := os.Stat(fname); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Infof("Patterns file does not exist:%s", fname)
			return FpContainer{}, nil
		}
	}

	f, err := os.Open(fileName)
	if err != nil {
		log.Errorf("error opening file:%s", fileName)
		return FpContainer{}, err
	}
	defer f.Close()

	scanner := CreateScannerFromReader(f)
	for scanner.Scan() {
		line := scanner.Text()
		var patternDetails FPatternCount
		if err := json.Unmarshal([]byte(line), &patternDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Read failed")
			return FpContainer{}, err
		}
		rawPatterns.Fpts = append(rawPatterns.Fpts, patternDetails)

	}
	return rawPatterns, nil

}

func ReadAndConvertHminePatterns(fname string) ([]ConditionalPattern, error) {
	results, err := ReadHminePatterns(fname)
	if err != nil {
		e := fmt.Errorf("unable to read hmine patters from results file:%s", fname)
		log.Error(e)
		return nil, e

	}
	pattern_res := ConvertHmineFptreeContainer(results)
	log.Debugf("read %d HuserPatters from file %s , total user properties pattens %d", len(results.Fpts), fname, len(pattern_res))
	return pattern_res, nil
}

type ft struct {
	key []string
	val int
}

func sortOnLength(trns [][]string) [][]string {

	var trnItms []ft = make([]ft, 0)
	var res [][]string = make([][]string, 0)

	for _, t := range trns {
		tmp := ft{t, len(t)}
		trnItms = append(trnItms, tmp)
	}

	sort.Slice(trnItms, func(i, j int) bool { return trnItms[i].val > trnItms[j].val })

	for _, t := range trnItms {
		res = append(res, t.key)
	}
	return res
}

func CheckAlignment(fname string, eventToverify string) (bool, error) {

	support_count := 1
	var eventDecoded string
	trns, err := ReadFileSliceStrings(fname)
	log.Infof("total number of transaction : %d", len(trns))
	if err != nil {
		return false, fmt.Errorf("Unable to read properties from file :%s", fname)
	}
	if len(trns) == 0 {
		return true, nil
	}
	encData, decoder := EncodeData(trns)
	for k, v := range decoder {
		if strings.Compare(v, eventToverify) == 0 {
			eventDecoded = k
		}
	}
	fm, err := CreateInitFmap(encData, support_count)
	if err != nil {
		return false, fmt.Errorf("unable to create fmap")
	}
	dbSorted := ResortTransOnPri(encData, fm.Pri)

	pDb := BuildHDB(dbSorted)
	log.Infof("eventPri :%v", fm.Pri)
	eventPri := fm.Pri[eventDecoded]
	log.Infof("event Priority: %s,%s,%d", eventToverify, eventDecoded, eventPri)
	for _, tr := range pDb.DB {
		var tmpPriSlice []int = make([]int, 0)
		for _, itm := range tr.Trans {

			tmpPri := fm.Pri[itm.ItemId]
			tmpPriSlice = append(tmpPriSlice, tmpPri)
		}
		log.Infof("priority:%v", tmpPriSlice)
	}
	log.Infof("Done with Testing")
	return true, nil

}

func ResortTransOnPri(trns [][]string, priMap map[string]int) [][]string {
	res := make([][]string, 0)
	for _, tr := range trns {
		var y []string
		r := SortOnPriority(tr, priMap, true)
		for _, sortedItem := range r {
			if _, ok := priMap[sortedItem]; ok {
				y = append(y, sortedItem)
			}
		}
		res = append(res, y)
	}

	return res
}

func FilterTwoLenPattern(fpc *FpContainer, fr *FPatternCount) {
	// check if atleast one item in len one patterns

	lenOnepts := make(map[string]bool)
	lenOneptsList := fpc.FptMap[1]
	for _, pt := range lenOneptsList {
		lenOnepts[pt.FpItm[0]] = true
	}
	addFlag := make([]string, 0)
	for _, itm := range fr.FpItm {
		if _, ok := lenOnepts[itm]; ok {
			addFlag = append(addFlag, itm)
		}
	}
	if len(addFlag) == 0 {
		fr.FpCounts = 0
	}
}
