CREATE TABLE IF NOT EXISTS prompt_embeddings (
    project_id bigint NOT NULL DEFAULT 0,
    prompt TEXT,
    query TEXT,
    embedding VECTOR(768, F32) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, prompt)
    );

ALTER TABLE prompt_embeddings ADD VECTOR INDEX idx_hnsw(embedding)
INDEX_OPTIONS '{
  "index_type": "HNSW_FLAT",
  "M": 30,
  "efConstruction": 40,
  "ef": 16,
  "metric_type":"DOT_PRODUCT"
}';
