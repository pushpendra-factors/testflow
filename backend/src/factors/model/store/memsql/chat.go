package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

func (store *MemSQL) DeleteEmbeddingsByProject(projectId int64) (int, string) {
	logFields := log.Fields{
		"method":     "DeleteEmbeddingsByProject",
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	// Delete all rows from the PromptEmbeddings table where the project ID matches the given project ID
	query := db.Where("project_id = ?", projectId).Delete(&model.PromptEmbeddings{})
	if err := query.Error; err != nil {
		log.WithError(err).Error("Failed deleting all prompt embeddings for project ID", projectId)
		return http.StatusInternalServerError, "Failed deleting all prompt embeddings"
	}

	if query.RowsAffected == 0 {
		log.Info("No embeddings found to delete for project ID", projectId)
	}

	return http.StatusAccepted, "Successfully deleted chat data"
}

func (store *MemSQL) AddAllEmbeddings(projectId int64, prompts []string, queries []string, embeddings [][]float32) (int, string) {
	logFields := log.Fields{
		"method": "AddAllEmbeddings",
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	// Check if the lengths of prompts, queries, and embeddings match
	if len(prompts) != len(embeddings) || len(prompts) != len(queries) {
		return http.StatusBadRequest, "Mismatched lengths of prompts, queries, and embeddings"
	}

	// Loop through prompts, queries, and embeddings and create records
	for i := range prompts {
		prompt := prompts[i]
		query := queries[i]
		embedding := embeddings[i]

		// Serialize the embedding slice into a JSON string
		embeddingJSON, err := json.Marshal(embedding)
		if err != nil {
			log.WithError(err).Error("Failed to marshal embedding to JSON")
			return http.StatusInternalServerError, "Failed to serialize embedding"
		}

		// Create a new record in the PromptEmbeddings table
		if err := db.Create(&model.PromptEmbeddings{ProjectID: projectId, Prompt: prompt, Query: query, Embedding: embeddingJSON}).Error; err != nil {
			log.WithError(err).Error("Failed to insert prompt and embedding")
			return http.StatusInternalServerError, "Failed to insert prompt and embedding"
		}
	}

	return http.StatusCreated, "Successfully inserted all prompt embeddings"
}

func (store *MemSQL) GetMatchingEmbeddings(projectId int64, queryEmbedding []float32) (int, string, model.PromptEmbeddingsPayload) {
	logFields := log.Fields{
		"method": "GetTopMatchingEmbeddings",
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// Serialize the embedding slice into a JSON string
	embeddingJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		log.WithError(err).Error("Failed to marshal embedding to JSON")
		return http.StatusInternalServerError, "Failed to retrieve top matching embeddings", model.PromptEmbeddingsPayload{}
	}

	db := C.GetServices().Db

	// Prepare the query
	query := `SELECT project_id, prompt, query, embedding,
		embedding <*> ? AS score
	FROM prompt_embeddings
	WHERE project_id = ?
	ORDER BY score DESC
	LIMIT 10;`

	// Execute the query and process the results
	rows, err := db.Raw(query, embeddingJSON, projectId).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to retrieve top matching embeddings")
		return http.StatusInternalServerError, "Failed to retrieve top matching embeddings", model.PromptEmbeddingsPayload{}
	}
	defer rows.Close()

	var embeddings []model.PromptEmbeddings
	for rows.Next() {
		var embeddingRecord model.PromptEmbeddings
		if err := db.ScanRows(rows, &embeddingRecord); err != nil {
			log.WithError(err).Error("Failed to scan top matching embeddings")
			return http.StatusInternalServerError, "Failed to scan top matching embeddings", model.PromptEmbeddingsPayload{}
		}
		embeddings = append(embeddings, embeddingRecord)
	}

	// Prepare the payload
	indexedPrompts := make([]string, 0, 10)
	indexedPromptEmbs := make([][]float32, 0, 10)
	indexedQueries := make([]string, 0, 10)

	for _, embeddingRecord := range embeddings {
		var embedding []float32
		if err := json.Unmarshal(embeddingRecord.Embedding, &embedding); err != nil {
			log.Printf("Failed to unmarshal embedding from JSON: %v", err)
			return http.StatusInternalServerError, "Failed to unmarshal embedding from JSON", model.PromptEmbeddingsPayload{}
		}

		indexedPrompts = append(indexedPrompts, embeddingRecord.Prompt)
		indexedPromptEmbs = append(indexedPromptEmbs, embedding)
		indexedQueries = append(indexedQueries, embeddingRecord.Query)
	}

	return http.StatusOK, "Successfully retrieved top matching embeddings", model.PromptEmbeddingsPayload{
		IndexedPrompts:    indexedPrompts,
		IndexedPromptEmbs: indexedPromptEmbs,
		IndexedQueries:    indexedQueries,
	}
}

func (store *MemSQL) GetMissingPromptsByProjectID(projectID int64, givenPrompts []string) (int, string, []string) {
	logFields := log.Fields{
		"method":    "GetMissingPromptsByProjectID",
		"projectID": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(givenPrompts) == 0 {
		return http.StatusOK, "No prompts given", nil
	}

	db := C.GetServices().Db

	// Create placeholders for the query
	placeholders := make([]string, len(givenPrompts))
	args := make([]interface{}, len(givenPrompts)+1)
	args[0] = projectID
	for i, prompt := range givenPrompts {
		placeholders[i] = "?"
		args[i+1] = prompt
	}

	// Prepare the query to find existing prompts
	query := `SELECT prompt FROM prompt_embeddings WHERE project_id = ? AND prompt IN (` + strings.Join(placeholders, ",") + `)`

	// Execute the query and process the results
	rows, err := db.Raw(query, args...).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to retrieve prompts by project ID")
		return http.StatusInternalServerError, "Failed to retrieve prompts by project ID", nil
	}
	defer rows.Close()

	existingPrompts := make(map[string]int)
	for rows.Next() {
		var prompt string
		if err := rows.Scan(&prompt); err != nil {
			log.WithError(err).Error("Failed to scan prompt")
			return http.StatusInternalServerError, "Failed to scan prompt", nil
		}
		existingPrompts[prompt] = 1
	}

	if err := rows.Err(); err != nil {
		log.WithError(err).Error("Error iterating over rows")
		return http.StatusInternalServerError, "Error iterating over rows", nil
	}

	// Determine missing prompts
	var missingPrompts []string
	for _, prompt := range givenPrompts {
		if _, found := existingPrompts[prompt]; !found {
			missingPrompts = append(missingPrompts, prompt)
		}
	}

	return http.StatusOK, "Successfully retrieved missing prompts", missingPrompts
}
