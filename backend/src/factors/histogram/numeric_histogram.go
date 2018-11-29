package histogram

import (
	"fmt"
	"math"
)

type NumericHistogram interface {
	Add(v []float64) error

	// When initialized with a template, can use dictionaries to add elements
	// to histogram.
	AddMap(m map[string]float64) error

	Mean() []float64

	Variance() []float64

	CDF(x []float64) float64

	Count() uint64
}

type NumericHistogramStruct struct {
	Bins      []numericBin              `json:"b"`
	Maxbins   int                       `json:"mb"`
	Total     uint64                    `json:"to"`
	Dimension int                       `json:"d"`
	Template  *NumericHistogramTemplate `json:"te"`
}

type NumericHistogramTemplateUnit struct {
	Name       string  `json:"n"`
	IsRequired bool    `json:"ir"`
	Default    float64 `json:"d"`
}

type NumericHistogramTemplate []NumericHistogramTemplateUnit

// New multidimensional Histogram with d dimensions and max n bins.
// TODO(aravind): Not returning the interface but the struct, since struct is
// required to marshal and unmarshal histograms. Fix it by adding methods to
// do it on the interface.
func NewNumericHistogram(n int, d int, t *NumericHistogramTemplate) (*NumericHistogramStruct, error) {
	if t != nil && len(*t) != d {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Mismatch in dimension %d and template length %d", d, len(*t)))
	}
	return &NumericHistogramStruct{
		Bins:      make([]numericBin, 0),
		Maxbins:   n,
		Total:     0,
		Dimension: d,
		Template:  t,
	}, nil
}

type numericBin struct {
	Count float64 `json:"c"`
	Min   vector  `json:"mn"`
	Max   vector  `json:"mx"`
}

// http://www.science.canterbury.ac.nz/nzns/issues/vol7-1979/duncan_b.pdf
func (b *numericBin) merge(o numericBin) numericBin {
	dimension := b.Min.Dimension()

	count := b.Count + o.Count

	min := make([]float64, dimension)
	max := make([]float64, dimension)

	for i := 0; i < dimension; i++ {
		if b.Min.Values[i] <= o.Min.Values[i] {
			min[i] = b.Min.Values[i]
		} else {
			min[i] = o.Min.Values[i]
		}

		if b.Max.Values[i] >= o.Max.Values[i] {
			max[i] = b.Max.Values[i]
		} else {
			max[i] = o.Max.Values[i]
		}

	}

	return numericBin{
		Count: count,
		Min:   NewVector(min),
		Max:   NewVector(max),
	}
}

func (h *NumericHistogramStruct) Add(values []float64) error {
	v := NewVector(values)
	if h.Dimension != v.Dimension() {
		return fmt.Errorf(
			fmt.Sprintf("Input dimension %d does not match histogram dimension %d",
				v.Dimension(), h.Dimension))
	}
	h.Total++
	h.Bins = append(h.Bins, numericBin{Count: 1, Min: v, Max: v})
	h.trim()
	return nil
}

func (h *NumericHistogramStruct) AddMap(keyValues map[string]float64) error {
	if h.Template == nil {
		return fmt.Errorf("Template not initialized")
	}
	seenKeys := map[string]bool{}
	vec := make([]float64, h.Dimension)
	template := *h.Template
	for i := range template {
		if value, ok := keyValues[template[i].Name]; ok {
			vec[i] = value
		} else if !template[i].IsRequired {
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

func (h *NumericHistogramStruct) Mean() []float64 {
	if h.Total == 0 {
		return []float64{}
	}

	sum := make([]float64, h.Dimension)

	for i := range h.Bins {
		for j := range sum {
			minIJ := h.Bins[i].Min.Values[j]
			maxIJ := h.Bins[i].Max.Values[j]
			meanIJ := minIJ + (maxIJ-minIJ)/2.0
			sum[j] += meanIJ * h.Bins[i].Count
		}
	}

	for k, s := range sum {
		s = s / float64(h.Total)
		sum[k] = s
	}
	return sum
}

func (h *NumericHistogramStruct) MeanMap() map[string]float64 {
	if h.Template == nil {
		return nil
	}
	mean := h.Mean()
	meanMap := make(map[string]float64)
	for i := 0; i < len(*h.Template); i++ {
		meanMap[(*h.Template)[i].Name] = mean[i]
	}
	return meanMap
}

func (h *NumericHistogramStruct) CDF(x []float64) float64 {
	xVec := NewVector(x)
	if xVec.Dimension() != h.Dimension {
		return -1
	}
	sum := 0.0
	for i := range h.Bins {
		count := h.Bins[i].Count
		for j := 0; j < h.Dimension; j++ {
			var (
				factor float64
				x      = xVec.Values[j]
				min    = h.Bins[i].Min.Values[j]
				max    = h.Bins[i].Max.Values[j]
			)
			if x < min {
				factor = 0
			} else if x >= max {
				factor = 1
			} else {
				factor = (x - min) / float64(max-min)
			}
			count *= factor
		}
		sum += count
	}

	return sum / float64(h.Total)
}

func (h *NumericHistogramStruct) CDFFromMap(xMap map[string]float64) float64 {
	x := make([]float64, h.Dimension)
	for i := 0; i < h.Dimension; i++ {
		eventName := (*h.Template)[i].Name
		if val, ok := xMap[eventName]; ok {
			x[i] = val
		} else {
			x[i] = math.MaxFloat64
		}
	}
	return h.CDF(x)
}

func (h *NumericHistogramStruct) trim() {
	for len(h.Bins) > h.Maxbins {
		// Find closest bins in terms of value
		minDelta := 1e99
		min_i := 0
		min_j := 0
		for i := range h.Bins {
			for j := range h.Bins {
				if j <= i {
					continue
				}

				vol_i := 1.0
				vol_j := 1.0
				vol := 1.0
				for k := 0; k < h.Dimension; k++ {
					val_max_i := h.Bins[i].Max.Values[k]
					val_min_i := h.Bins[i].Min.Values[k]

					val_max_j := h.Bins[j].Max.Values[k]
					val_min_j := h.Bins[j].Min.Values[k]

					vol_i *= val_max_i - val_min_i
					vol_j *= val_max_j - val_min_j
					vol *= max(val_max_i, val_max_j) - min(val_min_i, val_min_j)
				}

				count_i := h.Bins[i].Count
				count_j := h.Bins[j].Count

				// The propability of each data point occuring within bin boundaries is 1 / volBin, assuming it to be uniformly distributed.
				// The probability / likelihood of N such data points being in the bin is (1 / volBin)^N.
				// The log likelihood is -N * log(volBin)
				// The log likelihood of merged bin is -(N1 + N2) * log(mergedVol)
				// Select the bin whose bin1LogLikelihood + bin2LogLikelihood - mergedLogLikelihood is minimum.
				// i.e. the one which causes minimum drop in the overall likelihood as a result of merging.
				if delta := (count_i+count_j)*log(vol) - count_i*log(vol_i) - count_j*log(vol_j); delta < minDelta {
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

func (h *NumericHistogramStruct) Count() uint64 {
	return h.Total
}
