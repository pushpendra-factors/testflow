package gohistogram

// Copyright (c) 2013 VividCortex, Inc. All rights reserved.
// Please see the LICENSE file for applicable license terms.

import (
	"fmt"
)

type NumericHistogram struct {
	Bins    []Bin
	Maxbins int
	Total   uint64
}

// NewHistogram returns a new NumericHistogram with a maximum of n bins.
//
// There is no "optimal" bin count, but somewhere between 20 and 80 bins
// should be sufficient.
func NewHistogram(n int) *NumericHistogram {
	return &NumericHistogram{
		Bins:    make([]Bin, 0),
		Maxbins: n,
		Total:   0,
	}
}

func (h *NumericHistogram) Add(n float64) {
	defer h.trim()
	h.Total++
	for i := range h.Bins {
		if h.Bins[i].Value == n {
			h.Bins[i].Count++
			return
		}

		if h.Bins[i].Value > n {

			newbin := Bin{Value: n, Count: 1}
			head := append(make([]Bin, 0), h.Bins[0:i]...)

			head = append(head, newbin)
			tail := h.Bins[i:]
			h.Bins = append(head, tail...)
			return
		}
	}

	h.Bins = append(h.Bins, Bin{Count: 1, Value: n})
}

func (h *NumericHistogram) Quantile(q float64) float64 {
	count := q * float64(h.Total)
	for i := range h.Bins {
		count -= float64(h.Bins[i].Count)

		if count <= 0 {
			return h.Bins[i].Value
		}
	}

	return -1
}

// CDF returns the Value of the cumulative distribution function
// at x
func (h *NumericHistogram) CDF(x float64) float64 {
	count := 0.0
	for i := range h.Bins {
		if h.Bins[i].Value <= x {
			count += float64(h.Bins[i].Count)
		}
	}

	return count / float64(h.Total)
}

// Mean returns the sample mean of the distribution
func (h *NumericHistogram) Mean() float64 {
	if h.Total == 0 {
		return 0
	}

	sum := 0.0

	for i := range h.Bins {
		sum += h.Bins[i].Value * h.Bins[i].Count
	}

	return sum / float64(h.Total)
}

// Variance returns the variance of the distribution
func (h *NumericHistogram) Variance() float64 {
	if h.Total == 0 {
		return 0
	}

	sum := 0.0
	mean := h.Mean()

	for i := range h.Bins {
		sum += (h.Bins[i].Count * (h.Bins[i].Value - mean) * (h.Bins[i].Value - mean))
	}

	return sum / float64(h.Total)
}

func (h *NumericHistogram) Count() float64 {
	return float64(h.Total)
}

// trim merges adjacent bins to decrease the bin count to the maximum Value
func (h *NumericHistogram) trim() {
	for len(h.Bins) > h.Maxbins {
		// Find closest bins in terms of Value
		minDelta := 1e99
		minDeltaIndex := 0
		for i := range h.Bins {
			if i == 0 {
				continue
			}

			if delta := h.Bins[i].Value - h.Bins[i-1].Value; delta < minDelta {
				minDelta = delta
				minDeltaIndex = i
			}
		}

		// We need to merge bins minDeltaIndex-1 and minDeltaIndex
		totalCount := h.Bins[minDeltaIndex-1].Count + h.Bins[minDeltaIndex].Count
		mergedbin := Bin{
			Value: (h.Bins[minDeltaIndex-1].Value*
				h.Bins[minDeltaIndex-1].Count +
				h.Bins[minDeltaIndex].Value*
					h.Bins[minDeltaIndex].Count) /
				totalCount, // weighted average
			Count: totalCount, // summed heights
		}
		head := append(make([]Bin, 0), h.Bins[0:minDeltaIndex-1]...)
		tail := append([]Bin{mergedbin}, h.Bins[minDeltaIndex+1:]...)
		h.Bins = append(head, tail...)
	}
}

// String returns a string reprentation of the histogram,
// which is useful for printing to a terminal.
func (h *NumericHistogram) String() (str string) {
	str += fmt.Sprintln("Total:", h.Total)

	for i := range h.Bins {
		var bar string
		for j := 0; j < int(float64(h.Bins[i].Count)/float64(h.Total)*200); j++ {
			bar += "."
		}
		str += fmt.Sprintln(h.Bins[i].Value, "\t", bar)
	}

	return
}
