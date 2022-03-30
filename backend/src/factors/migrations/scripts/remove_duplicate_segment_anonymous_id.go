package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	projectId := flag.Uint64("project_id", 0, "")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	if *projectId == 0 {
		log.Fatal("Invalid project_id.")
	}

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to run migration. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	/*
		NOTICE: DEPRECATED - UPDATING SEGMENT ANONYMOUS ID WITH RANDOM STRING IS NOT RECOMMENDED ANY MORE.

		segmentUsers, err := getProjectSegmentDuplicateUsers(*projectId)
		if err != nil {
			log.WithError(err).Fatal("Failed to getProjectSegmentDuplicateUsers.")
		}
		log.Info("Got project segment duplicate users.")

		for _, segmentUser := range segmentUsers {
			users, err := getSegmentUsersToFix(segmentUser.ProjectId, segmentUser.SegmentAnonymousId)
			if err != nil {
				log.WithError(err).Fatal("Failed to getSegmentUsersToFix")
			}

			log.WithField("project_id", segmentUser.ProjectId).Info("Got segment users to fix.")

			for _, user := range users {
					err := db.Table("users").Where("project_id = ? AND id = ?",
						segmentUser.ProjectId, user.ID).Update("segment_anonymous_id",
						fmt.Sprintf("%s_%s", segmentUser.SegmentAnonymousId, U.RandomLowerAphaNumString(8))).Error
					if err != nil {
						log.WithField("project_id", segmentUser.ProjectId).WithError(err).Fatal(
							"Failed to getSegmentUsersToFix.")
					}
				log.WithField("project_id", segmentUser.ProjectId).Info("Updated user.")
			}
		}

		log.Info("Successfully updated duplicate segment anonymous id.")
	*/
}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func getSegmentUsersToFix(projectId uint64, segAnonId string) ([]model.User, error) {
	user, errCode := store.GetStore().GetUserBySegmentAnonymousId(projectId, segAnonId)
	if errCode != http.StatusFound {
		log.WithField("project_id", projectId).WithField("err_code",
			errCode).Info("Failed to GetUserBySegmentAnonymousId")
		return []model.User{}, errors.New("failed to get segment user")
	}

	users := make([]model.User, 0, 0)
	db := C.GetServices().Db
	err := db.Table("users").Select("id").Where("project_id=? AND segment_anonymous_id = ? AND id != ?",
		projectId, segAnonId, user.ID).Find(&users).Error
	if err != nil {
		log.WithField("project_id", projectId).WithField(
			"segment_anonymous_id", segAnonId).WithError(err).Error(
			"Failed to getSegmentUsersToFix")
		return users, err
	}

	return users, nil
}

type SegmentDuplicateUser struct {
	ProjectId          uint64
	SegmentAnonymousId string
	MinCreatedAt       string // using string to preserve nano seconds.
}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func getProjectSegmentDuplicateUsers(projectId uint64) ([]SegmentDuplicateUser, error) {
	db := C.GetServices().Db
	rows, err := db.Raw("SELECT * FROM (SELECT project_id, segment_anonymous_id, count(*) no_of_users, min(created_at) min_created_at FROM users WHERE project_id = ? AND segment_anonymous_id IS NOT NULL group by project_id, segment_anonymous_id) segment_duplicate_users WHERE no_of_users > 1", projectId).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to execute query on getSegmentDuplicateUsers")
		return nil, err
	}
	defer rows.Close()

	segmentUsers := make([]SegmentDuplicateUser, 0, 0)
	for rows.Next() {
		var segmentUser SegmentDuplicateUser
		if err := db.ScanRows(rows, &segmentUser); err != nil {
			log.WithError(err).Error("Failed scanning rows on getSegmentDuplicateUsers")
			return segmentUsers, err
		}
		segmentUsers = append(segmentUsers, segmentUser)
	}

	return segmentUsers, nil
}
