package main

import (
	billing "factors/billing/chargebee"
	C "factors/config"
	M "factors/model/model"
	"factors/model/store"
	"flag"
	"net/http"
	"sort"
	"time"

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
	solutionsAgentUUID := flag.String("solutions_agent_uuid", "", "")
	chargebeeApiKey := flag.String("chargebee_api_key", "dummy", "Chargebee api key")
	chargebeeSiteName := flag.String("chargebee_site_name", "dummy", "Chargebee site name")
	dryRun := flag.Bool("dry_run", false, "")
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
		PrimaryDatastore:  *primaryDatastore,
		ChargebeeApiKey:   *chargebeeApiKey,
		ChargebeeSiteName: *chargebeeSiteName,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}

	C.InitChargebeeObject(config.ChargebeeApiKey, config.ChargebeeSiteName)
	db := C.GetServices().Db
	defer db.Close()

	projectIds, _, _, err := store.GetStore().GetAllProjectIdsUsingPlanId(M.PLAN_ID_FREE)
	if err != nil {
		log.WithError(err).Error("Failed to pull free plan project ids")
		return
	}

	for _, projectID := range projectIds {
		// get project agent mappings
		// select billing admin (earliest created agent which is not solutions)// in case only solutions present as agent log the project
		// create the agent as customer on chargebee ( if not already exists )
		// map the customer id to agents (billing customer id)
		// create FREE USD MONTHLY subscription for that agent on chargebee
		// map the subscription id to projects table
		// map the billing account id of the agent to projects table
		// trigger sync should automatically happen (hence updating project plan mappings table )
		// DRY RUN MODE (do not create any customer/subscription, only LOG the billing admin and the project id)

		project, status := store.GetStore().GetProject(projectID)

		if project.EnableBilling {
			log.Info("Billing already enabled for project ", project)
			continue
		}

		pams, status := store.GetStore().GetProjectAgentMappingsByProjectId(projectID)

		if status != http.StatusFound {
			log.Error("Failed to pull project agent mappings for project ", projectID)
			continue
		}

		sort.Slice(pams, func(i, j int) bool {
			return pams[i].CreatedAt.Before(pams[j].CreatedAt)
		})

		for _, pam := range pams {
			if pam.AgentUUID == *solutionsAgentUUID {
				continue
			}

			agent, status := store.GetStore().GetAgentByUUID(pam.AgentUUID)

			if status != http.StatusFound {
				log.Error("Failed to get agent ", pam.AgentUUID, " for project ", projectID)
				break
			}

			bA, status := store.GetStore().GetBillingAccountByAgentUUID(agent.UUID)
			if status != http.StatusFound {
				log.Error("Failed to get billing account for agent ", pam.AgentUUID, " for project ", projectID)
				break
			}

			if agent.BillingCustomerID != "" {
				if *dryRun {
					log.Info("dry run")
					log.Info("agent skipping customer creation", agent.Email, agent.UUID)
				} else {
					customer, _, err := billing.CreateChargebeeCustomer(*agent)
					if err != nil {
						log.Error(err)
						log.Error("Skipping project due to failed customer creation", agent.UUID, projectID)
						break
					}
					// update agent
					agent.BillingCustomerID = customer.Id
					fields := make(map[string]interface{})

					fields["billing_customer_id"] = customer.Id
					fields["updated_at"] = time.Now()

					db = db.Model(&M.Agent{}).Where("uuid = ?", agent.UUID).Updates(fields)

					if db.Error != nil {
						log.WithError(db.Error).Error("UpdateAgent Failed", agent.UUID)
						break
					}
					if db.RowsAffected == 0 {
						log.WithError(db.Error).Error("UpdateAgent Failed No Rows affected", agent.UUID)
					}
				}
			}
			if *dryRun {
				log.Info("dry run enabled")
				log.Info("skipping project subscription creation", projectID, agent.UUID, agent.Email)
			} else {
				subscription, _, err := billing.CreateChargebeeSubscriptionForCustomer(projectID, agent.BillingCustomerID, M.DEFAULT_PLAN_ITEM_PRICE_ID)
				if err != nil {
					log.Error(err)
					log.Error("Skipping project due to failed subscrption creation", agent.UUID, projectID)
					break
				}

				// update project
				project.BillingSubscriptionID = subscription.Id
				project.EnableBilling = true
				project.BillingAccountID = bA.ID

				fields := make(map[string]interface{})
				fields["billing_subscription_id"] = subscription.Id
				fields["enable_billing"] = true
				fields["billing_account_id"] = project.BillingAccountID
				fields["billing_last_synced_at"] = time.Now()

				err = db.Model(&M.Project{}).Where("id = ?", projectID).Update(fields).Error
				if err != nil {
					log.WithError(err).Error(
						"Failed to execute query of update project", projectID)
					break
				}

				log.Info("Project ", projectID, " mapped with agent ", agent.Email, "with billing account ", bA.ID)
			}
		}
	}

}
