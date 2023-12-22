package model

const (
	DBSelect                = "SELECT "
	DBFrom                  = "FROM "
	DBWhere                 = "WHERE "
	DBGroupByConst          = "GROUP BY "
	DBOrderBy               = "ORDER BY "
	DBAscend                = "ASC "
	DBDescend               = "DESC "
	DBLimit                 = "LIMIT "
	WebsiteAggregationTable = "website_aggregation"
	PredefinedWebAggrLimit  = "2500"

	ChartTypeBarChart 				= "pb"
	ChartTypeLineChart 				= "pl"
	ChartTypeTable 					= "pt"
	ChartTypeSparkLines 			= "pc"
	ChartTypeStackedAread 			= "pa"
	ChartTypeStackedBar 			= "ps"
	ChartTypeScatterPlot 			= "sp"
	ChartTypeHorizontalBarChart 	= "hb"
	ChartTypePivotChart 			= "pi"
	ChartTypeMetricChart 			= "mc"
	ChartTypeFunnelChart 			= "fc"
	PresentationTypeChart 			= "chart"
	PresentationTypeTable 			= "table"
)

var MapOfOperatorToExpression = map[string]string{
	"Division": "/",
	"Multiply": "*",
}

type PredefinedDashboard struct {
	InternalID  int64  `json:"inter_id"`
	Name        string `json:"na"`
	DisplayName string `json:"d_na"`
	Description string `json:"desc"`
}

type PredefinedDashboardConfig struct {
	InternalID int64              `json:"inter_id"`
	Properties [][]string         `json:"pr"`
	Widgets    []PredefinedWidget `json:"wid"`
}

type PredefinedDashboardProperty struct {
	Name            string `json:"na"`
	DisplayName     string `json:"d_na"`
	DataType        string `json:"d_t"`
	SourceEventName string
	SourceEntity    string
	SourceProperty  string
}

type PredefinedWidget struct {
	Name        string              `json:"na"`
	DisplayName string              `json:"d_na"`
	Metrics     []PredefinedMetric  `json:"me"`
	GroupBy     []PredefinedGroupBy `json:"g_by"`
	InternalID  int64               `json:"inter_id"`
	Setting		ChartSetting		`json:"chart_setting"` 
}

type PredefinedMetric struct {
	Name              	string `json:"na"`
	DisplayName       	string `json:"d_na"`
	InternalEventType 	string `json:"inter_e_type"`
	Type				string `json:"ty"`
}

type PredefinedGroupBy struct {
	Name        string `json:"na"`
	DisplayName string `json:"d_na"`
}

type ChartSetting struct {
	Type   	string	`json:"ty"`
	Presentation	string	`json:"pr"`
}

// interface for predefined dashboards.
// It provides methods for performing actions.
type PredefinedQueryGroup interface {
	IsValid() (bool, string)
}

type PredefinedFilter struct {
	PropertyName     string `json:"pr_na"`
	PropertyDataType string `json:"pr_da_ty"`
	Condition        string `json:"co"`
	Value            string `json:"va"`
	LogicalOp        string `json:"l_op"`
}

var PredefinedDashboards = []PredefinedDashboard{
	{
		InternalID: 1, Name: "Traffic Dashboard", DisplayName: "Traffic Dashboard", Description: "Traffic Dashboard",
	},
}

var MapfOfPredefinedDashboardIDToConfig = map[int64]PredefinedDashboardConfig{
	PredefinedDashboardConfigs[0].InternalID: PredefinedDashboardConfigs[0],
}

var PredefinedDashboardConfigs = []PredefinedDashboardConfig{
	{
		InternalID: 1,
		Properties: convertToArrayPropertiesFormat(predefinedWebsiteAggregationProperties),
		Widgets:    predefinedWebsiteAggregationWidgets,
	},
}

// This is used in filter Values. TODO later move to predefined_website_aggregation.
var MapOfPredefinedDashboardToPropertyNameToProperties = map[int64]map[string]PredefinedDashboardProperty{
	1: {
		predefinedWebsiteAggregationProperties[0].Name:  predefinedWebsiteAggregationProperties[0],
		predefinedWebsiteAggregationProperties[1].Name:  predefinedWebsiteAggregationProperties[1],
		predefinedWebsiteAggregationProperties[2].Name:  predefinedWebsiteAggregationProperties[2],
		predefinedWebsiteAggregationProperties[3].Name:  predefinedWebsiteAggregationProperties[3],
		predefinedWebsiteAggregationProperties[4].Name:  predefinedWebsiteAggregationProperties[4],
		predefinedWebsiteAggregationProperties[5].Name:  predefinedWebsiteAggregationProperties[5],
		predefinedWebsiteAggregationProperties[6].Name:  predefinedWebsiteAggregationProperties[6],
		predefinedWebsiteAggregationProperties[7].Name:  predefinedWebsiteAggregationProperties[7],
		predefinedWebsiteAggregationProperties[8].Name:  predefinedWebsiteAggregationProperties[8],
		predefinedWebsiteAggregationProperties[9].Name:  predefinedWebsiteAggregationProperties[9],
		predefinedWebsiteAggregationProperties[10].Name: predefinedWebsiteAggregationProperties[10],
		predefinedWebsiteAggregationProperties[11].Name: predefinedWebsiteAggregationProperties[11],
	},
}

// Where is it used? Should it be used as complete map or individual maps. i.e.website_aggregation.
var MapOfPredefinedDashboardToNameToWidgets = map[int64]map[string]PredefinedWidget{
	1: {
		predefinedWebsiteAggregationWidgets[0].Name: predefinedWebsiteAggregationWidgets[0],
	},
}
