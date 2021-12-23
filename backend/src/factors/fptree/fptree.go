package fptree

import (
	U "factors/util"
	"fmt"
	"reflect"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

// canRemoveNode check if node can ber removed
func (t Tree) canRemoveNode(nd *node) bool {
	log.Debugf("checking to remove node :%s", nd.Item)
	if len(nd.NextMap) != 0 {
		log.Debugf("Unable to delete next map is not 0:%v,%d,parent is %s,%d", nd.NextMap, len(nd.NextMap), nd.ParentNode.Item, nd.ParentNode.Counter)
		for _, v := range nd.NextMap {
			log.Debugf("next map needs to be deleted: %s,%d", v.Item, v.Counter)
		}
		return false
	}
	if nd.Counter > 0 {
		log.Debug("counter is greater than 0")
		return false
	}
	return true
}

// DeleteNode delete node from tree
func (t Tree) DeleteNode(nd *node) error {
	var prev *node
	var err error
	log.Debugf(" inside Delete Node :%s,%d", nd.Item, nd.Counter)
	if !nd.TombstoneMarker {
		return fmt.Errorf("unable to delete node, TombstoneMarker not set:%s", nd.Item)
	}

	if val, ok := t.HeadMap[nd.Item]; ok {
		if val == nd {
			log.Debug("Items to be deleted , available in head map")
		}
	}

	if val, ok := t.TailMap[nd.Item]; ok {
		if val == nd {
			log.Debug("Items to be deleted , available in tail map")
		}
	}

	for k, v := range nd.NextMap {
		//need to check if nd.counter or 0
		if v.Counter > nd.Counter {
			return fmt.Errorf("unable to delete node, monotonic property violated, child has positive count:%s,%d", k, v.Counter)
		} else {
			v.ParentNode = nil
			delete(nd.NextMap, k)
		}
	}

	if t.HeadMap[nd.Item] == nd {
		log.Debugf("Node is found in head map for deleting:%s", nd.Item)
		if nd.AuxNode != nil {
			t.HeadMap[nd.Item] = nd.AuxNode
			nd.AuxNode = nil
		} else {
			delete(t.HeadMap, nd.Item)
			log.Debugf("Node deleted from headMap :%s,%d,%v", nd.Item, nd.Counter, t.HeadMap)
		}

	} else {
		//unlinking aux node
		log.Debugf("Getting previous node while deleteing:%v", nd)
		prev, err = t.GetPreviousNode(nd)
		if err != nil {
			return err
		}
		if prev != nil {
			log.Debugf("Previous is not nil:%v", prev)
			log.Debugf("node to be shorted:%v", nd)
			prev.AuxNode = nd.AuxNode
			// nd.AuxNode = nil
			log.Debugf("Previous node after short circuititn:%v", prev)

			if t.TailMap[nd.Item] == nd {
				log.Debug("node available in tail map, hence setting prev in tail map:%v,%v", nd, prev)
				t.TailMap[nd.Item] = prev
			}

		}

	}

	if tailNode, ok := t.TailMap[nd.Item]; ok {
		if prev != nil && tailNode == nd {
			log.Debugf("setting the previous node in tail map")
			t.TailMap[nd.Item] = prev
		}
		if tailNode == nd {
			log.Debugf("node in tail map for deleting:%v", nd)
			log.Debug("Removing from tail map")
			delete(t.TailMap, nd.Item)
		}

	}

	if nd.ParentNode != nil {
		log.Debugf("unlinking from parent :%v", nd.ParentNode)
		delete(nd.ParentNode.NextMap, nd.Item)
		nd.ParentNode = nil
	}
	if t.canRemoveNode(nd) {
		log.Debugf("Deleting node after checking :%s,%v", nd.Item, nd)
		nd.TombstoneMarker = false
		nd.Item = ""
		nd = nil
		if reflect.ValueOf(nd).IsNil() {
			log.Debugf("Deleted node :%v", nd)
		}

	}
	return nil
}

// checkNodeInHeadMap check if node in head map
func (t Tree) checkNodeInHeadMap(nd *node) error {
	head := t.HeadMap[nd.Item]
	if strings.Compare(head.Item, nd.Item) != 0 {
		log.Errorf("illegal access of headMap , node and nodeItem not matching")
		return fmt.Errorf("illegal access of headMap , node and nodeItem not matching")
	}
	return nil
}

// checkNodeInTailMap check if node in tail map
func (t Tree) checkNodeInTailMap(nd *node) error {
	tail := t.TailMap[nd.Item]
	if strings.Compare(tail.Item, nd.Item) != 0 {
		log.Errorf("illegal access of TailMap , node and nodeItem not matching")
		return fmt.Errorf("illegal access of tailMap , node and nodeItem not matching")
	}
	return nil
}

// GetPreviousNode get previous node before inserting new node
func (t Tree) GetPreviousNode(n *node) (*node, error) {
	log.Debugf("Inside GetPreviousNode:%v", n)
	var prev *node
	head := t.HeadMap[n.Item]
	tail := t.TailMap[n.Item]
	log.Debugf("tail node for:%s is %v", n.Item, tail)

	if head == nil {
		log.Debug("Head node is nil, when accessing previous node")
		return nil, nil
	}

	err := t.checkNodeInHeadMap(n)
	if err != nil {
		return nil, err
	}

	// mainHead := head
	if head == n {
		log.Debugf("item found in head map,hence no prev node:%s", n.Item)
		return nil, nil
	}
	if head.AuxNode == nil {
		log.Debugf("aux node of head is nil, hence no next node:%v", head)
		return nil, nil
	}
	for head != nil {
		log.Debugf("looping :%v", head)
		if head == n {
			log.Debug("found current node")
			return prev, nil
		}
		if head == tail && tail != nil {
			log.Debug("Reached tail node from head, hence setting prev as tail")
			return tail, nil
		}
		prev = head
		head = head.AuxNode
	}

	if n.Item == "" && n.AuxNode == nil && n.Counter == 0 && n.IsRoot == false {
		log.Errorf("Unable to reach node empty node from %v", n)
		return nil, fmt.Errorf("unable to reach node as node is empty:%v", n)
	}
	log.Errorf("Unable to reach node:%v from %v", n, t.HeadMap[n.Item])
	return nil, fmt.Errorf("unable to reach node:%s", n.Item)

}

// InitNode initialize a new node
func InitNode(itm string, count int) node {
	log.Debugf("Initializing node:%s,%d", itm, count)
	var nd node
	nd.Item = itm
	nd.Counter = count
	nd.ParentNode = nil
	nd.NextMap = make(map[string]*node)
	nd.IsRoot = false
	nd.AuxNode = nil

	nd.ParentNode = nil
	nd.TombstoneMarker = false
	return nd
}

// InitTree Initialize an empty tree
func InitTree() Tree {
	var tr Tree
	tr.CountMap = make(map[string]int)
	tr.HeadMap = make(map[string]*node)
	tr.TailMap = make(map[string]*node)
	tr.ValueMap = make(map[int][]string)
	tr.NumNodes = make([]int, 1)
	tr.NumInsertion = make([]int, 1)
	log.Debug("empty tree is initalized")

	r := InitNode("@", 0)
	r.IsRoot = true
	r.ParentNode = nil
	tr.Root = r
	log.Debug("initalized root for tree")

	return tr
}

// InsertItemsIntoTree add items to tree
func (t Tree) InsertItemsIntoTree(itms []string) error {
	log.Debug("Inserting items into tree")
	if !U.CheckUniqueTrans(itms) {
		return fmt.Errorf("Items are not unique:%v", itms)
	}

	curr := &t.Root
	for _, itm := range itms {
		t.NumInsertion[0] += 1
		if v, ok := curr.NextMap[itm]; ok {
			if v.TombstoneMarker == true {
				log.Errorf("TombstoneMarker is set , unable to insert")
				return fmt.Errorf("TombstoneMarker is set , unable to insert")
			}
			v.Counter += 1
			curr = v
		} else {

			newNode := InitNode(itm, 1)

			newNode.NextMap = make(map[string]*node)
			newNode.AuxNode = nil
			//set Parent Node
			newNode.ParentNode = curr
			log.Debug("getting aux when inserting items into tree")
			err := t.UpdateHeaderTable(&newNode)
			if err != nil {
				return err
			}
			curr.NextMap[itm] = &newNode
			curr = curr.NextMap[itm]
			t.NumNodes[0] += 1
		}

		log.Debugf("Inserting %s:%d", curr.Item, curr.Counter)

	}

	return nil
}

// GetNumNodes get total number of nodes in tree
func (t Tree) GetNumNodes() int {
	tm := t.NumNodes[0]
	return tm
}

// getNumInsertion get total number of insertions in tree
func (t Tree) getNumInsertion() int {
	tm := t.NumInsertion[0]
	return tm
}

// OrderAndInsertCondTrans insert patterns in conditional tree
func (t Tree) OrderAndInsertCondTrans(trn ConditionalPattern) error {
	transaction := trn.Items
	count := trn.Count

	for idx := 0; idx < count; idx++ {
		err := t.OrderAndInsertTrans(transaction)
		if err != nil {
			return err
		}
	}
	return nil
}

// OrderAndInsertTrans insert items into tree
func (t Tree) OrderAndInsertTrans(trns []string) error {

	//sort transcation
	sort.Strings(trns)
	sortedList := U.SortOnPriority(trns, t.CountMap, false)
	log.Debugf("sorted list for inserting :%v", sortedList)

	log.Debugf("raw trans:%v ,sorted trans:%v", trns, sortedList)
	for _, tr := range sortedList {
		if currCount, ok := t.CountMap[tr]; ok {

			updatedCount := currCount + 1
			t.CountMap[tr] = updatedCount
			losserList := t.UpdateValueMap(currCount, tr)
			t.ValueMap[currCount] = losserList

			// check if losserlist is not  and call reconstruct when required
			if len(losserList) > 0 {
				log.Debugf("Calling reconstruction winner:%s \n losser :%v", tr, losserList)
				winnerItem := tr
				err := t.Reconstruct(winnerItem, losserList)
				if err != nil {
					log.Errorf("error in reconstruct :%v", err)
					return err
				}
			}
			//update valueCount
			t.ValueMap[updatedCount] = append(t.ValueMap[updatedCount], tr)
			log.Debugf("valueMap :%v", t.ValueMap)
		} else {
			// update countsValue for tr with 1
			log.Debugf("Adding trans :%s", tr)
			t.CountMap[tr] = 1
			t.ValueMap[1] = append(t.ValueMap[1], tr)

			log.Debugf("Added trans countValue is 1 : countsMap:%v, transMap:%v as ", t.CountMap, t.ValueMap)
			log.Debugf("Added HeadMap:%v as ", t.HeadMap)

		}
	}

	err := t.InsertItemsIntoTree(sortedList)
	if err != nil {
		return err
	}
	// to check for err call checkReachability here
	for k, v := range t.ValueMap {
		if len(v) == 0 {
			delete(t.ValueMap, k)
		}
	}

	return nil
}

// UpdateValueMap update the value map on each insertion
func (t Tree) UpdateValueMap(curr int, tr string) []string {
	//remove value from valueMap
	tmpValList := t.ValueMap[curr]
	losserList := U.RemoveString(tmpValList, tr)
	if len(t.ValueMap[curr]) == 0 {
		delete(t.ValueMap, curr)
	}
	return losserList
}

// isEmpty check if node is null node
func isEmpty(nd *node) bool {

	if nd.Item == "" && nd.Counter == 0 && len(nd.NextMap) == 0 && nd.ParentNode == nil && nd.AuxNode == nil {
		return true
	}
	return false
}

// UpdateHeaderTable update the head table when a node is deleted
func (t Tree) UpdateHeaderTable(n *node) error {
	log.Debugf("Updating header table with :%v", n)

	if n.AuxNode != nil {
		log.Debugf("Aux node already set for node:%v", n)
		return nil
	}
	if n.Item == "" {
		log.Errorf("node item not set and trying to access the header table:%v", n)
		return fmt.Errorf("accessing header table, without node item")
	}

	if headNode, ok := t.HeadMap[n.Item]; ok {
		log.Debug("Checking if head node is empty")
		if isEmpty(headNode) {
			log.Debug("Deleting head node is empty")
			delete(t.HeadMap, n.Item)
		}
	}

	if headNode, ok := t.HeadMap[n.Item]; !ok {
		log.Debug("adding node to head map,aux node will be empty")
		// the value of head map will only be used in fpgrowth
		if _, ok := t.TailMap[n.Item]; ok {
			log.Error("Tail map already set")
			return fmt.Errorf("Trying to set headmap when tail is not empty")
		}
		t.HeadMap[n.Item] = n
		n.AuxNode = nil
	} else {
		if strings.Compare(headNode.Item, n.Item) != 0 {
			log.Error("Node in headmap is not matching with item:head :%v,n:%v", headNode, n)
			return fmt.Errorf("Node in headmap is not matching with item")
		}
		log.Debug("set tail map")
		if tnode, ok := t.TailMap[n.Item]; ok {
			log.Debugf("setting tail map , following tail node:%v", tnode)

			if tnode.Counter == 0 && tnode.IsRoot == false && tnode.Item == "" {
				log.Debug("reachine empty node")

			}
			prev, err := t.GetPreviousNode(tnode)
			if err != nil {
				return err
			}
			if prev != nil {
				log.Debugf("Able to tail reach tail node while setting header table")
			} else {
				log.Debugf("unable to reach previous node while updating header table")
				log.Debugf("count in coundtsMap :%d", t.CountMap[tnode.Item])
			}

			// assign value of tail map to node
			log.Debugf("Tail node before resettinging:%v", tnode)
			tnode.AuxNode = n
			log.Debugf("Tail node after resettinging:%v", tnode)

			//reset tail map vale with address of node
			t.TailMap[n.Item] = n
			log.Debugf("Tail map after  resettinging:%v", t.TailMap[n.Item])

		} else {
			log.Debugf("tailMap is nil,set first time:%s", n.Item)
			// this is the case the tail node is populated first time
			// get the value from head , update the auxnode of head node
			// then insert into tail map
			log.Debugf("current head map:%v", t.HeadMap[n.Item])
			hn := t.HeadMap[n.Item]
			log.Debug("head map value :%v", hn)
			// check of head nodes aux is nil
			if hn.AuxNode == nil {
				log.Debug("set aux of head node")
				hn.AuxNode = n
				log.Debug("head map value after setting :%v", hn)

			} else {
				log.Debugf("header aux node is already set")
				lastNode, err := t.getLastNode(hn)
				if err != nil {
					return err
				}
				if lastNode.AuxNode != nil {
					lastNode.AuxNode = n
				}
			}
			// set tail node
			t.TailMap[n.Item] = n
			log.Debug("tail after setting head:%v", t.TailMap[n.Item])

		}
	}

	return nil

}

// getLastNode : get the last node by traversing from header map and check for cycle.
func (t Tree) getLastNode(nd *node) (*node, error) {
	testMap := make(map[*node]bool)
	var prev *node
	head := nd
	for head != nil {
		if _, ok := testMap[head]; ok {
			log.Debug("finding cycle")
			return prev, nil
		} else {
			testMap[head] = true
		}
		prev = head
		head = head.AuxNode
	}
	return prev, nil
}
