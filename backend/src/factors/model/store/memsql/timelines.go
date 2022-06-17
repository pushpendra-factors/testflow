package memsql

import (
	C "factors/config"
	"factors/model/model"
	"factors/util"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
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
	type MinMaxTime struct {
		MinUpdatedAt time.Time `json:"min_updated_at"`
		MaxUpdatedAt time.Time `json:"max_updated_at"`
	}
	var minMax MinMaxTime
	err := db.Raw(`SELECT MIN(updated_at) AS min_updated_at, 
		MAX(updated_at) AS max_updated_at 
		FROM 
		(
		SELECT updated_at 
		FROM users 
		WHERE project_id = ? 
		AND (is_group_user=0 OR is_group_user IS NULL)
		AND updated_at < ? LIMIT 1000
		)`, projectId, gorm.NowFunc()).
		Scan(&minMax).Error
	if err != nil {
		log.Error(err)
		return nil, http.StatusInternalServerError
	}

	var profileUsers []model.Contact
	err = db.Table("users").
		Select(`COALESCE(customer_user_id, id) AS identity,
		ISNULL(customer_user_id) AS is_anonymous,
		JSON_EXTRACT_STRING(properties, ?) AS country,
		MAX(updated_at) AS last_activity`, util.UP_COUNTRY).
		Where(`project_id = ? 
		AND (is_group_user=0 OR is_group_user IS NULL) 
		AND updated_at BETWEEN ? AND ?`, projectId, minMax.MinUpdatedAt, minMax.MaxUpdatedAt).
		Group("identity").
		Order("last_activity DESC").
		Limit(1000).
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

	userId := "customer_user_id"
	if isAnonymous == "true" {
		userId = "id"
	}

	db := C.GetServices().Db
	var uniqueUser model.ContactDetails
	err := db.Table("users").Select(`COALESCE(customer_user_id,id) AS user_id,
		ISNULL(customer_user_id) AS is_anonymous,
		JSON_EXTRACT_STRING(properties, ?) AS name,
		JSON_EXTRACT_STRING(properties, ?) AS company,
		JSON_EXTRACT_STRING(properties, ?) AS email,
		JSON_EXTRACT_STRING(properties, ?) AS country,
		JSON_EXTRACT_STRING(properties, ?) AS web_sessions_count,
		JSON_EXTRACT_STRING(properties, ?) AS time_spent_on_site,
		JSON_EXTRACT_STRING(properties, ?) AS number_of_page_views,
		GROUP_CONCAT(group_1_id) IS NOT NULL AS group_1,
		GROUP_CONCAT(group_2_id) IS NOT NULL AS group_2,
		GROUP_CONCAT(group_3_id) IS NOT NULL AS group_3,
		GROUP_CONCAT(group_4_id) IS NOT NULL AS group_4`,
		util.UP_NAME, util.UP_COMPANY, util.UP_EMAIL, util.UP_COUNTRY, util.UP_SESSION_COUNT, util.UP_TOTAL_SPENT_TIME, util.UP_PAGE_COUNT).
		Where("project_id=? AND "+userId+"=?", projectID, identity).
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

	str := []string{`SELECT event_names.name AS event_name, events.timestamp AS timestamp 
		FROM (SELECT project_id, user_id, event_name_id, timestamp FROM events 
		WHERE project_id=? 
		AND timestamp < ? 
		LIMIT 5000) AS events
		LEFT JOIN event_names 
		ON events.event_name_id=event_names.id 
		WHERE events.project_id=? 
		AND user_id 
		IN (
		SELECT id FROM users 
		WHERE project_id=? 
		AND`, userId, `= ?
		) ORDER BY timestamp DESC;`}
	eventsQuery := strings.Join(str, " ")
	rows, err := db.Raw(eventsQuery, projectID, gorm.NowFunc().Unix(), projectID, projectID, identity).Rows()
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
