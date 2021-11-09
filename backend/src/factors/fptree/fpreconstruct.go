package fptree

import (
	U "factors/util"
	"fmt"
	"math"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetFarAncestor get the fartest reaching Ancestor whicle ascending path
func (t Tree) GetFarAncestor(wNode *node, losserList []string) (*node, int, error) {
	log.Debugf("inside far anscestor:%s, loserList :%v", wNode.Item, losserList)
	var farLosserNode *node
	nodeCount := 0
	pnode := wNode.ParentNode
	pathLength := 0
	pos := float64(0)
	for pnode != nil && !pnode.IsRoot {
		log.Debugf("Node item and count:%s,%d", pnode.Item, pnode.Counter)
		pathLength += 1
		if U.In(losserList, pnode.Item) {
			farLosserNode = pnode
			pos = math.Max(float64(pos), float64(pathLength))
		}
		if nodeCount <= pnode.Counter {
			nodeCount = pnode.Counter
		} else {
			eString := fmt.Sprintf("node count is greater than child , monotonic property broken: prev->%d,curr->%d", nodeCount, pnode.Counter)
			e := fmt.Errorf(eString)
			log.Error(eString)
			return nil, 0, e
		}
		pnode = pnode.ParentNode
	}

	if pnode != nil && pnode.IsRoot == false {
		eString := fmt.Sprintf("unable to reach the root")
		e := fmt.Errorf(eString)
		log.Error(e)
		return nil, 0, e
	}

	if farLosserNode != nil {
		log.Debugf("farthest ansector node for %s is %s, lenght of path is %d, parent is %v", wNode.Item, farLosserNode.Item, int(pos), farLosserNode.ParentNode)
		return farLosserNode, int(pos), nil
	} else {
		log.Debug("no anscestor found")
	}
	return nil, 0, nil
}

// Reconstruct rebalance the tree when each insertion happens and tree is not balanced
func (t Tree) Reconstruct(winnerItem string, losserList []string) error {
	log.Debugf("Reconstruct for winner :%s", winnerItem)
	wNode := t.HeadMap[winnerItem]
	auxCount := 1
	for wNode != nil && strings.Compare(wNode.Item, winnerItem) == 0 && wNode.TombstoneMarker == false {

		wiNextNode := wNode.AuxNode
		ansNode, lenPath, err := t.GetFarAncestor(wNode, losserList)
		if err != nil {
			return err
		}

		if ansNode != nil && ansNode.TombstoneMarker == false {

			log.Debugf("ansecestor node of: %s,%d is %s,%d ,pathLen:%d", wNode.Item, wNode.Counter, ansNode.Item, ansNode.Counter, lenPath)
			log.Debugf("ansecestor node: %v", ansNode)
			if tailNodeItm, ok := t.TailMap[wNode.Item]; ok {
				log.Debugf("tail node in tailMap:%v", tailNodeItm)
			} else {
				log.Debugf("tail node in tailMap is empty:%s", wNode.Item)

			}
			_, winParentNode, winNextMap, err := t.isolateWinNode(wNode)
			if err != nil {
				return err
			}

			pathNodes, ansParentNode, err := t.ReduceCount(winParentNode, ansNode, wNode.Counter, lenPath)
			if err != nil {
				return err
			}

			log.Debug("ansecestor parent node:%v", ansParentNode)
			log.Debug("Creating path")
			head, tail, err := t.createPath(pathNodes, wNode)
			if err != nil {
				return err
			}

			log.Debugf("Rewiring parent and child node:%v,%v", ansParentNode, wNode)
			winNode, err := t.rewire(wNode, head, tail, winNextMap)
			if err != nil {
				return err
			}

			if ansParentNode != nil {
				log.Debug("parent node is not nil:%v", ansParentNode)
			} else {
				log.Debug(" ans parent is empty")
			}
			log.Debug("ansecestor parent node for merging:%v", ansParentNode)
			log.Debug("merging to node:%v", ansParentNode.NextMap[winNode.Item])

			if ansNextNode, ok := ansParentNode.NextMap[winNode.Item]; !ok {
				log.Debugf("Updating header table with winner node")

				err := t.UpdateHeaderTable(winNode)
				if err != nil {
					return err
				}

				ansParentNode.NextMap[wNode.Item] = wNode
				wNode.ParentNode = ansParentNode

				log.Debugf("wNode after merging:%v", wNode)

			} else {
				if strings.Compare(ansNextNode.Item, wNode.Item) == 0 {

					log.Debug("merge nodes as both have common keys")

					if wNode.AuxNode == nil {
						log.Debug("setting aux node as winner node aux is emptyx")

						err := t.UpdateHeaderTable(wNode)
						if err != nil {
							return err
						}

					}
					log.Debugf("Need to merge nodes:%s,%s", ansNextNode.Item, wNode.Item)
					log.Debugf("Need to merge winner node:%v", wNode)
					log.Debugf("Need to ans node:%v", ansNextNode)
					prevNWin, err := t.GetPreviousNode(wNode)
					if prevNWin != nil {
						log.Debugf("Prev nodeis available for winner node,%v", prevNWin)
					} else {
						log.Debugf("Prev nodeis not  available for winner node,%v", wNode)
					}
					if err != nil {
						return err
					}

					err = t.mergeNodes(ansNextNode, wNode)
					if err != nil {
						return err
					}

				}

			}

		} else {
			break
		}

		wNode = wiNextNode
		auxCount += 1

	}
	return nil
}

// isolateWinNode Remove all connection to nextnode, parent node and prev auxillary isolateWinNode
func (t Tree) isolateWinNode(nd *node) (*node, *node, map[string]*node, error) {
	log.Debug("Inside isolate winned node")
	// remove connection to prev node
	// remove connection to next node
	// remove connection to parent node
	if nd.TombstoneMarker {
		log.Error("unable to isolate as TombstoneMarker is true")
		return nil, nil, nil, fmt.Errorf("unable to isolate as TombstoneMarker is true")
	}
	if nd.IsRoot {
		log.Error("unable to isolate, as root node")
		return nil, nil, nil, fmt.Errorf("unable to isolate as root node")
	}

	winNextMap := make(map[string]*node)
	winNextNode := nd.AuxNode
	winParentNode := nd.ParentNode

	// remove connection from parent to child
	delete(winParentNode.NextMap, nd.Item)

	//remove connection winner to child
	for k, v := range nd.NextMap {
		winNextMap[k] = v
		v.ParentNode = nil
		delete(nd.NextMap, k)
	}

	err := t.updateAuxillaryNode(nd)
	if err != nil {
		return nil, nil, nil, err
	}

	if headNode, ok := t.HeadMap[nd.Item]; ok {
		if headNode == nd {
			log.Debugf("Node found in head map , need to be updated ")
			if headNode.AuxNode != nil {
				if tailNode, ok := t.TailMap[nd.Item]; ok {
					if tailNode == headNode.AuxNode {
						log.Debugf("head node aux found in tailMap ,updating node in tail map ")

						t.HeadMap[nd.Item] = tailNode
						delete(t.TailMap, nd.Item)
					}
				} else {
					log.Debugf("updating headMap with aux node ")
					tmpNode := headNode.AuxNode
					delete(t.HeadMap, nd.Item)
					t.HeadMap[nd.Item] = tmpNode
				}
			} else {
				log.Debugf("aux node is empty hence deleting entry from head map")

				delete(t.HeadMap, nd.Item)
			}
		}
	}

	return winNextNode, winParentNode, winNextMap, nil
}

// updateAuxillaryNode update the aux node when updating or deleting the previous node
func (t Tree) updateAuxillaryNode(nd *node) error {
	// change aux node connection based on previous node
	// prev node should point to nil or be in HeadMap

	itm := nd.Item
	headNode := t.HeadMap[nd.Item]
	tailNode := t.TailMap[nd.Item]
	prev, err := t.GetPreviousNode(nd)
	if err != nil {
		log.Error("error while updating aux node")
		return err
	}

	// node is in head map
	if nd == headNode {
		log.Debug("short circuiting -Node found in head map")
		if nd.AuxNode != nil {
			log.Debugf("setting aux node as head node, and removing the aux connection:%v", nd.AuxNode)
			//case aux node not nil
			nxt := nd.AuxNode
			t.HeadMap[itm] = nxt
			if nxt == tailNode {
				log.Debugf("Deleting tail map entry , as tail is set to head")
				delete(t.TailMap, itm)
			}
			//delete aux node connection from orig node
			nd.AuxNode = nil

		} else {
			//case aux node is nil
			if tailNode == nil {
				// log.Debugf("deleting head node as,  aux and tail node are empty (no other node next)")
				// case aux node is nil and tail node is nil
				delete(t.HeadMap, itm)
				log.Debugf("deleting head node as,  aux and tail node are empty (no other node next):%v", t.HeadMap[itm])

			} else {
				// case aux node is nil and tail node is not nil
				log.Debugf("setting tail node to head map as aux node is empty")
				t.HeadMap[itm] = tailNode
				delete(t.TailMap, itm)
			}
		}
		return nil
	}

	if nd == tailNode {
		if prev.AuxNode == nd {
			log.Debug("Reachable from previous node")
			prev.AuxNode = nil
			t.TailMap[itm] = prev
		} else {
			log.Error("unable to reach from previous node")
			return fmt.Errorf("while setting aux node, unable to reach from previous node")
		}
	}
	if prev != nil {
		if prev.AuxNode == nd {
			log.Debug("inside previous node,Reachable from previous node")
			if nd.AuxNode != nil {
				prev.AuxNode = nd.AuxNode
				nd.AuxNode = nil
			} else {
				log.Debug("setting aux node to as nil , as curr aux is nil")
				prev.AuxNode = nil
			}
		}
		return nil
	}

	return nil
}

// ReduceCount reduce nodes by local count from start to end
func (t Tree) ReduceCount(start, end *node, localCount int, lenPath int) ([]string, *node, error) {
	// reduce lenght by given local count from start to end
	var pnodeParent *node
	log.Debugf("Inside ReduceCount startnode is %s ,endNode is %s, path length is : %d", start.Item, end.Item, lenPath)
	log.Debugf("local count to reduceby :%d", localCount)

	pathNodeItems := make([]string, 0)
	pnode := start
	ancestorParentNode := end.ParentNode

	for pnode != nil && lenPath > 0 {
		lenPath -= 1
		log.Debugf("Reducing item : %s count:%d  node:%v", pnode.Item, pnode.Counter, pnode)
		pnode.Counter = pnode.Counter - localCount
		pathNodeItems = append(pathNodeItems, pnode.Item)
		pnodeParent = pnode.ParentNode
		if pnode.Counter <= 0 {
			log.Debugf("setting tombstone for :%s, prev TombstoneMarker is :%v", pnode.Item, pnode.IsRoot)
			if len(pnode.NextMap) > 1 {
				log.Debugf("next map is greater than 0")
			}
			pnode.TombstoneMarker = true

			err := t.DeleteNode(pnode)
			if err != nil {
				return nil, nil, err
			}
			//log.Debugf("Delete node inside Reduce Count:%v", pnode)
		}
		pnode = pnodeParent
	}

	return pathNodeItems, ancestorParentNode, nil
}

// createPath create a linked list of nodes based on trans and winner node's count
func (t Tree) createPath(trns []string, wNode *node) (*node, *node, error) {

	var prevNode *node
	var tail *node
	var head *node
	localCount := wNode.Counter
	log.Debugf("number of nodes to create : %d", len(trns))
	// create individual nodes
	for _, tr := range trns {
		log.Debugf("creating node :%s", tr)
		tmpNode := InitNode(tr, localCount)

		// set tail map in header table
		err := t.UpdateHeaderTable(&tmpNode)
		if err != nil {
			return nil, nil, err
		}

		if prevNode != nil {
			tmpNode.NextMap[prevNode.Item] = prevNode
			prevNode.ParentNode = &tmpNode
		} else {
			tail = &tmpNode
		}

		prevNode = &tmpNode
		head = &tmpNode

		log.Debugf("created node :%v", tmpNode)
	}
	tmp := tail
	for tmp != nil {
		log.Debugf("path node created and appended :%s,%v", tmp.Item, tmp)
		tmp = tmp.ParentNode
	}

	return head, tail, nil
}

// rewire rewire parent pointer when deleting the node
func (t Tree) rewire(winNode, head, tail *node, tailNextMap map[string]*node) (*node, error) {
	log.Debug("inside rewire node")
	log.Debugf("tail node in tailMap :%v", t.TailMap[winNode.Item])
	for k, v := range tailNextMap {
		log.Debugf("Attaching %s to tail:%s", k, tail.Item)
		tail.NextMap[k] = v
		v.ParentNode = tail
		log.Debugf("parent of tail :%v", v)
	}

	//connect head with winNode
	log.Debugf("connect ->winner node:%v", winNode)
	log.Debugf("connect ->head node:%v", head)

	if _, ok := winNode.NextMap[head.Item]; !ok {
		winNode.NextMap[head.Item] = head
		head.ParentNode = winNode
		log.Debugf("win node after rewiring :%v", winNode)
	} else {
		log.Debugf("Winnode nextmap is already occupied")
	}
	log.Debugf("tail node in tailMap :%v", t.TailMap[winNode.Item])
	return winNode, nil
}

type Stack struct {
	container map[int][2]*node
	currPos   int
}

// push add to stack data struct
func (s *Stack) push(newNode [2]*node) error {
	log.Debugf("inside stack push")
	pos := s.currPos
	newpos := pos + 1
	s.container[newpos] = newNode
	s.currPos = s.currPos + 1
	log.Debugf("stack :%d,%v", s.currPos, s.container)
	if len(s.container) != s.currPos {
		return fmt.Errorf("stack property violated")
	}
	return nil
}

func (s *Stack) pop() ([2]*node, error) {
	log.Debugf("inside pop stack")

	pos := s.currPos
	log.Debugf("pos:%d", pos)
	value := s.container[pos]

	delete(s.container, pos)
	s.currPos = pos - 1
	log.Debugf("curr pos:%d", s.currPos)

	if len(s.container) != s.currPos {
		return [2]*node{}, fmt.Errorf("stack property violated,%d,%d", len(s.container), s.currPos)
	}
	if s.currPos < 0 {
		return [2]*node{}, fmt.Errorf("stack property violated,lenght cannot be negative")
	}
	return value, nil
}
func (s *Stack) lenOfStack() (int, error) {
	if s.currPos < 0 {
		return 0, fmt.Errorf("stack property violated,lenght cannot be negative")
	}
	if len(s.container) != s.currPos {
		return 0, fmt.Errorf("stack property violated")

	}

	return s.currPos, nil
}

// checkReachability check for checkReachability of all leaf nodes to roo
func (t Tree) checkReachability() error {
	log.Debug("Inside check Reachability")
	for k, v := range t.HeadMap {
		if v == nil {
			log.Debugf("node is head map is nil:%s", k)
		}
		err := t.checkNodeInHeadMap(v)
		if err != nil {
			return err
		}
		log.Debugf("Checking Reachability for %s,%v,%d", k, v, t.CountMap[k])
		if tailVal, ok := t.TailMap[k]; ok {
			head := v
			count := 0
			for head != nil {
				count += 1
				log.Debugf("Loop head for Reachability :%v,%d", head, count)
				if head == tailVal {
					break
				}
				head = head.AuxNode
			}
			if head == tailVal {
				log.Debugf("tail reachable :%s", k)
			} else {
				log.Debugf("tail not reachable :%s", k)
			}
		} else {
			log.Debug("checking Reachability , tail map not set")
		}
	}
	return nil
}

// mergeNodes merge nodes in  reconstruction
func (t Tree) mergeNodes(src, tgt *node) error {
	log.Debug("Inside Merge Node")
	var st Stack
	st.container = make(map[int][2]*node)
	var visited = make(map[[2]*node]int)
	var visitedInt = make(map[int][2]*node)

	rootNodesPair := [2]*node{src, tgt}
	st.push(rootNodesPair)

	for st.currPos > 0 {

		tos, err := st.pop()
		if err != nil {
			log.Errorf("unable to pop stack")
			return err
		}
		if _, okVisit := visited[tos]; !okVisit {
			lval := len(visited)
			visited[tos] = lval
			visitedInt[lval] = tos
		}
		srcNode := tos[0]
		tgtNode := tos[1]

		for kTgt, vTgt := range tgtNode.NextMap {

			if vSrc, ok := srcNode.NextMap[kTgt]; ok {
				log.Debug("processing as nodes are available in src and tgt")
				nd := [2]*node{vSrc, vTgt}
				if _, okVisit := visited[nd]; !okVisit {
					log.Debug("Pushing to stack:%v", nd)
					err := st.push(nd)
					if err != nil {
						return err
					}
				}
			} else {
				log.Debug("processing as, nodes are available in tgt but not in src")
				srcNode.NextMap[kTgt] = vTgt
				vTgt.ParentNode = srcNode
				delete(tgtNode.NextMap, kTgt)
			}

		}

	}

	log.Debugf("Processing visited table, length of table is :%d", len(visited))
	log.Debugf("Processing visited table, length of table is :%d", len(visitedInt))
	log.Debugf("visitedInt:%v", visitedInt)
	for idx := len(visited) - 1; idx >= 0; idx-- {
		log.Debugf("Processing node idx:%d", idx)
		if v, ok := visitedInt[idx]; ok {

			log.Debug("joining nodes")
			src, tgt := v[0], v[1]
			src.Counter = src.Counter + tgt.Counter
			tgt.Counter = 0
			// tgt.Item = ""
			tgt.TombstoneMarker = true
			err := t.DeleteNode((tgt))
			if err != nil {
				log.Error("err in deleting node,while merging:%v", tgt)
				return err
			}
			delete(visited, v)
			delete(visitedInt, idx)
		}

	}
	log.Debug("len of nodes in visited left :%d", len(visited))
	l, _ := st.lenOfStack()
	log.Debug("len of stack :%d", l)

	return nil
}

// checkCycle checck for cycle in tree
func (t Tree) checkCycle() error {

	for _, itm := range t.HeadMap {
		itmCount := t.CountMap[itm.Item]
		itmMap := make(map[*node]bool)
		countNodes := 0
		counter := 0
		for itm != nil {
			countNodes += 1
			counter += itm.Counter
			if _, ok := itmMap[itm]; !ok {
				itmMap[itm] = true
			} else {
				log.Fatal("Cycle found for itm :%s,%d,%d,%d", itm.Item, itmCount, countNodes, counter)
				e := fmt.Errorf("cycle found for itm :%s", itm.Item)
				return e
			}
			if itm.TombstoneMarker {
				log.Fatalf("Traversing itm with tombstone marker :%s", itm.Item)

			}
			itm = itm.AuxNode
		}
		// if counter != itmCount {
		// 	log.Errorf("Counts Mismatch :%s", itm.Item)
		// 	e := fmt.Errorf("counts Mismatch :%s,%d,%d", itm.Item, counter, itmCount)
		// 	return e
		// }
	}

	return nil
}
