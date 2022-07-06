package bigquery

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"factors/filestore"
	"factors/model/model"
	"factors/model/store"

	"cloud.google.com/go/bigquery"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	BIGQUERY_TABLE_EVENTS string = "f_events"
	BIGQUERY_TABLE_USERS  string = "f_users"
)

var bigqueryArchivalTables = map[string]model.EventFileFormat{
	BIGQUERY_TABLE_EVENTS: model.ArchiveEventTableFormat{},
	BIGQUERY_TABLE_USERS:  model.ArchiveUsersTableFormat{},
}

var bqTaskID = "Service#Bigquery"
var bqLog = log.WithFields(log.Fields{
	"Prefix": bqTaskID,
})

// CreateBigqueryClient Creates and returns a new bigquery client.
func CreateBigqueryClient(ctx *context.Context, bigquerySetting *model.BigquerySetting) (*bigquery.Client, error) {
	client, err := bigquery.NewClient(*ctx, bigquerySetting.BigqueryProjectID,
		option.WithCredentialsJSON([]byte(bigquerySetting.BigqueryCredentialsJSON)))
	if err != nil {
		bqLog.WithError(err).Error("Failed to create bigquery client")
		return nil, err
	}
	return client, nil
}

// CreateBigqueryClientForProject Creates and returns a bigquery client for the given projectID.
func CreateBigqueryClientForProject(ctx *context.Context, projectID int64) (*bigquery.Client, error) {
	bigquerySetting, status := store.GetStore().GetBigquerySettingByProjectID(projectID)
	if status == http.StatusInternalServerError {
		return nil, fmt.Errorf("Failed to get bigquery setting for project_id %d", projectID)
	} else if status == http.StatusNotFound {
		return nil, fmt.Errorf("No BigQuery configuration found for project id %d in database", projectID)
	}

	client, err := CreateBigqueryClient(ctx, bigquerySetting)
	if err != nil {
		bqLog.WithError(err).Error("Failed to create bigquery client")
		return nil, err
	}
	return client, nil
}

// ExecuteQuery Executes a given query on Biquery for given client. Writes output to writer.
func ExecuteQuery(ctx *context.Context, client *bigquery.Client, query string, resultSet *[][]string) error {
	bqLog.Infof("Executing '%s' on Bigquery", query)
	q := client.Query(query)
	job, err := q.Run(*ctx)
	if err != nil {
		return err
	}
	status, err := job.Wait(*ctx)
	if err != nil {
		return err
	}
	if err := status.Err(); err != nil {
		return err
	}
	it, err := job.Read(*ctx)
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		var rowString []string
		for _, column := range row {
			rowString = append(rowString, fmt.Sprintf("%v", column))
		}
		*resultSet = append(*resultSet, rowString)
	}
	return nil
}

// CreateBigqueryArchivalTables Creates tables required in Bigquery.
// To be called at the time of onboarding new project to Bigquery.
func CreateBigqueryArchivalTables(projectID int64) error {
	bigquerySetting, status := store.GetStore().GetBigquerySettingByProjectID(projectID)
	if status == http.StatusInternalServerError {
		return fmt.Errorf("Failed to get bigquery setting for project_id %d", projectID)
	} else if status == http.StatusNotFound {
		return fmt.Errorf("No BigQuery configuration found for project id %d in database", projectID)
	}

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, bigquerySetting.BigqueryProjectID,
		option.WithCredentialsJSON([]byte(bigquerySetting.BigqueryCredentialsJSON)))
	if err != nil {
		log.WithError(err).Error("Failed to create bigquery client")
		return err
	}
	defer client.Close()

	// Bigquery project id and dataset must already be present.
	for tableName, tableSchemaObject := range bigqueryArchivalTables {
		eventsSchema, err := bigquery.InferSchema(tableSchemaObject)
		if err != nil {
			log.WithError(err).Error("Failed to infer schema for table", tableName)
			return err
		}
		tableMetadata := &bigquery.TableMetadata{
			Schema:      eventsSchema,
			Description: "Table generated from Factors.Ai",
			TimePartitioning: &bigquery.TimePartitioning{
				Field: tableSchemaObject.GetEventTimestampColumnName(),
			},
		}

		log.Infof("Creating table %s", tableName)
		t := client.Dataset(bigquerySetting.BigqueryDatasetName).Table(tableName)
		if err = t.Create(ctx, tableMetadata); err != nil {
			log.WithError(err).Errorf("Failed to create table %s. Table might be present already", tableName)
			continue
		}
	}
	return nil
}

// UploadFileToBigQuery Uploads a given file in cloudManager to specified Bigquery table.
func UploadFileToBigQuery(ctx context.Context, client *bigquery.Client, archiveFile string, bigquerySetting *model.BigquerySetting,
	tableName string, pbLog *log.Entry, cloudManager *filestore.FileManager) (*bigquery.JobStatus, error) {

	pbLog.Infof("Uploading file %s", archiveFile)
	filePath, fileName := getFilePathAndName(archiveFile)
	fileReader, err := (*cloudManager).Get(filePath, fileName)
	if err != nil {
		pbLog.WithError(err).Errorf("Failed to get file %s", archiveFile)
		return nil, err
	}

	pbLog.Infof("Creating loader.")
	sourceReader := bigquery.NewReaderSource(fileReader)
	sourceReader.SourceFormat = bigquery.JSON
	loader := client.Dataset(bigquerySetting.BigqueryDatasetName).Table(tableName).LoaderFrom(sourceReader)
	loader.CreateDisposition = bigquery.CreateNever
	job, err := loader.Run(ctx)
	if err != nil {
		pbLog.WithError(err).Error("Failed to load.")
		return nil, err
	}

	pbLog.Infof("Waiting for the upload job to finish.")
	status, err := job.Wait(ctx)
	if err != nil {
		pbLog.WithError(err).Error("Failed to wait.")
		return nil, err
	}
	if status.Err() != nil {
		pbLog.WithError(status.Err()).Error("Status error.")
		return status, status.Err()
	}

	return status, nil
}

func getFilePathAndName(fullPath string) (string, string) {
	fullPathSplit := strings.Split(fullPath, "/")
	filePath := strings.Join(fullPathSplit[:len(fullPathSplit)-1], "/")
	fileName := fullPathSplit[len(fullPathSplit)-1]
	return filePath, fileName
}
