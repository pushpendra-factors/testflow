package internal

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DataServiceLinkedinGetProjectSettings(c *gin.Context) {
	linkedinProjectSettings, status := store.GetStore().GetLinkedinEnabledProjectSettings()
	c.JSON(status, linkedinProjectSettings)
}

type LinkedInProjectIDs struct {
	ProjectIDs []string `json:"project_ids"`
}

func DataServiceLinkedinGetProjectSettingsForProjects(c *gin.Context) {
	r := c.Request

	var linkedinProjectIDs LinkedInProjectIDs
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&linkedinProjectIDs); err != nil {
		log.WithError(err).Error("Failed to decode Json request on linkedin get project settings handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}
	linkedinProjectSettings, status := store.GetStore().GetLinkedinEnabledProjectSettingsForProjects(linkedinProjectIDs.ProjectIDs)
	c.JSON(status, linkedinProjectSettings)
}

func DataServiceLinkedinAddDocumentHandler(c *gin.Context) {
	r := c.Request

	var linkedinDocument model.LinkedinDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&linkedinDocument); err != nil {
		log.WithError(err).Error("Failed to decode Json request on linkedin upsert document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	errCode := store.GetStore().CreateLinkedinDocument(linkedinDocument.ProjectID, &linkedinDocument)
	if errCode != http.StatusCreated {
		log.WithField("document", linkedinDocument).Error("Failed to insert the linkedin document on create document.")
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed to upsert linkedin document."})
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully upserted linkedin document."})
}

// DataServiceLinkedinAddMultipleDocumentsHandler can help bulk insert of 10
func DataServiceLinkedinAddMultipleDocumentsHandler(c *gin.Context) {
	r := c.Request

	var linkedinDocuments []model.LinkedinDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&linkedinDocuments); err != nil {
		log.WithError(err).Error("Failed to decode JSON request")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to decode JSON request."})
		return
	}
	errCode := store.GetStore().CreateMultipleLinkedinDocument(linkedinDocuments)
	if errCode == http.StatusConflict {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Duplicate documents."})
		return
	}

	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed creating linkedin document."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created linkedin document."})
}

func DataServiceLinkedinGetCampaignGroupInfoHandler(c *gin.Context) {
	r := c.Request

	var campaignGroupInfoRequestPayload model.LinkedinCampaignGroupInfoRequestPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&campaignGroupInfoRequestPayload); err != nil {
		log.WithError(err).Error("Failed to decode JSON request")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to decode JSON request."})
		return
	}
	linkedinDocuments, statusCode := store.GetStore().GetCampaignGroupInfoForGivenTimerange(campaignGroupInfoRequestPayload)
	c.JSON(statusCode, linkedinDocuments)
}
func DataServiceLinkedinValidationHandler(c *gin.Context) {
	r := c.Request

	var validationRequestPayload model.LinkedinValidationRequestPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&validationRequestPayload); err != nil {
		log.WithError(err).Error("Failed to decode JSON request")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to decode JSON request."})
		return
	}
	isValid, statusCode := store.GetStore().GetValidationForGivenTimerangeAndJobType(validationRequestPayload)
	c.JSON(statusCode, isValid)
}

// DataServiceLinkedinDeleteDocumentsHandler deletes the db insertions of one doc_type of given timestamp
func DataServiceLinkedinDeleteDocumentsHandler(c *gin.Context) {
	r := c.Request

	var deleteDocumentsPayload model.LinkedinDeleteDocumentsPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&deleteDocumentsPayload); err != nil {
		log.WithError(err).Error("Failed to decode JSON request")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to decode JSON request."})
		return
	}
	errCode := store.GetStore().DeleteLinkedinDocuments(deleteDocumentsPayload)

	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed deleting linkedin documents"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Deleted linkedin documents"})
}

type LinkedinUpdateAccessToken struct {
	ProjectID   int64  `json:"project_id"`
	AccessToken string `json:"access_token"`
}

func DataServiceLinkedinUpdateAccessToken(c *gin.Context) {
	r := c.Request
	var linkedinUpdateAccessToken LinkedinUpdateAccessToken
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&linkedinUpdateAccessToken); err != nil {
		log.WithError(err).Error("Failed to decode Json request on linkedin update access token handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}
	_, errCode := store.GetStore().UpdateProjectSettings(linkedinUpdateAccessToken.ProjectID, &model.ProjectSetting{IntLinkedinAccessToken: linkedinUpdateAccessToken.AccessToken})
	if errCode != http.StatusAccepted {
		log.Error("Failed to update access token.")
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed to update access token."})
		return
	}
	c.JSON(errCode, gin.H{"message": "Successfully updated access token."})
}

/*
it fetches last sync info for all doc types:
ad account, campaign group (meta + insights), campaign (meta + insights), company engagements (insights + last_backfill for weekly data)
*/

func DataServiceLinkedinGetLastSyncInfoHandler(c *gin.Context) {
	r := c.Request

	var payload model.LinkedinLastSyncInfoPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on linkedin get last sync info handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}
	projectID, _ := strconv.ParseInt(payload.ProjectID, 10, 64)
	lastSyncInfo, status := store.GetStore().GetLinkedinLastSyncInfo(projectID, payload.CustomerAdAccountID)
	c.JSON(status, lastSyncInfo)
}

/*
In V1 version we are separating out Ads and company enagagements
the following fetches last sync info for:
ad account, campaign group (meta + insights), campaign (meta + insights)
*/
func DataServiceLinkedinAdsGetLastSyncInfoV1Handler(c *gin.Context) {
	r := c.Request

	var payload model.LinkedinLastSyncInfoPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on linkedin get last sync info handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}
	projectID, _ := strconv.ParseInt(payload.ProjectID, 10, 64)
	lastSyncInfo, status := store.GetStore().GetLinkedinAdsLastSyncInfoV1(projectID, payload.CustomerAdAccountID)
	c.JSON(status, lastSyncInfo)
}

/*
here we fetch 3 types of last sync infos:
member company insights
t8
t22
Details explaination present in memsql/linkedin_documents.go GetLinkedinCompanyLastSyncInfoV1
*/
func DataServiceLinkedinCompanyGetLastSyncInfoV1Handler(c *gin.Context) {
	r := c.Request

	var payload model.LinkedinLastSyncInfoPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on linkedin get last sync info handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}
	projectID, _ := strconv.ParseInt(payload.ProjectID, 10, 64)
	lastSyncInfo, status := store.GetStore().GetLinkedinCompanyLastSyncInfoV1(projectID, payload.CustomerAdAccountID)
	c.JSON(status, lastSyncInfo)
}
