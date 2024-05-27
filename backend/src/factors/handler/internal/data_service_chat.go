package internal

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func DataServiceAddEmbeddingsFromScratch(c *gin.Context) {
	r := c.Request

	var promptEmbeddings model.PromptEmbeddingsPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&promptEmbeddings); err != nil {
		log.WithError(err).Error("Failed to decode Json request")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	errCode, msg := store.GetStore().DeleteAllEmbeddings()
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": msg})
		return
	}

	errCode, msg = store.GetStore().AddAllEmbeddings(promptEmbeddings.IndexedPrompts, promptEmbeddings.IndexedQueries, promptEmbeddings.IndexedPromptEmbs)
	if errCode != http.StatusCreated {
		log.WithField("document", promptEmbeddings).Error("Failed to insert the promptEmbeddings")
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed to add prompt embedding"})
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully added all prompt embeddings from scratch"})
}

func DataServiceAddNewEmbeddings(c *gin.Context) {
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

	errCode, msg := store.GetStore().AddAllEmbeddings(promptEmbeddings.IndexedPrompts, promptEmbeddings.IndexedQueries, promptEmbeddings.IndexedPromptEmbs)
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

	statusCode, msg, embeddings := store.GetStore().GetMatchingEmbeddings(payload.QueryEmbedding)
	if statusCode != http.StatusOK {
		log.Error("Failed to retrieve embeddings")
		c.JSON(statusCode, gin.H{"error": msg})
		return
	}

	c.JSON(statusCode, gin.H{"message": msg, "data": embeddings})
}

func DataServiceGetDBPrompts(c *gin.Context) {
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

	statusCode, msg, prompts := store.GetStore().GetDBPromptsByProjectID(payload.ProjectID)
	if statusCode != http.StatusOK {
		log.Error("Failed to retrieve embeddings")
		c.JSON(statusCode, gin.H{"error": msg})
		return
	}

	c.JSON(statusCode, gin.H{"message": msg, "data": prompts})
}
