package main

import (
	"bufio"
	"flag"
	"fmt"
	"strings"

	"os"

	U "datasets/util"

	log "github.com/sirupsen/logrus"
)

/*
{
  "na": "loantap.in/check-my-rate/home/emi-free/",
  "co": 1,
  "pr": {
    "$ip": "27.59.138.140",
    "$rawURL": "https://loantap.in/check-my-rate/home/emi-free/?utm_campaign=smart_display\u0026utm_medium=cpc\u0026utm_source=google\u0026utm_campaign_id=843823084",
    "$referrer": "https://news-manikarthik-com.cdn.ampproject.org/v/s/news.manikarthik.com/loantap-review-emi-less-loans-india/money/loans/amp/?amp_js_v=0.1\u0026usqp=mq331AQCKAE%3D",
    "$pageTitle": "Home - LoanTap - Lending Platform for Salaried Professionals",
    "$qp_utm_medium": "cpc",
    "$qp_utm_source": "google",
    "$qp_utm_campaign": "smart_display",
    "$qp_utm_campaign_id": 843823084
  },
  "ti": 1561915704,
  "uid": "a150d6e1-c1eb-4faf-a5c0-736a473cdd10",
  "cuid": null,
  "ujt": 1561915644,
  "upr": {
    "$os": "Android",
    "$city": "Hyderabad",
    "$browser": "Chrome",
    "$country": "India",
    "$joinTime": 1561915644,
    "$platform": "web",
    "$osVersion": "6.0.1",
    "$screenWidth": 360,
    "$screenHeight": 640,
    "$browserVersion": "75.0.3770.101"
  }
}
*/

// Masks each line of Json string.
func maskJson(srcJson string) string {
	s1 := strings.Replace(srcJson, "loantap.in", "lend.go", -1)
	s2 := strings.Replace(s1, "loantap", "lending_company", -1)
	s3 := strings.Replace(s2, "LoanTap", "LendingCompany", -1)
	s4 := strings.Replace(s3, "Loantap", "LendingCompany", -1)
	s5 := strings.Replace(s4, "LOANTAP", "LENDING_COMPANY", -1)
	return s5
}

func mask(srcFilePath string, maskedFilePath string) error {
	srcFile, err := os.OpenFile(srcFilePath, os.O_RDONLY, 0444)
	if err != nil {
		log.WithError(err).Error("Failed opening src events file")
		return err
	}
	scanner := bufio.NewScanner(srcFile)

	maskedFile, err := os.Create(maskedFilePath)
	if err != nil {
		log.WithError(err).Error("Failed creating events file : " + maskedFilePath)
		return err
	}
	defer maskedFile.Close()

	// Adjust scanner buffer capacity upto 10MB per line.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Bytes()
		maskedJson := maskJson(string(line))
		_, err := maskedFile.WriteString(fmt.Sprintf("%s\n", maskedJson))
		if err != nil {
			log.WithError(err).Error("Failed writing masked json to file.")
			return err
		}
	}

	err = scanner.Err()

	return err
}

func main() {
	var dir = flag.String("dir", "", "Directory at which src_events file exist and masked_events file to be kept.")
	var dry = flag.Bool("dry", true, "Dry run events masking without ingestion.")
	var apiHost = flag.String("api_host", "http://localhost:8080", "Host for API request.")
	var apiToken = flag.String("api_token", "", "Token for API request.")

	flag.Parse()

	srcFile := U.GetFilePath(*dir, U.SourceEventsFileName)
	maskedFile := U.GetFilePath(*dir, U.MaskedEventsFileName)

	log.Infof("Masking given source events from %s ..", srcFile)

	err := mask(srcFile, maskedFile)
	if err != nil {
		log.WithError(err).Fatal("Failed masking events")
	}

	log.Infof("Masked events and written to file : %s", maskedFile)

	if *dry {
		log.Info("Not ingested. Use --dry=false for ingesting.")
		os.Exit(0)
	}

	if *apiHost == "" || *apiToken == "" {
		log.Fatal("Failed to ingest. api_host or api_token is missing.")
	}

	eventPropertiesRenameMap := map[string]string{
		"$rawURL":    "$page_raw_url",
		"_$rawURL":    "$page_raw_url",
		"$pageRawURL": "$page_raw_url",
		"_$pageRawURL" : "$page_raw_url",
		"_$page_raw_url" : "$page_raw_url",
		"$pageTitle": "$page_title",
		"_$pageTitle": "$page_title",
		"$pageURL": "$page_url",
		"_$pageURL": "$page_url",
		"$pageDomain": "$page_domain",
		"_$pageDomain": "$page_domain",
		"$referrerURL": "$referrer_url",
		"_$referrerURL": "$referrer_url",
		"$referrerDomain": "$referrer_domain",
		"_$referrerDomain": "$referrer_domain",
	}

	userPropertiesRenameMap := map[string]string{
		"$osVersion":      "$os_version",
		"_$osVersion":      "$os_version",
		"$screenWidth":    "$screen_width",
		"_$screenWidth":    "$screen_width",
		"$screenHeight":   "$screen_height",
		"_$screenHeight":   "$screen_height",
		"$browserVersion": "$browser_version",
		"_$browserVersion": "$browser_version",
	}

	excludeEventNamePrefixes := []string{
		"dev-", // exclude event names with 'dev-' prefix.
	}

	// customer_user_id to user_id cache.
	var clientUserIdToUserIdMap map[string]string = make(map[string]string, 0)
	err = U.IngestEventsFromFile(maskedFile, *apiHost, *apiToken, &clientUserIdToUserIdMap, 
		excludeEventNamePrefixes, &eventPropertiesRenameMap, &userPropertiesRenameMap)
	if err != nil {
		log.WithError(err).Fatal("Failed to ingest from file.")
	}
}
