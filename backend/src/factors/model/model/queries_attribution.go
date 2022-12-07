package model

import "github.com/jinzhu/gorm/dialects/postgres"

type QueryAndDashboardUnit struct {
	Query         Queries       `json:"query"`
	DashboardUnit DashboardUnit `json:"dashboard_unit"`
}

// CreateQueryAndSaveToDashboardInfo is struct for post request to create attribution query and save to dashboard
type CreateQueryAndSaveToDashboardInfo struct {
	Title                     string          `json:"title"`
	Type                      int             `json:"type"`
	Query                     *postgres.Jsonb `json:"query"`
	Settings                  *postgres.Jsonb `json:"settings"`
	CreatedBy                 string          `gorm:"type:varchar(255)" json:"created_by"`
	DashboardUnitDescription  string          `json:"description"`
	DashboardUnitPresentation string          `json:"presentation"`
}

// RemoveAttributionV1Queries takes list of queries as input and returns list of queries by removing attribution v1 queries
func RemoveAttributionV1Queries(queries []Queries) []Queries {
	var finalQueries []Queries
	for _, query := range queries {
		if query.Type != QueryTypeAttributionV1Query {
			finalQueries = append(finalQueries, query)
		}
	}
	return finalQueries

}

// RemoveAttributionV1dashboard takes list of dashboards as input and returns list of dashboards by removing attribution v1 dashboard
func RemoveAttributionV1dashboard(dashboards []Dashboard) []Dashboard {
	var finalDashboards []Dashboard
	for _, dashboard := range dashboards {
		if dashboard.Type != DashboardTypeAttributionV1 {
			finalDashboards = append(finalDashboards, dashboard)
		}
	}
	return finalDashboards
}
