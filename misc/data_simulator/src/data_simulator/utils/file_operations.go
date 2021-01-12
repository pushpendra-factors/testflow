package utils

/*
Util for File based operations
*/

import (
	"compress/gzip"
	Log "data_simulator/logger"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var file *os.File
var m sync.Mutex

func GetAllUnreadFiles(FileRootPath string, filenameprefix string, extension string) []string {

	var files []string
	Log.Debug.Printf("Searching for all the files in path %s with extension %s and File name prefix %s", FileRootPath, extension, filenameprefix)

	err := filepath.Walk(FileRootPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != extension {
			return nil
		}
		if !strings.HasPrefix(info.Name(), filenameprefix) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		Log.Error.Fatal(err)
	}
	return files
}

func GetAllFiles(FileRootPath string, filenameprefix string) []string {

	var files []string
	Log.Debug.Printf("Searching for all the files in path %s", FileRootPath)

	err := filepath.Walk(FileRootPath, func(path string, info os.FileInfo, err error) error {
		if filenameprefix != "" {
			if !strings.HasPrefix(info.Name(), filenameprefix) {
				return nil
			}
		}
		if !(filepath.Ext(path) == ".log" || filepath.Ext(path) == ".gz") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		Log.Error.Fatal(err)
	}
	return files
}

func GetFileHandle(FilePath string) *os.File {

	Log.Debug.Printf("ReadingFile %s", FilePath)
	handle, err := os.Open(FilePath)
	if err != nil {
		Log.Error.Fatal(err)
	}
	return handle
}

func GetFileHandlegz(FilePath string) *gzip.Reader {

	Log.Debug.Printf("ReadingFile %s", FilePath)
	handle, err := os.Open(FilePath)
	if err != nil {
		Log.Error.Fatal(err)
	}
	zipReader, err := gzip.NewReader(handle)
	if err != nil {
		Log.Error.Fatal(err)
	}
	defer zipReader.Close()
	return zipReader
}

func ReadFile(FileName string) []byte {

	//workingDirectory, err := os.Getwd()
	data, err := ioutil.ReadFile(FileName)
	// exec, _ := os.Executable()
	// files := GetAllFiles("go/bin", false)
	// fmt.Println(files)
	// fmt.Println(exec)
	// fmt.Println(data)
	if err != nil {
		Log.Error.Fatal(err)
	}
	return data

}

type FileWriter struct{}

func (f FileWriter) Write(data string) {
	m.Lock()
	file.WriteString(data + "\n")
	m.Unlock()
}

func DoesFileExist(fileName string) bool {

	workingDirectory, _ := os.Getwd()
	path := workingDirectory + "/" + fileName

	var _, err = os.Stat(path)
	return !os.IsNotExist(err)
}

func DeleteFile(fileName string) {
	workingDirectory, _ := os.Getwd()
	path := workingDirectory + "/" + fileName
	e := os.Remove(path)
	if e != nil {
		Log.Debug.Printf("Failed to remove file or doesn't exit %s", path)
	}
}

func CreateFile(fileName string) bool {

	workingDirectory, _ := os.Getwd()
	path := workingDirectory + "/" + fileName

	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		Log.Debug.Printf("Creating File %s", path)
		var _, err = os.Create(path)
		if err != nil {
			Log.Error.Fatal(err)
			return false
		}
		return true
	}
	return false
}

func (f FileWriter) RegisterOutputFile(FileName string) {

	workingDirectory, _ := os.Getwd()
	path := workingDirectory + "/" + FileName

	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		Log.Debug.Printf("Creating File %s", path)
		var _, err = os.Create(path)
		if err != nil {
			Log.Error.Fatal(err)
		}
	}

	file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		Log.Error.Fatal(err)
	}
}

func CreateDirectoryIfNotExists(folderPath string) {
	_, err := os.Stat(folderPath)

	if os.IsNotExist(err) {
		dir := os.MkdirAll(folderPath, 0755)
		if dir != nil {
			Log.Error.Fatal(err)
		}
	}
}

func MoveFiles(old string, new string) {
	err := os.Rename(old, new)
	if err != nil {
		Log.Error.Fatal(err)
	}
}
