package model

import (
	C "factors/config"
	"time"
)

func OverrideCacheDateRangeForProjects(projectID uint64) time.Time {
	seedDate, ok := C.GetConfig().CacheLookUpRangeProjects[projectID]
	var currentDate time.Time
	if ok == true {
		currentDate = seedDate
	} else {
		currentDate = time.Now().UTC()
	}
	return currentDate
}
