package histogram

import (
	"fmt"
)

type vector struct {
	Values []float64
}

func NewVector(v []float64) vector {
	return vector{Values: v}
}

func (v *vector) Dimension() int {
	return len(v.Values)
}

func (v *vector) String() string {
	return fmt.Sprintf("%v", v.Values)
}

func (v *vector) Equals(o vector) bool {
	if v.Dimension() != o.Dimension() {
		return false
	}
	for i := range v.Values {
		if v.Values[i] != o.Values[i] {
			return false
		}
	}
	return true
}
