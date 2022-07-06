package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesBillingAccountForeignConstraints(ba model.BillingAccount) int {
	logFields := log.Fields{
		"ba": ba,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if _, errCode := store.GetAgentByUUID(ba.AgentUUID); errCode != http.StatusFound {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func (store *MemSQL) createBillingAccount(planCode string, AgentUUID string) (*model.BillingAccount, int) {
	logFields := log.Fields{
		"plan_code":  planCode,
		"agent_uuid": AgentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if planCode == "" || AgentUUID == "" {
		return nil, http.StatusBadRequest
	}

	plan, errCode := store.GetPlanByCode(planCode)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	bA := &model.BillingAccount{
		ID:        U.GetUUID(),
		PlanID:    plan.ID,
		AgentUUID: AgentUUID,
	}

	if errCode := store.satisfiesBillingAccountForeignConstraints(*bA); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}

	db := C.GetServices().Db
	if bA.PlanID == 0 {
		log.Errorf("Error Creating Billing Account for agent: %s, missing planID", AgentUUID)
		return nil, http.StatusBadRequest
	}

	if err := db.Create(bA).Error; err != nil {
		log.WithError(err).Error("createBillingAccount Failed")
		return nil, http.StatusInternalServerError
	}

	return bA, http.StatusCreated
}

func (store *MemSQL) existsBillingAccountByID(billingAccountID string) bool {
	logFields := log.Fields{
		"billin_account_id": billingAccountID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var billingAccount model.BillingAccount
	err := db.Limit(1).Where("id = ?", billingAccountID).Select("id").Find(&billingAccount).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.WithField("ba_id", billingAccountID).Error("Failed to check if billing account exists")
		}
		return false
	}

	if billingAccount.ID != "" {
		return true
	}
	return false
}

func (store *MemSQL) GetBillingAccountByProjectID(projectID int64) (*model.BillingAccount, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		return nil, http.StatusBadRequest
	}

	bA := model.BillingAccount{}

	db := C.GetServices().Db
	if err := db.Joins("JOIN project_billing_account_mappings ON project_billing_account_mappings.billing_account_id = billing_accounts.id").Where("project_billing_account_mappings.project_id = ? ", projectID).Find(&bA).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &bA, http.StatusFound
}

func (store *MemSQL) GetBillingAccountByAgentUUID(AgentUUID string) (*model.BillingAccount, int) {
	logFields := log.Fields{
		"agent_uuid": AgentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if AgentUUID == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	bA := model.BillingAccount{}
	if err := db.Limit(1).Where("agent_uuid = ?", AgentUUID).Find(&bA).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &bA, http.StatusFound
}

func (store *MemSQL) UpdateBillingAccount(id string, planId uint64, orgName, billingAddr, pinCode, phoneNo string) int {
	logFields := log.Fields{
		"id":           id,
		"plan_id":      planId,
		"org_name":     orgName,
		"billing_addr": billingAddr,
		"pincode":      pinCode,
		"phone_no":     phoneNo,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if id == "" || planId == 0 {
		log.WithFields(log.Fields{
			"id":      id,
			"plan_id": planId,
		}).Errorln("Missing params UpdateBillingAccount")
		return http.StatusBadRequest
	}

	if planId != model.FreePlanID {
		if orgName == "" || billingAddr == "" || pinCode == "" {
			log.WithFields(log.Fields{
				"id":          id,
				"planId":      planId,
				"orgName":     orgName,
				"billingAddr": billingAddr,
				"pincode":     pinCode,
			}).Errorln("Missing params UpdateBillingAccount")
			return http.StatusBadRequest
		}
	}

	db := C.GetServices().Db
	db = db.Model(&model.BillingAccount{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"organization_name": orgName,
			"billing_address":   billingAddr,
			"pincode":           pinCode,
			"phone_no":          phoneNo,
			"plan_id":           planId,
		})

	if db.Error != nil {
		log.WithError(db.Error).Error("UpdateBillingAccount Failed")
		return http.StatusInternalServerError
	}

	if db.RowsAffected == 0 {
		return http.StatusNoContent
	}
	return http.StatusAccepted
}

func (store *MemSQL) GetProjectsUnderBillingAccountID(ID string) ([]model.Project, int) {
	logFields := log.Fields{
		"id": ID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	projects := make([]model.Project, 0, 0)

	if err := db.Joins("JOIN project_billing_account_mappings ON project_billing_account_mappings.project_id = projects.id").Where("project_billing_account_mappings.billing_account_id = ? ", ID).Find(&projects).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	return projects, http.StatusFound
}

func (store *MemSQL) GetAgentsByProjectIDs(projectIDs []int64) ([]*model.Agent, int) {
	logFields := log.Fields{
		"project_ids": projectIDs,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	agents := make([]*model.Agent, 0, 0)

	if err := db.Joins("INNER JOIN project_agent_mappings ON project_agent_mappings.agent_uuid = agents.uuid").Where("project_agent_mappings.project_id IN (?)", projectIDs).Group("agents.uuid").Find(&agents).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	return agents, http.StatusFound
}

func (store *MemSQL) GetAgentsUnderBillingAccountID(ID string) ([]*model.Agent, int) {
	logFields := log.Fields{
		"id": ID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	agents := make([]*model.Agent, 0, 0)

	err := db.Joins("JOIN project_agent_mappings ON project_agent_mappings.agent_uuid = agents.uuid").
		Joins("JOIN project_billing_account_mappings ON project_billing_account_mappings.project_id = project_agent_mappings.project_id").Where("project_billing_account_mappings.billing_account_id = ? ", ID).
		Group("agents.uuid").
		Find(&agents).Error
	if err != nil {
		return agents, http.StatusInternalServerError
	}
	return agents, http.StatusFound
}

func (store *MemSQL) IsNewProjectAgentMappingCreationAllowed(projectID int64, emailOfAgentToAdd string) (bool, int) {
	logFields := log.Fields{
		"project_id":            projectID,
		"email_of_agent_to_add": emailOfAgentToAdd,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	billingAccount, errCode := store.GetBillingAccountByProjectID(projectID)
	if errCode != http.StatusFound {
		return false, errCode
	}

	agents, errCode := store.GetAgentsUnderBillingAccountID(billingAccount.ID)
	for _, agent := range agents {
		if agent.Email == emailOfAgentToAdd {
			// agent already present in billing account
			return true, http.StatusOK
		}
	}

	plan, errCode := store.GetPlanByID(billingAccount.PlanID)
	if len(agents) >= plan.MaxNoOfAgents {
		return false, http.StatusOK
	}

	return true, http.StatusOK
}
