package fptree

type HeaderItm struct {
	ItemId       string
	SupportCount int
	HyperLink    *PrjTrans
	LastTr       *PrjTrans
}
type HeadTable struct {
	Hmap map[string]*HeaderItm
}

type PrjTrans struct {
	Trans []*PrjTransItm
}
type PrjDB struct {
	DB []*PrjTrans
}

type PrjTransItm struct {
	ItemId string
	Count  int
	HLink  *PrjTrans
}

type Fmap struct {

	// priority kept in ascending order 0=Highest Priority , N=least priority
	Count  map[string]int
	InvPri map[int]string
	Pri    map[string]int
}

type FPatternCount struct {
	FpItm    []string `json:"fi"`
	FpCounts int      `json:"fc"`
}

type FpContainer struct {
	Fpts   []FPatternCount         `json:"fct"`
	FptMap map[int][]FPatternCount `json:"fpm"`
}

type FpHProperties struct {
	Prop []string `json:"rp"`
}
