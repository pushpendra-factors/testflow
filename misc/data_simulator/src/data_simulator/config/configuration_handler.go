package config

/*
Reads the input config and applies required customization on object
*/

import(
	"data_simulator/config/parser"
	"data_simulator/utils"
	"time"
	Log "data_simulator/logger"
	"fmt"
)

var ConfigV2 ConfigurationV2
func GenerateInputConfigV2(parserInstance parser.IParser, FileName string, seedtime time.Time){

	Log.Debug.Printf("Processing input config - %s", FileName)
	InputConfig := parserInstance.Parse(utils.ReadFile(FileName), ConfigV2)
	ConfigV2 = *InputConfig.(*ConfigurationV2)
	ConfigV2.User_id_prefix = fmt.Sprintf("%s_%v_",ConfigV2.User_id_prefix, seedtime.Unix())
	for item, element := range ConfigV2.User_segments {
		element.Start_Time = seedtime
        ConfigV2.User_segments[item] = element
	}
	Log.Debug.Printf("Operating with config %v", ConfigV2)
}