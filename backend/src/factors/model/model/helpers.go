package model

import (
	C "factors/config"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
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

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func IsPasswordAndHashEqual(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func LogOnSlowExecutionWithParams(starttime time.Time, params *log.Fields) {
	// Should be enabled using flag on need basis.
	if !C.IsSlowDBQueryLoggingEnabled() {
		return
	}

	timeTakenInMillsecs := time.Now().Sub(starttime).Milliseconds()
	logCtx := log.WithField("app_name", C.GetConfig().AppName).
		WithField("tag", "slow_exec").
		WithField("time_taken_in_ms", timeTakenInMillsecs).
		WithField("params", params)

	pc, _, _, _ := runtime.Caller(1)
	if fn := runtime.FuncForPC(pc); fn != nil {
		name := fn.Name()
		logCtx = logCtx.WithField("function_full", name)

		nameOnlySplit := strings.Split(name, ".")
		if len(nameOnlySplit) > 0 {
			logCtx = logCtx.WithField("function", nameOnlySplit[len(nameOnlySplit)-1])
		}
	}

	// Log based on threshold.
	if timeTakenInMillsecs > 50 {
		logCtx.Info("Slow query or method execution.")
	}
}
