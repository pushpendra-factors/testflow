package main

import (
	"factors/model/model"
	U "factors/util"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func main() {

	fmt.Println("Hello fren!")
	payload := getPayloadDetails()
	// url := `https://zeus.useparagon.com/projects/60ef58ab-2e11-4f44-a388-b81aafca37ad/sdk/triggers/488ab709-d79b-405a-9827-ba8db84dba24`
	// token := `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIyIiwiZXhwIjo5MjIzMzcyMDM2ODU0Nzc1ODA3LCJpYXQiOjE3MDUwNTQwNTR9.as7Uz5TqQjdwrPqJuHzCA7eOiKkfqP_0-4OSWVOuXimxf4yhB5cqncDTNIYYr3FcUEyeGNYKLtjap5ky_NhkkH2T3dSv6Ac-o2P0ot1UvbFlPvmnee3xDUESxMsSwXckiVsbtksLRebm5yl9lJG5YubR7T0BgkBgDiwSwQTZ56e-dlyFoA08VzpNL69-o1CP71XqkE-2m2lT7cPcBuJfoS54mbH0e_mxNlt33p63oPt2uq7gcXDR073yThNQcP14ZYcqBhAh3HhoAb7Kx6AQeqdGXM9EDXDOdGxGIiWetTDjMxpZfQAXhr2UKvB-xseI7JUsxk5q8h6GhFHCK5aBb9FjJSR7ffWRf4l8ibFxMRP6_nhWjskZomThhn6X34VnbU6Ix_WWXnf0W9LyECxR5z8KtEbdYrJGIPGpsgDfuhe6GCu_ZXZAMrMgJjNjC18Iwh-AgyhEv-F1UKGBgJHo6YtjPDHaJx6zG99OjVEargY2evfJgPceosWQ3FD0G5_Gv-Cz5g3U7jiW1Jd6fOxOccDYn4DOmJvj4Ueb3pOhMOwPaBGeAhUjMeknEwtIip8bnGZoI5jPbpsljht8o6s2McrefpgqKeG53v7yIH87RnYb7epyhLnUXsBYMPkxoKNv6S_GZisH5fBvF3uMU4qd9jWd3Gj3YE_mLfnqThAmX_E`

	// response, err := paragon.SendParagonEventRequest(url, token, payload)
	// if err != nil {
	// 	log.WithError(err).Error("failed to make request for paragon event")
	// }
	fmt.Println(payload)

	// fmt.Println(response)
}

func getPayloadDetails() interface{} {

	js := `
	{
		"action_performed": "action_event",
		"alert_limit": 5,
		"cool_down_time": 1800,
		"event": "$session",
		"event_level": "active",
		"filters": [
			{
				"en": "user",
				"grpn": "user",
				"lop": "AND",
				"op": "equals",
				"pr": "Country",
				"ty": "categorical",
				"va": "OTHERS"
			},
			{
				"en": "user",
				"grpn": "user",
				"lop": "OR",
				"op": "equals",
				"pr": "Country",
				"ty": "categorical",
				"va": "US"
			},
			{
				"en": "user",
				"grpn": "user",
				"lop": "OR",
				"op": "equals",
				"pr": "Country",
				"ty": "categorical",
				"va": "$none"
			}
		],
		"notifications": false,
		"repeat_alerts": true,
		"template_title": "Factors → HubSpot company",
		"template_id": "",
		"title": "Test workflow - 22Apr24",
		"message_properties": {
			"mandatory_properties": [
				{
					"factors": "$hubspot_company_domain",
					"others": "domain"
				},
				{
					"factors": "$hubspot_company_name",
					"others": "name"
				}
			],
			"additional_properties": [
				{
					"factors": "$hubspot_company_description",
					"others": "description"
				}
			],
			"mapping_details": [
				"hubspot"
			]
		}
	}`

	userProperties := getUserPropeties()
	var wf model.WorkflowAlertBody
	err := U.DecodeJSONStringToStructType(js, &wf)
	if err != nil {
		log.Error("failed to decode struct")
	}

	messageProperties, err := U.DecodePostgresJsonb(wf.MessageProperties)
	if err != nil {
		log.Error("got error in mp")
	}

	addProps, ok := (*messageProperties)["additional_properties"].(map[string]interface{})
	if !ok {
		log.Error("got error in addprops")
	}
	var ap, mp []model.WorkflowPropertiesMapping
	err = U.DecodeInterfaceMapToStructType(addProps, &ap)
	if err != nil {
		log.Error("got error in ap")
	}
	mandatoryProps, ok := (*messageProperties)["mandatory_properties"].(map[string]interface{})
	if !ok {
		log.Error("got error in mandprops")
	}
	err = U.DecodeInterfaceMapToStructType(mandatoryProps, &mp)
	if err != nil {
		log.Error("got error in mp")
	}

	manPld := make(model.WorkflowPayloadProperties)
	for _, prop := range mp {
		manPld[prop.Others] = userProperties[prop.Factors]
	}

	addPld := make(model.WorkflowPayloadProperties)
	for _, prop := range ap {
		addPld[prop.Others] = userProperties[prop.Factors]
	}

	payload := model.WorkflowParagonPayload{
		MandatoryPropsCompany:  manPld,
		AdditionalPropsCompany: addPld,
	}

	return payload
}

func getUserPropeties() map[string]interface{} {
	userProperties := map[string]interface{}{
		"$6Signal_name":   "Factors.ai",
		"$6Signal_domain": "factors.ai",
		"$6Signal_description": "I am batman!",
	}

	return userProperties
}
