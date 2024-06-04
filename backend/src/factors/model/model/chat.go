package model

import "time"

type PromptEmbeddings struct {
	ProjectID int64     `gorm:"not null; primary_key:true" json:"project_id"`
	Prompt    string    `gorm:"primary_key" json:"prompt"`
	Query     string    `json:"query"`
	Embedding []byte    `gorm:"not null" json:"embedding"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PromptEmbeddingsPayload struct {
	ProjectId         int64       `json:"Project_id"`
	IndexedPrompts    []string    `json:"indexed_prompts"`
	IndexedQueries    []string    `json:"indexed_queries"`
	IndexedPromptEmbs [][]float32 `json:"indexed_prompt_embs"`
}

type QueryEmbeddingPayload struct {
	ProjectId      int64     `json:"project_id"`
	QueryEmbedding []float32 `json:"query_embedding"`
}
