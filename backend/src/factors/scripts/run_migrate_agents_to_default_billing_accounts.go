package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/util"
	"flag"
	"fmt"

	"github.com/jinzhu/gorm"

	log "github.com/sirupsen/logrus"
)

// go run run_migrate_agents_to_default_billing_accounts.go --env=development --db_host=localhost --db_port=5432 --db_user=autometa --db_name=autometa --db_pass=@ut0me7a
func main() {
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#MigrateAgents", *env)

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}

	C.InitConf(config.Env)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	agents, err := getAllAgents(db)
	if err != nil {
		panic(err)
	}
	log.Printf("No Of Agents: %+v", len(agents))

	failedAgentUUIDs := make([]string, 0, 0)

	for _, agent := range agents {
		err := createDefaultBillingAccount(db, agent)
		if err != nil {
			failedAgentUUIDs = append(failedAgentUUIDs, agent.UUID)
		}
	}

	if len(failedAgentUUIDs) == 0 {
		log.Println("Success")
	} else {
		log.Printf("Failed to create billing account for following agents:\n %+v\n", failedAgentUUIDs)
	}
}

func createDefaultBillingAccount(db *gorm.DB, agent model.Agent) error {
	billingAcc := &model.BillingAccount{PlanID: model.FreePlanID, AgentUUID: agent.UUID}
	return db.Create(billingAcc).Error
}

func getAllAgents(db *gorm.DB) ([]model.Agent, error) {
	agents := make([]model.Agent, 0, 0)
	err := db.Find(&agents).Error
	if err != nil {
		return agents, err
	}
	return agents, nil
}
