package registration

/*
Registering config Reader, ouput writer
*/

import(
	"data_simulator/config"
	"data_simulator/config/parser"
	"data_simulator/utils"
	"data_simulator/adaptors"
	Log "data_simulator/logger"
	"fmt"
    "data_simulator/constants"
    "time"
)

var WriterInstance adaptors.Writer

func RegisterHandlers(filePattern string, env string, dataConfig string, seedTime time.Time){	
    
	Log.Debug.Println("Registering Handlers")

	Log.Debug.Println("Registering Yaml Parser")
	var _parser parser.IParser
	_parser = parser.YamlParser{}
	if(env != "development"){
		config.GenerateInputConfigV2(_parser,fmt.Sprintf("go/bin/%s.yaml", dataConfig), seedTime)
	} else {
		config.GenerateInputConfigV2(_parser,fmt.Sprintf("../config/%s.yaml", dataConfig), seedTime)
	}
	// log.Println("Registering Output to File Writer")
	// WriterInstance = utils.FileWriter{}
	// WriterInstance.RegisterOutputFile(config.ConfigV2.Output_file_name)
	Log.Debug.Println("Registering Output to Log Writer")
	WriterInstance = utils.LogWriter{}
    WriterInstance.RegisterOutputFile(fmt.Sprintf("%s/%s_%s.log",
        constants.LOCALOUTPUTFOLDER,
		dataConfig, 
		filePattern))

	Log.Debug.Println("Registering UserData to Log Writer")
    WriterInstance.RegisterUserDataFile(fmt.Sprintf("%s/%s_%s.log",
		constants.LOCALOUTPUTFOLDER,	
		config.ConfigV2.User_data_file_name_prefix,
		filePattern))
	
	Log.Debug.Println("Registration Done !!!")
}