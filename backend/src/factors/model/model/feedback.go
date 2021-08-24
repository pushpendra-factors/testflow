package model

import (
	"github.com/jinzhu/gorm/dialects/postgres"
	"time"
)

const VOTE_TYPE_UPVOTE int = 1
const VOTE_TYPE_DOWNVOTE int = 2

type Feedback struct {
	ID        string          `json:"id"`
	ProjectID uint64          `json:"project_id"`
	Feature   string          `json:"feature"`
	Property  *postgres.Jsonb `json:"property"`
	//1 for upvote 2 for downvote
	VoteType  int        `json:"vote_type"`
	CreatedBy string     `json:"created_by"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
type WeeklyInsightsProperty struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	QueryID     uint64 `json:"query_id"`
	Order       int    `json:"order"`
	Entity      string `json:"entity"`
	IsIncreased bool   `json:"is_increased"`
	Date        string `json:"date"`
}
type ExplainProperty struct {
	EventsWithProperties FactorsGoalRule `json:"ewp"`
	Type                 string          `json:"type"`
	Key                  string          `json:"key"`
	Value                string          `json:"value"`
	IsIncreased          bool            `json:"is_increased"`
	Date                 string          `json:"date"`
}
