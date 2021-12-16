package fptree

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type customQueue struct {
	// https://play.golang.org/p/goRAMx0l4jQ
	queue []*node
	lock  sync.RWMutex
}

func (c *customQueue) Enqueue(n *node) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.queue = append(c.queue, n)
}

func (c *customQueue) Dequeue() error {
	if len(c.queue) > 0 {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.queue = c.queue[1:]
		return nil
	}
	return fmt.Errorf("Pop Error: Queue is empty")
}

func (c *customQueue) Front() (*node, error) {
	if len(c.queue) > 0 {
		c.lock.Lock()
		defer c.lock.Unlock()
		return c.queue[0], nil
	}
	return nil, fmt.Errorf("Peep Error: Queue is empty")
}

func (c *customQueue) Size() int {
	return len(c.queue)
}

func (c *customQueue) isEmpty() bool {
	return len(c.queue) == 0
}

func (c *customQueue) DequeFront() (*node, error) {
	top_node, err := c.Front()
	if err != nil {
		return nil, err
	}
	err = c.Dequeue()
	if err != nil {
		return nil, err
	}
	return top_node, nil
}

func make_Tree_Node(n node) TreeNode {
	var t TreeNode
	if n.IsRoot == true {
		t.Item = ""
		t.Count = -1
	} else {
		t.Item = n.Item
		t.Count = n.Counter
	}
	log.Debugf("tree node :%v", t)
	return t
}

func make_string_from_node(n node) (string, error) {
	// take node as input convert to tree_node and return json as string

	tnode := make_Tree_Node(n)
	bytes, err := json.Marshal(tnode)
	if err != nil {
		log.Errorf("unable to marshall string :%v", n)
		return "", err
	}
	return string(bytes), nil
}

func (t Tree) Serialize() ([]string, error) {
	var nodeString []string = make([]string, 0)
	var nodeListString []string = make([]string, 0)
	root := t.Root

	if t.Root.NextMap == nil {
		return []string{}, nil
	}

	treeNodeList, err := serializeHelper(&root)
	if err != nil {
		return []string{}, err
	}

	for _, nd := range treeNodeList {

		nodeString = append(nodeString, nd.Item)
		ndString, err := make_string_from_node(*nd)
		if err != nil {
			return nil, err
		}
		nodeListString = append(nodeListString, ndString)
	}
	log.Debugf("Node string :%v", nodeString)
	log.Debugf("Node list  string :%v", nodeListString)

	return nodeListString, nil
}

func serializeHelper(root *node) ([]*node, error) {

	var nodeQueue = &customQueue{queue: make([]*node, 0)}
	var nodeList []*node = make([]*node, 0)

	sentinel_node := InitNode("#", -2)
	sentinel_node.ParentNode = nil
	child_node := InitNode("$", -2)
	child_node.ParentNode = nil
	nodeQueue.Enqueue(root)
	nodeQueue.Enqueue(&sentinel_node)

	for nodeQueue.Size() > 0 {

		frontNode, err := nodeQueue.DequeFront()
		if err != nil {
			return nil, err
		}

		if strings.Compare(frontNode.Item, "#") == 0 {
			nodeList = append(nodeList, &sentinel_node)

			if nodeQueue.Size() != 0 {
				nodeQueue.Enqueue(&sentinel_node)
			}
		} else if strings.Compare(frontNode.Item, "$") == 0 {
			nodeList = append(nodeList, &child_node)
		} else {
			nodeList = append(nodeList, frontNode)
			for _, nd := range frontNode.NextMap {
				nodeQueue.Enqueue(nd)
			}
			fd, err := nodeQueue.Front()
			if err != nil {
				return nil, err
			}
			if strings.Compare(fd.Item, "#") != 0 {
				nodeQueue.Enqueue(&child_node)
			}

		}

	}
	return nodeList, nil
}

func WriteTreeToFile(fname string, nodes []string) error {

	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)

	for _, nd := range nodes {

		lineWrite := string(nd)
		if _, err := w.WriteString(fmt.Sprintf("%s\n", lineWrite)); err != nil {
			log.WithFields(log.Fields{"line": nd, "err": err}).Error("Unable to write to file.")
			return err
		}

	}
	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func SerializeTreeToFile(tr Tree, fname string) error {

	nodeStrings, err := tr.Serialize()
	if err != nil {
		log.Error("Unable to serialize tree")
		return err
	}
	err = WriteTreeToFile(fname, nodeStrings)
	if err != nil {
		log.Error("Unable to write serialized tree to file")
		return err
	}

	return nil
}

func ReadNodesFromFile(fname string) ([]node, error) {
	var nodes []node = make([]node, 0)
	file, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		tmp_node, err := create_node_from_treeNodeString(line)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, tmp_node)
	}
	log.Debugf("Number of nodes to process:%d", len(nodes))
	return nodes, nil
}

func create_node_from_treeNodeString(data string) (node, error) {
	var t TreeNode
	if err := json.Unmarshal([]byte(data), &t); err != nil {
		log.WithFields(log.Fields{"line": data, "err": err}).Error("Read failed")
		return node{}, err
	}
	nd := InitNode(t.Item, t.Count)

	return nd, nil
}

func deserialize(data []node) (Tree, error) {

	var t Tree
	if len(data) == 0 {
		return t, nil
	}
	log.Debugf("Entering deserialize :%d", len(data))
	t = InitTree()
	time.Sleep(time.Second * 15)
	err := deserializeHelper(data, t)
	if err != nil {
		return Tree{}, err
	}

	return t, nil

}

func deserializeHelper(data []node, tr Tree) error {

	var prevLevel = &customQueue{queue: make([]*node, 0)}
	var currentLevel = &customQueue{queue: make([]*node, 0)}
	var parentNode *node
	var err error
	root := tr.Root
	currentLevel.Enqueue(&root)
	parentNode = &root
	log.Debugf("Total number of nodes to be inserted:%d", len(data))
	for idx := 1; idx < len(data); idx++ {
		tmp_node := data[idx]
		if strings.Compare(tmp_node.Item, "#") == 0 {
			if prevLevel.Size() == 0 {
				for currentLevel.Size() != 0 {
					log.Debugf("deque a")
					tmp_node, err := currentLevel.DequeFront()
					if err != nil {
						return err
					}
					prevLevel.Enqueue(tmp_node)
				}
			} else {
				log.Errorf("Prev level not empty")
				return fmt.Errorf("Prev level not empty :%d", prevLevel.Size())
			}

			if currentLevel.Size() != 0 {
				log.Errorf("Current level is not zero")
				return fmt.Errorf("current level is not zero")
			}

			// parentNode = prevLevel.popleft() if prevLevel else None;

			if prevLevel.Size() > 0 {
				parentNode, err = prevLevel.DequeFront()
				if err != nil {
					return err
				}
			}

		} else if strings.Compare(tmp_node.Item, "$") == 0 {
			log.Debugf("$$$$ -> Inserting node_b in $:%s", tmp_node.Item)

			if prevLevel.Size() > 0 {
				parentNode, err = prevLevel.DequeFront()
				if err != nil {
					return err
				}
			}

		} else {

			// childNode := InitNode(tmp_node.Item, tmp_node.Counter)
			currentLevel.Enqueue(&tmp_node)
			parentNode.NextMap[tmp_node.Item] = &tmp_node
			tmp_node.ParentNode = parentNode

			if _, ok := tr.CountMap[tmp_node.Item]; ok {
				tr.CountMap[tmp_node.Item] += 1
			} else {
				tr.CountMap[tmp_node.Item] = 1
			}

			if hVal, ok := tr.HeadMap[tmp_node.Item]; !ok {
				tr.HeadMap[tmp_node.Item] = &tmp_node
			} else {
				if tVal, ok := tr.TailMap[tmp_node.Item]; ok {
					tVal.AuxNode = &tmp_node
					tr.TailMap[tmp_node.Item] = &tmp_node
				} else {
					hVal.AuxNode = &tmp_node
					tr.TailMap[tmp_node.Item] = &tmp_node
				}
			}

		}

	}
	return nil
}

func CreateTreeFromFile(fname string) (Tree, error) {

	nodes, err := ReadNodesFromFile(fname)
	if err != nil {
		return Tree{}, err
	}
	tree, err := deserialize(nodes)
	if err != nil {
		return Tree{}, err
	}
	return tree, nil

}
