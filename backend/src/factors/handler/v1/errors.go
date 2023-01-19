package v1

import "strings"

var INVALID_PROJECT string = "INVALID_PROJECT"
var INVALID_INPUT string = "INVALID_INPUT"
var PROCESSING_FAILED string = "PROCESSING_FAILED"
var DUPLICATE_RECORD string = "DUPLICATE_RECORD"
var DEPENDENT_RECORD_PRESENT string = "DEPENDENT_RECORD_PRESENT"

var ErrorMessages = map[string]string{
	INVALID_PROJECT:   "Project Id is Invalid",
	INVALID_INPUT:     "Input Params are incorrect",
	PROCESSING_FAILED: "Processing Failed",
	DUPLICATE_RECORD:  "Entity with the same name exists, please try a different name.",
}

// Takes the objects which depend on the entity to build an error message 
// Currently being used in KPI - Custom Metrics and Property Mappings delete to build a error message when dependent objects are present
func BuildDependentsErrorMessage(entity string, dependentObjectLists [][]string, dependentObjectTypeNames []string) string {
	if len(dependentObjectLists) != len(dependentObjectTypeNames) {
		return "Error in building error message"
	}

	errorMessage := "This " + entity + " is part of \""
	IsPrevious := false
	length := 0

	for index, dependentObjects := range dependentObjectLists {
		if len(dependentObjects) > 0 {
			length += len(dependentObjects)
			if IsPrevious {
				errorMessage = errorMessage + " and \""
			}
			errorMessage = errorMessage + strings.Join(dependentObjects, "\", \"") + "\" " + dependentObjectTypeNames[index]
			if len(dependentObjects) > 1 {
				errorMessage = errorMessage + "s"
			}
			IsPrevious = true
		}
	}

	pronoun := "it"
	if length > 1 {
		pronoun = "them"
	}
	errorMessage = errorMessage + ". Please remove " + pronoun + " first."
	return errorMessage
}