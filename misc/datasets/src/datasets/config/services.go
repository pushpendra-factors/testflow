package config

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	// postgres dialect for gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Services struct {
	Db *gorm.DB
}

// global services object.
var services = &Services{}

func InitDB(dbHost string, dbPort int, dbUser string, dbName string, dbPass string) {
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbName, dbPass))
	if err != nil {
		log.Fatal("Failed initializing DB")
	}

	// Connection Pooling and Logging.
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.LogMode(true)

	services.Db = db
}

func GetServices() *Services {
	return services
}
