package model

import (
	C "factors/config"
	"net/http"
	"strings"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	SALT_LEN = 32
)

type Agent struct {
	UUID string `gorm:"primary_key:true;type:varchar(255);default:uuid_generate_v4()" json:"uuid"`

	FirstName string `gorm:"type:varchar(100)" json:"first_name"`
	LastName  string `gorm:"type:varchar(100)" json:"last_name"`

	Email           string `gorm:"type:varchar(100)" json:"email"`
	IsEmailVerified bool   `json:"is_email_verified"`

	Salt              string     `gorm:"type:varchar(100)" json:"-"` // Should we add a unique on salt ?
	Password          string     `gorm:"type:varchar(100)" json:"-"`
	PasswordCreatedAt *time.Time `json:"password_created_at"`

	InvitedBy *string `gorm:"type:uuid" json:"invited_by"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`

	LastLoggedInAt *time.Time `json:"last_logged_in_at"`
	LoginCount     uint64     `json:"login_count"`
}

// AgentInfo - Exposable Info.
type AgentInfo struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}

func (a *Agent) BeforeCreate(scope *gorm.Scope) error {
	a.Salt = U.RandomString(SALT_LEN)
	return nil
}

// TODO: Make index name a constant and read it
const error_Duplicate_email_error = "pq: duplicate key value violates unique constraint \"agent_email_unique_idx\""

func CreateAgent(agent *Agent) (*Agent, int) {
	if agent.Email == "" {
		log.Error("CreateAgent Failed. Email not provided.")
		return nil, http.StatusBadRequest
	}

	agent.Email = strings.ToLower(agent.Email)

	db := C.GetServices().Db

	if err := db.Create(agent).Error; err != nil {
		if err.Error() == error_Duplicate_email_error {
			log.WithError(err).Error("CreateAgent Failed, duplicate email")
			return nil, http.StatusBadRequest
		}
		log.WithError(err).Error("CreateAgent Failed")
		return nil, http.StatusInternalServerError
	}

	return agent, http.StatusCreated
}

func GetAgentByEmail(email string) (*Agent, int) {

	if email == "" {
		log.Error("GetAgentByEmail Failed. Email not provided.")
		return nil, http.StatusBadRequest
	}

	email = strings.ToLower(email)

	db := C.GetServices().Db

	var agent Agent
	if err := db.Limit(1).Where("email = ?", email).Find(&agent).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &agent, http.StatusFound
}

func GetAgentByUUID(uuid string) (*Agent, int) {
	if uuid == "" {
		log.Error("GetAgentyUUID Failed. UUID not provided.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var agent Agent

	if err := db.Limit(1).Where("uuid = ?", uuid).Find(&agent).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		log.WithError(err).Error("GetAgentyUUID Failed.")
		return nil, http.StatusInternalServerError
	}

	return &agent, http.StatusFound
}

func GetAgentInfo(uuid string) (*AgentInfo, int) {
	agent, errCode := GetAgentByUUID(uuid)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	agentInfo := &AgentInfo{FirstName: agent.FirstName, Email: agent.Email}
	return agentInfo, errCode
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func IsPasswordAndHashEqual(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func UpdateAgentPassword(uuid, plainTextPassword string, passUpdatedAt time.Time) int {

	if uuid == "" || plainTextPassword == "" {
		log.Error("UpdateAgentPassword Failed. Missing params")
		return http.StatusBadRequest
	}

	hashedPassword, err := HashPassword(plainTextPassword)
	if err != nil {
		return http.StatusInternalServerError
	}

	db := C.GetServices().Db

	db = db.Model(&Agent{}).Where("uuid = ?", uuid).
		Updates(map[string]interface{}{
			"password":            hashedPassword,
			"password_created_at": passUpdatedAt,
			"salt":                U.RandomString(SALT_LEN),
		})

	if db.Error != nil {
		log.WithError(db.Error).Error("UpdateAgentPassword Failed")
		return http.StatusInternalServerError
	}

	if db.RowsAffected == 0 {
		return http.StatusNoContent
	}

	return http.StatusAccepted
}

func UpdateAgentLastLoginInfo(email string, ts time.Time) int {
	if email == "" {
		log.Error("UpdateAgentLastLoginInfo Failed. Missing params")
		return http.StatusBadRequest
	}

	email = strings.ToLower(email)

	db := C.GetServices().Db

	db = db.Model(&Agent{}).Where("email = ?", email).Updates(map[string]interface{}{
		"last_logged_in_at": ts,
		"login_count":       gorm.Expr("login_count + ? ", 1)})

	if db.Error != nil {
		log.WithError(db.Error).Error("UpdateAgentLastLoginInfo Failed")
		return http.StatusInternalServerError
	}

	if db.RowsAffected == 0 {
		return http.StatusNoContent
	}

	return http.StatusAccepted
}

func UpdateAgentVerificationDetails(agentUUID, password, firstName, lastName string, verified bool, passUpdatedAt time.Time) int {

	if agentUUID == "" || firstName == "" || password == "" {
		log.Error("UpdateAgentVerificationDetails Failed. Missing params")
		return http.StatusBadRequest
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return http.StatusInternalServerError
	}

	db := C.GetServices().Db

	db = db.Model(&Agent{}).Where("uuid = ?", agentUUID).Updates(map[string]interface{}{
		"is_email_verified":   verified,
		"first_name":          firstName,
		"last_name":           lastName,
		"password":            hashedPassword,
		"password_created_at": passUpdatedAt,
	})

	if db.Error != nil {
		log.WithError(db.Error).Error("UpdateAgentVerificationDetails Failed")
		return http.StatusInternalServerError
	}

	if db.RowsAffected == 0 {
		return http.StatusNoContent
	}

	return http.StatusAccepted
}
