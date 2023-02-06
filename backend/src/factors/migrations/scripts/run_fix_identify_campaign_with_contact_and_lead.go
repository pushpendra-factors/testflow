package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/sdk"
	"factors/util"
	"flag"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	projectID := flag.Int64("project_id", 0, "Project Id.")
	from := flag.Int64("start_timestamp", 0, "Staring timestamp from events search.")
	to := flag.Int64("end_timestamp", 0, "Ending timestamp from events search. End timestamp will be excluded")
	wetRun := flag.Bool("wet", false, "Wet run")
	flag.Parse()
	defer util.NotifyOnPanic("Task#run_fix_identify_campaign_members_with_contact_and_lead.", *env)

	taskID := "run_fix_identify_campaign_members_with_contact_and_lead"
	if *projectID == 0 {
		log.Error("projectId not provided")
		os.Exit(1)
	}

	if *from <= 0 || *to <= 0 {
		log.Panic("Invalid range.")
	}

	// verify timezone file exist
	if _, err := os.Stat("/usr/local/go/lib/time/zoneinfo.zip"); os.IsNotExist(err) {
		log.Panic("missing timezone info file.")
	}

	config := &C.Configuration{
		AppName: taskID,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		SentryDSN:           *sentryDSN,
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	if !*wetRun {
		log.Info("Running in dry run")
	} else {
		log.Info("Running in wet run")
	}

	status := identifyCampaignMembersWithContactAndLead(*projectID, *wetRun, *from, *to)
	if status != http.StatusOK {
		log.WithFields(log.Fields{"wet_run": *wetRun}).WithError(err).Error("Failed to identify campaign_members with contact and lead.")
		os.Exit(1)
	}
}

func identifyCampaignMembersWithContactAndLead(projectID int64, wetRun bool, from int64, to int64) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	campaignMemberDocuments, _ := store.GetStore().GetSalesforceDocumentsByTypeAndAction(projectID, model.SalesforceDocumentTypeCampaignMember, model.SalesforceDocumentCreated, from, to)

	campaignMembers := make(map[string]*model.User)
	for i := range campaignMemberDocuments {
		user, status := store.GetStore().GetUser(projectID, campaignMemberDocuments[i].UserID)
		if status != http.StatusFound {
			logCtx.WithFields(log.Fields{"user_id": campaignMemberDocuments[i].UserID}).Error("Failed to get user from user ID.")
			continue
		}
		campaignMembers[campaignMemberDocuments[i].ID] = user
	}

	toBeUpdated := make(map[string]string)

	for campaignMemberID := range campaignMembers {
		userProperties, err := util.DecodePostgresJsonb(&campaignMembers[campaignMemberID].Properties)
		if err != nil {
			logCtx.WithFields(log.Fields{"user_id": campaignMembers[campaignMemberID].ID}).Error("Failed to decode user properties.")
			continue
		}

		if (*userProperties)["$salesforce_lead_id"] == nil && (*userProperties)["$salesforce_contact_id"] == nil {
			toBeUpdated[campaignMemberID] = campaignMembers[campaignMemberID].ID
		}
	}

	for campaignMemberID, _ := range toBeUpdated {
		latestDoc, errCode := store.GetStore().GetLatestSalesforceDocumentByID(projectID, []string{campaignMemberID}, model.SalesforceDocumentTypeCampaignMember, time.Now().Unix())
		if errCode != http.StatusFound {
			logCtx.WithFields(log.Fields{"campaign_member_id": campaignMemberID}).Error("Failed to get latest salesforce document by user id.")
			continue
		}

		doc := latestDoc[len(latestDoc)-1]
		value, err := util.DecodePostgresJsonb(doc.Value)
		if err != nil {
			logCtx.WithFields(log.Fields{"document": doc}).Error("Failed to decode user properties as properties map.")
			continue
		}

		customerUserID := ""

		if (*value)["ContactId"] != nil {
			contactID, err := util.GetValueAsString((*value)["ContactID"])
			if err != nil {
				logCtx.WithFields(log.Fields{"ContactID": (*value)["ContactId"]}).WithError(err).Error("Failed to get latest contact document by user id.")
				continue
			}

			latestContactDocs, errCode := store.GetStore().GetLatestSalesforceDocumentByID(projectID, []string{contactID}, model.SalesforceDocumentTypeContact, time.Now().Unix())
			if errCode != http.StatusFound {
				logCtx.WithFields(log.Fields{"campaign_member_id": campaignMemberID, "contact_id": contactID}).Error("Failed to get latest contact document by user id.")
				continue
			}

			latestContact := latestContactDocs[len(latestContactDocs)-1]
			user, status := store.GetStore().GetUser(projectID, latestContact.UserID)
			if status != http.StatusFound {
				logCtx.WithFields(log.Fields{"latest_contact_userId": latestContact.UserID}).Error("Failed to get user from contact user id.")
				continue
			}

			customerUserID = user.CustomerUserId
			if wetRun == true {
				if user.CustomerUserId != "" {
					status, _ := sdk.Identify(projectID, &sdk.IdentifyPayload{
						UserId: user.ID, CustomerUserId: user.CustomerUserId, RequestSource: model.UserSourceSalesforce}, false)
					if status != http.StatusOK {
						logCtx.WithField("customer_user_id", user.CustomerUserId).Error(
							"Failed to identify user on salesforce.")
						return http.StatusInternalServerError
					}
				} else {
					logCtx.Warning("Skipped user identification on salesforce. No customer_user_id on properties.")
				}
			}
		} else if (*value)["LeadId"] != nil {
			leadID, err := util.GetValueAsString((*value)["LeadID"])
			if err != nil {
				logCtx.WithFields(log.Fields{"LeadID": (*value)["LeadId"]}).WithError(err).Error("Failed to get latest lead document by user id.")
				continue
			}

			latestLeadDocs, errCode := store.GetStore().GetLatestSalesforceDocumentByID(projectID, []string{leadID}, model.SalesforceDocumentTypeLead, time.Now().Unix())
			if errCode != http.StatusFound {
				logCtx.WithFields(log.Fields{"campaign_member_id": campaignMemberID, "lead_id": leadID}).Error("Failed to get latest lead document by user id.")
				continue
			}

			latestLead := latestLeadDocs[len(latestLeadDocs)-1]
			user, status := store.GetStore().GetUser(projectID, latestLead.UserID)
			if status != http.StatusFound {
				logCtx.WithFields(log.Fields{"latest_lead_userId": latestLead.UserID}).Error("Failed to get user from lead user id.")
				continue
			}

			customerUserID = user.CustomerUserId
			if wetRun == true {
				if user.CustomerUserId != "" {
					status, _ := sdk.Identify(projectID, &sdk.IdentifyPayload{
						UserId: user.ID, CustomerUserId: user.CustomerUserId, RequestSource: model.UserSourceSalesforce}, false)
					if status != http.StatusOK {
						logCtx.WithField("customer_user_id", user.CustomerUserId).Error(
							"Failed to identify user on salesforce.")
						return http.StatusInternalServerError
					}
				} else {
					logCtx.Warning("Skipped user identification on salesforce. No customer_user_id on properties.")
				}
			}
		} else {
			logCtx.WithField("campaign_member_id", campaignMemberID).Error("No contact or lead document found.")
		}

		if wetRun == false {
			log.Infof("Campaign member id : %s, Customer User ID : ", campaignMemberID, customerUserID)
		}
	}
	return http.StatusOK
}
