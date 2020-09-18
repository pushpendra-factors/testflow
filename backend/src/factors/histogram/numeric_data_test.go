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

func sum(array []int) int {
	result := 0
	for _, v := range array {
		result += v
	}
	return result
}

type testData struct {
	table []TestTable
	data  [][]float64
}

func TestNumericalCuts(t *testing.T) {
	//sample usuage : go test -run ^TestNumericalCuts$
	// need to add outlier data

	// can add more testcases here
	testCases := []testData{
		testData{dataMM6TestTable, dataMM6}, //uniform
		testData{dataMM5TestTable, dataMM5}, //multivariate multimodal
		testData{dataMM4TestTable, dataMM4}, //multivariate
	}

	var maxErrorThreshold = 0.04
	var regions = []int{0, 1, 2, 3, 4, 5, 6, 7} // total number of regions the data is split into
	var bins = []int{8, 16, 32, 64, 128}

	for testCaseIdx, testCase := range testCases {
		data := testCase.data
		dataTable := testCase.table
		var dims = len(data[0])
		var numPoints = []int{0, 0, 0, 0, 0, 0, 0, 0}

		for _, region := range regions {

			for _, numBins := range bins {
				hist := buildNumericHistogramFromData(numBins, dims, data)

				var evalPoints = [][]float64{}

				// calculate the min and max in each region while reading the points
				var maxPoint = []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
				var minPoint = []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
				for tableIdx := 0; tableIdx < len(dataTable); tableIdx++ {
					if dataTable[tableIdx].region == region {
						var tmpPoint = dataTable[tableIdx].point
						for dim := 0; dim < dims; dim++ {
							maxPoint[dim] = math.Max(maxPoint[dim], tmpPoint[dim])
							minPoint[dim] = math.Min(minPoint[dim], tmpPoint[dim])

						}

						evalPoints = append(evalPoints, dataTable[tableIdx].point)
					}
				}

				numPoints[region] = len(evalPoints)
				sampleCDFMin := computeCDFUsingData(data, minPoint)
				sampleCDFMax := computeCDFUsingData(data, maxPoint)

				histcdf := hist.CDF(maxPoint) - hist.CDF(minPoint)
				samplecdf := math.Abs(sampleCDFMax - sampleCDFMin)
				cumPDF := float64(sum(numPoints[0:region+1])) / float64(len(data))
				percentPoints := float64(len(evalPoints)) / float64(len(data))

				// num : Id of region
				// Bin : Bin size used for hisogram
				// Hist CDF : CDF calculated using hist function
				// sample CDF : CDF calculated from CDF
				// cumPDF : CDF calculated from adding (total number of points in each region)/ total #of Points
				// fracPoints : Percentage of points in the region
				fmt.Println(fmt.Sprintf("num: %d | Bin: %d | HistCDF:%.4f | sampleCDF:%.4f | cumPDF:%.4f | fracPoints:%.4f", region, numBins, histcdf, samplecdf, cumPDF, percentPoints))

				for pointIdx := 0; pointIdx < dims; pointIdx++ {
					maxPoint[pointIdx] = maxPoint[pointIdx] + 1e+10
					minPoint[pointIdx] = minPoint[pointIdx] - 1e+10
				}

				// Testing for diff in histogram CDF and sample CDF
				if math.Abs(histcdf-samplecdf) > maxErrorThreshold {
					t.Logf(fmt.Sprintf("High Sample CDF error : num: %d | Bin: %d | HistCDF:%.4f | sampleCDF:%.4f | diff: %.4f | CasesIdx:%d", region, numBins, histcdf, samplecdf, math.Abs(histcdf-samplecdf), testCaseIdx))
					t.Fail()
				}

				// cdf of point in min P(x<xmin,y<ymin,z<zmin)
				if hist.CDF(minPoint) != 0 {
					t.Logf("PDF of min point in region %d  and bin %d is not 0 case: %d", region, numBins, testCaseIdx)
					t.Fail()
				}

				// cdf of point in max p(x>xmax)

				if hist.CDF(maxPoint) != 0 {
					t.Logf("PDF of max point in region %d  and bin %d is not 0 in cases:%d", region, numBins, testCaseIdx)
					t.Fail()
				}

			}

		}

	}
}
