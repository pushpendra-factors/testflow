package histogram

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Benchmarking and tests for data in sample_data.go
// Data contains 10000 random samples each from symmetrical Gaussian distributions
// of dimension 1, 2, 3, 4 and 5. Each Gaussian distribution haz zero mean and a
// variance 10000. The data is available in the variables dataDimension1,
// dataDimension2,  dataDimension3,  dataDimension4 and  dataDimension5
// respectively.

func computeCDFUsingData(data [][]float64, evalPoint []float64) float64 {
	sum := 0.0
	count := float64(len(data))
	for i := range data {
		if less(data[i], evalPoint) {
			sum += 1
		}
	}
	return sum / count
}

func buildNumericHistogramFromData(maxBins int, dimensions int, data [][]float64) NumericHistogram {
	h, _ := NewNumericHistogram(maxBins, dimensions, nil)

	for _, val := range data {
		h.Add(val)
	}
	return h
}

func getEvaluationPointsAndActualCDFs(mean []float64, variance []float64) ([][]float64, []float64) {
	dimensions := len(mean)
	sd := sqrt(variance)
	// First Evaluation Point is the mean. [u1, u2, ... ud]
	evaluationPoints := [][]float64{mean}
	// CDFs are determined from the Gaussian CDF functions.
	// Since no such library found in Golang the values are hardcoded using
	// values from scipy. Ex: For 2 dimension at CDF at [u1 - sd, u2- sd]
	// from scipy.stats import mvn
	// import numpy as np
	// mu = np.array([0.0, 0.0])
	// S = np.array([[1.0,0.0],[0.0,1.0]])
	// low = np.array([-1000.0, -1000.0])  # Very low value. Ideally (-inf, -inf)
	// high = np.array([-1.0, -1.0])
	// cdf = mvn.mvnun(low, high, mu, S)[0]
	dimensionValues := []float64{0.0, 0.5, 0.25, 0.125, 0.0625, 0.0312}
	actualCDFs := []float64{dimensionValues[dimensions]}

	// Second evaluation point is mean - sd. [u1-sd1, u2-sd2, .. ud-sdd]
	evaluationPoints = append(evaluationPoints, subtract(mean, sd))
	dimensionValues = []float64{0.0, 0.15865, 0.02517, 0.004, 0.0006, 0.0001}
	actualCDFs = append(actualCDFs, dimensionValues[dimensions])

	// Third evaluation point is mean - 0.5*sd. [u1-0.5*sd1, u2-0.5*sd2, .. ud-0.5*sdd]
	evaluationPoints = append(evaluationPoints, subtract(mean, multiply(0.5, sd)))
	dimensionValues = []float64{0.0, 0.3085, 0.0952, 0.0294, 0.0091, 0.0028}
	actualCDFs = append(actualCDFs, dimensionValues[dimensions])

	// Fourth evaluation point is mean - 0.25*sd.
	evaluationPoints = append(evaluationPoints, subtract(mean, multiply(0.25, sd)))
	dimensionValues = []float64{0.0, 0.4012, 0.161, 0.0646, 0.0259, 0.0104}
	actualCDFs = append(actualCDFs, dimensionValues[dimensions])

	// Fifth evaluation point is mean + 0.25*sd.
	evaluationPoints = append(evaluationPoints, add(mean, multiply(0.25, sd)))
	dimensionValues = []float64{0.0, 0.5987, 0.3584, 0.2146, 0.1285, 0.0769}
	actualCDFs = append(actualCDFs, dimensionValues[dimensions])

	// Sixth evaluation point is mean + 0.5*sd.
	evaluationPoints = append(evaluationPoints, add(mean, multiply(0.5, sd)))
	dimensionValues = []float64{0.0, 0.6915, 0.4781, 0.3306, 0.2286, 0.1581}
	actualCDFs = append(actualCDFs, dimensionValues[dimensions])

	// Seventh evaluation point is mean + sd.
	evaluationPoints = append(evaluationPoints, add(mean, sd))
	dimensionValues = []float64{0.0, 0.8413, 0.7079, 0.5955, 0.5010, 0.4216}
	actualCDFs = append(actualCDFs, dimensionValues[dimensions])

	return evaluationPoints, actualCDFs
}

func TestNumericalSampleData(t *testing.T) {
	for _, data := range [][][]float64{dataDimension1, dataDimension2,
		dataDimension3, dataDimension4, dataDimension5} {

		dimensions := len(data[0])
		var actualMean []float64 = make([]float64, dimensions)
		for i := range actualMean {
			actualMean[i] = 0.0
		}
		var actualVariance []float64 = make([]float64, dimensions)
		for i := range actualVariance {
			actualVariance[i] = 10000.0
		}
		evaluationPoints, actualCDFs := getEvaluationPointsAndActualCDFs(actualMean, actualVariance)
		numDataSamples := len(data)

		// Evaluate with different bin sizes.
		for j, numBins := range []int{8, 16, 32, 64, 128} {
			fmt.Println("------------------------------------------------------------")
			fmt.Println(fmt.Sprintf("EVALUATING FOR DIMENSIONS:%d  BINS:%d NUM_SAMPLES:%d",
				dimensions, numBins, numDataSamples))
			fmt.Println("------------------------------------------------------------")

			hist := buildNumericHistogramFromData(numBins, dimensions, data)
			assert.Equal(t, uint64(numDataSamples), hist.Count(), "Mismatch in number of samples.")

			// Empirically determined upper bound for error.
			// j+3 represents log2(numBins).
			// With 8 bins and 1 dimensions the maximum error is 0.08
			// With 128 bins and 1 dimensions the maximum error is 0.035
			// With 8 bins and 5 dimensions the maximum error is 0.41
			// With 128 bins and 5 dimensions the maximum error is 0.17
			maxError := 0.25 * float64(dimensions) / float64(j+3)
			for k, evalPoint := range evaluationPoints {
				histCDF := hist.CDF(evalPoint)
				actualCDF := actualCDFs[k]
				actualCDFError := actualCDF - histCDF
				sampleCDF := computeCDFUsingData(data, evalPoint)
				sampleCDFError := sampleCDF - histCDF
				fmt.Println(fmt.Sprintf(
					"EVALPOINT%d:%v, ACTUAL_CDF:%.2f, SAMPLE_CDF:%.2f, HIST_CDF:%.2f, ACTUAL_CDF_ERROR:%.2f, SAMPLE_CDF_ERROR:%.2f",
					k+1, evalPoint, actualCDF, sampleCDF, histCDF, actualCDFError, sampleCDFError))
				assert.InDelta(t, actualCDF, histCDF, maxError, "High Histogram CDF error")
				assert.InDelta(t, sampleCDF, histCDF, maxError, "High Sample CDF error")
			}

			fmt.Println("---------------------------------------------------")
		}
	}
}
