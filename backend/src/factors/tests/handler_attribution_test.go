package tests

import (
	"encoding/json"
	"factors/model/model"
	U "factors/util"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"testing"

	"fmt"
)

func TestAttributionDecode(t *testing.T) {

	// adding json attribution rule
	attributionQuery := postgres.Jsonb{RawMessage: json.RawMessage(`{"cl":"attribution","meta":{"metrics_breakdown":true},"query":{"analyze_type":"hs_deals","kpi_query_group":{"cl":"kpi","qG":[{"ca":"profiles","pgUrl":"","dc":"hubspot_deals","me":["KPI group updated 3"],"fil":[],"gBy":[],"fr":1647109800,"to":1647714599,"tz":"Asia/Kolkata"},{"ca":"profiles","pgUrl":"","dc":"hubspot_deals","me":["KPI group updated 3"],"fil":[],"gBy":[],"gbt":"date","fr":1647109800,"to":1647714599,"tz":"Asia/Kolkata"}],"gGBy":[{"gr":"","prNa":"$hubspot_deal_hs_object_id","prDaTy":"numerical","en":"user","objTy":"","gbty":"raw_values"}],"gFil":[]},"cm":["Impressions","Clicks","Spend"],"ce":{"na":"$sf_contact_updated","pr":[]},"attribution_key":"Campaign","attribution_key_f":[],"query_type":"EngagementBased","lbw":30,"tactic_offer_type":"TacticOffer","from":1647714600,"to":1648146599,"attribution_key_dimensions":["channel_name","campaign_name"],"attribution_key_custom_dimensions":[]}}`)}
	var attrQuery model.AttributionQuery
	err := U.DecodePostgresJsonbToStructType(&attributionQuery, &attrQuery)
	fmt.Println(attrQuery)
	assert.Nil(t, err)
}
