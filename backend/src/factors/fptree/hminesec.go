package fptree

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

const MAX_PATTERNS = 20000
const MAX_LENGTH = 4

func HmineHelper(prefix []string, htb HeadTable, fc Fmap, support_count int, fpc *FpContainer, fmain Fmap) error {

	prefixItemId := prefix[len(prefix)-1]
	if stopHmine(fc.Count, fpc) {
		log.Debugf("stopping recursion,%d", prefix)
		return nil
	}

	fpmap_helper, err := initFmapHelper(prefix, htb, fc, support_count, fmain)
	if err != nil {
		return err
	}

	if _, ok := htb.Hmap[prefixItemId]; !ok {
		return nil
	}

	if len(prefix) < MAX_LENGTH {
		var FrPattern FPatternCount
		FrPattern.FpItm = prefix
		FrPattern.FpCounts = fc.Count[prefixItemId]

		if len(FrPattern.FpItm) == 2 {
			FilterTwoLenPattern(fpc, &FrPattern)
		}

		if FrPattern.FpCounts > 0 {
			i := len(FrPattern.FpItm)
			if _, ok := fpc.FptMap[i]; ok {
				fpc.FptMap[i] = append(fpc.FptMap[i], FrPattern)
			} else {
				fpc.FptMap[i] = make([]FPatternCount, 0)
				fpc.FptMap[i] = append(fpc.FptMap[i], FrPattern)
			}
		}
	}

	htbItem := htb.Hmap[prefixItemId]

	htbNew := buildHtableHelper(fpmap_helper, htbItem)
	if len(fpmap_helper.Count) == 0 && len(htb.Hmap) == 0 {
		return nil
	}

	tmpLog_a := make([]string, 0)
	for k, _ := range htbNew.Hmap {
		tmpLog_a = append(tmpLog_a, k)
	}

	totalCountOfSubitems := len(fpmap_helper.Count) //check for of by one error
	for prIdx := 0; prIdx < totalCountOfSubitems; prIdx++ {
		tmpPrefix := make([]string, 0)
		for _, pi := range prefix {
			tmpPrefix = append(tmpPrefix, pi)
		}

		prItm := fpmap_helper.InvPri[prIdx]
		tmpPrefix = append(tmpPrefix, prItm)
		// log.Debugf("mining subseq :%v,%d", tmpPrefix, len(fpc.Fpts))

		if len(tmpPrefix) > MAX_LENGTH {
			return nil
		}

		err := HmineHelper(tmpPrefix, htbNew, fpmap_helper, support_count, fpc, fmain)
		if err != nil {
			return err
		}

		if stopHmine(fpmap_helper.Count, fpc) == true {
			return nil
		}

		err = SecondaryLinkSubseq(tmpPrefix, prItm, htbNew, fc, support_count)
		if err != nil {
			return err
		}

	}

	return nil
}

// initFmapHelper create head table and frequency map
func initFmapHelper(prefix []string, htb HeadTable, fc Fmap, support_count int, fmain Fmap) (Fmap, error) {
	// create frequency map for hmine helper

	var fcNew Fmap
	fcNew.Count = make(map[string]int)
	fcNew.Pri = make(map[string]int)
	fcNew.InvPri = make(map[int]string)
	var htbItem *HeaderItm
	var itmPriority int

	itmString := prefix[len(prefix)-1]

	htbItem = htb.Hmap[itmString]
	if htbItem == nil {
		return Fmap{}, nil
	}
	itmPriority = fmain.Pri[itmString]
	plink := htbItem.HyperLink
	loopDetect := make(map[*PrjTrans]int)

	for plink != nil {
		if _, ok := loopDetect[plink]; ok {
			for _, v := range plink.Trans {
				if strings.Compare(v.ItemId, itmString) == 0 {
					v.HLink = nil
				}
			}
		} else {
			loopDetect[plink] = 1
		}
		ftr := plink.Trans
		for _, itm := range ftr {
			itmTranString := itm.ItemId
			if itmP, ok := fmain.Pri[itmTranString]; ok {
				if itmP > itmPriority {
					fcNew.Count[itmTranString] += 1
				}
			}
		}
		for _, v := range plink.Trans {
			if strings.Compare(v.ItemId, itmString) == 0 {
				plink = v.HLink
			}
		}
	}

	filterSetPrioritySecondary(fcNew, fc, support_count)

	//--------checks before passing to build stage-------
	// assert prefix items not in new fmap
	for _, prefixItem := range prefix {
		if _, ok := fcNew.Count[prefixItem]; ok {
			log.Fatalf("prefix item in new fmap:%s,%v", prefixItem, fcNew.Count)
		}
		if _, ok := fcNew.Pri[prefixItem]; ok {
			log.Fatalf("prefix item in new fmap priority:%s,%v", prefixItem, fcNew.Pri)
		}
	}

	// log.Debugf("new fmap after counting :%v", fc)

	return fcNew, nil
}

func buildHtableHelper(fmp Fmap, htbItem *HeaderItm) HeadTable {
	// log.Debugf("building subsequence header table:%s", htbItem.ItemId)

	var htb HeadTable
	htb.Hmap = make(map[string]*HeaderItm)

	hprev := make(map[string]*PrjTrans)
	pItem := htbItem.ItemId
	pLink := htbItem.HyperLink

	for pLink != nil {

		pTrans := pLink.Trans
		for _, pi := range pTrans {

			if _, ok := fmp.Count[pi.ItemId]; ok {
				// log.Debugf("pi item:%s", pi.ItemId)

				if _, ok := htb.Hmap[pi.ItemId]; !ok {
					var itmHmp HeaderItm
					itmHmp.ItemId = pi.ItemId
					itmHmp.SupportCount = 1
					itmHmp.HyperLink = pLink
					itmHmp.LastTr = nil

					htb.Hmap[pi.ItemId] = &itmHmp
					hprev[pi.ItemId] = pLink
				} else {
					if hmPrev, ok := hprev[pi.ItemId]; ok {
						nxTrans := hmPrev.Trans
						for _, prItm := range nxTrans {
							if strings.Compare(prItm.ItemId, pi.ItemId) == 0 {
								htb.Hmap[pi.ItemId].SupportCount += 1
								prItm.HLink = pLink
								// prItm.HLink = pLink
								hprev[pi.ItemId] = pLink
							}
						}
					}
				}
				break
			}
		}
		pLink = SetNextLink(pItem, pTrans)

	}

	for k, v := range hprev {
		htb.Hmap[k].LastTr = v
	}
	for k, v := range htb.Hmap {
		_, err := CountHlink(k, v.HyperLink)
		if err != nil {
			return HeadTable{}
		}
		// log.Debugf("item in htb %s,%d,%d", k, v.SupportCount, chl)

	}
	// log.Debugf("No loop till now :%d", len(htb.Hmap))
	return htb
}

func SecondaryLinkSubseq(prefix []string, itm string, htb HeadTable, fl Fmap, support_count int) error {
	// log.Debugf("Secondary link counting for itm:%v,%s", prefix, itm)

	tmpLog_a := make([]string, 0)
	for k, _ := range htb.Hmap {
		tmpLog_a = append(tmpLog_a, k)
	}
	// log.Debugf("item in new htb in sec subseq:%v,%v,%v", tmpLog_a, fl.Count, fl.Pri)

	tmpCount := make(map[string]int)

	if _, ok := htb.Hmap[itm]; !ok {
		return nil
	}

	plink := htb.Hmap[itm].HyperLink
	Hprev := make(map[string]*PrjTrans)
	tpm := make(map[string]int)
	nextItem, _ := GetNextItemInPriority(itm, fl)
	// log.Debugf("sec subseq for itm:%s nxtItem:%s curr_pri:%d ,next_pri:%d", itm, nextItem, fl.Pri[itm], nextPri)
	if _, ok := htb.Hmap[nextItem]; !ok {
		return nil
	}
	for k, v := range htb.Hmap {
		if strings.Compare(k, nextItem) == 0 {
			if v.HyperLink != nil {
				if v.LastTr == nil {
					v.LastTr = v.HyperLink
				}
			}
			if v != nil {
				Hprev[k] = v.LastTr
			}
		}
	}

	for plink != nil {
		// log.Debugf("sec linking processing  trans :%v", PrintTrans(plink))
		prTrns := plink.Trans
		var idx_itm int
		// var itm_prev string
		for idx, tr := range prTrns {

			if strings.Compare(tr.ItemId, itm) == 0 {
				// to find position of tr in transaction
				// itm_prev = tr.ItemId
				idx_itm = idx
			}

			if (idx-idx_itm) > 0 && strings.Compare(tr.ItemId, nextItem) == 0 {
				NextItm := tr
				tpm[NextItm.ItemId] += 1
				// log.Debugf("linking item :%s", NextItm.ItemId)
				if htb.Hmap[NextItm.ItemId].HyperLink == nil {
					htb.Hmap[NextItm.ItemId].HyperLink = plink
					htb.Hmap[NextItm.ItemId].SupportCount += 1
					Hprev[NextItm.ItemId] = plink
					// log.Debugf("setting first time htb :%s", NextItm.ItemId)
				}

				if Hprev[NextItm.ItemId] != nil && Hprev[NextItm.ItemId] != plink {
					nLink := Hprev[NextItm.ItemId]
					htb.Hmap[NextItm.ItemId].SupportCount += 1
					for _, v := range nLink.Trans {
						if strings.Compare(v.ItemId, NextItm.ItemId) == 0 {
							v.HLink = plink
							Hprev[NextItm.ItemId] = plink
						}
					}
				}
			}
		}

		plink = SetNextLink(itm, prTrns)
	}

	for k, v := range Hprev {
		htb.Hmap[k].LastTr = v
	}

	for k, v := range htb.Hmap {
		if v.SupportCount > 0 {
			tmpCount[k] = v.SupportCount
		}
	}
	// log.Debugf("support in htb :%v", tmpCount)

	return nil
}

func stopHmine(f map[string]int, fc *FpContainer) bool {
	// stop hmine counting based on various conditions
	// stop if MAX_PATTERNS is reached
	// stop counting

	if len(f) == 0 {
		return true
	}

	if len(fc.Fpts) > MAX_PATTERNS {
		return true
	}
	for _, v := range f {
		if v > 0 {
			return false
		}
	}

	return true
}
