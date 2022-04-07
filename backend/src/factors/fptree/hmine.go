package fptree

import (
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Algorithm implemented : H-mine from spmf
// https://www.philippe-fournier-viger.com/spmf/Hmine.pdf

/*	the entry function for hmine algo
	sort the rows of transaction based on count
	build header table based on from higest to lowest priority
	have a container to collect all counts
	do a depth first search starting on item highest support_count

	T_0 : item_1,item_2, item_3
	T_1 : item_1,item_4, item_5,item_8
	T_2 : item_1,item_3, item_9,item_10,item_6
	T_3 : item_1,item_7, item_6
	total number of transactions : 4
	total number of items : 10
	count of items : item_1:4 , item_2:1 , item_3:2, item_4:1, item_5:1,item_6:2,item_7:1,item_8:1,item_9:1,item_10:1
	sort each transaction based on counts
	create header table of item counts
	link the first item in transaction to head_table
		- the idea is the the most priority item will be found only in the first position and the item wih second most
		  priority can be in first or second position and so on
	starting from largest item recurse depth wise to count n-length patterns
	once an item is counted link the second item in transaction to the header table
*/

func CountAndWriteResultsFromPattern(trns [][]string, fname string, support_percentage float32, count_all int, topk_patterns []int) (int, int, FpContainer, error) {
	// the entry function to the counting logic, properties are given as list
	// fnameOfProps : path containing  properties
	// fnameRes : path to which results are Writen
	// support_percentage : percentage below which items are filtered
	// count_all : -1 if all items are to be caputured , no ranking and cutoff
	// topk_patterns : list containing the number of items to be counted for each pattern length

	var support_count int
	var ProcessResults FpContainer
	ProcessResults.Fpts = make([]FPatternCount, 0)

	if len(trns) == 0 {
		return 0, 0, FpContainer{}, nil
	}
	encData, decoder := EncodeData(trns)
	support_count = computeSupport(support_percentage, len(encData))
	trnsSorted := ResortTrans(encData, support_count)
	results, err := Hmine(trnsSorted, support_count)
	if err != nil {
		log.Fatal("unable to compute Hmine")
		return -1, 0, FpContainer{}, err
	}
	ProcessResults, err = ProcessHmineResults(decoder, results, count_all, topk_patterns)
	if err != nil {
		log.Errorf("unable to decode results:%v", err)
		return -1, 0, FpContainer{}, err
	}

	if len(ProcessResults.Fpts) > 0 {

		err = WriteHmineResultsToFile(fname, ProcessResults)
		if err != nil {
			log.Fatalf("unable to write hmine results:%v", err)
			return -1, 0, FpContainer{}, err
		}
		log.Infof("written results to file :%d, %s, %d", len(ProcessResults.Fpts), fname, len(trns))
	}

	return len(trns), len(ProcessResults.Fpts), ProcessResults, nil
}

func CountAndWriteResultsToFile(fnameOfProps, fnameRes string, support_percentage float32, count_all int, topk_patterns []int) (int, int, FpContainer, error) {
	// the entry function to the counting logic, read properties from file
	// fnameOfProps : path containing  properties
	// fnameRes : path to which results are Writen
	// support_percentage : percentage below which items are filtered
	// count_all : -1 if all items are to be caputured , no ranking and cutoff
	// topk_patterns : list containing the number of items to be counted for each pattern length

	var support_count int
	var trns [][]string = make([][]string, 0)
	var ProcessResults FpContainer
	ProcessResults.Fpts = make([]FPatternCount, 0)
	if _, err := os.Stat(fnameOfProps); err == nil {
		trns, err = ReadFileSliceStrings(fnameOfProps)
		if err != nil {
			return -1, 0, FpContainer{}, fmt.Errorf("Unable to read properties from file :%s", fnameOfProps)
		}
		if len(trns) == 0 {
			return 0, 0, FpContainer{}, nil
		}
		// log.Debugf("read properties from file : %d", len(trns))
		encData, decoder := EncodeData(trns)

		support_count = computeSupport(support_percentage, len(encData))
		trnsSorted := ResortTrans(encData, support_count)

		results, err := Hmine(trnsSorted, support_count)
		if err != nil {
			log.Fatal("unable to compute Hmine")
			return -1, 0, FpContainer{}, err
		}
		ProcessResults, err = ProcessHmineResults(decoder, results, count_all, topk_patterns)
		if err != nil {
			log.Errorf("unable to decode results:%v", err)
			return -1, 0, FpContainer{}, err
		}

		if len(ProcessResults.Fpts) > 0 {

			err = WriteHmineResultsToFile(fnameRes, ProcessResults)
			if err != nil {
				log.Fatalf("unable to write hmine results:%v", err)
				return -1, 0, FpContainer{}, err
			}
			log.Debugf("written results to file :%d, %s, %d", len(ProcessResults.Fpts), fnameRes, len(trns))
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		log.Infof("Properties file does not exist: %s", fnameOfProps)
		return 0, 0, FpContainer{}, nil

	}
	return len(trns), len(ProcessResults.Fpts), ProcessResults, nil
}

func Hmine(db [][]string, support_count int) (FpContainer, error) {

	log.Debugf("inside hmine:%d", len(db))
	flist, err := CreateInitFmap(db, support_count)
	if err != nil {
		return FpContainer{}, err
	}
	dbSorted := ResortTransOnPri(db, flist.Pri)
	pDb := BuildHDB(dbSorted)
	var fpatterns FpContainer
	fpatterns.FptMap = make(map[int][]FPatternCount)
	htable, err := InitHmine(pDb, support_count, flist)
	if err != nil {
		return FpContainer{}, err
	}
	TotalItemCount := len(flist.Count)

	if len(flist.Count) == 0 {
		return FpContainer{}, nil
	}

	for idx := 0; idx < TotalItemCount; idx++ {
		countingItem := flist.InvPri[idx]
		prefix := make([]string, 0)
		prefix = append(prefix, countingItem)
		err := HmineHelper(prefix, htable, flist, support_count, &fpatterns, flist)
		if err != nil {
			return FpContainer{}, err
		}
		if stopHmine(flist.Count, &fpatterns) {
			return fpatterns, nil
		}
		err = SecondaryLink(prefix[len(prefix)-1], htable, flist, support_count)
		if err != nil {
			return FpContainer{}, err
		}

		ct := make(map[string]int)
		for k, _ := range flist.Pri {
			t, _ := CountHlink(k, htable.Hmap[k].HyperLink)
			ct[k] = t
		}
	}

	for _, v := range fpatterns.FptMap {
		fpatterns.Fpts = append(fpatterns.Fpts, v...)
	}

	log.Infof("number of patterns computed in hmine :%d", len(fpatterns.Fpts))
	return fpatterns, nil
}

func InitHmine(pdb PrjDB, support_count int, fPriorityMap Fmap) (HeadTable, error) {
	header_table := createInitialHmap(pdb, fPriorityMap)
	return header_table, nil
}

func CreateInitFmap(pdb [][]string, support_count int) (Fmap, error) {
	// create initial Fmap of counts of items and filter the items
	// based on support_count

	var fPriorityMap Fmap
	fPriorityMap.Count = make(map[string]int)
	fPriorityMap.Pri = make(map[string]int)
	fPriorityMap.InvPri = make(map[int]string)

	// count each item in transaction
	for _, trn := range pdb {
		for _, itm := range trn {
			fPriorityMap.Count[itm] += 1
		}
	}

	// remove keys from prioritymap based on counts
	err := filterSetPriority(fPriorityMap, support_count)
	if err != nil {
		return Fmap{}, err
	}
	return fPriorityMap, nil
}

func createInitialHmap(pdb PrjDB, fPriorityMap Fmap) HeadTable {
	// create  an intialize an initial header map
	var fpMap = make(map[string]*HeaderItm)
	var head_prev = make(map[string]*PrjTrans)
	var head_table HeadTable
	var pos int = 0

	for itmKey, _ := range fPriorityMap.Count {
		var headerItem HeaderItm
		headerItem.ItemId = itmKey
		headerItem.SupportCount = 0
		headerItem.HyperLink = nil
		fpMap[itmKey] = &headerItem
	}

	// traverse DB and fill HTable
	for _, txn := range pdb.DB {
		for idxItm, itm := range txn.Trans {
			if val, ok := fpMap[itm.ItemId]; ok {
				if idxItm == pos {
					val.SupportCount += 1

					if val.HyperLink == nil {
						val.HyperLink = txn
					}
					if txtHlink, ok := head_prev[val.ItemId]; !ok {
						head_prev[val.ItemId] = txn
					} else {
						txtHlink.Trans[pos].HLink = txn
						head_prev[itm.ItemId] = txn
					}
				}
			}
		}
	}
	head_table.Hmap = fpMap

	// fill the last transaction from hprev
	for ky, val := range head_table.Hmap {
		if v, ok := head_prev[ky]; ok {
			if val.HyperLink != v {
				val.LastTr = v
			}
		}
	}

	return head_table

}

func SecondaryLink(itm string, htb HeadTable, fl Fmap, support_count int) error {
	// once item is counted, pass through all the transactions for each current item link the next item bases on priority
	tmpCount := make(map[string]int)

	plink := htb.Hmap[itm].HyperLink
	hprev := make(map[string]*PrjTrans)
	tpm := make(map[string]int)

	// create and intialize a hprev
	for k, v := range htb.Hmap {
		if v.HyperLink != nil {
			if v.LastTr == nil {
				v.LastTr = v.HyperLink
			}
		}
		if v != nil {
			hprev[k] = v.LastTr
		}
	}

	for plink != nil {
		// log.Debugf("sec linking processing  trans :%v", PrintTrans(plink))
		prTrns := plink.Trans
		var idx_itm int
		var itm_prev string
		for idx, tr := range prTrns {

			if strings.Compare(tr.ItemId, itm) == 0 {
				// to find position of tr in transaction
				itm_prev = tr.ItemId
				idx_itm = idx
			}

			if (idx-idx_itm) == 1 && strings.Compare(tr.ItemId, itm) != 0 && strings.Compare(itm_prev, itm) == 0 && fl.Count[tr.ItemId] >= support_count {
				nextItm := tr
				tpm[nextItm.ItemId] += 1

				if htb.Hmap[nextItm.ItemId].HyperLink == nil {
					// init item for first time
					htb.Hmap[nextItm.ItemId].HyperLink = plink
					htb.Hmap[nextItm.ItemId].SupportCount += 1
					hprev[nextItm.ItemId] = plink
				}

				if hprev[nextItm.ItemId] != nil && hprev[nextItm.ItemId] != plink {
					nLink := hprev[nextItm.ItemId]
					htb.Hmap[nextItm.ItemId].SupportCount += 1
					for _, v := range nLink.Trans {
						if strings.Compare(v.ItemId, nextItm.ItemId) == 0 {
							v.HLink = plink
							hprev[nextItm.ItemId] = plink
						}
					}
				}
			}
		}

		plink = SetNextLink(itm, prTrns)
	}

	for k, v := range hprev {
		htb.Hmap[k].LastTr = v
	}

	for k, v := range htb.Hmap {
		if v.SupportCount > 0 {
			tmpCount[k] = v.SupportCount
		}
	}

	return nil
}
