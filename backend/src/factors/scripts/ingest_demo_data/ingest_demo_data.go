package main

import (
	"bufio"
	"time"

	//	"database/sql"
	"encoding/json"
	"errors"
	C "factors/config"
	M "factors/model/model"
	"factors/util"
	"flag"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var jobMap = map[string][]string{
	"dev_setup": []string{"adwords_documents", "facebook_documents", "linkedin_documents"},
}
var folderName string = "demo_data"

var numberOfLinesCopied int = 0
var numberOfLinesFailed int = 0

func main() {

	env := flag.String("env", C.DEVELOPMENT, "")
	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	flag.Parse()

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}
	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()
	if !(*env == "staging" || *env == "development" || *env == "docker") {
		log.Error("this can be run only in staging or development or docker")
		return
	}
	currentJobConfig := jobMap["dev_setup"]
	writeData(currentJobConfig, "dev_setup", *env)
}

func writeData(tables []string, jobType string, env string) {
	db := C.GetServices().Db
	defer db.Close()
	for _, tableNameValue := range tables {
		numberOfLinesCopied = 0
		numberOfLinesFailed = 0
		var path string
		if env == "development" {
			path = folderName + "/" + tableNameValue + ".txt"
		} else {
			path = "/go/bin/" + folderName + "/" + tableNameValue + ".txt"
		}

		file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Fatalf("open file error: %v", err)
			return
		}
		sc := bufio.NewScanner(file)
		for sc.Scan() {
			recordToBeInserted := sc.Text() // GET the line string
			var err error
			if tableNameValue == "adwords_documents" {
				adwordsObj := getAdwordsObj(recordToBeInserted)
				err = writeDataAdwords(adwordsObj, db)
			} else if tableNameValue == "facebook_documents" {
				facebookObj := getFacebookObj(recordToBeInserted)
				err = writeDataFacebook(facebookObj, db)
			} else if tableNameValue == "linkedin_documents" {
				linkedinObj := getLinkedinObj(recordToBeInserted)
				err = writeDataLinkedin(linkedinObj, db)
			}
			if err != nil {
				numberOfLinesFailed++
			} else {
				numberOfLinesCopied++
			}
		}
		log.Println(
			"succesfully inserted data into " + tableNameValue + " total number of records copied : " + strconv.Itoa(numberOfLinesCopied) + " total number of records failed : " + strconv.Itoa(numberOfLinesFailed))

	}
}

// returns the record object from string fetched from the text file.
func getAdwordsObj(recordToBeInserted string) M.AdwordsDocument {
	var adwordsObj M.AdwordsDocument
	json.Unmarshal([]byte(recordToBeInserted), &adwordsObj)
	adwordsObj.CreatedAt = time.Now()
	adwordsObj.UpdatedAt = time.Now()
	adwordsObj.Timestamp, _ = strconv.ParseInt(util.GetDateOnlyFromTimestampZ(util.TimeNowUnix()), 10, 64)
	return adwordsObj
}
func getFacebookObj(recordToBeInserted string) M.FacebookDocument {
	var facebookObj M.FacebookDocument
	json.Unmarshal([]byte(recordToBeInserted), &facebookObj)
	facebookObj.CreatedAt = time.Now()
	facebookObj.UpdatedAt = time.Now()
	facebookObj.Timestamp, _ = strconv.ParseInt(util.GetDateOnlyFromTimestampZ(util.TimeNowUnix()), 10, 64)
	return facebookObj
}
func getLinkedinObj(recordToBeInserted string) M.LinkedinDocument {
	var linkedinObj M.LinkedinDocument
	json.Unmarshal([]byte(recordToBeInserted), &linkedinObj)
	linkedinObj.CreatedAt = time.Now()
	linkedinObj.UpdatedAt = time.Now()
	linkedinObj.Timestamp, _ = strconv.ParseInt(util.GetDateOnlyFromTimestampZ(util.TimeNowUnix()), 10, 64)
	return linkedinObj
}

// pushes data to db tables when invoked.
func writeDataAdwords(data M.AdwordsDocument, db *gorm.DB) error {
	// function to write data to adwords_documents table
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into adwords_documents, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataFacebook(data M.FacebookDocument, db *gorm.DB) error {
	// function to write data to facebook_documents table
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into facebook_documents, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataLinkedin(data M.LinkedinDocument, db *gorm.DB) error {
	// function to write data to linked_documents table
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into linkedin_documents, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
