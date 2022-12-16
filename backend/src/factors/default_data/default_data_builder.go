package default_data

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"net/http"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func buildForCustomKPI(projectID int64, customMetrics []model.CustomMetric, customTransformations []model.CustomMetricTransformation,
	derivedTransformations []model.KPIQueryGroup) int {
	selectedStore := store.GetStore()
	postgresTransformations, statusCode := buildJsonTransformationsForCustomKPIs(customMetrics, customTransformations, derivedTransformations)
	if statusCode != http.StatusOK {
		log.WithField("projectId", projectID).Warn("Failed in building Json Transformations for hubspot")
		return http.StatusInternalServerError
	}

	for index := range customMetrics {
		customMetrics[index].ProjectID = projectID
		customMetrics[index].Transformations = &postgresTransformations[index]
	}

	for _, customMetric := range customMetrics {
		_, _, statusCode := selectedStore.CreateCustomMetric(customMetric)
		if statusCode == http.StatusConflict {
			continue
		}
		if statusCode != http.StatusCreated {
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

// TODO add error for the methods which are calling.
func buildJsonTransformationsForCustomKPIs(customMetrics []model.CustomMetric, profileTransformations []model.CustomMetricTransformation, derivedTransformations []model.KPIQueryGroup) ([]postgres.Jsonb, int) {

	resTransformations := make([]postgres.Jsonb, 0)
	indexOfProfileBasedTransformation := 0
	indexOfDerivedTransformation := 0
	for _, customMetric := range customMetrics {
		if customMetric.TypeOfQuery == 1 {
			transformation := profileTransformations[indexOfProfileBasedTransformation]
			jsonTransformation, err := json.Marshal(transformation)
			if err != nil {
				return make([]postgres.Jsonb, 0), http.StatusInternalServerError
			}
			postgresTransformation := postgres.Jsonb{json.RawMessage(jsonTransformation)}
			resTransformations = append(resTransformations, postgresTransformation)

			indexOfProfileBasedTransformation++
		} else {
			transformation := derivedTransformations[indexOfDerivedTransformation]
			jsonTransformation, err := json.Marshal(transformation)
			if err != nil {
				return make([]postgres.Jsonb, 0), http.StatusInternalServerError
			}
			postgresTransformation := postgres.Jsonb{json.RawMessage(jsonTransformation)}
			resTransformations = append(resTransformations, postgresTransformation)

			indexOfDerivedTransformation++
		}
	}
	return resTransformations, http.StatusOK
}

func CheckIfDefaultKPIDatasAreCorrect() bool {
	return CheckIfDefaultHubspotDatasAreCorrect() && CheckIfDefaultLeadSquaredDatasAreCorrect() &&
		CheckIfDefaultMarketoDatasAreCorrect() && CheckIfDefaultSalesforceDatasAreCorrect()
}

// We Follow the following order - Hubspot, leadSquared, Marketo, Salesforce.
func CheckIfFirstTimeIntegrationDone(projectID int64, integration string) (bool, int) {
	storeSelected := store.GetStore()
	integrationBits, statusCode := storeSelected.GetIntegrationBitsFromProjectSetting(projectID)
	if statusCode != http.StatusFound {
		return true, statusCode
	}
	if len(integrationBits) == 0 {
		return false, http.StatusFound
	}

	if integration == HubspotIntegrationName {
		return string(integrationBits[0]) == "1", http.StatusFound
	} else if integration == LeadSquaredIntegrationName {
		return string(integrationBits[1]) == "1", http.StatusFound
	} else if integration == model.MarketoIntegration {
		return string(integrationBits[2]) == "1", http.StatusFound
	} else {
		return string(integrationBits[3]) == "1", http.StatusFound
	}
}

func SetFirstTimeIntegrationDone(projectID int64, integration string) int {
	storeSelected := store.GetStore()
	integrationBits, statusCode := storeSelected.GetIntegrationBitsFromProjectSetting(projectID)
	if statusCode != http.StatusFound {
		return statusCode
	}
	if len(integrationBits) == 0 {
		integrationBits = model.DEFAULT_STRING_WITH_ZEROES_32BIT
	}

	byteArray := []byte(integrationBits)
	if integration == HubspotIntegrationName {
		byteArray[0] = 1
	} else if integration == LeadSquaredIntegrationName {
		byteArray[1] = 1
	} else if integration == model.MarketoIntegration {
		byteArray[2] = 1
	} else {
		byteArray[3] = 1
	}
	integrationBits = string(byteArray)
	statusCode2 := storeSelected.SetIntegrationBits(projectID, integrationBits)
	if statusCode2 != http.StatusAccepted {
		return statusCode2
	}
	return http.StatusOK
}
