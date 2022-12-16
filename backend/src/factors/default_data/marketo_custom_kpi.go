package default_data

import (
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var defaultMarketoCustomKPITransformations = []model.CustomMetricTransformation{
	model.CustomMetricTransformation{
		AggregateFunction:     model.UniqueAggregateFunction,
		AggregateProperty:     "",
		AggregatePropertyType: "categorical",
		Filters:               []model.KPIFilter{},
		DateField:             "$marketo_lead_created_at",
		EventName:             "",
		Entity:                model.UserEntity,
	},
}

var defaultMarketoCustomKPIs = []model.CustomMetric{
	model.CustomMetric{
		Name:        "Marketo Leads",
		Description: "All Marketo Leads timestamped at Lead Create Date",
		TypeOfQuery: 1,
		ObjectType:  model.MarketoLeadsDisplayCategory,
	},
}

var defaultMarketoDerivedKPITransformations = make([]model.KPIQueryGroup, 0)

type BuildDefaultMarketoCustomKPI struct {
}

func (buildDefault BuildDefaultMarketoCustomKPI) Build(projectID int64) int {
	if !CheckIfDefaultLeadSquaredDatasAreCorrect() {
		log.Warn("Failed because defaultDatas and transformations are of incorrect length: marketo.")
		return http.StatusInternalServerError
	}
	return buildForCustomKPI(projectID, defaultMarketoCustomKPIs, defaultMarketoCustomKPITransformations, defaultMarketoDerivedKPITransformations)
}

func CheckIfDefaultMarketoDatasAreCorrect() bool {
	return len(defaultMarketoCustomKPIs) == (len(defaultMarketoCustomKPITransformations) + len(defaultMarketoDerivedKPITransformations))
}
