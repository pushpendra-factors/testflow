package dd_attribution

import (
	"encoding/json"
	"factors/filestore"
	serviceDisk "factors/services/disk"
	"fmt"

	log "github.com/sirupsen/logrus"

	"encoding/json"
	"factors/filestore"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type SetFreqDist struct {
	maxItems	int
	setFreqDist   map[Set]interface{}
}

func computeTouchPointSynergies(projectId uint64, periodCode Period, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, deltaQuery Query, multiStepQuery MultiFunnelQuery, k int, unionOfFeatures *(map[string]map[string]bool), passId int, insightGranularity string, isEventOccurence bool, isMultiStep bool) error {
	scanner, err := GetEventFileScanner(projectId, periodCode, cloudManager, diskManager)
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Scanner initialization failed for period %v", periodCode))
		return err
	}
	lineNum := 0
	var convEvent string
	var tpSetFreqDist SetFreqDist
	for scanner.Scan() {
		lineNum++
		if lineNum%10000 == 0 {
			fmt.Printf("%d lines scanned\n", lineNum)
		}
		line := scanner.Text()
		var event P.CounterEventFormat
		json.Unmarshal([]byte(line), &event) // TODO: Add error check.
		sanitizeScreenSize(&event)
		if prevUserId == "" {
			prevUserId = event.UserId
		}
		if event.UserId == prevUserId { // If the same user's events are continuing...
			sessionId = updateSessions(&preSessionEvents, &sessions, sessionId, event)
			if sessionId == -1 {
				continue
			}
		} else { // If a new user's events have started coming...
			err = updateTouchPointSetFreq(&tpSetFreqDist, preSessionEvents, sessions, convEvent)
			}
			preSessionEvents = nil
			sessions = nil
			sessionId = -1
			sessionId = updateSessions(&preSessionEvents, &sessions, sessionId, event)
			prevUserId = event.UserId

		}
	}
	writeToFile(tpSetFreqDist)
	return nil
}

func computeShapleyValues() {
    synergy_dict = ComputeTouchPointSynergies()
    shapley = defaultdict(float)
    for coal, syn in synergy_dict.items():
        size = len(coal)
        for item in coal:
            shapley[item] += syn/size
    scores = shapley
    return scores
}
