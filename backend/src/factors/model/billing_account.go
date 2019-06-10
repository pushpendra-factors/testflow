package model

import (
	C "factors/config"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type BillingAccount struct {
	ID        uint64 `gorm:"primary_key:true;" json:"id"`
	PlanID    uint64 `gorm:"not null;" json:"plan_id"`
	AgentUUID string `gorm:"not null;" json:"agent_uuid"`

	OrganizationName string `json:"organization_name"`
	BillingAddress   string `json:"billing_address"`
	Pincode          string `json:"pincode"`
	PhoneNo          string `json:"phone_no"` // Optional

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func createBillingAccount(planCode string, AgentUUID string) (*BillingAccount, int) {

	if planCode == "" || AgentUUID == "" {
		return nil, http.StatusBadRequest
	}

	plan, errCode := GetPlanByCode(planCode)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	bA := &BillingAccount{
		PlanID:    plan.ID,
		AgentUUID: AgentUUID,
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

func GetBillingAccountByProjectID(projectID uint64) (*BillingAccount, int) {
	if projectID == 0 {
		return nil, http.StatusBadRequest
	}

	bA := BillingAccount{}

	db := C.GetServices().Db
	if err := db.Joins("JOIN project_billing_account_mappings ON project_billing_account_mappings.billing_account_id = billing_accounts.id").Where("project_billing_account_mappings.project_id = ? ", projectID).Find(&bA).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &bA, http.StatusFound
}

func GetBillingAccountByAgentUUID(AgentUUID string) (*BillingAccount, int) {
	if AgentUUID == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	bA := BillingAccount{}
	if err := db.Limit(1).Where("agent_uuid = ?", AgentUUID).Find(&bA).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &bA, http.StatusFound
}

func UpdateBillingAccount(id, planId uint64, orgName, billingAddr, pinCode, phoneNo string) int {
	if id == 0 || planId == 0 {
		log.WithFields(log.Fields{
			"id":      id,
			"plan_id": planId,
		}).Errorln("Missing params UpdateBillingAccount")
		return http.StatusBadRequest
	}

	if planId != FreePlanID {
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
	db = db.Model(&BillingAccount{}).Where("id = ?", id).
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

func GetProjectsUnderBillingAccountID(ID uint64) ([]Project, int) {
	db := C.GetServices().Db
	projects := make([]Project, 0, 0)

	if err := db.Joins("JOIN project_billing_account_mappings ON project_billing_account_mappings.project_id = projects.id").Where("project_billing_account_mappings.billing_account_id = ? ", ID).Find(&projects).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	return projects, http.StatusFound
}

func GetAgentsByProjectIDs(projectIDs []uint64) ([]*Agent, int) {
	db := C.GetServices().Db
	agents := make([]*Agent, 0, 0)

	if err := db.Joins("INNER JOIN project_agent_mappings ON project_agent_mappings.agent_uuid = agents.uuid").Where("project_agent_mappings.project_id IN (?)", projectIDs).Group("agents.uuid").Find(&agents).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	return agents, http.StatusFound
}

func GetAgentsUnderBillingAccountID(ID uint64) ([]*Agent, int) {
	db := C.GetServices().Db
	agents := make([]*Agent, 0, 0)

	err := db.Joins("JOIN project_agent_mappings ON project_agent_mappings.agent_uuid = agents.uuid").
		Joins("JOIN project_billing_account_mappings ON project_billing_account_mappings.project_id = project_agent_mappings.project_id").Where("project_billing_account_mappings.billing_account_id = ? ", ID).
		Group("agents.uuid").
		Find(&agents).Error
	if err != nil {
		return agents, http.StatusInternalServerError
	}
	return agents, http.StatusFound
}

func IsNewProjectAgentMappingCreationAllowed(projectID uint64, emailOfAgentToAdd string) (bool, int) {
	billingAccount, errCode := GetBillingAccountByProjectID(projectID)
	if errCode != http.StatusFound {
		return false, errCode
	}

	agents, errCode := GetAgentsUnderBillingAccountID(billingAccount.ID)
	for _, agent := range agents {
		if agent.Email == emailOfAgentToAdd {
			// agent already present in billing account
			return true, http.StatusOK
		}
	}

	plan, errCode := GetPlanByID(billingAccount.PlanID)
	if len(agents) >= plan.MaxNoOfAgents {
		return false, http.StatusOK
	}

	return true, http.StatusOK
}
