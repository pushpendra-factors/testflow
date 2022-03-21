package model

import (
	"encoding/json"
	"errors"
	"strings"

	// SDK "factors/sdk"
	U "factors/util"
	"time"
)

type LeadgenSettings struct {
	ProjectID      uint64    `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	Source         int       `gorm:"primary_key:true;auto_increment:false" json:"source"`
	SpreadsheetID  string    `json:"spreadsheet_id"`
	SheetName      string    `json:"sheet_name"`
	RowRead        int64     `json:"row_read"`
	SourceProperty string    `json:"source_property"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

var LeadgenToInternalProperty = map[string]string{
	"Form ID":       U.EP_FORM_ID,
	"Form Name":     U.EP_FORM_NAME,
	"Campaign ID":   U.EP_CAMPAIGN_ID,
	"Campaign Name": U.EP_CAMPAIGN,
	"Ad Group ID":   U.EP_ADGROUP_ID,
	"Ad Group Name": U.EP_ADGROUP,
	"Ad ID":         U.EP_AD_ID,
	"Ad Name":       U.EP_AD,
	"Email ID":      U.UP_EMAIL,
	"Mobile Number": U.UP_PHONE,
	"Created Time":  U.EP_TIMESTAMP,
}
var LeadgenCols = []string{"Form ID", "Campaign ID", "Campaign Name", "Ad Group ID", "Ad Group Name", "Ad ID", "Ad Name", "Email ID", "Mobile Number", "Created Time"}

var SourceAliasMapping = map[int]string{
	4: "Facebook",
	5: "Linkedin",
}

var LeadgenSourceAlias = map[string]int{
	"Facebook": 4,
	"Linkedin": 5,
}

type LeadgenDataPayload struct {
	FormID       string `json:"Form ID"`
	FormName     string `json:"Form Name"`
	CampaignID   string `json:"Campaign ID"`
	CampaignName string `json:"Campaign Name"`
	AdGroupID    string `json:"Ad Group ID"`
	AdGroupName  string `json:"Ad Group Name"`
	AdID         string `json:"Ad ID"`
	AdName       string `json:"Ad Name"`
	Email        string `json:"Email ID"`
	MobileNumber string `json:"Mobile Number"`
	CreatedTime  int64  `json:"Created Time"`
}

func TransformAndGenerateTrackPayload(record []interface{}, projectID uint64, source string) (map[string]interface{}, map[string]interface{}, int64, error) {
	if projectID == 0 {
		return nil, nil, 0, errors.New("incorrect project id")
	}
	if source == "" || (source != "Facebook" && source != "Linkedin") {
		return nil, nil, 0, errors.New("incorrect source")
	}
	if len(record) != len(LeadgenCols) {
		return nil, nil, 0, errors.New("incorrect data in records sent")
	}
	var finalLeadgenPayload *LeadgenDataPayload
	leadgenPayload, err := TransformDataArrayToLeadgenDataPayload(record, LeadgenCols)
	if err != nil {
		return nil, nil, 0, err
	}
	finalLeadgenPayload = leadgenPayload

	eventProperties, userProperties, timestamp, err := TransformLeadgenPayloadToPropertiesMap(*finalLeadgenPayload, source)
	return eventProperties, userProperties, timestamp, err
}
func TransformDataArrayToLeadgenDataPayload(record []interface{}, colsList []string) (*LeadgenDataPayload, error) {
	arrayToStruct := make(map[string]interface{})
	if record == nil || len(record) != 10 {
		return nil, errors.New("Empty or Invalid record")
	}
	if record[1] == "" || record[3] == "" || record[5] == "" || record[9] == "" || (record[7] == "" && record[8] == "") {
		return nil, errors.New("Invalid data in row")
	}
	for i := range record {
		if colsList[i] == "Created Time" {
			layout := "2006-01-02 15:04:05"
			str := record[i].(string)
			timestamp, err := time.Parse(layout, str)
			if err != nil {
				return nil, err
			}
			unixTimestamp := timestamp.Unix()
			arrayToStruct[colsList[i]] = unixTimestamp
		} else {
			arrayToStruct[colsList[i]] = record[i]
		}
	}
	jsonbody, err := json.Marshal(arrayToStruct)
	if err != nil {
		return nil, err
	}

	leadgenDataPayload := LeadgenDataPayload{}
	if err := json.Unmarshal(jsonbody, &leadgenDataPayload); err != nil {
		return nil, err
	}
	return &leadgenDataPayload, nil
}

func TransformLeadgenPayloadToPropertiesMap(leadgenDataPayload LeadgenDataPayload, source string) (map[string]interface{}, map[string]interface{}, int64, error) {
	eventProperties := make(U.PropertiesMap, 0)
	userProperties := make(U.PropertiesMap, 0)

	eventProperties[U.EP_FORM_ID] = leadgenDataPayload.FormID
	if source == "Linkedin" {
		campaignArray := strings.Split(leadgenDataPayload.CampaignID, ":")
		if len(campaignArray) != 4 {
			return nil, nil, 0, errors.New("invalid campaign ID")
		}
		eventProperties[U.EP_CAMPAIGN_ID] = campaignArray[3]
	} else {
		eventProperties[U.EP_CAMPAIGN_ID] = leadgenDataPayload.CampaignID
	}
	if strings.Contains(leadgenDataPayload.CampaignName, ":") {
		eventProperties[U.EP_CAMPAIGN] = ""
	} else {
		eventProperties[U.EP_CAMPAIGN] = leadgenDataPayload.CampaignName
	}
	eventProperties[U.EP_ADGROUP_ID] = leadgenDataPayload.AdGroupID
	eventProperties[U.EP_ADGROUP] = leadgenDataPayload.AdGroupName
	eventProperties[U.EP_AD_ID] = leadgenDataPayload.AdID
	eventProperties[U.EP_AD] = leadgenDataPayload.AdName
	eventProperties[U.EP_TYPE] = "Tactic"
	eventProperties[U.EP_CHANNEL] = "Paid Social"
	eventProperties[U.EP_SOURCE] = source + " Leadgen"

	userProperties[U.UP_EMAIL] = leadgenDataPayload.Email
	userProperties[U.UP_PHONE] = leadgenDataPayload.MobileNumber

	return eventProperties, userProperties, leadgenDataPayload.CreatedTime, nil
}
