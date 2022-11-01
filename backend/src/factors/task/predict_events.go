package task

import (
	"encoding/json"
	"factors/filestore"
	PP "factors/predict"
	serviceDisk "factors/services/disk"
	"fmt"

	log "github.com/sirupsen/logrus"
)

var prLog = log.WithField("prefix", "Task#PredictPullEvents")

const ONE_WEEK_EPOCH = 604800

func PredictPullData(configs map[string]interface{}) error {

	var prjProp PP.PredictProject

	prjProp.Project_id = configs["project_id"].(int64)
	prjProp.Base_event = configs["base_event"].(string)
	prjProp.Target_event = configs["target_event"].(string)
	prjProp.Target_ts = configs["target_time"].(int64)
	prjProp.Start_time = configs["start_time"].(int64)
	prjProp.End_time = configs["target_time"].(int64)
	prjProp.Buffer_time = configs["buffer_time"].(int64)
	prjProp.Target_value = configs["target_prop"].(string)
	prjProp.Model_id = configs["target_time"].(int64)

	prjProp.Buffer_time = prjProp.Buffer_time * ONE_WEEK_EPOCH

	prjByte, err := json.Marshal(prjProp)
	if err != nil {
		return err
	}
	prLog.Infof("running predict :%v", string(prjByte))

	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	cloudManager := configs["cloudManager"].(*filestore.FileManager)
	err = predict_pull_events_dataset(prjProp, cloudManager, diskManager)
	if err != nil {
		return fmt.Errorf("unable to pull events for preduct :%v", err)
	}
	return nil

}

func predict_pull_events_dataset(pr PP.PredictProject, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver) error {

	projectId := pr.Project_id
	model_id := pr.Model_id
	buffer_time := pr.Buffer_time
	projectPath := (diskManager).GetPredictProjectDir(projectId, model_id)
	err := serviceDisk.MkdirAll(projectPath)
	if err != nil {
		return fmt.Errorf("unable to create predict project path:%v", err)
	}
	projectDataPath := (diskManager).GetPredictProjectDataPath(projectId, model_id)
	err = serviceDisk.MkdirAll(projectDataPath)
	if err != nil {
		return fmt.Errorf("unable to create predict project path:%v", err)
	}

	err = PP.MakeTimestamps(&pr)
	if err != nil {
		e := fmt.Errorf("unable to make timestamps:%v", err)
		return e
	}

	prLog.Infof("Running pull events for project :%v", pr)
	base_event_id, err := PP.GetEventNameID(pr.Project_id, pr.Base_event)
	if err != nil {
		return err
	}
	prLog.Infof("event name and id :%s and %s", pr.Base_event, base_event_id)

	target_event_id, err := PP.GetEventNameID(pr.Project_id, pr.Target_event)
	if err != nil {
		return err
	}
	prLog.Infof("event name and id :%s and %s", pr.Target_event, target_event_id)

	userIdsOnTargetEvent, err := PP.GetAllUsersWithGroupsByCohort(projectId, model_id, pr.Target_event, pr.Start_time, pr.End_time, pr.Target_value, diskManager, cloudManager)
	if err != nil {
		prLog.Errorf("unable to fetch user cohort:%v", err)
	}

	start_base_time := pr.Start_time
	end_base_time := pr.End_time
	users_map, group_id_map, err := PP.GetAllBaseEvents(projectId, model_id, base_event_id, start_base_time, end_base_time, diskManager, cloudManager)
	if err != nil {
		e := fmt.Errorf("unable to get users and groupids on base event:%v", err)
		return e
	}

	prLog.Infof("total number of base events captured : %d", len(users_map))

	err = PP.GetGroupCountsOfGroupIds(projectId, model_id, group_id_map, diskManager, cloudManager)
	if err != nil {
		e := fmt.Errorf("unable to get users and groupids on base event:%v", err)
		return e
	}

	err = PP.GetEventsOfUsers(projectId, model_id, users_map, userIdsOnTargetEvent, start_base_time, end_base_time, buffer_time, diskManager, cloudManager)
	if err != nil {
		e := fmt.Errorf("unable to get users and groupids on base event:%v", err)
		return e
	}

	// usersOnEvent, err := PP.GetUsersWithEventAndFirstTimestamp(pr.Project_id, base_event_id, pr.Start_time, pr.End_time)
	// log.Infof("Users who did base events:%v", usersOnEvent)

	// err = PP.GetEventsOfUsers(projectId, model_id, base_event_id, usersOnEvent, diskManager, cloudManager)
	// if err != nil {
	// 	prLog.Errorf("unable to fetch events of users:%v", err)
	// }

	return nil

}
