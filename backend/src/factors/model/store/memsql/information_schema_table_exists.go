package memsql

import (
	C "factors/config"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type CountOfTableStruct struct{ CountOfTable int `gorm:"column:count_of_table"` }

const websiteAggregationTableExistsSql = "SELECT count(1) as count_of_table FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE='BASE TABLE' AND TABLE_NAME='website_aggregation'"

func (store *MemSQL) TableOfWebsiteAggregationExists() (bool, int) {

	var countOfTableStruct CountOfTableStruct
	db := C.GetServices().Db
	if err := db.Raw(websiteAggregationTableExistsSql).Scan(&countOfTableStruct).Error; err != nil {
		log.WithField("err", err).Error("Failed to execute website aggregation exists")
		return false, http.StatusInternalServerError
	}

	return countOfTableStruct.CountOfTable != 0, http.StatusOK
}
