package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetProfileUsersListByProjectId(projectId uint64) ([]model.Contact, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectId == 0 {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db
	var profileUsers []model.Contact
	err := db.Table("users").Select(`COALESCE(customer_user_id, id) AS identity,
		ISNULL(customer_user_id) AS is_anonymous,
		JSON_EXTRACT_STRING(properties, '$name') AS name,
		((CASE WHEN GROUP_CONCAT(group_1_id) IS NOT NULL THEN 1 ELSE 0 END)
		+ (CASE WHEN GROUP_CONCAT(group_2_id) IS NOT NULL THEN 1 ELSE 0 END) 
		+ (CASE WHEN GROUP_CONCAT(group_3_id) IS NOT NULL THEN 1 ELSE 0 END) 
		+ (CASE WHEN GROUP_CONCAT(group_4_id) IS NOT NULL THEN 1 ELSE 0 END)) AS groups,
		MAX(updated_at) AS last_activity`).
		Where("project_id = ? AND name IS NOT NULL AND updated_at >= DATE(NOW() - INTERVAL 240 HOUR)", projectId).
		Group("identity").
		Order("last_activity DESC").
		Limit(30).
		Find(&profileUsers).Error
	if err != nil {
		log.Error(err)
		return nil, http.StatusInternalServerError
	}

	return profileUsers, http.StatusFound
}
func (store *MemSQL) GetProfileUserDetailsByID(projectID uint64, identity string, isAnonymous string) (*model.ContactDetails, int) {
	logFields := log.Fields{
		"project_id":   projectID,
		"id":           identity,
		"is_anonymous": isAnonymous,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return nil, http.StatusBadRequest
	}
	if identity == "" {
		log.Error("Invalid user ID.")
		return nil, http.StatusBadRequest
	}
	if isAnonymous == "" {
		log.Error("Invalid user type.")
		return nil, http.StatusBadRequest
	}
	// log.Fatal("identity:::", identity)

	userId := "customer_user_id"
	if isAnonymous == "true" {
		userId = "id"
	}

	db := C.GetServices().Db
	var uniqueUser model.ContactDetails
	err := db.Table("users").Select(`COALESCE(customer_user_id,id) AS user_id,
			JSON_EXTRACT_STRING(properties, '$name') AS name,
			JSON_EXTRACT_STRING(properties, '$company') AS company,
			JSON_EXTRACT_STRING(properties, '$role') AS role,
			JSON_EXTRACT_STRING(properties, '$email') AS email,
			JSON_EXTRACT_STRING(properties, '$country') AS country,
			JSON_EXTRACT_STRING(properties, '$session_count') AS web_sessions_count,
			GROUP_CONCAT(group_1_id) IS NOT NULL AS group_1,
			GROUP_CONCAT(group_2_id) IS NOT NULL AS group_2,
			GROUP_CONCAT(group_3_id) IS NOT NULL AS group_3,
			GROUP_CONCAT(group_4_id) IS NOT NULL AS group_4`).
		Where("project_id=? AND name IS NOT NULL AND "+userId+"=?", projectID, identity).
		Group("user_id").
		Order("updated_at desc").
		Limit(1).
		Find(&uniqueUser).Error
	if err != nil {
		log.Error(err)
		return nil, http.StatusInternalServerError
	}

	groups, errCode := store.GetGroups(projectID)
	if errCode != http.StatusFound {
		log.Error(errCode)
		return nil, http.StatusNotFound
	}

	groupsMap := make(map[int]string)
	for _, value := range groups {
		groupsMap[value.ID] = value.Name
	}

	if uniqueUser.Group1 {
		uniqueUser.GroupInfos = append(uniqueUser.GroupInfos, model.GroupsInfo{GroupName: groupsMap[1]})
	}
	if uniqueUser.Group2 {
		uniqueUser.GroupInfos = append(uniqueUser.GroupInfos, model.GroupsInfo{GroupName: groupsMap[2]})
	}
	if uniqueUser.Group3 {
		uniqueUser.GroupInfos = append(uniqueUser.GroupInfos, model.GroupsInfo{GroupName: groupsMap[3]})
	}
	if uniqueUser.Group4 {
		uniqueUser.GroupInfos = append(uniqueUser.GroupInfos, model.GroupsInfo{GroupName: groupsMap[4]})
	}

	str := []string{`SELECT event_names.name as event_name, events.timestamp as timestamp 
			FROM events 
			LEFT JOIN event_names 
			ON events.event_name_id=event_names.id 
			WHERE events.project_id=? 
			AND user_id 
			IN (SELECT id FROM users 
				WHERE project_id=? 
				AND`, userId, `= ?) 
				AND timestamp >= DATE(NOW() - INTERVAL 1 YEAR)
			ORDER BY timestamp DESC LIMIT 10000;`}
	eventsQuery := strings.Join(str, " ")
	rows, err := db.Raw(eventsQuery, projectID, projectID, identity).Rows()
	if err != nil {
		log.WithError(err).Fatal("Failed to get events")
	}

	for rows.Next() {
		var contactActivity model.ContactActivity
		if err := db.ScanRows(rows, &contactActivity); err != nil {
			log.WithError(err).Fatal("Failed scanning events list")
		}
		uniqueUser.UserActivity = append(uniqueUser.UserActivity, contactActivity)
	}

	return &uniqueUser, http.StatusFound
}
