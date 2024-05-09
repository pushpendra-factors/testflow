package main

import (
	"encoding/json"
	C "factors/config"
	"factors/integration/paragon"
	"factors/model/model"
	"factors/model/store"
	"flag"
	"os"

	U "factors/util"
	"fmt"

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

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

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
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()

	payload := getPayloadDetails(1, "bbfc0369-87f2-45ce-a2f8-dc671ca9fa6d")
	url := `https://zeus.useparagon.com/projects/60ef58ab-2e11-4f44-a388-b81aafca37ad/sdk/triggers/8eee39fe-0f59-4981-937e-5e60726e47fc`
	token := `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIyIiwiZXhwIjo5MjIzMzcyMDM2ODU0Nzc1ODA3LCJpYXQiOjE3MDUwNTQwNTR9.as7Uz5TqQjdwrPqJuHzCA7eOiKkfqP_0-4OSWVOuXimxf4yhB5cqncDTNIYYr3FcUEyeGNYKLtjap5ky_NhkkH2T3dSv6Ac-o2P0ot1UvbFlPvmnee3xDUESxMsSwXckiVsbtksLRebm5yl9lJG5YubR7T0BgkBgDiwSwQTZ56e-dlyFoA08VzpNL69-o1CP71XqkE-2m2lT7cPcBuJfoS54mbH0e_mxNlt33p63oPt2uq7gcXDR073yThNQcP14ZYcqBhAh3HhoAb7Kx6AQeqdGXM9EDXDOdGxGIiWetTDjMxpZfQAXhr2UKvB-xseI7JUsxk5q8h6GhFHCK5aBb9FjJSR7ffWRf4l8ibFxMRP6_nhWjskZomThhn6X34VnbU6Ix_WWXnf0W9LyECxR5z8KtEbdYrJGIPGpsgDfuhe6GCu_ZXZAMrMgJjNjC18Iwh-AgyhEv-F1UKGBgJHo6YtjPDHaJx6zG99OjVEargY2evfJgPceosWQ3FD0G5_Gv-Cz5g3U7jiW1Jd6fOxOccDYn4DOmJvj4Ueb3pOhMOwPaBGeAhUjMeknEwtIip8bnGZoI5jPbpsljht8o6s2McrefpgqKeG53v7yIH87RnYb7epyhLnUXsBYMPkxoKNv6S_GZisH5fBvF3uMU4qd9jWd3Gj3YE_mLfnqThAmX_E`

	response, err := paragon.SendParagonEventRequest(url, token, payload)
	if err != nil {
		log.WithError(err).Error("failed to make request for paragon event")
	}
	fmt.Println(response)

	js, _ := json.Marshal(payload)
	fmt.Println(string(js))

}

func getPayloadDetails(projectID int64, id string) interface{} {

	wf, _, err := store.GetStore().GetWorkflowById(projectID, id)
	if err != nil {
		log.WithError(err).Error("get failed")
		return nil
	}

	userProperties := getUserPropeties()
	var workflow model.WorkflowAlertBody
	err = U.DecodePostgresJsonbToStructType(wf.AlertBody, &workflow)
	if err != nil {
		log.Error("failed to decode struct")
	}

	var messageProperties model.WorkflowMessageProperties
	err = U.DecodePostgresJsonbToStructType(workflow.MessageProperties, &messageProperties)
	if err != nil {
		log.Error("got error in mp")
	}

	ap := messageProperties.AdditionalPropertiesCompany
	fmt.Println(ap)

	apc := messageProperties.AdditionalPropertiesContact
	fmt.Println(apc)

	mp := messageProperties.MandatoryPropertiesCompany
	fmt.Println(mp)

	manPld := make(model.WorkflowPayloadProperties)
	for _, prop := range mp {
		manPld[prop.Others] = userProperties[prop.Factors]
	}

	addPld := make(model.WorkflowPayloadProperties)
	for _, prop := range ap {
		addPld[prop.Others] = userProperties[prop.Factors]
	}

	addCPld := make(model.WorkflowPayloadProperties)
	for _, prop := range apc {
		addCPld[prop.Others] = userProperties[prop.Factors]
	}

	config, err := U.DecodePostgresJsonb(workflow.AdditonalConfigurations)
	payload := model.WorkflowParagonPayload{
		MandatoryPropsCompany:  manPld,
		AdditionalPropsCompany: addPld,
		AdditionalPropsContact: addCPld,
		Configuration:          *config,
	}

	return payload
}

func getUserPropeties() map[string]interface{} {
	userProperties := map[string]interface{}{
		"$6Signal_name":                             "Factors.ai",
		"$6Signal_domain":                           "factors.ai",
		"$6Signal_description":                      "I am batman!",
		"$hubspot_company_employee_range___paragon": "10-49",
		"$hubspot_company_name":                     "test company2",
		"$hubspot_company_domain":                   "testcompany2.ai",
		"$hubspot_deal_hubspot_owner_assigneddate":  "1750500000",
		"$hubspot_deal_notes_last_updated":          "1750500000",
		"$hubspot_company_revenue_range___factors":  "10-20M",
	}

	return userProperties
}
