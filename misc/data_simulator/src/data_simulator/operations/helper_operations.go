package operations

import (
	"bufio"
	"compress/gzip"
	"context"
	"data_simulator/config"
	"data_simulator/constants"
	Log "data_simulator/logger"
	"data_simulator/utils"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"strings"

	"cloud.google.com/go/storage"
)

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func ExtractUserData(data string) (string, map[string]string) {
	split := strings.Split(data, " ")
	var op UserDataOutput
	if len(split) == 3 {
		json.Unmarshal([]byte(split[2]), &op)
	} else {
		//ignore: Fix this
	}
	return op.UserId, ConvertInterfaceToString(op.UserAttributes)
}

func LoadExistingUsers(env string) map[string]map[string]string {

	userData := make(map[string]map[string]string)

	var files []string
	if env == "development" || env == "docker" {
		files = utils.GetAllFiles(constants.LOCALOUTPUTFOLDER, config.ConfigV2.User_data_file_name_prefix)
	} else {
		files = utils.ListAllCloudFiles(fmt.Sprintf("%s/%s", constants.UNPROCESSEDFILESCLOUD, constants.LOCALOUTPUTFOLDER),
			constants.BUCKETNAME,
			config.ConfigV2.User_data_file_name_prefix)
	}
	for _, element := range files {
		var scanner *bufio.Scanner
		if strings.HasSuffix(element, ".gz") {
			var gzr *gzip.Reader
			if env == "development" || env == "docker" {
				gzr = utils.GetFileHandlegz(element)
			} else {
				_context := context.Background()
				_storageClient, _err := storage.NewClient(_context)
				if _err != nil {
					Log.Debug.Printf("%v", _err)
				}
				_bucket := _storageClient.Bucket(constants.BUCKETNAME)
				_object := _bucket.Object(element).ReadCompressed(true)
				_reader, _err := _object.NewReader(_context)
				if _err != nil {
					Log.Error.Fatal(_err)
				}
				defer _reader.Close()
				gzr, _err = gzip.NewReader(_reader)
				if _err != nil {
					Log.Error.Fatal(_err)
				}
				defer gzr.Close()
			}
			scanner = bufio.NewScanner(gzr)
		}
		if strings.HasSuffix(element, ".log") {
			if env == "development" || env == "docker" {
				_reader, _ := os.Open(element)
				scanner = bufio.NewScanner(_reader)
			} else {
				_context := context.Background()
				_storageClient, _err := storage.NewClient(_context)

				if _err != nil {
					Log.Debug.Printf("%v", _err)
				}
				_bucket := _storageClient.Bucket(constants.BUCKETNAME)
				_reader, _err := _bucket.Object(element).NewReader(_context)
				defer _reader.Close()
				scanner = bufio.NewScanner(_reader)
			}
		}
		for scanner.Scan() {
			s := scanner.Text()
			userId, attributes := ExtractUserData(s)
			userData[userId] = attributes
		}
	}
	Log.Debug.Printf("Existing User Count: %v\n", len(userData))
	return userData
}

func IsAllSegmentsDone(segmentStatus map[string]bool) bool {

	allSegmentsDone := true
	for _, element := range segmentStatus {
		if element == false {
			allSegmentsDone = false
			break
		}
	}
	return allSegmentsDone
}

func UserAlreadyExists(userId string, attributes map[string]map[string]string) bool {
	if attributes[userId] != nil {
		return true
	}
	return false
}
