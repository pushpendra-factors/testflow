package parser

/*
Takes Yaml byte stream and converts it to corresponding object
*/

import (
	"gopkg.in/yaml.v2"
	"reflect"
	Log "data_simulator/logger"
)

type YamlParser struct{}
func (y YamlParser) Parse(FileContents []byte, outputObj interface{}) (interface{}) {

	obj := reflect.New(reflect.TypeOf(outputObj)).Interface()
	err := yaml.Unmarshal(FileContents, obj)
	if(err != nil){
		Log.Error.Fatal(err)
	}
	return obj
}