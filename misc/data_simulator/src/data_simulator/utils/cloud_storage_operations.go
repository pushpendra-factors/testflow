package utils

/*
This file contains utils for cloud storage operations
*/

import(
	"cloud.google.com/go/storage"
	"os"
	"context"
	Log "data_simulator/logger"
	"fmt"
	"io"
	"google.golang.org/api/iterator"
)

func CopyFilesToCloud(folderPath string, bucketPath string, bucketName string, deleteSource bool){
	_context := context.Background()
	_storageClient, _err := storage.NewClient(_context)
	if _err != nil {
    	Log.Debug.Printf("%v", _err)
	}
    _bucket := _storageClient.Bucket(bucketName)
	files := GetAllFiles(folderPath, "")
    for _, element := range files {
        _file, _err := os.Open(element)
        if _err != nil {
            Log.Error.Fatal("os.Open: %v", _err)
        }
        defer _file.Close()
        _storageWriter := _bucket.Object(fmt.Sprintf("%v/%v",bucketPath, element)).NewWriter(_context)
        if _, _err = io.Copy(_storageWriter, _file); _err != nil {
            Log.Error.Fatal(_err)
        }
        if _err := _storageWriter.Close(); _err != nil {
			Log.Error.Fatal(_err)
		}
		if(deleteSource == true){
			os.Remove(element)
		}
    }
}

func MoveFilesInCloud(sourcePath string, destPath string, bucketName string){

	_context := context.Background()
	_storageClient, _err := storage.NewClient(_context)
	if _err != nil {
    	Log.Debug.Printf("%v", _err)
	}
    _bucket := _storageClient.Bucket(bucketName)
	_destinationObject := _bucket.Object(destPath)
	_sourceObject := _bucket.Object(sourcePath)

	if _, _err := _destinationObject.CopierFrom(_sourceObject).Run(_context); _err != nil {
		Log.Error.Fatal(_err)
	}
	if _err := _sourceObject.Delete(_context); _err != nil {
		Log.Error.Fatal(_err)
	}
}

func ListAllCloudFiles(path string, bucketName string, filePrefix string) []string{
	_context := context.Background()
	_storageClient, _err := storage.NewClient(_context)

	if _err != nil {
    	Log.Debug.Printf("%v", _err)
	}
	_bucket := _storageClient.Bucket(bucketName)
	
	it := _bucket.Objects(_context, &storage.Query{
		Prefix:    fmt.Sprintf("%s/%s", path, filePrefix),
		Delimiter: "/"})

	var _storageFileNames []string
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
				break
		}
		if err != nil {
				fmt.Errorf("%v", err)
		}
		_storageFileNames = append(_storageFileNames, attrs.Name)
	}
	return _storageFileNames
}