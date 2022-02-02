package memsql

import (
	"factors/model/model"
	"net/http"
	"time"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetPlanByID(planID uint64) (*model.Plan, int) {
	logFields := log.Fields{
		"plan_id": planID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for _, plan := range model.Plans {
		if plan.ID == planID {
			return &plan, http.StatusFound
		}
	}
	return nil, http.StatusNotFound
}

func (store *MemSQL) GetPlanByCode(Code string) (*model.Plan, int) {
	logFields := log.Fields{
		"code": Code,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for _, plan := range model.Plans {
		if plan.Code == Code {
			return &plan, http.StatusFound
		}
	}
	return nil, http.StatusNotFound
}
