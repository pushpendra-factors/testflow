package fivetran

import (
	"fmt"
	"net/http"

	U "factors/util"

	C "factors/config"

	log "github.com/sirupsen/logrus"
)

type ConfigSchema struct {
	Schema string `json:"schema"`
}

type FiveTranConnectorCreateRequest struct {
	Service           string       `json:"service"`
	GroupId           string       `json:"group_id"`
	TrustCertificates bool         `json:"trust_certificates"`
	RunSetupTests     bool         `json:"run_setup_tests"`
	Paused            bool         `json:"paused"`
	PauseAfterTrial   bool         `json:"pause_after_trial"`
	SyncFrequency     uint64       `json:"sync_frequency"`
	Config            ConfigSchema `json:"config"`
}

type FiveTranConnectorCreateResponseDetails struct {
	Id string `json:"id"`
}
type FiveTranConnectorCreateResponse struct {
	Data FiveTranConnectorCreateResponseDetails `json:"data"`
}

type FiveTranSchemaReloadRequest struct {
	ExcludeMode string `json:"exclude_mode"`
}

type FiveTranSchemaPatchRequest struct {
	Enabled bool                         `json:"enabled"`
	Tables  FiveTranBingAdsSchemaRequest `json:"tables"`
}

type FiveTranBingAdsSchemaRequest struct {
	CampaignPerformanceDailyReport FiveTranSchemaEnableRequest `json:"campaign_performance_daily_report"`
	AccountHistory                 FiveTranSchemaEnableRequest `json:"account_history"`
	AdGroupPerformanceDailyReport  FiveTranSchemaEnableRequest `json:"ad_group_performance_daily_report"`
	KeywordPerformanceDailyReport  FiveTranSchemaEnableRequest `json:"keyword_performance_daily_report"`
	CampaignHistory                FiveTranSchemaEnableRequest `json:"campaign_history"`
	AdGroupHistory                 FiveTranSchemaEnableRequest `json:"ad_group_history"`
	KeywordHistory                 FiveTranSchemaEnableRequest `json:"Keyword_History"`
}

type FiveTranSchemaEnableRequest struct {
	Enabled bool `json:"enabled"`
}

type FiveTranHistoricalSyncRequest struct {
	Paused           bool `json:"paused"`
	IsHistoricalSync bool `json:"is_historical_sync"`
}

func FiveTranCreateBingAdsConnector(projectId uint64) (int, string, string, string) {

	Authorization := map[string]string{
		"Authorization": C.GetFivetranLicenseKey(),
		"Content-Type":  "application/json",
	}
	GroupId := C.GetFivetranGroupId()

	connectorCreateRequest := FiveTranConnectorCreateRequest{
		Service:           "bingads",
		GroupId:           GroupId,
		TrustCertificates: true,
		RunSetupTests:     false,
		Paused:            true,
		PauseAfterTrial:   true,
		SyncFrequency:     1440,
		Config: ConfigSchema{
			Schema: fmt.Sprintf("%v_%v_%v", "bingads", projectId, U.TimeNowUnix()),
		},
	}
	statusCode, response, errResponse := HttpRequestWrapper("connectors", Authorization, connectorCreateRequest, "POST")
	log.Info(response)
	if statusCode == http.StatusCreated {
		connectorId := response["data"].(map[string]interface{})["id"].(string)
		schemaId := response["data"].(map[string]interface{})["schema"].(string)
		return statusCode, "", connectorId, schemaId // statuscode, errstring, connector_name, schema_name
	} else {
		return statusCode, errResponse.Code, "", ""
	}
}

func FiveTranReloadBingAdsConnectorSchema(ConnectorId string) (int, string) {
	Authorization := map[string]string{
		"Authorization": C.GetFivetranLicenseKey(),
		"Content-Type":  "application/json",
	}
	reloadRequest := FiveTranSchemaReloadRequest{
		ExcludeMode: "EXCLUDE",
	}
	statusCode, response, errResponse := HttpRequestWrapper(fmt.Sprintf("connectors/%v/schemas/reload", ConnectorId), Authorization, reloadRequest, "POST")
	log.Info(response)
	if statusCode == http.StatusOK {
		return statusCode, "" // statuscode, errstring
	} else {
		return statusCode, errResponse.Code
	}
}

func FiveTranPatchBingAdsConnectorSchema(ConnectorId string, SchemaName string) (int, string) {
	Authorization := map[string]string{
		"Authorization": C.GetFivetranLicenseKey(),
		"Content-Type":  "application/json",
	}
	schemaPatchRequest := FiveTranSchemaPatchRequest{
		Enabled: true,
		Tables: FiveTranBingAdsSchemaRequest{
			CampaignPerformanceDailyReport: FiveTranSchemaEnableRequest{
				Enabled: true,
			},
			AccountHistory: FiveTranSchemaEnableRequest{
				Enabled: true,
			},
			AdGroupPerformanceDailyReport: FiveTranSchemaEnableRequest{
				Enabled: true,
			},
			KeywordPerformanceDailyReport: FiveTranSchemaEnableRequest{
				Enabled: true,
			},
			CampaignHistory: FiveTranSchemaEnableRequest{
				Enabled: true,
			},
			AdGroupHistory: FiveTranSchemaEnableRequest{
				Enabled: true,
			},
			KeywordHistory: FiveTranSchemaEnableRequest{
				Enabled: true,
			},
		},
	}
	statusCode, response, errResponse := HttpRequestWrapper(fmt.Sprintf("connectors/%s/schemas/%s", ConnectorId, SchemaName), Authorization, schemaPatchRequest, "PATCH")
	log.Info(response)
	if statusCode == http.StatusOK {
		return statusCode, "" // statuscode, errstring
	} else {
		return statusCode, errResponse.Code
	}
}

func FiveTranCreateConnectorCard(ConnectorId string) (int, string, string) {
	Authorization := map[string]string{
		"Authorization": C.GetFivetranLicenseKey(),
		"Content-Type":  "application/json",
	}
	statusCode, response, errResponse := HttpRequestWrapper(fmt.Sprintf("connectors/%s/connect-card-token", ConnectorId), Authorization, nil, "POST")
	log.Info(response)
	if statusCode == http.StatusOK {
		fe_host := C.GetProtocol() + C.GetAPPDomain()
		redirectUri := fmt.Sprintf("https://fivetran.com/connect-card/setup?redirect_uri=%s&auth=%s", fe_host, response["token"].(string))
		return statusCode, "", redirectUri // statuscode, errstring, redirect_uri
	} else {
		return statusCode, errResponse.Code, ""
	}
}

func FiveTranPatchConnector(ConnectorId string) (int, string) {
	Authorization := map[string]string{
		"Authorization": C.GetFivetranLicenseKey(),
		"Content-Type":  "application/json",
	}
	request := FiveTranHistoricalSyncRequest{
		Paused:           false,
		IsHistoricalSync: true,
	}
	statusCode, response, errResponse := HttpRequestWrapper(fmt.Sprintf("connectors/%s", ConnectorId), Authorization, request, "PATCH")
	log.Info(response)
	if statusCode == http.StatusOK {
		return statusCode, ""
	} else {
		return statusCode, errResponse.Code // return error string - err code, err message
	}
}

func FiveTranDeleteConnector(ConnectorId string) (int, string) {
	Authorization := map[string]string{
		"Authorization": C.GetFivetranLicenseKey(),
		"Content-Type":  "application/json",
	}
	statusCode, response, errResponse := HttpRequestWrapper(fmt.Sprintf("connectors/%s", ConnectorId), Authorization, nil, "DELETE")
	log.Info(response)
	if statusCode == http.StatusOK {
		return statusCode, ""
	} else {
		return statusCode, errResponse.Code // return error string - status of connector, status code, err code
	}
}

func FiveTranGetConnector(ConnectorId string) (int, string, bool, string) {
	Authorization := map[string]string{
		"Authorization": C.GetFivetranLicenseKey(),
		"Content-Type":  "application/json",
	}
	statusCode, response, errResponse := HttpRequestWrapper(fmt.Sprintf("connectors/%s", ConnectorId), Authorization, nil, "GET")
	log.Info(response)
	if statusCode == http.StatusOK {
		paused := response["data"].(map[string]interface{})["paused"].(bool)
		var accounts []interface{}
		if paused == false {
			syncMode, exists := response["data"].(map[string]interface{})["config"].(map[string]interface{})["sync_mode"]
			if exists && syncMode == "AllAccounts" {
				accounts = make([]interface{}, 0)
				accounts = append(accounts, syncMode)
			} else {
				accountsObject, exists := response["data"].(map[string]interface{})["config"].(map[string]interface{})["accounts"]
				if exists {
					accounts = accountsObject.([]interface{})
				}
			}
		}
		accountArray := ""
		for _, account := range accounts {
			if accountArray == "" {
				accountArray = fmt.Sprintf("%v", account)
			} else {
				accountArray = fmt.Sprintf(",%v", account)
			}
		}
		return statusCode, "", !paused, accountArray // return accounts in comma seperated list - statuscode, errstring, status of connector, accounts
	} else {
		return statusCode, errResponse.Code, false, ""
	}
}
