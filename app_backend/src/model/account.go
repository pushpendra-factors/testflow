package model

import (
	C "config"
	"net/http"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type Account struct {
	ID        uint64    `gorm:"primary_key:true;" json:"id"`
	Name      string    `gorm:"not null;unique"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateAccount(account *Account) (*Account, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"account": &account}).Info("Creating account")

	// Input Validation. (ID is to be auto generated)
	if account.ID > 0 {
		return nil, http.StatusBadRequest
	}

	if err := db.Create(account).Error; err != nil {
		log.WithFields(log.Fields{"account": &account, "error": err}).Error("CreateAccount Failed")
		return nil, http.StatusInternalServerError
	} else {
		return account, -1
	}
}
