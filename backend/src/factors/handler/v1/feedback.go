package v1

import (
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type FeedbackRequestPayload struct {
	Feature  string          `json:"feature"`
	Property *postgres.Jsonb `json:"property"`
	VoteType int             `json:"vote_type"`
}

func GetPostFeedbackParams(c *gin.Context) (*FeedbackRequestPayload, error) {
	params := FeedbackRequestPayload{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}
func PostFeedbackHandler(c *gin.Context) {
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	params, err := GetPostFeedbackParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	errCode, errMsg := store.GetStore().PostFeedback(projectID, loggedInAgentUUID, params.Feature, params.Property, params.VoteType)
	if errCode != http.StatusCreated {
		if errMsg != "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errMsg})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	c.JSON(http.StatusCreated, response)

}
