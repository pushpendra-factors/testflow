package postgres

import (
	"factors/model/model"
	"net/http"
)

func (pg *Postgres) GetPlanByID(planID uint64) (*model.Plan, int) {
	for _, plan := range model.Plans {
		if plan.ID == planID {
			return &plan, http.StatusFound
		}
	}
	return nil, http.StatusNotFound
}

func (pg *Postgres) GetPlanByCode(Code string) (*model.Plan, int) {
	for _, plan := range model.Plans {
		if plan.Code == Code {
			return &plan, http.StatusFound
		}
	}
	return nil, http.StatusNotFound
}
