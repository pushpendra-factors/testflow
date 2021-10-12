package model

import (
	C "factors/config"
	"factors/util"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
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

func BuildSelectColumns(tableStruct interface{}, excludedColumns []string) string {
	if len(excludedColumns) == 0 {
		return "*"
	}

	columns := make([]string, 0, 0)
	val := reflect.ValueOf(tableStruct)
	for i := 0; i < val.NumField(); i++ {
		colName := gorm.ToColumnName(val.Type().Field(i).Name)

		// Override with column_name provided as tag.
		gormTag := val.Type().Field(i).Tag.Get("gorm")
		for _, k := range strings.Split(gormTag, ";") {
			subTag := strings.Split(k, ":")
			if len(subTag) == 2 && subTag[0] == "column" && subTag[1] != "" {
				colName = subTag[1]
			}
		}

		if util.StringValueIn(colName, excludedColumns) {
			continue
		}
		columns = append(columns, colName)
	}

	return strings.Join(columns, ",")
}
