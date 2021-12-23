package tests

import (
	fp "factors/fptree"
	"fmt"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetFarAncestor(t *testing.T) {
	tr := fp.InitTree()
	trns := []string{"A", "B", "C", "D", "E"}
	err := tr.InsertItemsIntoTree(trns)
	assert.Nil(t, err, "err should be nil")
	eNode := tr.HeadMap["E"]
	nd, lenPath, err := tr.GetFarAncestor(eNode, []string{"A", "B", "C", "D"})
	assert.Nil(t, err, "error in GetFarAncestor")
	assert.Equal(t, nd.Item, "A", "unable to reach ancestor")
	assert.Equal(t, lenPath, 4, "len of path should be 4")
}

func TestCreateFpTree(t *testing.T) {
	tree := fp.InitTree()
	assert.NotNil(t, tree)
}

func TestInsertItemsFpTree(t *testing.T) {
	tree := fp.InitTree()
	words := []string{"ABCD", "ABCD", "ABCD"}

	for _, w := range words {
		trns := strings.Split(w, "")
		log.Infof("Insering transaction :%v", trns)
		err := tree.InsertItemsIntoTree(trns)
		assert.Nil(t, err, "error in Insert Item")
	}

}

func TestUpdateValueMap(t *testing.T) {
	tr := fp.InitTree()
	tr.ValueMap[1] = []string{"A", "B", "C"}
	tr.ValueMap[2] = []string{"D", "E", "F", "H"}

	losserString := tr.UpdateValueMap(2, "E")
	assert.ElementsMatch(t, losserString, []string{"D", "F", "H"}, "Update value map")

}

func TestUpdateHeaderTable(t *testing.T) {
	tr := fp.InitTree()
	n1 := fp.InitNode("A", 1)
	n2 := fp.InitNode("A", 1)
	n3 := fp.InitNode("A", 1)

	err := tr.UpdateHeaderTable(&n1)
	assert.Nil(t, err, "error in update header table")
	assert.Nil(t, n1.AuxNode, "Aux node should be nil , as upddating header table")
	err = tr.UpdateHeaderTable(&n2)
	assert.Nil(t, err, "error in update header table for n2")
	assert.NotNil(t, n1.AuxNode, "Aux node should now be set , as upddating tail map")
	assert.NotNil(t, tr.TailMap["A"], "Aux node should now be set , as upddating header table")
	assert.Nil(t, n2.AuxNode, "Aux node should now be set , as upddating tail map")
	err = tr.UpdateHeaderTable(&n3)
	assert.Nil(t, err, "error in update header table for n2")
	assert.NotNil(t, n2.AuxNode, "Aux node should now be set , as upddating tail map")
	assert.Nil(t, n3.AuxNode, "Aux node should not be set , as upddating tail map")

}

func TestInsertItems(t *testing.T) {
	tr := fp.InitTree()
	trns := []string{"A", "B", "C", "D"}
	err := tr.InsertItemsIntoTree(trns)
	assert.Nil(t, err, "Error int insertItems to tree")
	assert.Equal(t, 4, len(tr.HeadMap), "number of items inserted")
	trns = []string{"B", "A", "C", "D"}
	err = tr.InsertItemsIntoTree(trns)
	assert.Nil(t, err, "Error int insertItems to tree")
	assert.Equal(t, 4, len(tr.HeadMap), "number of items inserted into head")
	assert.Equal(t, 4, len(tr.TailMap), "number of items inserted into tail")
	for _, v := range tr.HeadMap {
		assert.NotNil(t, v.AuxNode, "Aux node is not set")
	}
	for k, vTail := range tr.TailMap {
		vHead := tr.HeadMap[k]
		assert.NotNil(t, vHead.AuxNode, vTail, "Aux nodes should be equal")
	}

	for _, vTail := range tr.TailMap {
		assert.NotNil(t, vTail.ParentNode, "All parent nodes are set")
	}

}

func TestGetPreviousNode(t *testing.T) {
	tr := fp.InitTree()
	tr1 := []string{"A", "B", "C", "D"}
	tr2 := []string{"B", "C", "A", "D"}
	tr3 := []string{"B", "D", "A", "F"}
	tr4 := []string{"K", "L", "A", "C"}
	err := tr.InsertItemsIntoTree(tr1)
	assert.Nil(t, err)
	currA := tr.HeadMap["A"]
	err = tr.InsertItemsIntoTree(tr2)
	assert.Nil(t, err)
	prevC := tr.TailMap["C"]
	err = tr.InsertItemsIntoTree(tr3)
	assert.Nil(t, err)
	err = tr.InsertItemsIntoTree(tr4)
	assert.Nil(t, err)
	currC := tr.TailMap["C"]

	nd, err := tr.GetPreviousNode(currC)
	assert.Nil(t, err)
	assert.Equal(t, prevC, nd, "unable to reach prev node")
	nd, err = tr.GetPreviousNode(currA)
	assert.Nil(t, err)
	assert.Nil(t, nd)
}

func TestInsertTreeSimple(t *testing.T) {
	words1 := []string{"ABCD", "AD", "AC", "AK", "DKL"}
	words2 := []string{"ABCD", "ABDE", "AD", "ADF", "DFG", "GA", "GB", "GC", "AB", "KT"}
	words3 := []string{"DFG", "GA", "GC", "DA", "DB"}
	words4 := []string{"ABCDE", "KLME", "BF", "H", "KCT", "JKIN", "SUV", "BZ"}
	wordsColl := [][]string{words1, words2, words3, words4}
	for _, words := range wordsColl {
		tr := fp.InitTree()

		log.Info("=====new trans===")
		for _, w := range words {
			trns := strings.Split(w, "")
			log.Infof("Insering transaction :%v", trns)
			err := tr.OrderAndInsertTrans(trns)
			assert.Nil(t, err, fmt.Sprintf("unable to insert transaction:%v", words))
		}
	}

}

func TestSerializeTree(t *testing.T) {
	words := []string{"ABCD", "AD", "AC", "AK", "DKL"}
	tr := fp.InitTree()

	for _, word := range words {
		trns := strings.Split(word, "")
		err := tr.OrderAndInsertTrans(trns)
		assert.Nil(t, err, fmt.Sprintf("unable to insert transaction:%v", words))
	}
	fmt.Println("*****------****", tr.Root)
	_, err := tr.Serialize()
	assert.Nil(t, err)
	// fmt.Println(val_node_string)
}

func TestSerializeTree2(t *testing.T) {
	words := []string{"ABCD", "AD", "AC", "AK", "CK", "DL"}
	tr := fp.InitTree()

	for _, word := range words {
		trns := strings.Split(word, "")
		err := tr.OrderAndInsertTrans(trns)
		assert.Nil(t, err, fmt.Sprintf("unable to insert transaction:%v", words))
	}
	fmt.Println("*****------****", tr.Root)
	_, err := tr.Serialize()
	assert.Nil(t, err)
	// fmt.Println(val_node_string)
}

func TestSerializeTreeTofile(t *testing.T) {
	words := []string{"ABCD", "AD", "AC", "AK", "CK", "DL"}
	tr := fp.InitTree()

	for _, word := range words {
		trns := strings.Split(word, "")
		err := tr.OrderAndInsertTrans(trns)
		assert.Nil(t, err, fmt.Sprintf("unable to insert transaction:%v", words))
	}
	nodeString, err := tr.Serialize()
	assert.Nil(t, err)

	fname := "/tmp/fptree/test2.txt"
	err = fp.WriteTreeToFile(fname, nodeString)
	assert.Nil(t, err)
}

func TestReadTreeFromFile(t *testing.T) {
	words := []string{"ABCD", "AD", "AC", "AK", "CK", "DL"}
	tr := fp.InitTree()

	for _, word := range words {
		trns := strings.Split(word, "")
		err := tr.OrderAndInsertTrans(trns)
		assert.Nil(t, err, fmt.Sprintf("unable to insert transaction:%v", words))
	}
	nodeString, err := tr.Serialize()
	assert.Nil(t, err)

	fname := "/tmp/fptree/test2.txt"
	err = fp.WriteTreeToFile(fname, nodeString)
	assert.Nil(t, err)
	nodes, err := fp.ReadNodesFromFile(fname)
	assert.Nil(t, err)
	assert.Greater(t, len(nodes), 0)

}

func TestFptreeDeserialize(t *testing.T) {
	words := []string{"ABCD", "AD", "AC", "AK", "CK", "DL"}
	tr := fp.InitTree()

	for _, word := range words {
		trns := strings.Split(word, "")
		err := tr.OrderAndInsertTrans(trns)
		assert.Nil(t, err, fmt.Sprintf("unable to insert transaction:%v", words))
	}
	nodeString1, err := tr.Serialize()
	assert.Nil(t, err)

	fname := "/tmp/fptree/test2.txt"
	err = fp.WriteTreeToFile(fname, nodeString1)
	assert.Nil(t, err)
	tree, err := fp.CreateTreeFromFile(fname)
	assert.Nil(t, err)
	assert.NotNil(t, tree)
	nodeString2, err := tree.Serialize()
	assert.Nil(t, err)

	fname = "/tmp/fptree/test3.txt"
	err = fp.WriteTreeToFile(fname, nodeString2)
	log.Debugf("count map :%v", tree.CountMap)

}
