package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) UploadFilterFile(fileReference string, projectId int64){
	uploadFilterFile := model.UploadFilterFiles{
		FileReference:		fileReference,
		ProjectID: 			projectId,
		CreatedAt:          U.TimeNowZ(),
		UpdatedAt:          U.TimeNowZ(),
	}
	db := C.GetServices().Db

	if err := db.Create(&uploadFilterFile).Error; err != nil {
		log.WithError(err).Error("Failure to insert upload filereference")
	}

}