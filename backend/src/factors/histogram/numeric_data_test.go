package histogram

import (
	"fmt"
	"math"
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

func buildNumericHistogramFromData(maxBins int, dimensions int, data [][]float64) *NumericHistogramStruct {
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

func TestNumericTrimByBinSize(t *testing.T) {
	data := dataDimension5
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
	numBins := 128

	fmt.Println("------------------------------------------------------------")
	fmt.Println(fmt.Sprintf("EVALUATING FOR DIMENSIONS:%d  BINS:%d NUM_SAMPLES:%d",
		dimensions, numBins, numDataSamples))
	fmt.Println("------------------------------------------------------------")

	hist := buildNumericHistogramFromData(numBins, dimensions, data)
	assert.Equal(t, uint64(numDataSamples), hist.Count(), "Mismatch in number of samples.")
	assert.Equal(t, hist.numBins(), 128, "Mismatch in number of bins.")

	// Check error before trimming.
	// Empirically determined upper bound for error.
	// 7.0 represents log2(numBins).
	maxError := 0.25 * float64(dimensions) / float64(7.0)
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

	// Trim histogram.
	hist.TrimByBinSize(0.5)
	assert.Equal(t, uint64(numDataSamples), hist.Count(), "Mismatch in number of samples.")
	assert.Equal(t, hist.numBins(), 64, "Mismatch in number of bins.")
	// Check error after trimming.
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

	// Minimum 12 bin.
	hist.TrimByBinSize(0.01)
	assert.Equal(t, uint64(numDataSamples), hist.Count(),
		"Mismatch in number of samples.")
	assert.Equal(t, hist.numBins(), 12, "Mismatch in number of bins.")
}

func TestNumericalCuts(t *testing.T) {
	data := dataMM3
	dataT := dataMM3TestTable
	dims := len(data[0])
	// numBins := 8
	var maxError = 0.20

	var err [100]float64
	var errSample [100]float64
	fmt.Println("---------------------------------------------------")
	fmt.Println(fmt.Sprintf("Calculating for %d dims", dims))
	fmt.Println("---------------------------------------------------")

	for _, numBins := range []int{8, 16, 32, 64, 128} {

		hist := buildNumericHistogramFromData(numBins, dims, data)
		for idx := 0; idx < len(dataT); idx++ {
			histCDF := hist.CDF(dataT[idx].point)
			sampleCDF := computeCDFUsingData(data, dataT[idx].point)

			errAct := math.Abs(dataT[idx].probVal - histCDF)
			errorsample := math.Abs(sampleCDF - histCDF)

			err[dataT[idx].center] = err[dataT[idx].center] + errAct
			errSample[dataT[idx].center] = errSample[dataT[idx].center] + errorsample

			fmt.Println(fmt.Sprintf(
				"ACTUAL_CDF:%.2f, SAMPLE_CDF:%.2f, HIST_CDF:%.2f, ACTUAL_CDF_ERROR:%.2f, SAMPLE_CDF_ERROR:%.2f",
				dataT[idx].probVal, sampleCDF, histCDF, errAct, errorsample))

		}
		var totalAct = 0.0
		var totSample = 0.0
		for idx := 0; idx < 100; idx++ {
			//Average out for each region  div by 5 as I've
			//samples 5 points from each region
			err[idx] = err[idx] / 5.0
			errSample[idx] = errSample[idx] / 5.0
			totalAct += err[idx]
			totSample += errSample[idx]
		}
		totalAct = totalAct / 100.0
		totSample = totSample / 100.0

		fmt.Println("---------------------------------------------------")
		fmt.Println(fmt.Sprintf("Total Act-Hist Error:%.2f , Total sample-Hist: %.2f ", totalAct, totSample))
		fmt.Println("---------------------------------------------------")
		assert.InDelta(t, totalAct, totSample, maxError, "High Histogram CDF error")
		// assert.InDelta(t, sampleCDF, histCDF, maxError, "High Sample CDF error")
	}

}
