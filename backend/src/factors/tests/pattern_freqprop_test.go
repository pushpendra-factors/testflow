package tests

import (
	"bufio"
	"encoding/json"
	F "factors/fptree"
	U "factors/util"
	"os"
	"reflect"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const TEST_FILE_PATH = "../fptree/data/properties_test.txt"

func genFrequentPropertiesFromFile(path string) *F.FrequentPropertiesStruct {
	freqItems := make([]F.FrequentItemset, 0)
	allProps := make(map[F.PropertyNameType][]string)
	numProps := 0

	newfreq := F.NewFrequentPropertiesStruct()

	fq := F.FrequentItemset{}
	file, err := os.Open(TEST_FILE_PATH)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var pnc F.PropertyNamesCount
		line := scanner.Text()
		byteValue := []byte(line)
		json.Unmarshal(byteValue, &pnc)
		tmp := F.PropertyMapType{}
		tmp.PropertyMap = make(map[string]string)
		tmp.PropertyType = pnc.PropertyType
		fq.Frequency = int(pnc.PropertyCount)

		for _, pn := range pnc.PropertyNames {
			propkeyval := strings.Split(pn, "::")
			pk, pv := propkeyval[0], propkeyval[1]
			pnt := F.PropertyNameType{PropertyName: pk, PropertyType: pnc.PropertyType}
			tmp.PropertyMap[pk] = pv
			if _, ok := allProps[pnt]; !ok {
				allProps[pnt] = make([]string, 0)
			}
			if !U.In(allProps[pnt], pv) {
				allProps[pnt] = append(allProps[pnt], pv)
			}
		}
		fq.PropertyMapType = tmp
		freqItems = append(freqItems, fq)
		numProps++
	}

	newfreq.Total = uint64(numProps)
	newfreq.PropertyMap = allProps
	newfreq.FrequentItemsets = freqItems
	return newfreq
}
func TestGenFrequentPropertiesFromFile(t *testing.T) {
	fp := genFrequentPropertiesFromFile(TEST_FILE_PATH)
	freqitems := make([]F.FrequentItemset, 0)

	tmpMap := make(map[string]string)
	pmt := F.PropertyMapType{}
	tmpMap["$P1"] = "V11"
	tmpMap["$P2"] = "V22"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi := F.FrequentItemset{}
	fi.Frequency = 376
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	tmpMap = make(map[string]string)
	pmt = F.PropertyMapType{}
	tmpMap["$P2"] = "V22"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi = F.FrequentItemset{}
	fi.Frequency = 782
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	tmpMap = make(map[string]string)
	pmt = F.PropertyMapType{}
	tmpMap["$P2"] = "V22"
	tmpMap["$P1"] = "V11"
	tmpMap["$P3"] = "V31"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi = F.FrequentItemset{}
	fi.Frequency = 484
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	tmpMap = make(map[string]string)
	pmt = F.PropertyMapType{}
	tmpMap["$P4"] = "V41"
	tmpMap["$P5"] = "V51"
	tmpMap["$P7"] = "V71"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi = F.FrequentItemset{}
	fi.Frequency = 844
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	tmpMap = make(map[string]string)
	pmt = F.PropertyMapType{}
	tmpMap["$P6"] = "V61"
	tmpMap["$P7"] = "V72"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi = F.FrequentItemset{}
	fi.Frequency = 968
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	tmpMap = make(map[string]string)
	pmt = F.PropertyMapType{}
	tmpMap["$P6"] = "V61"
	tmpMap["$P1"] = "V12"
	tmpMap["$P3"] = "V32"
	tmpMap["$P2"] = "V21"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi = F.FrequentItemset{}
	fi.Frequency = 291
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	if reflect.DeepEqual(fp.FrequentItemsets, freqitems) {
		assert.Nil(t, nil)
	} else {
		assert.NotNil(t, nil)
	}
	assert.Equal(t, 7, len(fp.PropertyMap))
	assert.Equal(t, uint64(6), fp.Total)
}

func TestGetAllProperties(t *testing.T) {
	fp := genFrequentPropertiesFromFile(TEST_FILE_PATH)
	props := fp.GetAllProperties()
	assert.Equal(t, 7, len(props))
}

func TestGetFrequency(t *testing.T) {
	fp := genFrequentPropertiesFromFile(TEST_FILE_PATH)
	tmpMap := make(map[string]string)
	pmt := F.PropertyMapType{}

	//Case 1 [P2=V22]
	tmpMap["$P2"] = "V22"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	res, err := fp.GetFrequency(pmt)
	assert.Nil(t, err)
	assert.Equal(t, 782, res)

	//Case 2 [P1=V11,P2=V22]
	tmpMap = make(map[string]string)
	tmpMap["$P2"] = "V22"
	tmpMap["$P1"] = "V11"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	res, err = fp.GetFrequency(pmt)
	assert.Nil(t, err)
	assert.Equal(t, 376, res)

	//Case 3 [P6=V61,P1=V12,P3=V32,P2=V21]
	tmpMap = make(map[string]string)
	tmpMap["$P6"] = "V61"
	tmpMap["$P1"] = "V12"
	tmpMap["$P3"] = "V32"
	tmpMap["$P2"] = "V21"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	res, err = fp.GetFrequency(pmt)
	assert.Nil(t, err)
	assert.Equal(t, 291, res)

	//Case 4 [P1=V11]
	tmpMap = make(map[string]string)
	tmpMap["$P1"] = "V11"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	res, err = fp.GetFrequency(pmt)
	assert.NotNil(t, err)
}

func TestGetTopKCuts(t *testing.T) {
	fp := genFrequentPropertiesFromFile(TEST_FILE_PATH)
	res := fp.GetTopKCuts(3)
	freqitems := make([]F.FrequentItemset, 0)

	tmpMap := make(map[string]string)
	pmt := F.PropertyMapType{}
	tmpMap["$P6"] = "V61"
	tmpMap["$P7"] = "V72"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi := F.FrequentItemset{}
	fi.Frequency = 968
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	tmpMap = make(map[string]string)
	pmt = F.PropertyMapType{}
	tmpMap["$P4"] = "V41"
	tmpMap["$P5"] = "V51"
	tmpMap["$P7"] = "V71"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi = F.FrequentItemset{}
	fi.Frequency = 844
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	tmpMap = make(map[string]string)
	pmt = F.PropertyMapType{}
	tmpMap["$P2"] = "V22"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi = F.FrequentItemset{}
	fi.Frequency = 782
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	if reflect.DeepEqual(res, freqitems) {
		assert.Nil(t, nil)
	} else {
		assert.NotNil(t, nil)
	}
}

func TestGetTopKNameCuts(t *testing.T) {
	fp := genFrequentPropertiesFromFile(TEST_FILE_PATH)
	pnt1 := F.PropertyNameType{}
	pnt1.PropertyName = "$P2"
	pnt1.PropertyType = "event"
	lst := []F.PropertyNameType{pnt1}
	res := fp.GetTopKNameCuts(lst, 3)
	freqitems := make([]F.FrequentItemset, 0)

	tmpMap := make(map[string]string)
	pmt := F.PropertyMapType{}
	tmpMap["$P2"] = "V22"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi := F.FrequentItemset{}
	fi.Frequency = 782
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	tmpMap = make(map[string]string)
	pmt = F.PropertyMapType{}
	tmpMap["$P2"] = "V22"
	tmpMap["$P1"] = "V11"
	tmpMap["$P3"] = "V31"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi = F.FrequentItemset{}
	fi.Frequency = 484
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	tmpMap = make(map[string]string)
	pmt = F.PropertyMapType{}
	tmpMap["$P1"] = "V11"
	tmpMap["$P2"] = "V22"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi = F.FrequentItemset{}
	fi.Frequency = 376
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	if reflect.DeepEqual(res, freqitems) {
		assert.Nil(t, nil)
	} else {
		assert.NotNil(t, nil)
	}
}

func TestGetTopKNameValueCuts(t *testing.T) {
	fp := genFrequentPropertiesFromFile(TEST_FILE_PATH)
	tmpMap := make(map[string]string)
	tmpMap["$P2"] = "V22"
	tmpMap["$P1"] = "V11"
	pntv := F.PropertyMapType{}
	pntv.PropertyMap = tmpMap
	pntv.PropertyType = "event"
	res := fp.GetTopKNameValueCuts(pntv, 3)
	freqitems := make([]F.FrequentItemset, 0)

	tmpMap = make(map[string]string)
	pmt := F.PropertyMapType{}
	tmpMap["$P2"] = "V22"
	tmpMap["$P1"] = "V11"
	tmpMap["$P3"] = "V31"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi := F.FrequentItemset{}
	fi.Frequency = 484
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	tmpMap = make(map[string]string)
	pmt = F.PropertyMapType{}
	tmpMap["$P1"] = "V11"
	tmpMap["$P2"] = "V22"
	pmt.PropertyMap = tmpMap
	pmt.PropertyType = "event"
	fi = F.FrequentItemset{}
	fi.Frequency = 376
	fi.PropertyMapType = pmt
	freqitems = append(freqitems, fi)

	if reflect.DeepEqual(res, freqitems) {
		assert.Nil(t, nil)
	} else {
		assert.NotNil(t, nil)
	}
}
