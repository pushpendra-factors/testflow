package model

import (
	C "factors/config"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func OverrideCacheDateRangeForProjects(projectID uint64) time.Time {
	seedDate, ok := C.GetConfig().CacheLookUpRangeProjects[projectID]
	var currentDate time.Time
	if ok == true {
		currentDate = seedDate
	} else {
		currentDate = time.Now().UTC()
	}
	return currentDate
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func IsPasswordAndHashEqual(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
