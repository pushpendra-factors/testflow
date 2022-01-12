package pattern

import (
	"bufio"
	"encoding/json"
	fp "factors/fptree"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"math"
	"os"
	"path"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// CounterEventFormat Format to write files for build sequence job.
type CounterEventFormat struct {
	UserId            string                 `json:"uid"`
	UserJoinTimestamp int64                  `json:"ujt"`
	EventName         string                 `json:"en"`
	EventTimestamp    int64                  `json:"et"`
	EventCardinality  uint                   `json:"ecd"`
	EventProperties   map[string]interface{} `json:"epr"`
	UserProperties    map[string]interface{} `json:"upr"`
}

type PropertiesInfo struct {
	NumericPropertyKeys          map[string]bool
	CategoricalPropertyKeyValues map[string]map[string]bool
}

// Event name to corresponding properties Info.
type UserAndEventsInfo struct {
	UserPropertiesInfo     *PropertiesInfo
	EventPropertiesInfoMap *map[string]*PropertiesInfo
	ModelVersion           int
}

type PropertiesCount struct {
	PropertyName string
	PropertyType string
	Count        uint
}

type EvSameTs struct {
	// events with same timeStamp as collected in this
	// struct
	EventsNames    []string
	EventsMap      map[string]CounterEventFormat
	EventTimestamp int64
}

const CURRENT_MODEL_VERSION = 2.0
const topK_patterns = 50

func NewUserAndEventsInfo() *UserAndEventsInfo {
	eMap := make(map[string]*PropertiesInfo)
	userAndEventsInfo := &UserAndEventsInfo{
		UserPropertiesInfo: &PropertiesInfo{
			NumericPropertyKeys:          make(map[string]bool),
			CategoricalPropertyKeyValues: make(map[string]map[string]bool),
		},
		EventPropertiesInfoMap: &eMap,
		ModelVersion:           CURRENT_MODEL_VERSION,
	}
	return userAndEventsInfo
}

func GenCandidatesPair(p1 *Pattern, p2 *Pattern, userAndEventsInfo *UserAndEventsInfo) (*Pattern, *Pattern, bool) {
	p1Len := len(p1.EventNames)
	p2Len := len(p2.EventNames)
	if p1Len != p2Len || p1Len == 0 {
		return nil, nil, false
	}

	numDifferent := 0
	differentIndex := -1
	for i := 0; i < p1Len; i++ {
		if strings.Compare(p1.EventNames[i], p2.EventNames[i]) != 0 {
			numDifferent += 1
			differentIndex = i
			if numDifferent > 1 {
				// Candidates cannot be generated from patterns that differ
				// by more than one event.
				return nil, nil, false
			}
		}
	}

	if numDifferent != 1 {
		return nil, nil, false
	}

	c1String := make([]string, p1Len)
	copy(c1String, p1.EventNames)
	c2String := make([]string, p1Len)
	copy(c2String, p1.EventNames)
	// Insert the different event in p2 before and after differentIndex in p1.
	c1String = append(c1String[:differentIndex], append([]string{p2.EventNames[differentIndex]}, c1String[differentIndex:]...)...)
	c2String = append(c2String[:differentIndex+1], append([]string{p2.EventNames[differentIndex]}, c2String[differentIndex+1:]...)...)
	c1Pattern, _ := NewPattern(c1String, userAndEventsInfo)
	c2Pattern, _ := NewPattern(c2String, userAndEventsInfo)
	return c1Pattern, c2Pattern, true
}

//GenCandidatesPairRepeat Generate patterns with same eventNames
func GenCandidatesPairRepeat(p1 *Pattern, p2 *Pattern, userAndEventsInfo *UserAndEventsInfo) (*Pattern, *Pattern, bool) {
	p1Len := len(p1.EventNames)
	p2Len := len(p2.EventNames)
	if p1Len != p2Len || p1Len != 1 {
		log.WithFields(log.Fields{"p1": p1.EventNames,
			"p2": p2.EventNames,
		}).Error("In GenCandidatesPairRepeat , got patterns with lenght not equal to 1")
		return nil, nil, false
	}

	c1String := make([]string, 0)
	c2String := make([]string, 0)
	c1String = append(c1String, p1.EventNames[0])
	c1String = append(c1String, p2.EventNames[0])
	c2String = append(c2String, p1.EventNames[0])
	c2String = append(c2String, p2.EventNames[0])
	c1Pattern, _ := NewPattern(c1String, userAndEventsInfo)
	c2Pattern, _ := NewPattern(c2String, userAndEventsInfo)
	return c1Pattern, c2Pattern, true
}

func candidatesMapToSlice(candidatesMap map[string]*Pattern) []*Pattern {
	candidates := []*Pattern{}
	for _, v := range candidatesMap {
		candidates = append(candidates, v)
	}
	return candidates
}

func GenCandidates(currentPatterns []*Pattern, maxCandidates int, userAndEventsInfo *UserAndEventsInfo, repeatedEvents []string) (
	[]*Pattern, uint, error) {
	numPatterns := len(currentPatterns)
	var currentMinCount uint

	if numPatterns == 0 {
		return nil, currentMinCount, fmt.Errorf("Zero Patterns")
	}
	// Sort current patterns in decreasing order of frequency.
	sort.Slice(
		currentPatterns,
		func(i, j int) bool {
			return currentPatterns[i].PerUserCount > currentPatterns[j].PerUserCount
		})
	candidatesMap := make(map[string]*Pattern)
	// Candidates are formed in decreasing order of frequent patterns till maxCandidates.
	for i := 0; i < numPatterns; i++ {
		for j := i + 1; j < numPatterns; j++ {
			if c1, c2, ok := GenCandidatesPair(
				currentPatterns[i], currentPatterns[j], userAndEventsInfo); ok {
				currentMinCount = currentPatterns[j].PerUserCount
				candidatesMap[c1.String()] = c1
				if len(candidatesMap) >= maxCandidates {
					return candidatesMapToSlice(candidatesMap), currentMinCount, nil
				}
				candidatesMap[c2.String()] = c2
				if len(candidatesMap) >= maxCandidates {
					return candidatesMapToSlice(candidatesMap), currentMinCount, nil
				}
			}
		}
	}
	//TODO(Vinith):The ordering might change if there is a count field in Pattern object,
	//and there could be no repeated patterns.
	for i := 0; i < numPatterns; i++ {
		repeatedCandidatesMap, err := GenRepeatedEventCandidates(repeatedEvents, currentPatterns[i], userAndEventsInfo)
		if err != nil {
			log.Error("unable to create repeated Pair")
			return nil, 0, err
		}
		for cKey, cPt := range repeatedCandidatesMap {
			candidatesMap[cKey] = cPt
		}

	}

	if len(candidatesMap) > maxCandidates {
		log.Fatal("More than max candidates generated.")
	}
	return candidatesMapToSlice(candidatesMap), currentMinCount, nil
}

//GenRepeatedEventCandidates Generate patterns with one off repeated string based on cyclic events
func GenRepeatedEventCandidates(repeatedEvents []string, pt *Pattern, userAndEventsInfo *UserAndEventsInfo) (map[string]*Pattern, error) {
	repeatedCandidatesMap := make(map[string]*Pattern)
	cyclicSet := make(map[string]bool)
	for _, cy := range repeatedEvents {
		cyclicSet[cy] = true
	}

	if len(pt.EventNames) < 2 {
		log.Error("Creating Patterns for length less than 2")
		return nil, nil
	}
	eventNamesList := pt.EventNames
	repeatedPatterns := make([]*Pattern, 0)
	for idx, ename := range eventNamesList {
		if _, ok := cyclicSet[ename]; ok {
			tmpEventsList := append(eventNamesList[:0:0], eventNamesList...)
			tmpEventsList = U.InsertRepeated(tmpEventsList, idx+1, ename)
			tmpPattern, err := NewPattern(tmpEventsList, userAndEventsInfo)
			if err != nil {
				log.Error("Error creating Pattern for repeated Events")
				return nil, err
			}
			repeatedPatterns = append(repeatedPatterns, tmpPattern)

		}
	}

	for _, p := range repeatedPatterns {
		repeatedCandidatesMap[p.String()] = p
	}

	return repeatedCandidatesMap, nil

}

func GetPattEndWithGoal(combinationPatterns, goalPatterns []*Pattern) []*Pattern {
	// Filter combination ending in goal or repeated event
	goalsMap := make(map[string]bool, 0)
	allGoals := make([]*Pattern, 0)
	allEventsMap := make(map[string]bool, 0)
	for _, v := range goalPatterns {
		goalsMap[v.EventNames[0]] = true
	}

	for _, v := range combinationPatterns {

		if allEventsMap[strings.Join(v.EventNames, "_")] == false {
			if len(v.EventNames) != 2 {
				log.WithField("Events", v.EventNames).Error("len of eventnames in pattern not equal to 2")
			}

			if goalsMap[v.EventNames[1]] == true {
				allGoals = append(allGoals, v)
			}

			allEventsMap[strings.Join(v.EventNames, "_")] = true
		}
	}

	return allGoals

}

//GenSegmentsForTopGoals form candidated with topK goal events
func GenCombinationPatternsEndingWithGoal(currentPatterns, GoalPatterns []*Pattern, userAndEventsInfo *UserAndEventsInfo) (
	// create pairs of (startPattern, GoalPattern) to count from events file
	[]*Pattern, uint, error) {
	numPatterns := len(currentPatterns)
	numGoalPatterns := len(GoalPatterns)
	var currentMinCount uint

	if numPatterns == 0 {
		return nil, currentMinCount, fmt.Errorf("zero patterns")
	}
	// Sort current patterns in decreasing order of frequency.
	sort.Slice(
		currentPatterns,
		func(i, j int) bool {
			return currentPatterns[i].PerUserCount > currentPatterns[j].PerUserCount
		})
	candidatesMap := make(map[string]*Pattern)
	// Candidates are formed with TopK goal Events
	for i := 0; i < numPatterns; i++ {
		for j := 0; j < numGoalPatterns; j++ {

			cr := currentPatterns[i]
			gl := GoalPatterns[j]
			if strings.Compare(cr.EventNames[0], gl.EventNames[0]) != 0 {

				if len(cr.EventNames) > 1 || len(gl.EventNames) > 1 {
					return nil, currentMinCount, fmt.Errorf("Length of events more than 1")
				}

				if c1, c2, ok := GenCandidatesPair(
					cr, gl, userAndEventsInfo); ok {
					currentMinCount = gl.PerUserCount
					candidatesMap[c1.String()] = c1
					candidatesMap[c2.String()] = c2
				}
			}

		}
	}

	lenTwoPatterns := candidatesMapToSlice(candidatesMap)
	combinationGoalPatterns := GetPattEndWithGoal(lenTwoPatterns, GoalPatterns)
	log.Info("number of patterns ending in goal : ", len(combinationGoalPatterns))

	// removing max Candidates filtering condition
	return combinationGoalPatterns, currentMinCount, nil
}

//GenSegmentsForRepeatedEvents form candidated for repeated goal events
func GenSegmentsForRepeatedEvents(currentPatterns []*Pattern, userAndEventsInfo *UserAndEventsInfo, repeatedEvents []*Pattern) (
	// create pairs of (startPattern, GoalPattern) to count from events file
	[]*Pattern, uint, error) {
	numPatterns := len(currentPatterns)
	numRepeatPatterns := len(repeatedEvents)
	var currentMinCount uint

	if numPatterns == 0 {
		return nil, currentMinCount, fmt.Errorf("Zero Patterns")
	}
	// Sort current patterns in decreasing order of frequency.
	sort.Slice(
		currentPatterns,
		func(i, j int) bool {
			return currentPatterns[i].PerUserCount > currentPatterns[j].PerUserCount
		})
	candidatesMap := make(map[string]*Pattern)
	// Candidates are formed with repeated Events
	for i := 0; i < numPatterns; i++ {
		for j := 0; j < numRepeatPatterns; j++ {

			cr := currentPatterns[i]
			gl := repeatedEvents[j]
			if strings.Compare(cr.EventNames[0], gl.EventNames[0]) == 0 {
				if len(cr.EventNames) > 1 || len(gl.EventNames) > 1 {
					return nil, currentMinCount, fmt.Errorf("Length of events more than 1")
				}

				if c1, c2, ok := GenCandidatesPairRepeat(
					cr, gl, userAndEventsInfo); ok {
					currentMinCount = gl.PerUserCount
					candidatesMap[c1.String()] = c1
					candidatesMap[c2.String()] = c2
				}
			}

		}
	}

	// removing max Candidates filtering condition
	return candidatesMapToSlice(candidatesMap), currentMinCount, nil
}

func deletePatternFromSlice(patternArray []*Pattern, pattern *Pattern) []*Pattern {
	// Delete all occurrences of the pattern.
	j := 0
	for _, p := range patternArray {
		if p != pattern {
			patternArray[j] = p
			j++
		}
	}
	patternArray = patternArray[:j]
	return patternArray
}

func PatternPropertyKey(patternIndex int, propertyName string) string {
	// property names are scoped by index of the event in pattern,
	// since different events can have same properties.
	return fmt.Sprintf("%d.%s", patternIndex, propertyName)
}

// Collects event info for the events initilaized in userAndEventsInfo.
const max_SEEN_PROPERTIES = 20000
const max_SEEN_PROPERTY_VALUES = 1000

func CollectPropertiesInfo(projectID uint64, scanner *bufio.Scanner, userAndEventsInfo *UserAndEventsInfo) (*map[string]PropertiesCount, error) {
	lineNum := 0
	userPropertiesInfo := userAndEventsInfo.UserPropertiesInfo
	eventInfoMap := userAndEventsInfo.EventPropertiesInfoMap
	numUniqueEvents := len(*eventInfoMap)
	maxProperties := max_SEEN_PROPERTIES / (int(float64(numUniqueEvents)/150.0) + 1)
	maxPropertyValues := max_SEEN_PROPERTY_VALUES / (int(float64(numUniqueEvents)/150.0) + 1)

	log.Info("Maximum Properties ", maxProperties, " ")
	log.Info("Maximum Properties Values ", maxPropertyValues, " ")

	allProps := make(map[string]PropertiesCount)

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		var eventDetails CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return nil, err
		}
		for key, value := range eventDetails.UserProperties {
			propKey := key + "_UP"
			if Pval, ok := allProps[propKey]; ok {
				Pval.Count = Pval.Count + 1
				Pval.PropertyName = key
				Pval.PropertyType = "UP"
				allProps[propKey] = Pval
			} else {
				PrVal := PropertiesCount{key, "UP", 1}
				allProps[propKey] = PrVal
			}

			propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectID, "", key, value, true)
			if propertyType == U.PropertyTypeNumerical {
				if len(userPropertiesInfo.NumericPropertyKeys) > maxProperties {
					continue
				}
				userPropertiesInfo.NumericPropertyKeys[key] = true
			} else if propertyType == U.PropertyTypeCategorical {
				if len(userPropertiesInfo.CategoricalPropertyKeyValues) > maxProperties {
					continue
				}
				categoricalValue := U.GetPropertyValueAsString(value)
				cmap, ok := userPropertiesInfo.CategoricalPropertyKeyValues[key]
				if !ok {
					cmap = make(map[string]bool)
					userPropertiesInfo.CategoricalPropertyKeyValues[key] = cmap
				}
				if len(cmap) < maxPropertyValues {
					cmap[categoricalValue] = true
				}
			} else {
				log.WithFields(log.Fields{"property": key, "value": value, "line no": lineNum}).Debug(
					"Ignoring non string, non numeric user property.")
			}
		}
		eventName := eventDetails.EventName
		eInfo, ok := (*eventInfoMap)[eventName]
		if !ok {
			log.WithFields(log.Fields{"event": eventName, "line no": lineNum}).Info("Unexpected event. Ignoring")
			continue
		}
		for key, value := range eventDetails.EventProperties {
			propKey := key + "_EP"
			if Pval, ok := allProps[propKey]; ok {
				Pval.Count = Pval.Count + 1
				Pval.PropertyName = key
				Pval.PropertyType = "EP"
				allProps[propKey] = Pval
			} else {
				PrVal := PropertiesCount{key, "EP", 1}
				allProps[propKey] = PrVal
			}
			propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectID, eventDetails.EventName, key, value, false)
			if propertyType == U.PropertyTypeNumerical {
				if len(eInfo.NumericPropertyKeys) > maxProperties {
					continue
				}
				eInfo.NumericPropertyKeys[key] = true
			} else if propertyType == U.PropertyTypeCategorical {
				if len(eInfo.CategoricalPropertyKeyValues) > maxProperties {
					continue
				}
				categoricalValue := U.GetPropertyValueAsString(value)
				cmap, ok := eInfo.CategoricalPropertyKeyValues[key]
				if !ok {
					cmap = make(map[string]bool)
					eInfo.CategoricalPropertyKeyValues[key] = cmap
				}
				if len(cmap) < maxPropertyValues {
					cmap[categoricalValue] = true
				}
			} else {
				log.WithFields(log.Fields{"event": eventName, "property": key, "value": value, "line no": lineNum}).Debug(
					"Ignoring non string, non numeric event property.")
			}
		}
	}
	return &allProps, nil
}

func ComputeAllUserPropertiesHistogram(projectID uint64, scanner *bufio.Scanner, pattern *Pattern, countsVersion int) error {
	var seenUsers map[string]bool = make(map[string]bool)
	numEventsProcessed := 0
	var countUsers uint
	if pattern.PatternVersion == 2 {
		// set all histogram to null
		pattern.userTree = fp.InitTree()
		pattern.PerOccurrenceEventNumericProperties = nil
		pattern.PerOccurrenceEventCategoricalProperties = nil
		pattern.GenericPropertiesHistogram = nil
		pattern.PerOccurrenceUserNumericProperties = nil
		pattern.PerUserUserNumericProperties = nil
		pattern.PerUserEventNumericProperties = nil
		pattern.PerUserUserCategoricalProperties = nil
		pattern.PerUserEventCategoricalProperties = nil
		pattern.PerOccurrenceUserCategoricalProperties = nil

	}
	for scanner.Scan() {
		line := scanner.Text()
		var eventDetails CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return err
		}
		userId := eventDetails.UserId
		userProperties := eventDetails.UserProperties

		numEventsProcessed += 1
		if math.Mod(float64(numEventsProcessed), 100000.0) == 0.0 {
			log.Info(fmt.Sprintf("ComputeAllUserPropertiesHistogram. Processed %d events", numEventsProcessed))
		}

		_, isSeenUser := seenUsers[userId]
		if !isSeenUser {
			countUsers += 1
			if countsVersion == 1 {
				nMap := make(map[string]float64)
				cMap := make(map[string]string)

				// Histogram of all user properties as seen in their first event is tracked.
				AddNumericAndCategoricalProperties(projectID, eventDetails.EventName, 0, userProperties, nMap, cMap, true)
				if err := pattern.PerUserUserNumericProperties.AddMap(nMap); err != nil {
					return err
				}
				if err := pattern.PerUserUserCategoricalProperties.AddMap(cMap); err != nil {
					return err
				}
			}
			if countsVersion == 2 {
				userCatProperties := FilterCategoricalProperties(projectID, eventDetails.EventName, 0, userProperties, true)
				if len(userCatProperties) > 0 {
					err := pattern.userTree.OrderAndInsertTrans(userCatProperties)
					if err != nil {
						log.Infof("err inserting user properties into tree:%v", err)
					}
				}
			}

		}
		seenUsers[userId] = true
	}

	if countsVersion == 2 {
		cpUser, err := fp.MineTree(pattern.userTree, topK_patterns)
		if err != nil {
			return err
		}
		pattern.UserPropertiesPatterns = cpUser
		pattern.userTree = fp.InitTree()

	}
	if countsVersion == 1 {
		count := uint(pattern.PerUserUserCategoricalProperties.Count())
		pattern.TotalUserCount = count
		pattern.PerUserCount = count
		pattern.PerOccurrenceCount = count

	} else {
		count := countUsers
		pattern.TotalUserCount = count
		pattern.PerUserCount = count
		pattern.PerOccurrenceCount = count

	}
	return nil
}

func CountPatterns(projectID uint64, scanner *bufio.Scanner, patterns []*Pattern,
	shouldCountOccurence bool, countVersion int) error {
	var seenUsers map[string]bool = make(map[string]bool)
	var mem runtime.MemStats

	var prEventTimeStamp int64
	var prvUserId string
	var numEventsProcessed int64
	eventToPatternsMap := make(map[string][]*Pattern)
	// Initialize.
	for _, p := range patterns {
		p.PatternVersion = countVersion
		for _, event := range p.EventNames {
			if _, ok := eventToPatternsMap[event]; !ok {
				eventToPatternsMap[event] = []*Pattern{}
			}
			eventToPatternsMap[event] = append(eventToPatternsMap[event], p)

		}
	}
	prEventTimeStamp = 0
	eventDetailsList := make([]CounterEventFormat, 0)
	lineNum := 0
	start_time := time.Now()

	for scanner.Scan() {
		line := scanner.Text()
		lineNum += 1
		var eventDetails CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return err
		}

		if (prEventTimeStamp != eventDetails.EventTimestamp && prEventTimeStamp != 0) || (strings.Compare(prvUserId, eventDetails.UserId) != 0 && prvUserId != "") {
			// check if eventTimeStamp are same and put all same event time stamp in same container to process
			// break if userId is same, else break and process container

			err := CountPatternsWithTS(projectID, eventDetailsList, &numEventsProcessed, seenUsers, patterns,
				eventToPatternsMap, shouldCountOccurence, countVersion)
			if err != nil {
				log.Errorf("Error processing line:%d, %v", lineNum, err)
				return err
			}
			eventDetailsList = make([]CounterEventFormat, 0)
			eventDetailsList = append(eventDetailsList, eventDetails)
			prEventTimeStamp = eventDetails.EventTimestamp
			prvUserId = eventDetails.UserId
		} else {
			eventDetailsList = append(eventDetailsList, eventDetails)
			prEventTimeStamp = eventDetails.EventTimestamp
			prvUserId = eventDetails.UserId

		}

		if lineNum%1000 == 0 {
			log.Infof("Read lines :%d", lineNum)
		}

	}
	if len(eventDetailsList) > 0 {
		// process the last event
		err := CountPatternsWithTS(projectID, eventDetailsList, &numEventsProcessed, seenUsers, patterns,
			eventToPatternsMap, shouldCountOccurence, countVersion)
		if err != nil {
			log.Errorf("Error processing line:%d", lineNum)
			return err
		}
	}
	end_time := time.Now()
	time_elapsed := end_time.UnixNano() - start_time.UnixNano()
	log.Infof("Avg time taken to insert each line:%v", float64(time_elapsed)/float64(lineNum))
	log.Infof("num of patterns to mine:%d", len(patterns))
	basePathUser := path.Join("/", "tmp", "fptree", "user")
	basePathEvent := path.Join("/", "tmp", "fptree", "event")
	err := os.MkdirAll(basePathUser, os.ModePerm)
	if err != nil {
		log.Fatal("unable to create temp user directory")
	}
	err = os.MkdirAll(basePathEvent, os.ModePerm)
	if err != nil {
		log.Fatal("unable to create temp event directory")
	}

	if countVersion == 2 {

		for _, pt := range patterns {
			ptName := U.RandStringBytes(5) + ".json"
			treePathUser := path.Join(basePathUser, ptName)
			pt.userTreePath = treePathUser
			log.Infof("Writing pattern user tree to path:%s , %v", treePathUser, pt.EventNames)

			tr := pt.userTree
			err := fp.SerializeTreeToFile(tr, treePathUser)
			if err != nil {
				log.Errorf("unable to Write pattern user tree to path:%s", treePathUser)
				return err
			}

			treePathEvent := path.Join(basePathEvent, ptName)
			pt.eventTreePath = treePathEvent
			tr = pt.eventTree
			err = fp.SerializeTreeToFile(tr, treePathEvent)
			if err != nil {
				log.Errorf("unable to Write pattern event tree to path:%s", treePathEvent)
				return err
			}
			pt.eventTree = fp.Tree{}
			pt.userTree = fp.Tree{}

			log.Infof("Done writing all trees to file")
		}

		// calling garbage collector to free memory
		U.PrintStats(mem)
		time.Sleep(3 * time.Second)
		runtime.GC()
		U.PrintStats(mem)

		num_patterns := 0
		for _, p := range patterns {
			num_patterns += 1
			var err error
			log.Infof("reading user  tree from: %d,%s ,%v", num_patterns, p.userTreePath, p.EventNames)
			p.userTree, err = fp.CreateTreeFromFile(p.userTreePath)
			if err != nil {
				log.Errorf("Unable to reconstruct user tree :%v", p.EventNames)
			}
			log.Infof("mining user tree from:%s ,%v", p.userTreePath, p.EventNames)

			cpUser, err := fp.MineTree(p.userTree, topK_patterns)
			if err != nil {
				return err
			}
			p.UserPropertiesPatterns = cpUser

			log.Infof("reading event tree from:%d, %s ,%v", num_patterns, p.eventTreePath, p.EventNames)
			p.eventTree, err = fp.CreateTreeFromFile(p.eventTreePath)
			if err != nil {
				log.Errorf("Unable to reconstruct user tree :%v", p.EventNames)
			}
			log.Infof("mining event tree from:%s ,%v", p.eventTreePath, p.EventNames)
			cpEvent, err := fp.MineTree(p.eventTree, topK_patterns)
			if err != nil {
				return err
			}
			p.EventPropertiesPatterns = cpEvent
			if num_patterns%1000 == 0 {
				log.Infof("num of patterns mined:%d", num_patterns)
			}

			p.eventTree = fp.Tree{}
			p.userTree = fp.Tree{}

			e := os.Remove(p.userTreePath)
			if e != nil {
				log.Fatal(e)
			}
			e = os.Remove(p.eventTreePath)
			if e != nil {
				log.Fatal(e)
			}

			if num_patterns%100 == 0 {
				U.PrintStats(mem)
				time.Sleep(1 * time.Second)
				runtime.GC()
				U.PrintStats(mem)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	for _, p := range patterns {

		if err := p.ResetAfterLastUser(); err != nil {
			log.Fatal(err)
		}
	}

	if countVersion == 2 {
		if countVersion == 2 {
			for _, p := range patterns {
				// set all histogram properties to null
				p.PerOccurrenceEventNumericProperties = nil
				p.PerOccurrenceEventCategoricalProperties = nil
				p.GenericPropertiesHistogram = nil
				p.PerOccurrenceUserNumericProperties = nil
				p.PerUserUserNumericProperties = nil
				p.PerUserEventNumericProperties = nil
				p.PerUserUserCategoricalProperties = nil
				p.PerUserEventCategoricalProperties = nil
				p.PerOccurrenceUserCategoricalProperties = nil
			}

		}
	}

	log.Infof("Total number of events Processed: %d", numEventsProcessed)

	return nil
}

func CountPatternsWithTS(projectID uint64, eventsList []CounterEventFormat, numEventsProcessed *int64,
	seenUsers map[string]bool, patterns []*Pattern, eventToPatternsMap map[string][]*Pattern,
	shouldCountOccurence bool, countVersion int) error {

	// process all events with same timeStamp and userId
	var eventPatterns []*Pattern

	eventPatterns = make([]*Pattern, 0)
	eventNames := make([]string, 0)
	eventsMap := make(map[string]CounterEventFormat)

	eventTimeStamp := eventsList[0].EventTimestamp
	userId := eventsList[0].UserId
	userJoinTimestamp := eventsList[0].UserJoinTimestamp

	for _, v := range eventsList {
		eventNames = append(eventNames, v.EventName)
		eventsMap[v.EventName] = v
		eventPatterns = append(eventPatterns, eventToPatternsMap[v.EventName]...)
		*numEventsProcessed += 1
		if math.Mod(float64(*numEventsProcessed), 100000.0) == 0.0 {
			log.Info(fmt.Sprintf("Processed %d events", numEventsProcessed))
		}
	}

	// get all unique patterns
	eventPatternsMap := make(map[*Pattern]bool)
	var uniqueEventPatterns []*Pattern
	for _, ep := range eventPatterns {
		eventPatternsMap[ep] = true
	}
	for k, _ := range eventPatternsMap {
		uniqueEventPatterns = append(uniqueEventPatterns, k)
	}

	ets := EvSameTs{EventsNames: eventNames, EventsMap: eventsMap, EventTimestamp: eventTimeStamp}

	_, isSeenUser := seenUsers[userId]
	if !isSeenUser {
		for _, p := range patterns {
			if err := p.ResetForNewUser(userId, userJoinTimestamp); err != nil {
				log.Fatal(err)
			}

		}
	}

	for _, p := range uniqueEventPatterns {
		if err := p.CountForEvent(projectID, userId, userJoinTimestamp, shouldCountOccurence, ets, countVersion); err != nil {
			log.Error("Error when counting event")
		}

	}
	seenUsers[userId] = true

	return nil
}

// Special candidate generation method that generates upto maxCandidates with
// events that start with and end with the two event patterns.
func GenLenThreeCandidatePatterns(pattern *Pattern, startPatterns []*Pattern,
	endPatterns []*Pattern, maxCandidates int, userAndEventsInfo *UserAndEventsInfo, repeatedEvents []string) ([]*Pattern, error) {
	if len(pattern.EventNames) != 2 {
		return nil, fmt.Errorf(fmt.Sprintf("Pattern %s length is not two.", pattern.String()))
	}
	sLen := len(startPatterns)
	eLen := len(endPatterns)
	minLen := int(math.Min(float64(sLen), float64(eLen)))

	eventsWithStartMap := make(map[string]bool)
	eventsWithEndMap := make(map[string]bool)
	for i := 0; i < sLen; i++ {
		if len(startPatterns[i].EventNames) != 2 {
			return nil, fmt.Errorf("Start pattern %s of not length two.",
				startPatterns[i].String())
		}
		if strings.Compare(
			startPatterns[i].EventNames[0], pattern.EventNames[0]) != 0 {
			return nil, fmt.Errorf("Pattern %s does not match start event of %s",
				startPatterns[i].String(), pattern.String())
		}
		eventsWithStartMap[startPatterns[i].EventNames[1]] = true
	}
	for i := 0; i < eLen; i++ {
		if len(endPatterns[i].EventNames) != 2 {
			return nil, fmt.Errorf("End pattern %s of not length two.",
				endPatterns[i].String())
		}
		if strings.Compare(
			endPatterns[i].EventNames[len(endPatterns[i].EventNames)-1],
			pattern.EventNames[1]) == 0 {
			eventsWithEndMap[endPatterns[i].EventNames[0]] = true
		}
	}

	candidatesMap := make(map[string]*Pattern)
	var err error
	// Alternate between startsWith and endsWith till the end of one.
	// The ordering of the patterns should be taken care by the caller.
	// The ones at the beginning are given higher priority.
	for i := 0; i < (sLen + eLen); i++ {
		var candidate *Pattern
		cString := make([]string, 2)
		if i < 2*minLen {
			j := int(i / 2)
			if math.Mod(float64(i), 2) < 1.0 {
				if reflect.DeepEqual(pattern.EventNames, startPatterns[j].EventNames) {
					continue
				}
				if found, _ := eventsWithEndMap[startPatterns[j].EventNames[1]]; !found {
					continue
				}
				copy(cString, startPatterns[j].EventNames)
				cString = append(cString, pattern.EventNames[1])
				candidate, err = NewPattern(cString, userAndEventsInfo)
			} else {
				if reflect.DeepEqual(pattern.EventNames, endPatterns[j].EventNames) {
					continue
				}
				if found, _ := eventsWithStartMap[endPatterns[j].EventNames[0]]; !found {
					continue
				}
				cString = []string{pattern.EventNames[0]}
				cString = append(cString, endPatterns[j].EventNames...)
				candidate, err = NewPattern(cString, userAndEventsInfo)
			}
		} else {
			j := i - minLen
			if sLen > eLen {
				if reflect.DeepEqual(pattern.EventNames, startPatterns[j].EventNames) {
					continue
				}
				if found, _ := eventsWithEndMap[startPatterns[j].EventNames[1]]; !found {
					continue
				}
				copy(cString, startPatterns[j].EventNames)
				cString = append(cString, pattern.EventNames[1])
				candidate, err = NewPattern(cString, userAndEventsInfo)
			} else {
				if reflect.DeepEqual(pattern.EventNames, endPatterns[j].EventNames) {
					continue
				}
				if found, _ := eventsWithStartMap[endPatterns[j].EventNames[0]]; !found {
					continue
				}
				cString = []string{pattern.EventNames[0]}
				cString = append(cString, endPatterns[j].EventNames...)
				candidate, err = NewPattern(cString, userAndEventsInfo)
			}
		}
		if err != nil {
			return nil, err
		}
		candidatesMap[candidate.String()] = candidate
		if len(candidatesMap) >= maxCandidates {
			return candidatesMapToSlice(candidatesMap), nil
		}
	}
	//TODO(Vinith): Add quota for repeatedCandidate and candidates
	repeatedCandidateMap, err := GenRepeatedEventCandidates(repeatedEvents, pattern, userAndEventsInfo)
	for cKey, cPt := range repeatedCandidateMap {
		candidatesMap[cKey] = cPt
	}

	if err != nil {
		return nil, err
	}
	return candidatesMapToSlice(candidatesMap), nil
}

func GenCandidatesForGoals(patternA, patternB *Pattern,
	userAndEventsInfo *UserAndEventsInfo) ([]*Pattern, error) {

	// patterns := make(*P.Pattern, 0)
	pattLen := len(patternA.EventNames)
	var allPatterns []*Pattern
	if len(patternA.EventNames) != len(patternB.EventNames) {
		return nil, fmt.Errorf("len of patterns are not equal")
	}
	if (patternA.EventNames[pattLen-1]) != (patternB.EventNames[pattLen-1]) {
		return nil, nil
	}
	tmpStringA := make([]string, pattLen+1)
	tmpStringB := make([]string, pattLen+1)

	allstrings := make([][]string, 0)
	tmpStringA[pattLen] = patternA.EventNames[pattLen-1]
	tmpStringB[pattLen] = patternB.EventNames[pattLen-1]

	count := 0
	for i := 0; i < pattLen; i++ {
		if strings.Compare(patternA.EventNames[i], patternB.EventNames[i]) != 0 {
			count++
		}
	}

	if count == 1 {
		for i := 0; i < len(patternA.EventNames)-1; i++ {

			if strings.Compare(patternA.EventNames[i], patternB.EventNames[i]) == 0 {
				tmpStringA[i] = patternA.EventNames[i]
				tmpStringB[i] = patternB.EventNames[i]

			} else {
				tmpStringA[i] = patternA.EventNames[i]
				tmpStringB[i] = patternB.EventNames[i]

				tmpStringA[i+1] = patternB.EventNames[i]
				tmpStringB[i+1] = patternA.EventNames[i]

				for j := i + 2; j < pattLen+1; j++ {
					tmpStringA[j] = patternA.EventNames[j-1]
					tmpStringB[j] = patternB.EventNames[j-1]
				}
				tmpA := make([]string, pattLen+1)
				tmpB := make([]string, pattLen+1)

				for idx := 0; idx < pattLen+1; idx++ {
					tmpA[idx] = tmpStringA[idx]
					tmpB[idx] = tmpStringB[idx]

				}

				allstrings = append(allstrings, tmpA)
				allstrings = append(allstrings, tmpB)

			}
		}

		for _, tmpPattEventNames := range allstrings {
			tmpPatt, err := NewPattern(tmpPattEventNames, userAndEventsInfo)
			if err != nil {
				return []*Pattern{}, fmt.Errorf("unable to generate new n+1 len Pattern")
			}
			allPatterns = append(allPatterns, tmpPatt)

		}
	} else {
		log.Info("EventNames are not matching in more than One ", patternA.EventNames, patternB.EventNames)
		return nil, nil
	}
	return allPatterns, nil

}
