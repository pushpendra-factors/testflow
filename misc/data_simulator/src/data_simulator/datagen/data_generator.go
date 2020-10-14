package main

/*
main file of datagen job
*/

import(
	"data_simulator/registration"
	"data_simulator/operations"	
    Log "data_simulator/logger"
    "time"
    "fmt"
    "data_simulator/utils"
    "flag"
    "data_simulator/constants"
)

func main(){
    env := flag.String("env", "development", "")
    seedDate := flag.String("seed_date", "", "")
	dataConfig := flag.String("config","","")
    flag.Parse()
    var date time.Time
    if *seedDate == "" {
        date = time.Now()        
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
	operations.OperateV2(*env)
}  