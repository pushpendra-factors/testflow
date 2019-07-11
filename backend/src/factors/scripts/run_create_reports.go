package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	C "factors/config"
	M "factors/model"

	"factors/task/reports"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func getIds(str, delimiter string) ([]uint64, error) {
	if str == "" {
		return make([]uint64, 0, 0), nil
	}
	tokens := strings.Split(str, delimiter)
	ids := make([]uint64, len(tokens), len(tokens))
	for i, token := range tokens {

		id, err := strconv.ParseUint(strings.TrimSpace(token), 10, 64)
		if err != nil {
			return make([]uint64, 0, 0), err
		}
		ids[i] = id
	}
	return ids, nil
}

// go run run_create_reports.go --env=development --build_for_projects=1,2,3,4  --db_host=localhost --db_port=5432 --db_user=autometa --db_name=autometa --db_pass=@ut0me7a --aws_region=us-east-1 --aws_key=dummy --aws_secret=dummy --mail_reports=false
func main() {

	env := flag.String("env", "development", "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")

	buildForProjects := flag.String("build_for_projects", "", "")
	skipForProjects := flag.String("skip_for_projects", "", "")

	mailReports := flag.Bool("mail_reports", false, "")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	projectsToBuildFor, err := getIds(*buildForProjects, ",")
	if err != nil {
		panic(err)
	}
	projectsToSkipFor, err := getIds(*skipForProjects, ",")
	if err != nil {
		panic(err)
	}

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		EmailSender: "support@factors.ai",
	}

	C.InitConf(config.Env)
	C.InitSenderEmail(config.EmailSender)

	err = C.InitDB(config.DBInfo)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}

	C.InitMailClient(*awsAccessKeyId, *awsSecretAccessKey, *awsRegion)

	db := C.GetServices().Db
	defer db.Close()

	dashboards, errCode := fetchDashboards(db, 10000, 0, projectsToBuildFor, projectsToSkipFor)
	if errCode != http.StatusFound {
		panic("Failed to fetch dashboards")
		return
	}

	reports.BuildReports(*env, db, dashboards, *mailReports)
}

func fetchDashboards(gormDB *gorm.DB, limit, lastSeenID uint64, projectsToBuildFor, projectsToSkipFor []uint64) ([]*M.Dashboard, int) {
	var dashboards []*M.Dashboard

	db := gormDB.Limit(limit).Where("ID > ?", lastSeenID).Order("ID ASC")

	if len(projectsToBuildFor) > 0 {
		db = db.Where("project_id IN (?)", projectsToBuildFor)
	}

	if len(projectsToSkipFor) > 0 {
		db = db.Where("project_id NOT IN (?)", projectsToSkipFor)
	}

	err := db.Find(&dashboards).Error

	if err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(dashboards) == 0 {
		return nil, http.StatusNotFound
	}

	return dashboards, http.StatusFound
}
