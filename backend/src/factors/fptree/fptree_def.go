package fptree

type node struct {
	Item            string           `json:"it"`
	Counter         int              `json:"ct"`
	NextMap         map[string]*node `json:"np"`
	AuxNode         *node            `json:"an"`
	ParentNode      *node            `json:"pn"`
	TombstoneMarker bool             `json:"tm"`
	IsRoot          bool             `json:"ir"`
}

type Tree struct {
	Root         node             `json:"rt"`
	HeadMap      map[string]*node `json:"hm"`
	TailMap      map[string]*node `json:"tm"`
	CountMap     map[string]int   `json:"cm"`
	ValueMap     map[int][]string `json:"vm"`
	NumNodes     []int            `json:"nn"`
	NumInsertion []int            `json:"ni"`
}

type ConditionalPattern struct {
	Items    []string `json:"fit"`
	Count    int      `json:"fc"`
	CondItem []string `json:"fcon"`
}

type PropertyNamesCount struct {
	PropertyNames []string `json:"pn"`
	PropertyCount int64    `json:"pc"`
	PropertyType  string   `json:"pt"`
}

type PropertyNameType struct {
	PropertyName string
	PropertyType string
}

type PropertyMapType struct {
	PropertyMap  map[string]string
	PropertyType string
}

type FrequentItemset struct {
	PropertyMapType PropertyMapType
	Frequency       int
}

type FrequentPropertiesStruct struct {
	Total                 uint64
	FrequentItemsets      []FrequentItemset
	MaxFrequentProperties uint64
	PropertyMap           map[PropertyNameType][]string
}

type TreeNode struct {
	Item  string `json:"tni"`
	Count int    `json:"tnc"`
}
