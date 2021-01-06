package config

/*
Reads the input config and applies required customization on object
*/

import (
	"data_simulator/config/parser"
	Log "data_simulator/logger"
	"data_simulator/utils"
	"fmt"
	"math/rand"
	"time"
)

var ConfigV2 ConfigurationV2

func GenerateInputConfigV2(parserInstance parser.IParser, FileName string, seedtime time.Time) {
	rand.Seed(time.Now().UTC().UnixNano())
	Log.Debug.Printf("Processing input config - %s", FileName)
	InputConfig := parserInstance.Parse(utils.ReadFile(FileName), ConfigV2)
	ConfigV2 = *InputConfig.(*ConfigurationV2)
	ConfigV2.User_id_prefix = fmt.Sprintf("%s_%v_", ConfigV2.User_id_prefix, seedtime.Unix())
	for item, element := range ConfigV2.User_segments {
		element.Start_Time = seedtime
		if string(element.Start_Time.Weekday()) == "Saturday" || string(element.Start_Time.Weekday()) == "Sunday" {
			element.Number_of_users = int(element.Number_of_users/4) + rand.Intn(int(element.Number_of_users/2))
		} else {
			element.Number_of_users = int(element.Number_of_users/2) + rand.Intn(int(element.Number_of_users/2))
		}
		if element.Start_Time.Day() < 7 {
			element.Number_of_users = element.Number_of_users + int(element.Number_of_users/10) + rand.Intn(int(element.Number_of_users/10))
		}
		if element.Start_Time.Day() > 25 {
			element.Number_of_users = element.Number_of_users - int(element.Number_of_users/10) - rand.Intn(int(element.Number_of_users/10))
		}
		if element.Start_Time.Hour() > 20 || element.Start_Time.Hour() < 8 {
			element.Number_of_users = element.Number_of_users - int(element.Number_of_users/10)
		}
		ConfigV2.User_segments[item] = element
	}
	Log.Debug.Printf("Operating with config %v", ConfigV2)
}
