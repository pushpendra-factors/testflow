package v1

import (
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetFactorsAnalyticsHandler(c *gin.Context) {
	noOfDays := int(7)
	noOfDaysParam := c.Query("days")
	isHtmlRequired := c.Query("html")
	var err error
	if noOfDaysParam != "" {
		noOfDays, err = strconv.Atoi(noOfDaysParam)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	analytics, err := store.GetStore().GetEventUserCountsOfAllProjects(noOfDays)
	if err != nil {
		log.WithError(err).Error("GetEventUserCountsOfAllProjects")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if isHtmlRequired == "true" {
		ReturnReadableHtml(c, analytics)
		return
	}
	c.JSON(http.StatusOK, analytics)
}
func ReturnReadableHtml(c *gin.Context, analytics map[string][]*model.ProjectAnalytics) {
	keys := make([]string, 0, len(analytics))
	for k := range analytics {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	for _, date := range keys {
		heading := CreateHeadingTemplate(date)
		data := analytics[date]
		t := template.New("date")
		t, err := t.Parse(heading)
		if err != nil {
			log.Error(err)
			return
		}
		t.Execute(c.Writer, nil)

		var HtmlStr = ""
		for index, analyticsData := range data {
			if index%20 == 0 {
				HtmlStr = ""
				HtmlStr += CreateTable() //creating table for each 20 rows.
			}
			HtmlStr += InsertDataToTable(*analyticsData)
			if (index+1)%20 == 0 || index == len(data)-1 {
				HtmlStr += "</table><br>" //closing table after every 20th row or before data ends.
				t := template.New("table")
				t, err := t.Parse(HtmlStr)
				if err != nil {
					log.Error(err)
					return
				}
				t.Execute(c.Writer, nil)
			}
		}

	}

}

func CreateHeadingTemplate(date string) string {
	formated_date := date[0:4] + "-" + date[4:6] + "-" + date[6:8] //yyyy-mm-dd
	html := fmt.Sprintf(`<h1> Date: &nbsp; %s</h1>
						`, formated_date)
	return html
}
func CreateTable() string {
	html := `<table border="1px">
				<tr>
					<th>Project ID</th>
					<th>Project Name</th>
					<th>Adwords</th>
					<th>Facebook</th>
					<th>Hubspot</th>
					<th>LinkedIn</th>
					<th>Salesforce</th>
					<th>Total Events</th>
					<th>Total Unique Events</th>
					<th>Total Unique Users</th>
					<th>6Signal Domain Enrichment Count</th>
					<th>6Signal Total API Hits</th>
				</tr>
			` // not closing the table as data is pending to be inserted
	return html
}
func InsertDataToTable(data model.ProjectAnalytics) string {
	html := fmt.Sprintf(`<tr>
		<td width="150px">&nbsp;%s</td>
		<td width="20%%">&nbsp;%s</td>
		<td width="150px" align="center">&nbsp;%s</td>
		<td width="150px" align="center">&nbsp;%s</td>
		<td width="150px" align="center">&nbsp;%s</td>
		<td width="150px" align="center">&nbsp;%s</td>
		<td width="150px" align="center">&nbsp;%s</td>
		<td width="150px" align="center">&nbsp;%s</td>
		<td width="150px" align="center">&nbsp;%s</td>
		<td width="150px" align="center">&nbsp;%s</td>
		<td width="150px" align="center">&nbsp;%s</td>
		<td width="150px" align="center">&nbsp;%s</td>
		</tr>  
		`, strconv.Itoa(int(data.ProjectID)), data.ProjectName, strconv.Itoa(int(data.AdwordsEvents)), strconv.Itoa(int(data.FacebookEvents)), strconv.Itoa(int(data.HubspotEvents)), strconv.Itoa(int(data.LinkedinEvents)), strconv.Itoa(int(data.SalesforceEvents)), strconv.Itoa(int(data.TotalEvents)), strconv.Itoa(int(data.TotalUniqueEvents)), strconv.Itoa(int(data.TotalUniqueUsers)), strconv.Itoa(int(data.SixSignalAPIHits)), strconv.Itoa(int(data.SixSignalAPITotalHits)))
	return html
}
