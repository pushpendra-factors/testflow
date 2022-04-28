package delta

import (
	"bufio"
	M "factors/model/model"

	log "github.com/sirupsen/logrus"
)

var salesforceMetricToFunc = map[string]func(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error){}

func GetSalesforceMetrics(metricNames []string, queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (map[string]*MetricInfo, error) {
	metricsInfoMap := make(map[string]*MetricInfo)
	for _, metric := range metricNames {
		if _, ok := salesforceMetricToFunc[metric]; !ok {
			continue
		}
		if info, err := salesforceMetricToFunc[metric](queryEvent, scanner, propFilter, propsToEval); err != nil {
			log.WithError(err).Error("error GetHubpotMetrics")
		} else {
			metricsInfoMap[metric] = info
		}
	}
	return metricsInfoMap, nil
}
