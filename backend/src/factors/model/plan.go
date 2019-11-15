package model

import "net/http"

type Plan struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`

	BasePrice        float64 `json:"-"`
	MaxNoOfAgents    int     `json:"max_no_of_agents"`
	DaysToRetainData int     `json:"-"`

	// Soft Limit
	MonthlyNoOfEvents int64 `json:"-"`

	// Hard limit, track API will fail after exceeding this
	MaxMontlyNoOfEvents int64 `json:"-"`

	MonthlyNoOfFreeEvents int64   `json:"-"`
	NoOfEventsInBatch     int64   `json:"-"`
	PricePerBatch         float64 `json:"-"`
}

const (
	FreePlanCode       = "free"
	StartupPlanCode    = "startup"
	EnterprisePlanCode = "enterprise"
)
const (
	FreePlanID    = uint64(1)
	StartupPlanID = uint64(2)
)

var FreePlan = Plan{
	ID:               FreePlanID,
	Name:             "Free",
	Code:             FreePlanCode,
	BasePrice:        0,
	MaxNoOfAgents:    3,
	DaysToRetainData: 90,

	MonthlyNoOfEvents:     500000,
	MaxMontlyNoOfEvents:   625000,
	MonthlyNoOfFreeEvents: 625000,
	NoOfEventsInBatch:     100000,
	PricePerBatch:         0,
}
var StartupPlan = Plan{
	ID:               StartupPlanID,
	Name:             "Startup",
	Code:             StartupPlanCode,
	BasePrice:        49,
	MaxNoOfAgents:    50,
	DaysToRetainData: 365,

	MonthlyNoOfEvents:     5000000,
	MaxMontlyNoOfEvents:   6250000,
	MonthlyNoOfFreeEvents: 1000000,
	NoOfEventsInBatch:     100000,
	PricePerBatch:         10,
}

var plans = []Plan{FreePlan, StartupPlan}

func GetPlanByID(planID uint64) (*Plan, int) {
	for _, plan := range plans {
		if plan.ID == planID {
			return &plan, http.StatusFound
		}
	}
	return nil, http.StatusNotFound
}

func GetPlanByCode(Code string) (*Plan, int) {
	for _, plan := range plans {
		if plan.Code == Code {
			return &plan, http.StatusFound
		}
	}
	return nil, http.StatusNotFound
}
