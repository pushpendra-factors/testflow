package parser

/*
This interface can be extended to different types of parser implementation
For now, it is extended to YAML parser
YamlParser.go defines how to extend 
Registration.go defines how to instantiate
 */

type IParser interface{
	Parse([]byte, interface{})(interface{})
}