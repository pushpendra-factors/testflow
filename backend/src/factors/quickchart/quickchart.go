package quickchart

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	quickchartgo "github.com/henomis/quickchart-go"
	log "github.com/sirupsen/logrus"
)

type ChartConfig struct {
	Type string    `json:"type"`
	Data ChartData `json:"data"`
}
type ChartData struct {
	Labels   []interface{} `json:"labels"`
	DataSets []Dataset     `json:"datasets"`
}
type Dataset struct {
	Label       string        `json:"label"`
	Data        []interface{} `json:"data"`
	Fill        bool          `json:"fill"`
	LineTension float32           `json:"lineTension"`
}
type TableConfig struct {
	Title      string        `json:"title"`
	Columns    []Column      `json:"columns"`
	DataSource []interface{} `json:"dataSource"`
}
type Column struct {
	Width     int    `json:"width"`
	Title     string `json:"title"`
	DataIndex string `json:"dataIndex"`
}

func GetChartImageUrlForConfig(config ChartConfig) (url string, err error) {
	bytes, err := json.Marshal(config)
	if err != nil {
		log.Error("failed to marshal chart config")
		return "", errors.New("failed to get char url from quickchart")
	}
	qc := quickchartgo.New()
	qc.Config = string(bytes)
	url, error := qc.GetUrl()
	if error != nil {
		log.Error("failed to get char url from quickchart")
		return "", errors.New("failed to get char url from quickchart")
	}
	return url, nil
}

func GetTableURLfromTableConfig(config TableConfig) (string, error) {
	bytes, err := json.Marshal(config)
	if err != nil {
		return "", errors.New("Failed to marshal table config")
	}
	url := fmt.Sprintf("https://api.quickchart.io/v1/table?data=%s", url.QueryEscape(string(bytes)))
	return url, nil

}
