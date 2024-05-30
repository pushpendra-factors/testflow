package internal

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func DataServiceAddEmbeddings(c *gin.Context) {
	r := c.Request

	var promptEmbeddings model.PromptEmbeddingsPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&promptEmbeddings); err != nil {
		log.WithError(err).Error("Failed to decode JSON request")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request JSON."})
		return
	}

	errCode, msg := store.GetStore().AddAllEmbeddings(promptEmbeddings.ProjectId, promptEmbeddings.IndexedPrompts, promptEmbeddings.IndexedQueries, promptEmbeddings.IndexedPromptEmbs)
	if errCode != http.StatusCreated {
		log.WithField("document", promptEmbeddings).Error(msg)
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed to add new prompt embeddings"})
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully added new prompt embeddings"})
}

func DataServiceGetMatchingEmbeddings(c *gin.Context) {
	r := c.Request

	var payload model.QueryEmbeddingPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on get matching embeddings handler")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	statusCode, msg, embeddings := store.GetStore().GetMatchingEmbeddings(payload.ProjectId, payload.QueryEmbedding)
	if statusCode != http.StatusOK {
		log.Error("Failed to retrieve embeddings")
		c.JSON(statusCode, gin.H{"error": msg})
		return
	}

	c.JSON(statusCode, gin.H{"message": msg, "data": embeddings})
}

func DataServiceGetMissingPrompts(c *gin.Context) {
	r := c.Request

	type QueryPayload struct {
		ProjectID int64    `json:"project_id"`
		Prompts   []string `json:"indexed_prompts"`
	}

	var payload QueryPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on get matching embeddings handler")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	statusCode, msg, prompts := store.GetStore().GetMissingPromptsByProjectID(payload.ProjectID, payload.Prompts)
	if statusCode != http.StatusOK {
		log.Error("Failed to retrieve embeddings")
		c.JSON(statusCode, gin.H{"error": msg})
		return
	}

	c.JSON(statusCode, gin.H{"message": msg, "data": prompts})
}

func DataServiceDeleteDataByProjectId(c *gin.Context) {
	r := c.Request

	type QueryPayload struct {
		ProjectID int64 `json:"project_id"`
	}

	var payload QueryPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on get matching embeddings handler")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}
	errCode, msg := store.GetStore().DeleteEmbeddingsByProject(payload.ProjectID)
	if errCode != http.StatusAccepted {
		log.Error("Failed to retrieve embeddings")
		c.JSON(errCode, gin.H{"error": msg})
		return
	}

	c.JSON(errCode, gin.H{"message": msg})
}
