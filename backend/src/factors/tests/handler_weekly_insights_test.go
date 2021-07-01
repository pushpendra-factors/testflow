package tests

import (
	"factors/delta"
	H "factors/handler"
	v1 "factors/handler/v1"
	"testing"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)
type WeeklyInsightsParams struct{
	ProjectID		uint64		`json:"project_id"`	
	DashBoardUnitID	uint64 		`json:"dashboard_unit_id"`
	BaseStartTime	time.Time 	`json:"base_start_time"`
	CompStartTime	time.Time   `json:"comp_start_time"`
	InsightsType 	string		`json:"insights_type"`
	NumberOfRecords	int			`json:"number_of_records"`

}
func TestWeeklyInsights(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	var data v1.WeeklyInsightsParams
	data.ProjectID = 1
	data.QueryId = 1
	data.BaseStartTime = time.Unix(1625029668,0)
	data.CompStartTime = time.Now()
	data.InsightsType = "w"
	data.NumberOfRecords = 1
	response,err:= delta.GetWeeklyInsights(data.ProjectID,data.QueryId,&data.BaseStartTime,&data.CompStartTime,data.InsightsType,data.NumberOfRecords)
	assert.NotNil(t,response)
	assert.Nil(t,err)

	outputObj := response.(delta.WeeklyInsights)
	assert.Equal(t,outputObj.Base.W1,uint64(100))
	assert.Equal(t,outputObj.Base.W2,uint64(150))
	assert.Equal(t,outputObj.Base.IsIncreased,true)
	assert.Equal(t,outputObj.Base.Percentage,float64(50))
	assert.Equal(t,outputObj.Goal.W1,uint64(123))
	assert.Equal(t,outputObj.Goal.W2,uint64(246))
	assert.Equal(t,outputObj.Goal.IsIncreased,true)
	assert.Equal(t,outputObj.Goal.Percentage,float64(100))
	assert.Equal(t,outputObj.Conv.W1,uint64(1))
	assert.Equal(t,outputObj.Conv.W2,uint64(2))
	assert.Equal(t,outputObj.Conv.IsIncreased,true)
	assert.Equal(t,outputObj.Conv.Percentage,float64(100))

	assert.Equal(t,outputObj.Insights[0].Key,"browser")
	assert.Equal(t,outputObj.Insights[0].Value,"chrome")
	assert.Equal(t,outputObj.Insights[0].ActualValues.W1,uint64(1))
	assert.Equal(t,outputObj.Insights[0].ActualValues.W2,uint64(2))
	assert.Equal(t,outputObj.Insights[0].ActualValues.IsIncreased,true)
	assert.Equal(t,outputObj.Insights[0].ActualValues.Percentage,float64(100))
	assert.Equal(t,outputObj.Insights[0].ChangeInConversion.W1,uint64(1))
	assert.Equal(t,outputObj.Insights[0].ChangeInConversion.W2,uint64(2))
	assert.Equal(t,outputObj.Insights[0].ChangeInConversion.IsIncreased,true)
	assert.Equal(t,outputObj.Insights[0].ChangeInConversion.Percentage,float64(100))
	assert.Equal(t,outputObj.Insights[0].ChangeInPrevalance.W1,uint64(1))
	assert.Equal(t,outputObj.Insights[0].ChangeInPrevalance.W2,uint64(1))
	assert.Equal(t,outputObj.Insights[0].ChangeInPrevalance.IsIncreased,false)	
	assert.Equal(t,outputObj.Insights[0].ChangeInPrevalance.Percentage,float64(0))
	
	assert.Equal(t,outputObj.Insights[1].Key,"browser")
	assert.Equal(t,outputObj.Insights[1].Value,"chrome")
	assert.Equal(t,outputObj.Insights[1].ActualValues.W1,uint64(1))
	assert.Equal(t,outputObj.Insights[1].ActualValues.W2,uint64(2))
	assert.Equal(t,outputObj.Insights[1].ActualValues.IsIncreased,true)
	assert.Equal(t,outputObj.Insights[1].ActualValues.Percentage,float64(100))
	

}

