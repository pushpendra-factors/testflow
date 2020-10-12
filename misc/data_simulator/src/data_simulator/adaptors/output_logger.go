package adaptors

/*
This interface can be extended to support different types of output writers
For now, it is extended to File writer and Log Writer
*/

type Writer interface{
	WriteOutput(string)
	RegisterOutputFile(string)
	WriteUserData(string)
	RegisterUserDataFile(string)
}
