package main

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	projectIds := flag.String("project_ids", "", "Project Ids to runt he transformation for")
	appName := "free_plan_migration"

	flag.Parse()
	config := &C.Configuration{
		Env: *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}
	var ProjectIds []int64
	var successFullProjects []int64
	var failedProjects []int64

	if *projectIds == "*" {
		ProjectIds, _, _, err = store.GetStore().GetAllProjectIdsUsingPlanId(model.PLAN_ID_CUSTOM)
		if err != nil {
			log.Error(err)
			return
		}
	} else {
		ProjectIds = C.GetTokensFromStringListAsUint64(*projectIds)
	}

	for _, projectID := range ProjectIds {
		// validate if its mapped to custom plan

		planMapping, err := store.GetStore().GetProjectPlanMappingforProject(projectID)
		if err != nil {
			log.WithField("project_id", projectID).WithError(err).Error("Failed to get plan mapping")
			failedProjects = append(failedProjects, projectID)
			continue
		}

		if planMapping.PlanID != model.PLAN_ID_CUSTOM {
			log.WithField("project_id", projectID).Info("Project not a part of custom plan, skipping")
			continue
		}

		var oldOverWrite model.OverWriteOld

		if planMapping.OverWrite != nil {
			err = U.DecodePostgresJsonbToStructType(planMapping.OverWrite, &oldOverWrite)
			if err != nil {
				log.WithField("project_id", projectID).WithError(err).Error("Failed to decode over_writes")
				failedProjects = append(failedProjects, projectID)
				continue
			}
		}

		transformedOw := TransformFeatureListArrayToFeatureListMap(oldOverWrite)
		errMsg, err := store.GetStore().UpdateAddonsForProject(projectID, transformedOw)

		if err != nil {
			log.WithError(err).Error(errors.New(errMsg))
			failedProjects = append(failedProjects, projectID)
			continue
		}

		successFullProjects = append(successFullProjects, projectID)

	}

	log.Info("Failed to transform over_writes for projects ", failedProjects)
	log.Info("Succesfully transformed over_writes for projects ", successFullProjects)
}

func TransformFeatureListArrayToFeatureListMap(oldList model.OverWriteOld) model.OverWrite {
	transformedList := make(map[string]model.FeatureDetails)

	for _, ow := range oldList {
		transformedList[ow.Name] = model.FeatureDetails{
			Limit:            ow.Limit,
			Granularity:      ow.Granularity,
			IsEnabledFeature: ow.IsEnabledFeature,
			Expiry:           ow.Expiry,
		}
	}
	return transformedList
}
