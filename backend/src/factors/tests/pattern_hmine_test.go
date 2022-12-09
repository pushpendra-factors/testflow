package tests

import (
	fp "factors/fptree"
	"fmt"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestBuildPrDB(t *testing.T) {

	trns_a := []string{"A", "D", "C", "B"}
	trns_b := []string{"C", "B", "E"}
	trns_c := []string{"D", "E"}
	trns_d := []string{"D", "C"}
	trns_e := []string{"A", "E"}
	trns_f := []string{"A", "F"}
	support_count := 2

	trns := [][]string{trns_a, trns_b, trns_c, trns_d, trns_d, trns_e, trns_f}
	flist, err := fp.CreateInitFmap(trns, support_count)

	hdb := fp.BuildHDB(trns)
	assert.Equal(t, len(hdb.DB), len(trns))
	htable, err := fp.InitHmine(hdb, support_count, flist)
	assert.Nil(t, err)
	fmt.Println(flist.Count)
	assert.Equal(t, len(htable.Hmap), len(flist.Count))
}

func TestHmineHelper(t *testing.T) {

	trns_a := []string{"A", "D", "C", "B"}
	trns_b := []string{"C", "B", "E"}
	trns_c := []string{"D", "E"}
	trns_d := []string{"D", "C"}
	trns_e := []string{"A", "E"}
	trns_f := []string{"A", "F"}
	trns_g := []string{"A", "F", "C"}

	support_count := 2

	trns := [][]string{trns_a, trns_b, trns_c, trns_d, trns_e, trns_f, trns_g}
	flist, err := fp.CreateInitFmap(trns, support_count)

	hdb := fp.BuildHDB(trns)
	assert.Equal(t, len(hdb.DB), len(trns))
	htable, err := fp.InitHmine(hdb, support_count, flist)
	assert.Nil(t, err)

	assert.Equal(t, len(htable.Hmap), len(flist.Count))
	prefix := []string{"A"}
	var f fp.FpContainer
	f.Fpts = make([]fp.FPatternCount, 0)
	f.FptMap = make(map[int][]fp.FPatternCount)

	err = fp.HmineHelper(prefix, htable, flist, 1, &f, flist)
	assert.Nil(t, err)
	fcountMap := make(map[string]int)
	fcountMap["A"] = 4
	fcountMap["AD"] = 1
	fcountMap["AE"] = 1
	fcountMap["AF"] = 2
	fcountMap["ADB"] = 1
	for _, ky := range f.FptMap {
		for _, fc := range ky {
			fstr := strings.Join(fc.FpItm, "")
			assert.Equal(t, fcountMap[fstr], fc.FpCounts)
		}
	}
}

func TestCountHlink(t *testing.T) {

	trns_a := []string{"A", "D", "C", "B"}
	trns_b := []string{"C", "B", "E"}
	trns_c := []string{"D", "E"}
	trns_d := []string{"D", "C"}
	trns_e := []string{"A", "E"}
	trns_f := []string{"A", "F"}
	trns_g := []string{"A", "D", "C"}

	support_count := 0

	trns := [][]string{trns_a, trns_b, trns_c, trns_d, trns_e, trns_f, trns_g}
	f, err := fp.CreateInitFmap(trns, support_count)

	hdb := fp.BuildHDB(trns)
	assert.Equal(t, len(hdb.DB), len(trns))
	htable, err := fp.InitHmine(hdb, support_count, f)
	assert.Nil(t, err)
	count, err := fp.CountHlink("A", htable.Hmap["A"].HyperLink)
	fmt.Println("count A :", count)
	assert.Nil(t, err)
	assert.Equal(t, count, 4)
	count, err = fp.CountHlink("D", htable.Hmap["D"].HyperLink)
	fmt.Println("count D :", count)
	assert.Nil(t, err)
	assert.Equal(t, count, 2)
}

func TestSecondaryLink(t *testing.T) {

	trns_a := []string{"A", "D", "C", "B"}
	trns_b := []string{"C", "B", "E"}
	trns_c := []string{"D", "E"}
	trns_d := []string{"D", "C"}
	trns_e := []string{"A", "E"}
	trns_f := []string{"A", "F"}
	trns_g := []string{"A", "D", "C"}

	support_count := 0

	trns := [][]string{trns_a, trns_b, trns_c, trns_d, trns_e, trns_f, trns_g}
	trns_sorted := fp.ResortTrans(trns, support_count)
	f, err := fp.CreateInitFmap(trns_sorted, support_count)
	assert.Nil(t, err)
	hdb := fp.BuildHDB(trns_sorted)

	assert.Equal(t, len(hdb.DB), len(trns))
	htable, err := fp.InitHmine(hdb, support_count, f)
	assert.Nil(t, err)
	err = fp.SecondaryLink("A", htable, f, support_count)
	assert.Nil(t, err)
	count_d, err := fp.CountHlink("D", htable.Hmap["D"].HyperLink)
	assert.Nil(t, err)
	assert.Equal(t, 4, count_d)
	count_e, err := fp.CountHlink("E", htable.Hmap["E"].HyperLink)
	assert.Nil(t, err)
	assert.Equal(t, 1, count_e)

	err = fp.SecondaryLink("D", htable, f, support_count)
	assert.Nil(t, err)
	count_d, err = fp.CountHlink("D", htable.Hmap["D"].HyperLink)
	assert.Nil(t, err)
	assert.Equal(t, 4, count_d)
	count_e, err = fp.CountHlink("E", htable.Hmap["E"].HyperLink)
	assert.Nil(t, err)
	assert.Equal(t, 2, count_e)

	count_c, err := fp.CountHlink("C", htable.Hmap["C"].HyperLink)
	assert.Nil(t, err)
	assert.Equal(t, 4, count_c)

}

func TestHmine(t *testing.T) {

	trns_a := []string{"A", "D", "C", "B"}
	trns_b := []string{"C", "B", "E"}
	trns_c := []string{"D", "E"}
	trns_d := []string{"D", "C"}
	trns_e := []string{"A", "E"}
	trns_f := []string{"A", "F"}
	trns_g := []string{"A", "D", "C"}

	trns := [][]string{trns_a, trns_b, trns_c, trns_d, trns_e, trns_f, trns_g}
	support_count := 0
	trnsSorted := fp.ResortTrans(trns, support_count)
	_, err := fp.Hmine(trnsSorted, support_count)
	assert.Nil(t, err)

}

func createMapFromContainer(fpc fp.FpContainer) map[string]int {
	countsMap := make(map[string]int)
	for _, ptns := range fpc.FptMap {
		for _, pt := range ptns {
			pts := strings.Join(pt.FpItm, "_")
			cnt := pt.FpCounts
			countsMap[pts] = cnt
		}
	}
	return countsMap
}

func TestHmine2(t *testing.T) {

	trns_a := []string{"A", "D", "C", "B"}
	trns_b := []string{"C", "B", "E"}
	trns_c := []string{"D", "E"}
	trns_d := []string{"D", "C"}
	trns_e := []string{"A", "E"}
	trns_f := []string{"A", "F"}
	trns_g := []string{"A", "D", "C"}
	trns_g2 := []string{"A", "D", "C", "E"}
	trns_g3 := []string{"D", "C", "E", "B"}

	trns := [][]string{trns_a, trns_b, trns_c, trns_d, trns_e, trns_f, trns_g, trns_g2, trns_g3}
	itemCounts := make(map[string]int)
	itemCounts["A"] = 5
	itemCounts["B"] = 3
	itemCounts["C"] = 6
	itemCounts["D"] = 6
	itemCounts["E"] = 5
	itemCounts["F"] = 1
	itemCounts["A_B"] = 1
	itemCounts["C_B"] = 3
	itemCounts["A_F"] = 1
	support_count := 0
	trnsSorted := fp.ResortTrans(trns, support_count)
	log.Info(trnsSorted)
	fpc, err := fp.Hmine(trnsSorted, support_count)
	fpcMap := createMapFromContainer(fpc)

	for k, item_val := range itemCounts {
		if _, ok := fpcMap[k]; ok {
			assert.Equal(t, item_val, fpcMap[k], k)
		}
	}

	assert.Nil(t, err)

}

func TestSortOnPriority(t *testing.T) {

	strMap := make(map[string]int)
	strMap["B"] = 2
	strMap["A"] = 2
	strMap["C"] = 3
	strMap["D"] = 4
	strMap["E"] = 5
	strMap["K"] = 2

	p := fp.SortOnPriorityTable(strMap, false)
	fmt.Println(p)
	assert.Equal(t, len(p), 6)

}

func getCountFromTransTable(trns map[int]map[string]int, pat []string) int {

	num_trans := 0

	for _, single_tran := range trns {
		l := 0
		for _, p := range pat {
			if _, ok := single_tran[p]; ok {
				l += 1
			}
		}
		if l == len(pat) {
			num_trans += 1
		}
	}
	return num_trans
}

func TestReadDataAndCompute2(t *testing.T) {
	// get properties from file and write it to file
	// can be replaced with anyother properties file to verify results

	fname := "data/hmine_test_data.json"
	fname_res := "data/res.txt"
	topk := []int{1, 20, 20, 20}
	trans_map := make(map[int]map[string]int)
	support := 1
	data, err := fp.ReadFileSliceStrings(fname)
	for idx, tr := range data {
		tmpMap := make(map[string]int)
		for _, t := range tr {
			tmpMap[t] = 1
		}
		trans_map[idx] = tmpMap
	}

	assert.Nil(t, err)
	encDb, enc := fp.EncodeData(data)
	trnsSorted := fp.ResortTrans(encDb, support)
	fres_raw, err := fp.Hmine(trnsSorted, support)
	assert.Nil(t, err)
	assert.Greater(t, len(fres_raw.Fpts), 0)
	fres_proc, err := fp.ProcessHmineResults(enc, fres_raw, -1, topk)
	assert.Nil(t, err)
	err = fp.WriteHmineResultsToFile(fname_res, fres_proc)
	assert.Nil(t, err)

	for k, itm_patterns := range fres_proc.FptMap {
		if k < 3 {
			for _, pt := range itm_patterns {
				num_trans := getCountFromTransTable(trans_map, pt.FpItm)
				s := fmt.Sprintf("counts not matching : %v", pt.FpItm)
				assert.Equal(t, pt.FpCounts, num_trans, s)
			}
		}

	}

}
