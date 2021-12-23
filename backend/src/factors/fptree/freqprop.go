package fptree

import (
	U "factors/util"
	"fmt"
	"reflect"
	"strconv"

	log "github.com/sirupsen/logrus"
)

//{Items:[pk1::pv1],Count:count, CondItem:[pk2::pv2]}
func NewFrequentPropertiesStruct() *FrequentPropertiesStruct {
	return &FrequentPropertiesStruct{
		Total:                 0,                          //no of FrequentItemsets
		FrequentItemsets:      make([]FrequentItemset, 0), // frequent_itemset = [{map["pk2":pk2, "pk1":pk1],type}, count]
		MaxFrequentProperties: 0,
		PropertyMap:           make(map[PropertyNameType][]string), // pk1, pk2 added to propertyMap
	}
}

// Get map of all frequent properties (properties(keys) present in at least one frequent itemset)
func (fp *FrequentPropertiesStruct) GetAllProperties() []PropertyNameType {
	var props = make([]PropertyNameType, 0)
	for pnt, _ := range fp.PropertyMap {
		props = append(props, pnt)
	}
	return props
}

// Get frequency of an itemset among the frequent itemsets
func (fp *FrequentPropertiesStruct) GetFrequency(pntv PropertyMapType) (int, error) {
	if fp == nil {
		log.Error("pattern freqprops not loaded or null")
		return 0, fmt.Errorf("pattern freqprops not loaded or null")
	}
	for _, freqitem := range fp.FrequentItemsets {
		if reflect.DeepEqual(freqitem.PropertyMapType, pntv) {
			return freqitem.Frequency, nil
		}
	}
	return 0, fmt.Errorf("not found - %v", pntv)
}

// Sort Frequent itemsets by frequency and get top k
func (fp *FrequentPropertiesStruct) GetTopKCuts(k int) []FrequentItemset {

	//create index to frequency map
	indexFreqMap := make(map[string]int)
	for i, freqitem := range fp.FrequentItemsets {
		indexFreqMap[strconv.Itoa(i)] = freqitem.Frequency
	}

	//sort
	rankedList := U.RankByWordCount(indexFreqMap)

	// check if at least k itemsets present (if not, take number of present as k)
	if len(rankedList) < k {
		k = len(rankedList)
	}

	// get top k
	req := make([]FrequentItemset, 0)
	for _, pair := range rankedList[:k] {
		if ind, err := strconv.Atoi(pair.Key); err == nil {
			req = append(req, fp.FrequentItemsets[ind])
		}
	}
	return req
}

// Filter Frequent itemsets to contain given properties(keys),Sort by frequency and get top k
func (fp *FrequentPropertiesStruct) GetTopKNameCuts(pnts []PropertyNameType, k int) []FrequentItemset {

	//create index to frequency map and add based on filter
	indexFreqMap := make(map[string]int)
	for i, freqitem := range fp.FrequentItemsets {
		toBeAdded := true
		for _, pnt := range pnts {
			pn := pnt.PropertyName
			pt := pnt.PropertyType
			pntv := freqitem.PropertyMapType
			if _, ok := pntv.PropertyMap[pn]; pntv.PropertyType != pt || !ok {
				toBeAdded = false
			}
		}
		if toBeAdded {
			indexFreqMap[strconv.Itoa(i)] = freqitem.Frequency
		}
	}

	// sort
	rankedList := U.RankByWordCount(indexFreqMap)

	// check if at least k itemsets present (if not, take number of present as k)
	if len(rankedList) < k {
		k = len(rankedList)
	}

	// get top k
	req := make([]FrequentItemset, 0)
	for _, pair := range rankedList[:k] {
		if ind, err := strconv.Atoi(pair.Key); err == nil {
			req = append(req, fp.FrequentItemsets[ind])
		}
	}
	return req
}

// Filter Frequent itemsets to contain given properties(keys and values),Sort by frequency and get top k
func (fp *FrequentPropertiesStruct) GetTopKNameValueCuts(pntv PropertyMapType, k int) []FrequentItemset {

	//create index to frequency map and add based on filter
	indexFreqMap := make(map[string]int)
	for i, freqitem := range fp.FrequentItemsets {
		toBeAdded := true
		for pn, pv := range pntv.PropertyMap {
			pt := pntv.PropertyType
			pntv := freqitem.PropertyMapType
			if val, ok := pntv.PropertyMap[pn]; pntv.PropertyType != pt || !ok || val != pv {
				toBeAdded = false
			}
		}
		if toBeAdded {
			indexFreqMap[strconv.Itoa(i)] = freqitem.Frequency
		}
	}

	// sort
	rankedList := U.RankByWordCount(indexFreqMap)

	// check if at least k itemsets present (if not, take number of present as k)
	if len(rankedList) < k {
		k = len(rankedList)
	}

	// get top k
	req := make([]FrequentItemset, 0)
	for _, pair := range rankedList[:k] {
		if ind, err := strconv.Atoi(pair.Key); err == nil {
			req = append(req, fp.FrequentItemsets[ind])
		}
	}
	return req
}

func (fp *FrequentPropertiesStruct) GetPropertyValues(propName, propType string) []string {
	pnt := PropertyNameType{propName, propType}
	if vals, ok := fp.PropertyMap[pnt]; ok {
		return vals
	} else {
		return []string{}
	}
}

func (fp *FrequentPropertiesStruct) GetPropertiesOfType(propType string) ([]string, error) {
	if propType != "event" && propType != "user" {
		return []string{}, fmt.Errorf("wrong property type: %v", propType)
	}
	var propNames = make([]string, 0)
	for pnt, _ := range fp.PropertyMap {
		if pnt.PropertyType == propType {
			propNames = append(propNames, pnt.PropertyName)
		}
	}
	return propNames, nil
}

func (fp *FrequentPropertiesStruct) GetPropertiesWithCount(propType string) map[string]uint64 {
	var propertyCountMap = make(map[string]uint64)
	for _, itemset := range fp.FrequentItemsets {
		if itemset.PropertyMapType.PropertyType == propType || propType == "all" {
			if len(itemset.PropertyMapType.PropertyMap) == 1 {
				for key, _ := range itemset.PropertyMapType.PropertyMap {
					propertyCountMap[key] = uint64(itemset.Frequency)
				}
			}
		}
	}
	return propertyCountMap
}
