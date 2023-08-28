package six_signal

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/model/store/memsql"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"
)

// SendSixSignalEmailForSubscribe sends mail to all the email for which the sixsignal-report is subscribed on a weekly basis. The list of email id is fetched from the DB.
func SendSixSignalEmailForSubscribe(projectIdArray []int64) map[int64][]string {

	projectIdToFailSendEmailIdsMap := make(map[int64][]string)
	for _, projectId := range projectIdArray {

		logCtx := log.WithFields(log.Fields{
			"project_id": projectId,
		})

		//Changing all the report generated share type to public and generating public url
		//Fetching date range and timezone for generating hash key and query
		timezone, _ := store.GetStore().GetTimezoneForProject(projectId)
		from, to, _ := U.GetQueryRangePresetLastWeekIn(timezone)

		data := fmt.Sprintf("%d%s%d%d", projectId, timezone, from, to)
		query := model.SixSignalQuery{From: from, To: to, Timezone: timezone}
		queryJson, _ := json.Marshal(query)

		queryRequest := &model.Queries{
			Query:     postgres.Jsonb{RawMessage: json.RawMessage(queryJson)},
			Title:     "Six Signal Report",
			CreatedBy: "",
			Settings:  postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
			IdText:    U.HashKeyUsingSha256Checksum(data),
			Type:      model.QueryTypeSixSignalQuery,
		}

		queryId, errCode, errMsg := CreateSixSignalShareableURL(queryRequest, projectId, "")
		if errCode != http.StatusCreated {
			logCtx.Error(errMsg)
			continue
		}

		publicUrlParams := fmt.Sprintf("/reports/visitor_report?queryId=%v&pId=%d&version=v1", queryId, projectId)
		publicURL := C.GetProtocol() + C.GetAPPDomain() + publicUrlParams

		//Fetching emailIds from database and converting the datatype to array
		emailIdsString, errCode1 := store.GetStore().GetSixsignalEmailListFromProjectSetting(projectId)
		if errCode1 != http.StatusFound || emailIdsString == "" {
			logCtx.Error("No email Ids for sixsignal report subscription is found.")
			continue
		}
		emailIds := strings.Split(emailIdsString, ",")
		project, _ := store.GetStore().GetProject(projectId)
		reqPayload := model.SixSignalEmailAndMessage{
			EmailIDs: emailIds,
			Url:      publicURL,
			Domain:   project.Name,
			From:     from,
			To:       to,
			Timezone: timezone,
		}

		_, failToSendEmailIds := memsql.SendSixSignalReportViaEmail(reqPayload)
		if len(failToSendEmailIds) > 0 {
			projectIdToFailSendEmailIdsMap[projectId] = failToSendEmailIds
		}
	}
	return projectIdToFailSendEmailIdsMap
}

// CreateSixSignalShareableURL saves the query to the queries table and generate the queryID for public-URL for the given queryRequest and projectId
func CreateSixSignalShareableURL(queryRequest *model.Queries, projectId int64, agentUUID string) (string, int, string) {

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
		"query":      queryRequest,
	})

	isShared, _ := isReportShared(projectId, queryRequest.IdText)
	if isShared {
		logCtx.Info("Shareable query already shared.")
		return queryRequest.IdText, http.StatusCreated, "Shareable Query already shared"
	}

	queries, errCode, errMsg := store.GetStore().CreateQuery(projectId, queryRequest)
	if errCode != http.StatusCreated {
		return "", errCode, errMsg
	}

	shareableUrlRequest := &model.ShareableURL{
		QueryID:    queries.IdText,
		EntityType: model.ShareableURLEntityTypeSixSignal,
		EntityID:   queries.ID,
		ShareType:  model.ShareableURLShareTypePublic,
		ProjectID:  projectId,
		CreatedBy:  agentUUID,
		ExpiresAt:  time.Now().AddDate(0, 3, 0).Unix(),
	}

	share, err := store.GetStore().CreateShareableURL(shareableUrlRequest)
	if err != http.StatusCreated {
		logCtx.Error("Failed to create shareable query")
		errCode, errMsg = store.GetStore().DeleteQuery(projectId, queries.ID)
		if errCode != http.StatusAccepted {
			logCtx.Warn("Failed to Delete Query in CreateSixSignalShareableURLHandler: ", errMsg)
		}
		return "", http.StatusInternalServerError, "Shareable query creation failed."
	}

	return share.QueryID, http.StatusCreated, "Shareable Query creation successful"
}

// isReportShared checks if the report has been already made public
func isReportShared(projectID int64, idText string) (bool, string) {

	share, err := store.GetStore().GetShareableURLWithShareStringWithLargestScope(projectID, idText, model.ShareableURLEntityTypeSixSignal)
	if err == http.StatusBadRequest || err == http.StatusInternalServerError {
		return false, "Shareable query fetch failed. DB error."
	} else if err == http.StatusFound {
		if share.ShareType == model.ShareableURLShareTypePublic {
			return true, "Shareable url already exists."
		}
	}
	return false, "Shareable url doesn't exist"

}
