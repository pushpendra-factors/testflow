package task

import (
	"encoding/json"
	"factors/filestore"
	"factors/model/store"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

func NumericalBucketing(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {
	//env := configs["env"].(string)
	//db := configs["db"].(*gorm.DB)
	cloudManager := configs["cloudManager"].(*filestore.FileManager)
	//etcdClient := configs["etcdClient"].(*serviceEtcd.EtcdClient)
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	//bucketName := configs["bucketName"].(string)
	startTimestamp := configs["startTimestamp"].(int64)
	//endTimestamp := configs["endTimestamp"].(int64)
	modelType := configs["modelType"].(string)
	status := make(map[string]interface{})

	efCloudPath, efCloudName := (*cloudManager).GetModelEventsFilePathAndName(projectId, startTimestamp, modelType)
	efTmpPath, efTmpName := diskManager.GetModelEventsFilePathAndName(projectId, startTimestamp, modelType)
	log.WithFields(log.Fields{"eventFileCloudPath": efCloudPath,
		"eventFileCloudName": efCloudName}).Info("Downloading events file from cloud.")
	eReader, err := (*cloudManager).Get(efCloudPath, efCloudName)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
			"eventFileName": efCloudName}).Error("Failed downloading events file from cloud.")
		status["error"] = "Failed downloading events file"
		return status, false
	}
	err = diskManager.Create(efTmpPath, efTmpName, eReader)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
			"eventFileName": efCloudName}).Error("Failed creating events file in local.")
		status["error"] = "Failed creating events file in local."
		return status, false
	}
	tmpEventsFilepath := efTmpPath + efTmpName
	log.Info("Successfuly downloaded events file from cloud.", tmpEventsFilepath, efTmpPath, efTmpName)

	//Download master file
	masterNumBucketCloudPath, masterNumBucketCloudName := (*cloudManager).GetMasterNumericalBucketsFile(projectId)
	masterNumBucketLocalPath, masterNumBucketLocalName := diskManager.GetMasterNumericalBucketsFile(projectId)
	weekNumBucketCloudPath, weekNumBucketCloudName := (*cloudManager).GetModelEventsNumericalBucketsFile(projectId, startTimestamp, modelType)
	weekNumBucketLocalPath, weekNumBucketLocalName := diskManager.GetModelEventsNumericalBucketsFile(projectId, startTimestamp, modelType)
	eReaderMaster, err := (*cloudManager).Get(masterNumBucketCloudPath, masterNumBucketCloudName)
	existingEPMasterBuckets, existingUPMasterBuckets := make(map[string][]BucketRange), make(map[string][]BucketRange)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "masterFile": masterNumBucketCloudPath,
			"numericalBucketFile": masterNumBucketCloudName}).Error("Failed downloading numerical bucket file from cloud.")
	} else {
		err = diskManager.Create(masterNumBucketLocalPath, masterNumBucketLocalName, eReaderMaster)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "numBucketsFilePath": masterNumBucketCloudPath,
				"numericalBucketFile": masterNumBucketCloudName}).Error("Failed creating numerical bucket file from cloud.")
		}
		tmpNumBucketFilepath := masterNumBucketLocalPath + masterNumBucketLocalName
		log.Info("Successfuly downloaded num bucket file from cloud.", tmpNumBucketFilepath, masterNumBucketLocalPath, masterNumBucketLocalName)
		existingEPMasterBuckets, existingUPMasterBuckets = GetMasterBuckets(tmpNumBucketFilepath)
	}

	eventProperties, userProperties, err := GetNumericalPropertiesWithDatarange(tmpEventsFilepath, projectId)
	if err != nil {
		log.WithError(err).Error("Failed finding min/max for  numerical properties")
		status["error"] = "Failed finding min/max for  numerical properties"
		return status, false
	}

	EPRanges := make(map[string][]BucketRange)
	UPRanges := make(map[string][]BucketRange)
	for key, ranges := range U.PREDEFINED_BIN_RANGES_FOR_PROPERTY {
		if !strings.HasPrefix(key, "UP") {
			EPRanges[key] = make([]BucketRange, 0)
			for _, values := range ranges {
				EPRanges[key] = append(EPRanges[key], BucketRange{
					Min:   int64(values[0]),
					Max:   int64(values[1]),
					Label: fmt.Sprintf("%v-%v", int(values[0]), int(values[1])),
				})
			}
		}
	}
	for key, ranges := range U.PREDEFINED_BIN_RANGES_FOR_PROPERTY {
		if !strings.HasPrefix(key, "EP") {
			UPRanges[key] = make([]BucketRange, 0)
			for _, values := range ranges {
				UPRanges[key] = append(UPRanges[key], BucketRange{
					Min:   int64(values[0]),
					Max:   int64(values[1]),
					Label: fmt.Sprintf("%v-%v", int(values[0]), int(values[1])),
				})
			}
		}
	}

	for key, bucketRange := range existingEPMasterBuckets {
		_, ok := EPRanges[key]
		if !ok {
			EPRanges[key] = bucketRange
		}
	}
	for key, bucketRange := range existingUPMasterBuckets {
		_, ok := UPRanges[key]
		if !ok {
			UPRanges[key] = bucketRange
		}
	}

	eventPropBuckets := BucketizeNumericalPropertiesUsingPercentile(eventProperties)
	userPropBuckets := BucketizeNumericalPropertiesUsingPercentile(userProperties)

	for key, bucketRange := range eventPropBuckets {
		_, ok := EPRanges[key]
		if !ok {
			EPRanges[key] = bucketRange
		}
	}
	for key, bucketRange := range userPropBuckets {
		_, ok := UPRanges[key]
		if !ok {
			UPRanges[key] = bucketRange
		}
	}

	fPath, fName := diskManager.GetModelEventsBucketingFilePathAndName(projectId, startTimestamp, modelType)
	serviceDisk.MkdirAll(fPath) // create dir if not exist.
	tmpBucketedEventsFile := fPath + fName
	WriteBucketedEvents(efTmpPath+efTmpName, EPRanges, UPRanges, tmpBucketedEventsFile)
	WriteBuckets(weekNumBucketLocalPath+weekNumBucketLocalName, EPRanges, UPRanges)
	WriteBuckets(masterNumBucketLocalPath+masterNumBucketLocalName, EPRanges, UPRanges)

	tmpOutputFile, err := os.Open(tmpBucketedEventsFile)
	if err != nil {
		log.WithField("error", err).Error("Failed to read tmp bucketed eventsfailed.")
		status["error"] = "Failed to read tmp bucketed eventsfailed."
		return status, false
	}

	cDir, cName := (*cloudManager).GetModelEventsBucketingFilePathAndName(projectId, startTimestamp, modelType)
	err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
	if err != nil {
		log.WithField("error", err).Error("Failed to write bucketed events. Upload failed.")
		status["error"] = "Failed to write bucketed events. Upload failed."
		return status, false
	}

	tmpNumBucketWeekFile, err := os.Open(weekNumBucketLocalPath + weekNumBucketLocalName)
	if err != nil {
		log.WithField("error", err).Error("Failed to read tmp num buckets failed.")
		status["error"] = "Failed to read tmp num buckets failed."
		return status, false
	}

	err = (*cloudManager).Create(weekNumBucketCloudPath, weekNumBucketCloudName, tmpNumBucketWeekFile)
	if err != nil {
		log.WithField("error", err).Error("Failed to write num buckets. Upload failed.")
		status["error"] = "Failed to write num buckets. Upload failed."
		return status, false
	}

	tmpNumBucketMasterFile, err := os.Open(masterNumBucketLocalPath + masterNumBucketLocalName)
	if err != nil {
		log.WithField("error", err).Error("Failed to read tmp num buckets - master failed.")
		status["error"] = "Failed to read tmp num buckets - master failed."
		return status, false
	}

	err = (*cloudManager).Create(masterNumBucketCloudPath, masterNumBucketCloudName, tmpNumBucketMasterFile)
	if err != nil {
		log.WithField("error", err).Error("Failed to write num buckets - master. Upload failed.")
		status["error"] = "Failed to write num buckets - master. Upload failed."
		return status, false
	}
	return nil, true
}

type MinMaxTuple struct {
	Min     float64
	Max     float64
	Numbers []float64
}

type BucketRange struct {
	Min   int64
	Max   int64
	Label string
}

// This method bucketizes the whole data set into 10 buckets ad
func BucketizeNumericalPropertiesUsingPercentile(props map[string]MinMaxTuple) map[string][]BucketRange {
	propBuckets := make(map[string][]BucketRange)
	for property, ranges := range props {
		_, ok := propBuckets[property]
		if ok == false {
			propBuckets[property] = make([]BucketRange, 0)
		}
		if ranges.Min != ranges.Max {
			propBuckets[property] = append(propBuckets[property], BucketRange{
				Min:   math.MinInt64,
				Max:   int64(ranges.Min),
				Label: fmt.Sprintf("lesserOrEquals %v", int64(ranges.Min)),
			})
			sort.Slice(ranges.Numbers, func(i, j int) bool {
				return ranges.Numbers[i] < ranges.Numbers[j]
			})
			bucketCount := int(1)
			for i := 1; i <= 10; i++ {
				partitionIndex := (float64(len(ranges.Numbers)) / 10) * float64(i)
				if int(partitionIndex) >= len(ranges.Numbers) {
					break
				}
				min := propBuckets[property][bucketCount-1].Max + 1
				max := int64(ranges.Numbers[int(partitionIndex)])
				multiplier := FindMultiplier(max - min)
				if max%multiplier != 0 {
					max = max + (multiplier - max%multiplier)
				}
				if min <= max {
					propBuckets[property] = append(propBuckets[property], BucketRange{
						Min:   int64(min),
						Max:   int64(max),
						Label: fmt.Sprintf("%v-%v", min, max),
					})
					bucketCount++
				}
			}
			propBuckets[property] = append(propBuckets[property], BucketRange{
				Min:   int64(propBuckets[property][len(propBuckets[property])-1].Max) + 1,
				Max:   math.MaxInt64,
				Label: fmt.Sprintf("greaterOrEquals %v", int64(propBuckets[property][len(propBuckets[property])-1].Max+1)),
			})
		} else {
			propBuckets[property] = append(propBuckets[property], BucketRange{
				Min:   math.MinInt64,
				Max:   int64(ranges.Min) - 1,
				Label: fmt.Sprintf("lesserOrEquals %v", ranges.Min-1),
			})
			propBuckets[property] = append(propBuckets[property], BucketRange{
				Min:   int64(ranges.Min),
				Max:   int64(ranges.Max),
				Label: fmt.Sprintf("%v-%v", ranges.Min, ranges.Max),
			})
			propBuckets[property] = append(propBuckets[property], BucketRange{
				Min:   int64(ranges.Max + 1),
				Max:   math.MaxInt64,
				Label: fmt.Sprintf("greaterOrEquals %v", ranges.Max+1),
			})
		}
	}
	return propBuckets
}

func FindMultiplier(n int64) int64 {
	multiplier := int64(1)
	for {
		if n/10 == 0 {
			break
		}
		multiplier = multiplier * 10
		n = n / 10
	}
	return multiplier
}

func GetMasterBuckets(tmpNumBucketFilePath string) (map[string][]BucketRange, map[string][]BucketRange) {
	scanner, err := OpenEventFileAndGetScanner(tmpNumBucketFilePath)
	epBuckets, upBuckets := make(map[string][]BucketRange), make(map[string][]BucketRange)
	if err != nil {
		return epBuckets, upBuckets
	}
	i := 0
	for scanner.Scan() {
		line := scanner.Text()
		if i == 0 {
			json.Unmarshal([]byte(line), &epBuckets)
		} else {
			json.Unmarshal([]byte(line), &upBuckets)
		}
	}
	return epBuckets, upBuckets
}

// This method takes all the numerical properties and returns the list of values seen so that bucketing can be done.
func GetNumericalPropertiesWithDatarange(tmpEventsFilePath string, projectID int64) (map[string]MinMaxTuple, map[string]MinMaxTuple, error) {
	scanner, err := OpenEventFileAndGetScanner(tmpEventsFilePath)
	if err != nil {
		return nil, nil, err
	}
	propertyNumRangeEventProperties := make(map[string]MinMaxTuple)
	propertyNumRangeUserProperties := make(map[string]MinMaxTuple)
	for scanner.Scan() {
		line := scanner.Text()
		var event P.CounterEventFormat
		json.Unmarshal([]byte(line), &event)
		for property, value := range event.EventProperties {
			var numValue float64
			ptype := store.GetStore().GetPropertyTypeByKeyValue(projectID, event.EventName, property, value, false)
			if ptype == U.PropertyTypeNumerical {
				numValue, err = U.GetPropertyValueAsFloat64(value)
				if err != nil {

					// try removing comma separated number
					cleanedValue := strings.ReplaceAll(U.GetPropertyValueAsString(value), ",", "")
					numValue, err = U.GetPropertyValueAsFloat64(cleanedValue)
					if err != nil {
						continue
					}
				}
			} else {
				continue
			}
			numRange, ok := propertyNumRangeEventProperties[property]
			if ok == false {
				propertyNumRangeEventProperties[property] = MinMaxTuple{
					Min: numValue,
					Max: numValue,
				}
			} else {
				if numRange.Min > numValue {
					numRange.Min = numValue
				}
				if numRange.Max < numValue {
					numRange.Max = numValue
				}
				propertyNumRangeEventProperties[property] = numRange
			}
		}
		for property, value := range event.UserProperties {
			var numValue float64
			ptype := store.GetStore().GetPropertyTypeByKeyValue(projectID, "", property, value, true)
			if ptype == U.PropertyTypeNumerical {
				numValue, err = U.GetPropertyValueAsFloat64(value)
				if err != nil {

					// try removing comma separated number
					cleanedValue := strings.ReplaceAll(U.GetPropertyValueAsString(value), ",", "")
					numValue, err = U.GetPropertyValueAsFloat64(cleanedValue)
					if err != nil {
						continue
					}
				}

			} else {
				continue
			}
			numRange, ok := propertyNumRangeUserProperties[property]
			if ok == false {
				propertyNumRangeUserProperties[property] = MinMaxTuple{
					Min:     numValue,
					Max:     numValue,
					Numbers: []float64{numValue},
				}
			} else {
				if numRange.Min > numValue {
					numRange.Min = numValue
				}
				if numRange.Max < numValue {
					numRange.Max = numValue
				}
				numRange.Numbers = append(numRange.Numbers, numValue)
				propertyNumRangeUserProperties[property] = numRange
			}
		}
	}
	return propertyNumRangeEventProperties, propertyNumRangeUserProperties, nil
}

func WriteBuckets(filePath string, eventBuckets map[string][]BucketRange, userBuckets map[string][]BucketRange) {
	file, err := os.Create(filePath)
	defer file.Close()
	lineBytes, err := json.Marshal(eventBuckets)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal ep buckets.")
	}
	transformedline := string(lineBytes)
	if _, err := file.WriteString(fmt.Sprintf("%s\n", transformedline)); err != nil {
		log.WithFields(log.Fields{"transformedline": transformedline, "err": err}).Error("Unable to write to file.")
	}
	lineBytes, err = json.Marshal(userBuckets)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal up buckets.")
	}
	transformedline = string(lineBytes)
	if _, err := file.WriteString(fmt.Sprintf("%s\n", transformedline)); err != nil {
		log.WithFields(log.Fields{"transformedline": transformedline, "err": err}).Error("Unable to write to file.")
	}
}

func WriteBucketedEvents(tmpEventsFilePath string, eventBuckets map[string][]BucketRange, userBuckets map[string][]BucketRange, tmpBucketedEventsFile string) {
	file, err := os.Create(tmpBucketedEventsFile)
	defer file.Close()

	scanner, err := OpenEventFileAndGetScanner(tmpEventsFilePath)
	if err != nil {
		return
	}
	for scanner.Scan() {
		line := scanner.Text()
		var event P.CounterEventFormat
		json.Unmarshal([]byte(line), &event)
		for property, value := range event.EventProperties {
			bucketRanges, ok := eventBuckets[property]
			if ok == false {
				continue
			}
			numValue, err := U.GetPropertyValueAsFloat64(value)
			if err != nil {
				continue
			}
			numValueInt := int64(numValue)
			for _, bucket := range bucketRanges {
				if numValueInt >= bucket.Min && numValueInt <= bucket.Max {
					event.EventProperties[property] = bucket.Label
				}
			}
		}
		for property, value := range event.UserProperties {
			bucketRanges, ok := userBuckets[property]
			if ok == false {
				continue
			}
			numValue, err := U.GetPropertyValueAsFloat64(value)
			if err != nil {
				continue
			}
			numValueInt := int64(numValue)
			for _, bucket := range bucketRanges {
				if numValueInt >= bucket.Min && numValueInt <= bucket.Max {
					event.UserProperties[property] = bucket.Label
				}
			}
		}
		lineBytes, err := json.Marshal(event)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event.")
		}
		transformedline := string(lineBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", transformedline)); err != nil {
			log.WithFields(log.Fields{"transformedline": transformedline, "err": err}).Error("Unable to write to file.")
		}
	}
}
