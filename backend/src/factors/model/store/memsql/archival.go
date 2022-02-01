package memsql

import (
	"fmt"
	"net/http"
	"time"

	"factors/model/model"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

// GetNextArchivalBatches Returns a list of EventsArchivalBatch from the give startTime to 1 day before.
func (store *MemSQL) GetNextArchivalBatches(projectID uint64, startTime int64, maxLookbackDays int, hardStartTime, hardEndTime time.Time) ([]model.EventsArchivalBatch, error) {
	var eventsArchivalBatches []model.EventsArchivalBatch
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"max_look_back_days": maxLookbackDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var completedBatches map[int64]int64
	endTime := time.Unix(U.GetBeginningOfDayTimestampZ(U.TimeNowUnix())-1, 0)

	maxLookBackTime := U.GetBeginningOfDayTimestampZ(U.TimeNowZ().AddDate(0, 0, -maxLookbackDays).Unix())
	if !hardStartTime.IsZero() {
		startTime = U.GetBeginningOfDayTimestampZ(hardStartTime.Unix())
		endTime = hardEndTime.Add(time.Second * time.Duration(1))
		var status int
		completedBatches, status = store.GetCompletedArchivalBatches(projectID, hardStartTime, hardEndTime)
		if status == http.StatusInternalServerError {
			return eventsArchivalBatches, fmt.Errorf("Failed to get completed batches")
		}
	} else if maxLookBackTime > startTime {
		// If startTime older than maxLookback allowed, set it to oldest allowed date.
		startTime = maxLookBackTime
	}

	if startTime > endTime.Unix() {
		// Start time > end of yesterday's time. No batches to be processed.
		log.WithFields(log.Fields{"project_id": projectID, "start": startTime, "end": endTime.Unix()}).Info("Invalid date range")
		return eventsArchivalBatches, nil
	}

	batchTime := time.Unix(startTime, 0).UTC()
	for batchTime.Before(endTime) {
		nextBatchTime := batchTime.AddDate(0, 0, 1)
		batchStartTime := U.GetBeginningOfDayTimestampZ(batchTime.Unix())
		batchEndTime := U.GetBeginningOfDayTimestampZ(nextBatchTime.Unix()) - 1
		if !hardStartTime.IsZero() {
			_, found := completedBatches[batchStartTime]
			if found {
				// Start date part of completed batches. Skip adding to the new batches.
				batchTime = nextBatchTime
				continue
			}
		}
		eventsArchivalBatches = append(eventsArchivalBatches, model.EventsArchivalBatch{
			StartTime: batchStartTime,
			EndTime:   batchEndTime,
		})

		batchTime = nextBatchTime
	}

	return eventsArchivalBatches, nil
}
