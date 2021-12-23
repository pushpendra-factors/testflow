package histogram

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

const fMAP_MAX_SIZE = 200
const fMAP_MIN_SIZE = 50
const CHIST_MIN_BIN_SIZE = 1
const fMAP_OTHER_KEY = "__OTHER__"

type CategoricalHistogram interface {
	Add([]string) error

	// When initialized with a template, can use dictionaries to add elements
	// to histogram.
	AddMap(m map[string]string) error

	PDF(x []string) (float64, error)

	Count() uint64
	TrimByFmapSize(float64) error
	TrimByBinSize(float64) error

	// Internal for testing
	totalBinCount() uint64
	frequency(symbol string) uint64
	numBins() int
	maxFmapSize() int
}

type CategoricalHistogramStruct struct {
	Bins                            []categoricalBin              `json:"b"`
	Maxbins                         int                           `json:"mb"`
	MaxFmapSize                     int                           `json:"mf"`
	Total                           uint64                        `json:"to"`
	Dimension                       int                           `json:"d"`
	Template                        *CategoricalHistogramTemplate `json:"te"`
	binLogLikelihoodCacheLock       sync.RWMutex
	binLogLikelihoodCache           map[string]float64
	mergedBinLogLikelihoodCacheLock sync.RWMutex
	mergedBinLogLikelihoodCache     map[string]map[string]float64
}

type CategoricalHistogramTemplateUnit struct {
	Name       string `json:"n"`
	IsRequired bool   `json:"ir"`
	Default    string `json:"d"`
}

type CategoricalHistogramTemplate []CategoricalHistogramTemplateUnit

// New Categorical Histogram with d categoriacal variables and max n bins.
// TODO(aravind): Not returning the interface but the struct, since struct is
// required to marshal and unmarshal histograms. Fix it by adding methods to
// do it on the interface.
func NewCategoricalHistogram(
	n int, d int, t *CategoricalHistogramTemplate) (*CategoricalHistogramStruct, error) {
	if t != nil && len(*t) != d {
		return nil, fmt.Errorf(
			"Mismatch in dimension %d and template length %d", d, len(*t))
	}
	return &CategoricalHistogramStruct{
		Bins:                        make([]categoricalBin, 0),
		Maxbins:                     n,
		MaxFmapSize:                 fMAP_MAX_SIZE,
		Total:                       0,
		Dimension:                   d,
		Template:                    t,
		binLogLikelihoodCache:       make(map[string]float64),
		mergedBinLogLikelihoodCache: make(map[string]map[string]float64),
	}, nil
}

// Each bin has a separate frequency map for each of the d categorical variables.
type categoricalBin struct {
	FrequencyMaps []frequencyMap `json:"fm"`
	Count         uint64         `json:"c"`
	// uuid required for internal caching.
	uuid string
}
type frequencyMap struct {
	Fmap  map[string]uint64 `json:"fm"`
	Count uint64            `json:"c"`
}

func (h *CategoricalHistogramStruct) PDF(x []string) (float64, error) {
	if h.Dimension != len(x) {
		return 0.0, fmt.Errorf(
			"Input dimension %d not matching histogram dimension %d.",
			len(x), h.Dimension)
	}
	if h.Total < 1 {
		return 0.0, nil
	}
	// Assume there are k Bins B1, B2 ... Bk
	// Assume there are N items with A1, A2, .. Ak items in each bin
	// such that A1 + A2 + ... Ak = N
	// The final probability of one variable P(X=x1) = (n1 / A1 + n2 / A2 + ... nk / Ak)
	// where n1 is the number of time x1 is seen.
	// TODO(If count for x1 is missing in bin, it should be considered from the _OTHER_ bin rather than zero).
	totalProb := 0.0
	for i := range h.Bins {
		binProb := 1.0
		fMaps := h.Bins[i].FrequencyMaps
		for j := 0; j < h.Dimension; j++ {
			if x[j] == "" {
				continue
			}
			var operator string
			values := strings.Split(x[j], ",")
			if values[0] == "!=" {
				operator = values[0]
				values = values[1:]
			}
			if fMaps[j].Count < 1 {
				binProb = 0.0
				break
			}
			var varFreq uint64 = 0
			for _, value := range values {
				if count, ok := fMaps[j].Fmap[value]; ok {
					varFreq += count
				}
			}
			if operator == "!=" {
				varFreq = (fMaps[j].Count - varFreq)
			}
			binProb *= float64(varFreq) / float64(h.Bins[i].Count)
		}
		binFraction := float64(h.Bins[i].Count) / float64(h.Total)
		totalProb += (binFraction * binProb)
	}
	return totalProb, nil
}

func (h *CategoricalHistogramStruct) PDFFromMap(xMap map[string]string) (float64, error) {
	x := make([]string, h.Dimension)
	numOfFilterValues := int(0)
	for i := 0; i < h.Dimension; i++ {
		eventName := (*h.Template)[i].Name
		if val, ok := xMap[eventName]; ok {
			x[i] = val
			numOfFilterValues++
		} else {
			x[i] = ""
		}
	}
	if numOfFilterValues != len(xMap) {
		return 0.0, nil
	}
	return h.PDF(x)
}

func (h *CategoricalHistogramStruct) Add(values []string) error {
	if h.Dimension != len(values) {
		return fmt.Errorf(
			"Input dimension %d not matching histogram dimension %d.",
			len(values), h.Dimension)
	}
	h.Total++
	binFrequencyMaps := make([]frequencyMap, h.Dimension)
	for i := 0; i < h.Dimension; i++ {
		binFrequencyMaps[i].Fmap = make(map[string]uint64)
		binFrequencyMaps[i].Count = 0
		if values[i] != "" {
			// If the value of a variable is empty, it is assumed to be missing.
			// Hence each frequencyMap has it's own separate total count.
			binFrequencyMaps[i].Fmap[values[i]] = 1
			binFrequencyMaps[i].Count = 1
		}
	}
	h.Bins = append(h.Bins, categoricalBin{FrequencyMaps: binFrequencyMaps, Count: 1, uuid: randomLowerAphaNumString(32)})
	h.trim()
	return nil
}

func (h *CategoricalHistogramStruct) AddMap(keyValues map[string]string) error {
	if h.Template == nil {
		return fmt.Errorf("Template not initialized")
	}
	seenKeys := map[string]bool{}
	vec := make([]string, h.Dimension)
	template := *h.Template
	for i := range template {
		if value, ok := keyValues[template[i].Name]; ok {
			vec[i] = value
		} else if !template[i].IsRequired {
			// If Default value is not set it is set to "", which is
			// assumed to be missing.
			vec[i] = template[i].Default
		} else {
			return fmt.Errorf("Missing required key %s in %v",
				template[i].Name, keyValues)
		}
		seenKeys[template[i].Name] = true
	}
	for k := range keyValues {
		if _, ok := seenKeys[k]; !ok {
			return fmt.Errorf(
				"Unexpected key %s in %v ,seenValues  %v ", k, keyValues, seenKeys)
		}
	}
	return h.Add(vec)
}

func (h *CategoricalHistogramStruct) getBinLogLikelihood(bin *categoricalBin) float64 {

	h.binLogLikelihoodCacheLock.RLock()
	if l, ok := h.binLogLikelihoodCache[bin.uuid]; ok {
		h.binLogLikelihoodCacheLock.RUnlock()
		return l
	}
	h.binLogLikelihoodCacheLock.RUnlock()

	l := bin.logLikelihood()
	h.binLogLikelihoodCacheLock.Lock()
	h.binLogLikelihoodCache[bin.uuid] = l
	h.binLogLikelihoodCacheLock.Unlock()
	return l
}

func (h *CategoricalHistogramStruct) getMergedBinLogLikelihood(
	bin1 *categoricalBin, bin2 *categoricalBin, maxFmapSize int) float64 {
	h.mergedBinLogLikelihoodCacheLock.RLock()
	if lMap, ok1 := h.mergedBinLogLikelihoodCache[bin1.uuid]; ok1 {
		if l, ok2 := lMap[bin2.uuid]; ok2 {
			h.mergedBinLogLikelihoodCacheLock.RUnlock()
			return l
		}
	} else if lMap, ok1 := h.mergedBinLogLikelihoodCache[bin2.uuid]; ok1 {
		if l, ok2 := lMap[bin1.uuid]; ok2 {
			h.mergedBinLogLikelihoodCacheLock.RUnlock()
			return l
		}
	}
	h.mergedBinLogLikelihoodCacheLock.RUnlock()
	mergedbin := (*bin1).merge(*bin2, maxFmapSize)
	mergedLikelihood := mergedbin.logLikelihood()
	h.mergedBinLogLikelihoodCacheLock.Lock()
	// Add to cache;
	lMap, ok := h.mergedBinLogLikelihoodCache[bin1.uuid]
	if !ok {
		lMap = make(map[string]float64)
	}
	lMap[bin2.uuid] = mergedLikelihood
	h.mergedBinLogLikelihoodCache[bin1.uuid] = lMap
	h.mergedBinLogLikelihoodCacheLock.Unlock()
	return mergedLikelihood
}

func (h *CategoricalHistogramStruct) cleanCache() {
	// Clear cached entries of non existent bins.
	currentIDS := make(map[string]bool)
	for _, bin := range h.Bins {
		if bin.uuid != "" {
			currentIDS[bin.uuid] = true
		}
	}

	// Delete from binLogLikelihoodCache.
	idsToDelete := []string{}
	for id, _ := range h.binLogLikelihoodCache {
		if _, ok := currentIDS[id]; !ok {
			idsToDelete = append(idsToDelete, id)
		}
	}
	for _, id := range idsToDelete {
		delete(h.binLogLikelihoodCache, id)
	}

	// Delete from mergedBinLogLikelihoodCache.
	primaryIDSToDelete := []string{}
	for pid, secMap := range h.mergedBinLogLikelihoodCache {
		if _, ok1 := currentIDS[pid]; !ok1 {
			primaryIDSToDelete = append(primaryIDSToDelete, pid)
			continue
		}
		secondaryIDSToDelete := []string{}
		for sid, _ := range secMap {
			if _, ok2 := currentIDS[sid]; !ok2 {
				secondaryIDSToDelete = append(secondaryIDSToDelete, sid)
				continue
			}
		}
		for _, id := range secondaryIDSToDelete {
			delete(secMap, id)
		}
	}
	for _, id := range primaryIDSToDelete {
		delete(h.mergedBinLogLikelihoodCache, id)
	}
}

func (h *CategoricalHistogramStruct) trim() {
	h.concurrentTrim()
	h.cleanCache()
}

func (h *CategoricalHistogramStruct) linearTrim() {
	for len(h.Bins) > h.Maxbins {
		// Find closest bins in terms of value
		minDelta := 1e99
		min_i := 0
		min_j := 0
		for i := range h.Bins {
			if h.Bins[i].uuid == "" {
				h.Bins[i].uuid = randomLowerAphaNumString(32)
			}

			binILikelihood := h.getBinLogLikelihood(&h.Bins[i])
			for j := range h.Bins {
				if j <= i {
					continue
				}

				if h.Bins[j].uuid == "" {
					h.Bins[j].uuid = randomLowerAphaNumString(32)
				}

				binJLikelihood := h.getBinLogLikelihood(&h.Bins[j])
				mergedLikelihood := h.getMergedBinLogLikelihood(&h.Bins[i], &h.Bins[j], h.MaxFmapSize)

				// Select the bin whose bin1LogLikelihood + bin2LogLikelihood - mergedLogLikelihood is minimum.
				// i.e. the one which causes minimum drop in the overall likelihood as a result of merging.
				if delta := binILikelihood + binJLikelihood - mergedLikelihood; delta < minDelta {
					minDelta = delta
					min_i = i
					min_j = j
				}
			}
		}

		// We need to merge bins min_i-1 and min_j
		mergedbin := h.Bins[min_i].merge(h.Bins[min_j], h.MaxFmapSize)

		// Remove min_i and min_j bins
		min, max := sortTuple(min_i, min_j)

		head := h.Bins[0:min]
		mid := h.Bins[min+1 : max]
		tail := h.Bins[max+1:]

		h.Bins = append(head, mid...)
		h.Bins = append(h.Bins, tail...)

		h.Bins = append(h.Bins, mergedbin)
	}
}

type result struct {
	i, j                                             int
	binILikelihood, binJLikelihood, mergedLikelihood float64
}

func (h *CategoricalHistogramStruct) concurrentTrim() {

	for len(h.Bins) > h.Maxbins {
		var wg sync.WaitGroup
		collectResult := make(chan result, (len(h.Bins)*len(h.Bins))/2)

		for i := range h.Bins {
			if h.Bins[i].uuid == "" {
				h.Bins[i].uuid = randomLowerAphaNumString(32)
			}
			for j := i + 1; j < len(h.Bins); j++ {
				if h.Bins[j].uuid == "" {
					h.Bins[j].uuid = randomLowerAphaNumString(32)
				}
				wg.Add(1)
				go func(x int, y int) {
					defer wg.Done()
					h.calculateLikelihood(x, y, collectResult)
				}(i, j)
			}
		}

		wg.Wait()
		close(collectResult)

		minDelta := 1e99
		min_i := 0
		min_j := 0

		for result := range collectResult {
			if delta := result.binILikelihood + result.binJLikelihood - result.mergedLikelihood; delta < minDelta {
				minDelta = delta
				min_i = result.i
				min_j = result.j
			}
		}

		// We need to merge bins min_i-1 and min_j
		mergedbin := h.Bins[min_i].merge(h.Bins[min_j], h.MaxFmapSize)

		// Remove min_i and min_j bins
		min, max := sortTuple(min_i, min_j)

		head := h.Bins[0:min]
		mid := h.Bins[min+1 : max]
		tail := h.Bins[max+1:]

		h.Bins = append(head, mid...)
		h.Bins = append(h.Bins, tail...)

		h.Bins = append(h.Bins, mergedbin)

	}
}

func (h *CategoricalHistogramStruct) calculateLikelihood(i, j int, publishResults chan result) {
	binILikelihood := h.getBinLogLikelihood(&h.Bins[i])
	binJLikelihood := h.getBinLogLikelihood(&h.Bins[j])
	mergedLikelihood := h.getMergedBinLogLikelihood(&h.Bins[i], &h.Bins[j], h.MaxFmapSize)

	publishResults <- result{
		i:                i,
		j:                j,
		binILikelihood:   binILikelihood,
		binJLikelihood:   binJLikelihood,
		mergedLikelihood: mergedLikelihood,
	}
}

func (b *categoricalBin) merge(o categoricalBin, maxFmapSize int) categoricalBin {
	dimension := len(b.FrequencyMaps)
	// Initialize merged Frequency Maps.
	mergedFmaps := make([]frequencyMap, dimension)
	for i := 0; i < dimension; i++ {
		mergedFmaps[i] = frequencyMap{Fmap: make(map[string]uint64), Count: 0}
	}
	for i := 0; i < dimension; i++ {
		mFmap := &mergedFmaps[i]
		mFmap.Count = o.FrequencyMaps[i].Count + b.FrequencyMaps[i].Count
		for k, bCount := range b.FrequencyMaps[i].Fmap {
			if _, ok := mFmap.Fmap[k]; !ok {
				// Add to mergedMap if not already merged.
				var oCount uint64 = 0
				if ov, ok := o.FrequencyMaps[i].Fmap[k]; ok {
					oCount = ov
				}
				mFmap.Fmap[k] = bCount + oCount
			}
		}
		// Loop over o to merge ones not present in b.
		for k, oCount := range o.FrequencyMaps[i].Fmap {
			if _, ok := mFmap.Fmap[k]; !ok {
				// Add to mergedMap if not already merged.
				var bCount uint64 = 0
				if bv, ok := b.FrequencyMaps[i].Fmap[k]; ok {
					bCount = bv
				}
				mFmap.Fmap[k] = bCount + oCount
			}
		}
		// Trim the frequency maps to maxFmapSize.
		if len(mFmap.Fmap) > maxFmapSize {
			mFmap.Fmap = trimFrequencyMap(mFmap.Fmap, maxFmapSize)
		}
	}
	return categoricalBin{
		FrequencyMaps: mergedFmaps,
		Count:         b.Count + o.Count,
		uuid:          randomLowerAphaNumString(32),
	}
}

func trimFrequencyMap(fmap map[string]uint64, maxSize int) map[string]uint64 {
	if len(fmap) < maxSize {
		return fmap
	}

	type kv struct {
		key   string
		value uint64
	}
	var ss []kv
	for k, v := range fmap {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		// Move fmap_OTHER_KEY to the end.
		if ss[i].key == fMAP_OTHER_KEY {
			return false
		} else if ss[j].key == fMAP_OTHER_KEY {
			return true
		}
		// Remaining are sorted in descending order.
		return ss[i].value > ss[j].value
	})

	trimmedFmap := map[string]uint64{}
	for i, kv := range ss {
		if i < maxSize-1 {
			trimmedFmap[kv.key] = kv.value
		} else {
			if count, ok := trimmedFmap[fMAP_OTHER_KEY]; ok {
				trimmedFmap[fMAP_OTHER_KEY] = count + kv.value
			} else {
				trimmedFmap[fMAP_OTHER_KEY] = kv.value
			}
		}
	}

	return trimmedFmap
}

func (b *categoricalBin) logLikelihood() float64 {
	// The multinomial log likelihood of a frequency distribution
	// F = (f1, f2, ...fk) with f1 + f2 .. + fk = N is
	// L(F) = f1*logf1 + f2*logf2 .. + fk*logfk - N * logN
	// The log likelihood of d frequency distributions F1, F2, .. Fd assuming independence is.
	// L(F1) + L(F2) ... + L(Fd)
	var totalLh float64
	for i := 0; i < len(b.FrequencyMaps); i++ {
		N := float64(b.FrequencyMaps[i].Count)
		lh := -N * log(N)
		for _, intF := range b.FrequencyMaps[i].Fmap {
			f := float64(intF)
			lh += f * log(f)
		}
		totalLh += lh
	}
	return totalLh
}

func (h *CategoricalHistogramStruct) Count() uint64 {
	return h.Total
}

func (h *CategoricalHistogramStruct) frequency(symbol string) uint64 {
	var symbolCount uint64 = 0
	for l := range h.Bins {
		for m := range h.Bins[l].FrequencyMaps {
			fm := h.Bins[l].FrequencyMaps[m]
			if c, ok := fm.Fmap[symbol]; ok {
				symbolCount += c
			}
		}
	}
	return symbolCount
}

func (h *CategoricalHistogramStruct) totalBinCount() uint64 {
	var c uint64 = 0
	for l := range h.Bins {
		c += h.Bins[l].Count
	}
	return c
}

func (h *CategoricalHistogramStruct) numBins() int {
	return len(h.Bins)
}

func (h *CategoricalHistogramStruct) maxFmapSize() int {
	return h.MaxFmapSize
}

func (h *CategoricalHistogramStruct) TrimByFmapSize(trimFraction float64) error {
	if trimFraction <= 0 || trimFraction > 1.0 {
		return fmt.Errorf("Unexpected value of trimFraction: %f", trimFraction)
	}
	newMaxFMapSize := int(math.Max(float64(h.MaxFmapSize)*trimFraction, fMAP_MIN_SIZE))
	if newMaxFMapSize >= h.MaxFmapSize {
		return fmt.Errorf(
			"No trimming required, h.MaxFmapSize:%d, newMaxFMapSize: %d, trimFraction: %f",
			h.MaxFmapSize, newMaxFMapSize, trimFraction)
	}
	for l := range h.Bins {
		for m := range h.Bins[l].FrequencyMaps {
			h.Bins[l].FrequencyMaps[m].Fmap = trimFrequencyMap(
				h.Bins[l].FrequencyMaps[m].Fmap, newMaxFMapSize)
		}
	}
	h.MaxFmapSize = newMaxFMapSize
	return nil
}

func (h *CategoricalHistogramStruct) TrimByBinSize(trimFraction float64) error {
	if trimFraction <= 0 || trimFraction > 1.0 {
		return fmt.Errorf("Unexpected value of trimFraction: %f", trimFraction)
	}
	newMaxbins := int(math.Max(float64(h.Maxbins)*trimFraction, CHIST_MIN_BIN_SIZE))
	if newMaxbins >= h.Maxbins {
		return fmt.Errorf(
			"No trimming required, h.MaxBins:%d, newMaxbins: %d, trimFraction: %f",
			h.Maxbins, newMaxbins, trimFraction)
	}
	h.Maxbins = newMaxbins
	h.trim()
	return nil
}

func (h *CategoricalHistogramStruct) GetBinValues(key string) []string {
	ranges := make([]string, 0)
	dim := -1
	for i := 0; i < h.Dimension; i++ {
		if (*h.Template)[i].Name == key {
			dim = i
			break
		}
	}
	if dim < 0 {
		return ranges
	}
	valueCountMap := make(map[string]uint64)
	for _, bin := range h.Bins {
		for value, count := range bin.FrequencyMaps[dim].Fmap {
			valueCountMap[value] += count
		}
	}
	for value, _ := range valueCountMap {
		ranges = append(ranges, value)
	}
	return ranges
}
