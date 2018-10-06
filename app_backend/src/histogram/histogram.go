package histogram

import (
	"fmt"
)

type Histogram interface {
	Add(vector []float64)

	Mean() []float64

	Variance() []float64

	CDF(x []float64) float64

	String() (str string)

	Count() float64
}

type histogram struct {
	Bins      []bin
	Maxbins   int
	Total     uint64
	Dimension int
}

// New multidimensional Histogram with d dimensions and max n bins.
func NewHistogram(n int, d int) Histogram {
	return &histogram{
		Bins:      make([]bin, 0),
		Maxbins:   n,
		Total:     0,
		Dimension: d,
	}
}

type bin struct {
	Mean     vector
	Variance vector
	Count    float64
	Min      vector
	Max      vector
}

// http://www.science.canterbury.ac.nz/nzns/issues/vol7-1979/duncan_b.pdf
func (b *bin) Merge(o bin) bin {
	dimension := b.Mean.Dimension()

	count := b.Count + o.Count

	mean := make([]float64, dimension)
	variance := make([]float64, dimension)
	min := make([]float64, dimension)
	max := make([]float64, dimension)

	for i := 0; i < dimension; i++ {
		mean[i] = (b.Count*b.Mean.Values[i] + o.Count*o.Mean.Values[i]) / float64(count)

		variance[i] =
			((b.Count*(b.Variance.Values[i]+b.Mean.Values[i]*b.Mean.Values[i]) +
				o.Count*(o.Variance.Values[i]+o.Mean.Values[i]*o.Mean.Values[i])) / float64(count)) - mean[i]*mean[i]

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

	return bin{
		Mean:     NewVector(mean),
		Variance: NewVector(variance),
		Count:    count,
		Min:      NewVector(min),
		Max:      NewVector(max),
	}
}

func (h *histogram) Add(values []float64) {
	m := NewVector(values)
	v := NewVector(make([]float64, len(values)))

	if h.Dimension != m.Dimension() {
		return
	}
	h.Total++
	for i := range h.Bins {
		if h.Bins[i].Mean.Equals(v) {
			h.Bins[i].Count++
			return
		}
	}
	h.Bins = append(h.Bins, bin{Count: 1, Mean: m, Variance: v, Min: m, Max: m})
	h.trim()
}

func (h *histogram) Mean() []float64 {
	if h.Total == 0 {
		return []float64{}
	}

	sum := make([]float64, h.Dimension)

	for i := range h.Bins {
		for j := range sum {
			sum[j] += h.Bins[i].Mean.Values[j] * h.Bins[i].Count
		}
	}

	for k, s := range sum {
		s = s / float64(h.Total)
		sum[k] = s
	}
	return sum
}

// http://www.science.canterbury.ac.nz/nzns/issues/vol7-1979/duncan_b.pdf
func (h *histogram) Variance() []float64 {
	if h.Total == 0 {
		return []float64{}
	}

	sum := make([]float64, h.Dimension)
	mean := h.Mean()

	for i := range h.Bins {
		for j := range sum {
			sum[j] += (h.Bins[i].Count * (h.Bins[i].Variance.Values[j] + h.Bins[i].Mean.Values[j]*h.Bins[i].Mean.Values[j]))
		}
	}

	for k, _ := range sum {
		sum[k] = sum[k] / float64(h.Total)
		sum[k] = sum[k] - mean[k]*mean[k]
	}
	return sum
}

func (h *histogram) CDF(x []float64) float64 {
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

func (h *histogram) String() (str string) {
	str += fmt.Sprintln("Total:", h.Total)

	for i := range h.Bins {
		var bar string
		for j := 0; j < int(h.Bins[i].Count); j++ {
			bar += "."
		}
		str += fmt.Sprintln(h.Bins[i].Mean.String(), h.Bins[i].Min.String(), h.Bins[i].Max.String(), "\t", h.Count())
	}
	return
}

func (h *histogram) Count() float64 {
	return float64(h.Total)
}

func (h *histogram) trim() {
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

				if delta := (count_i+count_j)*log(vol) - count_i*log(vol_i) - count_j*log(vol_j); delta < minDelta {
					minDelta = delta
					min_i = i
					min_j = j
				}

			}
		}

		// We need to merge bins min_i-1 and min_j
		mergedbin := h.Bins[min_i].Merge(h.Bins[min_j])

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
