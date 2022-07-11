package task

import (
	"factors/filestore"
	serviceDisk "factors/services/disk"

	"github.com/jinzhu/gorm"
)

func ComputeStreamingMetrics(
	db *gorm.DB, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, bucketName string,
	projectId int64, modelId uint64) error {
	return nil
}
