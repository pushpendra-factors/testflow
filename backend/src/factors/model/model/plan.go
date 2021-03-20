package model

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
	FreePlanCode = "free"
	FreePlanID   = uint64(1)

	StartupPlanCode = "startup"
	StartupPlanID   = uint64(2)

	EnterprisePlanCode = "enterprise"
)

var FreePlan = Plan{
	ID:               FreePlanID,
	Name:             "Free",
	Code:             FreePlanCode,
	BasePrice:        0,
	MaxNoOfAgents:    10000,
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
	MaxNoOfAgents:    10000,
	DaysToRetainData: 365,

	MonthlyNoOfEvents:     5000000,
	MaxMontlyNoOfEvents:   6250000,
	MonthlyNoOfFreeEvents: 1000000,
	NoOfEventsInBatch:     100000,
	PricePerBatch:         10,
}

var Plans = []Plan{FreePlan, StartupPlan}
