package main

/*
main file of datagen job
*/

import (
	"data_simulator/constants"
	Log "data_simulator/logger"
	"data_simulator/operations"
	"data_simulator/registration"
	"data_simulator/utils"
	"flag"
	"fmt"
	"time"
)

func main() {
	env := flag.String("env", "development", "")
	seedDate := flag.String("seed_date", "", "")
	dataConfig := flag.String("config", "", "")
	offsetInHours := flag.Int("offset_hours_past", 0, "")
	endpoint := flag.String("endpoint", "http://localhost:8085", "")
	authToken := flag.String("projectkey", "", "")

	flag.Parse()
	var date time.Time
	if *seedDate == "" {
		date = time.Now()
		if *offsetInHours != 0 {
			date = time.Now().Add(time.Hour * time.Duration(*offsetInHours))
		}
	} else {
		date, _ = time.Parse(constants.TIMEFORMAT, *seedDate)
	}
	filePattern := fmt.Sprintf("%v_%v_%v_%v",
		date.Year(),
		date.Month(),
		date.Day(),
		date.Hour())
	utils.CreateDirectoryIfNotExists(constants.LOCALOUTPUTFOLDER)
	Log.RegisterLogFiles(filePattern, "datagen")
	registration.RegisterHandlers(filePattern, *env, *dataConfig, date)
	fmt.Println(*env)
	fmt.Println(*authToken)
	fmt.Println(*endpoint)
	fmt.Println(*seedDate)
	operations.OperateV2(*env, endpoint, authToken)

}
