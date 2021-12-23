package fptree

import (
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

// createCondTree create conditional tree while mining
func createCondTree(trns []ConditionalPattern) (Tree, error) {

	t := InitTree()
	var err error
	log.Debugf("Number of trans inserted:%d", len(trns))

	for _, tr := range trns {
		err = t.OrderAndInsertCondTrans(tr)
		if err != nil {
			log.Debugf("Error in tree insertion:%v", err)
			return Tree{}, err
		}
	}

	return t, nil
}

// MineTree Main entry point for mining fptree
func MineTree(mtr Tree, topk int) ([]ConditionalPattern, error) {
	basePat := U.SortOnPriorityTable(mtr.CountMap, false)
	var patterns []ConditionalPattern = make([]ConditionalPattern, 0)
	for idx, patt := range basePat {
		log.Debugf("Mining base patt:%s", patt)
		if idx < topk {
			log.Debugf("taking top patt:%s at %d", patt, idx)
			pattContainer := make([]ConditionalPattern, 0)
			bp := mtr.HeadMap[patt]
			cp, err := findPrefixPath(bp.Item, bp)
			if err != nil {
				return nil, err
			}
			log.Debugf("prefixpaths:%v", cp)

			tr, err := createCondBaseTree(cp)
			if err != nil {
				return nil, err
			}
			var t ConditionalPattern
			t.Items = []string{patt}
			t.Count = mtr.CountMap[patt]
			pattContainer = append(pattContainer, t)
			err = mineBTree(tr, []string{patt}, &pattContainer)
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, pattContainer...)
		}
		var t ConditionalPattern
		t.Items = []string{patt}
		t.Count = mtr.CountMap[patt]
		patterns = append(patterns, t)

	}
	return patterns, nil
}

// createCondBaseTree based on patterns while ascending, generate a conditional fp-tree
func createCondBaseTree(trns []ConditionalPattern) (Tree, error) {

	t := InitTree()
	var err error

	for _, tr := range trns {
		for i := 0; i < tr.Count; i++ {
			err = t.OrderAndInsertTrans(tr.Items)
			if err != nil {
				log.Infof("Error in tree insertion:%v", err)
				return Tree{}, err
			}
		}
	}

	return t, nil
}

// mineBTree create condition tree and mine patterns
func mineBTree(tr Tree, pr []string, container *[]ConditionalPattern) error {
	basePat := U.SortOnPriorityTable(tr.CountMap, false)
	for _, p := range basePat {
		base := make([]string, 0)
		for _, s := range pr {
			base = append(base, s)
		}
		base = append(base, p)
		var t ConditionalPattern
		t.Items = base
		t.Count = tr.CountMap[p]
		if t.Count > 10 {
			*container = append(*container, t)
		}
		condPatt, err := findPrefixPath(p, tr.HeadMap[p])
		if err != nil {
			return err
		}
		cTr, err := createCondTree(condPatt)
		if err != nil {
			return err
		}
		if len(cTr.CountMap) > 0 {
			mineBTree(cTr, base, container)
		}

	}

	return nil
}

// findPrefixPath traverse the prefix paths in treee while mining to get conditional patterns
func findPrefixPath(basePat string, treeNode *node) ([]ConditionalPattern, error) {
	condPattern := make([]ConditionalPattern, 0)
	treeNodeMap := make(map[*node]bool)
	for treeNode != nil {
		prefixPath := make([]string, 0)
		ascendFpTree(treeNode, &prefixPath)
		if len(prefixPath) > 1 {
			var c ConditionalPattern
			c.Items = append(c.Items, prefixPath[1:]...)
			c.Count = treeNode.Counter
			condPattern = append(condPattern, c)
		}
		treeNode = treeNode.AuxNode
		if _, ok := treeNodeMap[treeNode]; !ok {
			treeNodeMap[treeNode] = true
		} else {
			log.Errorf("found cycle while ascending")
			break
		}
	}
	return condPattern, nil
}

// ascendFpTree ascend tree from leaf node to root
func ascendFpTree(n *node, prefixPath *[]string) {
	if n.ParentNode != nil {
		*prefixPath = append(*prefixPath, n.Item)
		ascendFpTree(n.ParentNode, prefixPath)
	}
}
