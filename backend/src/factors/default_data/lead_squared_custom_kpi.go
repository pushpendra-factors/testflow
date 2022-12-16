package default_data

import (
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type BuildDefaultLeadSquaredCustomKPI struct {
}

var defaultLeadSquaredCustomKPITransformations = []model.CustomMetricTransformation{
	model.CustomMetricTransformation{
		AggregateFunction:     model.UniqueAggregateFunction,
		AggregateProperty:     "",
		AggregatePropertyType: "categorical",
		Filters:               []model.KPIFilter{},
		DateField:             "$leadsquared_lead_createdon",
		EventName:             "",
		Entity:                model.UserEntity,
	},
}

var defaultLeadSquaredCustomKPIs = []model.CustomMetric{
	model.CustomMetric{
		Name:        "LeadSquared Leads",
		Description: "All LeadSquared Leads timestamped at Lead Create Date",
		TypeOfQuery: 1,
		ObjectType:  model.LeadSquaredLeadsDisplayCategory,
	},
}

var defaultLeadSquaredDerivedKPITransformations = make([]model.KPIQueryGroup, 0)

func (buildDefault BuildDefaultLeadSquaredCustomKPI) Build(projectID int64) int {
	if !CheckIfDefaultLeadSquaredDatasAreCorrect() {
		log.Warn("Failed because defaultDatas and transformations are of incorrect length: leadsquared.")
		return http.StatusInternalServerError
	}
	return buildForCustomKPI(projectID, defaultLeadSquaredCustomKPIs, defaultLeadSquaredCustomKPITransformations, defaultLeadSquaredDerivedKPITransformations)
}

func CheckIfDefaultLeadSquaredDatasAreCorrect() bool {
	return len(defaultLeadSquaredCustomKPIs) == (len(defaultLeadSquaredCustomKPITransformations) + len(defaultLeadSquaredDerivedKPITransformations))
}
