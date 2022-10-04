package v1

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
