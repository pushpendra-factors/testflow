package histogram

import (
	"fmt"
	"sort"
)

type CategoricalHistogram interface {
	Add([]string) error

	// When initialized with a template, can use dictionaries to add elements
	// to histogram.
	AddMap(m map[string]string) error

	PDF(x []string) (float64, error)

	Count() uint64

	// Internal for testing
	totalBinCount() uint64
	frequency(symbol string) uint64
}

type CategoricalHistogramStruct struct {
	Bins      []categoricalBin
	Maxbins   int
	Total     uint64
	Dimension int
	Template  *CategoricalHistogramTemplate
}

type CategoricalHistogramTemplateUnit struct {
	Name       string
	IsRequired bool
	Default    string
}

type CategoricalHistogramTemplate []CategoricalHistogramTemplateUnit

// New Categorical Histogram with d categoriacal variables and max n bins.
// TODO(aravind): Not returning the interface but the struct, since struct is
// required to marshal and unmarshal histograms. Fix it by adding methods to
// do it on the interface.
func NewCategoricalHistogram(
	n int, d int, t *CategoricalHistogramTemplate) (*CategoricalHistogramStruct, error) {
	if t != nil && len(*t) != d {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Mismatch in dimension %d and template length %d", d, len(*t)))
	}
	return &CategoricalHistogramStruct{
		Bins:      make([]categoricalBin, 0),
		Maxbins:   n,
		Total:     0,
		Dimension: d,
		Template:  t,
	}, nil
}

// Each bin has a separate frequency map for each of the d categorical variables.
type categoricalBin struct {
	FrequencyMaps []frequencyMap
	Count         uint64
}
type frequencyMap struct {
	Fmap  map[string]uint64
	Count uint64
}

const fMAP_MAX_SIZE = 5000
const fMAP_OTHER_KEY = "__OTHER__"

func (h *CategoricalHistogramStruct) PDF(x []string) (float64, error) {
	if h.Dimension != len(x) {
		return 0.0, fmt.Errorf(fmt.Sprintf(
			"Input dimension %d not matching histogram dimension %d.",
			len(x), h.Dimension))
	}
	totalProb := 0.0
	for i := range h.Bins {
		binProb := 1.0
		fMaps := h.Bins[i].FrequencyMaps
		for j := 0; j < h.Dimension; j++ {
			var varFreq uint64 = 0
			if count, ok := fMaps[j].Fmap[x[j]]; ok {
				varFreq = count
			}
			binProb *= float64(varFreq) / float64(fMaps[j].Count)
		}
		binFraction := float64(h.Bins[i].Count) / float64(h.Total)
		totalProb += (binFraction * binProb)
	}
	return totalProb, nil
}

func (h *CategoricalHistogramStruct) PDFMap(xMap map[string]string) (float64, error) {
	x := make([]string, h.Dimension)
	for i := 0; i < h.Dimension; i++ {
		eventName := (*h.Template)[i].Name
		if val, ok := xMap[eventName]; ok {
			x[i] = val
		} else {
			x[i] = ""
		}
	}
	return h.PDF(x)
}

func (h *CategoricalHistogramStruct) Add(values []string) error {
	if h.Dimension != len(values) {
		return fmt.Errorf(fmt.Sprintf(
			"Input dimension %d not matching histogram dimension %d.",
			len(values), h.Dimension))
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
	h.Bins = append(h.Bins, categoricalBin{FrequencyMaps: binFrequencyMaps, Count: 1})
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
			return fmt.Errorf(fmt.Sprintf("Missing required key %s in %v",
				template[i].Name, keyValues))
		}
		seenKeys[template[i].Name] = true
	}
	for k, _ := range keyValues {
		if _, ok := seenKeys[k]; !ok {
			return fmt.Errorf(fmt.Sprintf(
				"Unexpected key %s in %v", k, keyValues))
		}
	}
	return h.Add(vec)
}

func (h *CategoricalHistogramStruct) trim() {
	for len(h.Bins) > h.Maxbins {
		// Find closest bins in terms of value
		minDelta := 1e99
		min_i := 0
		min_j := 0
		for i := range h.Bins {
			binILikelihood := h.Bins[i].logLikelihood()
			for j := range h.Bins {
				if j <= i {
					continue
				}

				binJLikelihood := h.Bins[j].logLikelihood()
				mergedbin := h.Bins[i].merge(h.Bins[j])
				mergedLikelihood := mergedbin.logLikelihood()

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
		mergedbin := h.Bins[min_i].merge(h.Bins[min_j])

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

func (b *categoricalBin) merge(o categoricalBin) categoricalBin {
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
		// Trim the frequency maps to fMAP_MAX_SIZE.
		if len(mFmap.Fmap) > fMAP_MAX_SIZE {
			mFmap.Fmap = trimFrequencyMap(mFmap.Fmap, fMAP_MAX_SIZE)
		}
	}
	return categoricalBin{
		FrequencyMaps: mergedFmaps,
		Count:         b.Count + o.Count,
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
