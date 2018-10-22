package histogram

import (
	"fmt"
)

type CategoricalHistogram interface {
	Add([]string) error

	PDF(x []string) (float64, error)

	Count() uint64

	// Internal for testing
	totalBinCount() uint64
	frequency(symbol string) uint64
}

type categoricalHistogram struct {
	Bins      []categoricalBin
	Maxbins   int
	Total     uint64
	Dimension int
}

// New Categorical Histogram with d categoriacal variables and max n bins.
func NewCategoricalHistogram(n int, d int) CategoricalHistogram {
	return &categoricalHistogram{
		Bins:      make([]categoricalBin, 0),
		Maxbins:   n,
		Total:     0,
		Dimension: d,
	}
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

func (h *categoricalHistogram) PDF(x []string) (float64, error) {
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

func (h *categoricalHistogram) Add(values []string) error {
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

func (h *categoricalHistogram) trim() {
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
		min, max := sort(min_i, min_j)

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
	}
	return categoricalBin{
		FrequencyMaps: mergedFmaps,
		Count:         b.Count + o.Count,
	}
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

func (h *categoricalHistogram) Count() uint64 {
	return h.Total
}

func (h *categoricalHistogram) frequency(symbol string) uint64 {
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

func (h *categoricalHistogram) totalBinCount() uint64 {
	var c uint64 = 0
	for l := range h.Bins {
		c += h.Bins[l].Count
	}
	return c
}
