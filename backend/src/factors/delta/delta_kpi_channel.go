package delta

import (
	"bufio"
	M "factors/model/model"

	log "github.com/sirupsen/logrus"
)

var channelMetricToFunc = map[string]func(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error){
	"impressions": GetChannelImpressions,
	"clicks":      GetChannelClicks,
	"spend":       GetChannelSpend,
}

func GetChannelMetrics(metricNames []string, queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (map[string]*MetricInfo, error) {
	metricsInfoMap := make(map[string]*MetricInfo)
	for _, metric := range metricNames {
		if _, ok := channelMetricToFunc[metric]; !ok {
			continue
		}
		if info, err := channelMetricToFunc[metric](queryEvent, scanner, propFilter, propsToEval); err != nil {
			log.WithError(err).Error("error GetChannelMetrics")
		} else {
			metricsInfoMap[metric] = info
		}
	}
	return metricsInfoMap, nil
}

func GetChannelImpressions(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	return &MetricInfo{}, nil
}

func GetChannelClicks(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	return &MetricInfo{}, nil
}

func GetChannelSpend(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	return &MetricInfo{}, nil
}
