package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)



func segmentFolderReq(t *testing.T, r *gin.Engine, method string, request interface{}, projectId int64, agent *model.Agent, folder_type string, folderID int64, segmentID string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	path := fmt.Sprintf("/projects/%d/segment_folders", projectId)

	if(segmentID != ""){
		path = fmt.Sprintf("%s_item/%s", path, segmentID)
	}else if(folderID != 0) {
		path = fmt.Sprintf("%s/%d", path, folderID)
	}

	path = fmt.Sprintf("%s?type=%s", path, folder_type)
	
	t.Logf(path, "PV Path", request)
	rb := C.NewRequestBuilderWithPrefix(method,path)
	if(method == http.MethodPost || method == http.MethodPut){
		rb = rb.WithPostParams(request)
	}
	
	rb = rb.WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create segment req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func CreateTestSegmentFolder(t *testing.T, r *gin.Engine, method string, request interface{}, projectId int64, agent *model.Agent, folder_type string, folderID int64, segmentID string){
	// Create SegmentFolder
	w := segmentFolderReq(t, r, method, request, projectId, agent, folder_type,0, "")
	assert.Equal(t, http.StatusCreated, w.Code)

	// Creating SegmentFolder with name, this must throw an error
	w = segmentFolderReq(t, r, method, request, projectId, agent, folder_type,0, "")
	assert.NotEqual(t, http.StatusCreated, w.Code)
}
func handlerGetAllTestSegmentsFolder(t *testing.T, r *gin.Engine, method string, projectId int64, agent *model.Agent, folder_type string) ([]model.SegmentFolder, int) {
	w  := segmentFolderReq(t, r, method, nil, projectId, agent, folder_type, 0, "")

	if w.Code != http.StatusFound {
		return nil, w.Code
	}

	body := w.Body.Bytes()
	var jsonData []model.SegmentFolder

	err := json.Unmarshal(body, &jsonData)
	if err != nil {
		return nil, http.StatusBadRequest
	}
	return jsonData, w.Code 
}

func TestGetSegmentFolders(t *testing.T){
	var w *httptest.ResponseRecorder

	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	createSegmentPayload := model.SegmentFolderPayload{Name: "Accounts Demo Folder"}

	// 1. Create SegmentFolder
	CreateTestSegmentFolder(t, r, http.MethodPost, createSegmentPayload, project.ID, agent, "account",0, "")
	
	// 2. Get All SegmentsFolder
	jsonData, errCode := handlerGetAllTestSegmentsFolder(t, r, http.MethodGet, project.ID, agent, "account")
	assert.Equal(t, http.StatusFound, errCode)

	// 3. Update SegmentFolder
	createSegmentPayload.Name = "Accounts Demo Updated 1"
	w = segmentFolderReq(t, r, http.MethodPut, createSegmentPayload, project.ID, agent, "account", jsonData[0].Id, "")
	assert.Equal(t, http.StatusAccepted, w.Code)

	//  UPDATE TEST: It should not Accept Name = ""
	createSegmentPayload.Name = ""
	w = segmentFolderReq(t, r, http.MethodPut, createSegmentPayload, project.ID, agent, "account", jsonData[0].Id, "")
	assert.NotEqual(t, http.StatusFound, w.Code)


	// 4. Move segment to Existing Folder
	// First Create Segment and Then Move it
	// Create new segment.
	events := make([]model.QueryEventWithProperties, 1)
	properties := make([]model.QueryProperty, 1)
	prop := model.QueryProperty{
		Value:     "1",
		Property:  "prop1",
		Operator:  "op1",
		LogicalOp: "logicalop1",
		Type:      "type1",
	}
	properties[0] = prop
	event := model.QueryEventWithProperties{
		Name:       "eventName1",
		Properties: properties,
	}
	events[0] = event
	querySegment := model.Query{
		EventsWithProperties: events,
		GlobalUserProperties: properties,
	}

	segment := &model.SegmentPayload{
		Name:        "Name1",
		Description: "dummy info",
		Query:       querySegment,
		Type:        "event",
	}

	w = sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Segment created Now move it
	// Now fetch the segment ID to do further operations
	w = sendAllSegmentGetReq(r, project.ID, agent)
	assert.Equal(t, http.StatusFound, w.Code)
	body := w.Body.Bytes()
	var segmentsData map[string][]model.Segment

	err = json.Unmarshal(body, &segmentsData)

	assert.Nil(t, err)
	
	// 5. Move newly created Segment to prev Folder
	moveFolderPayload := model.MoveSegmentFolderItemPayload{FolderID: jsonData[0].Id}
	w = segmentFolderReq(t, r, http.MethodPut, moveFolderPayload, project.ID, agent, "account", jsonData[0].Id, segmentsData["$domains"][0].Id)
	assert.Equal(t, http.StatusAccepted, w.Code)
	// Adding Segment to Some Random Segment Folder
	moveFolderPayload = model.MoveSegmentFolderItemPayload{FolderID: 9999999}
	w = segmentFolderReq(t, r, http.MethodPut, moveFolderPayload, project.ID, agent, "account", 9999999, segmentsData["$domains"][0].Id)
	assert.NotEqual(t, http.StatusAccepted, w.Code)


	// 6. MoveSegment To New Folder
	createSegmentPayload = model.SegmentFolderPayload{Name: "NewFolder"}
	w = segmentFolderReq(t, r, http.MethodPost, createSegmentPayload, project.ID, agent, "account", 0, segmentsData["$domains"][0].Id)
	assert.Equal(t, http.StatusAccepted, w.Code)

	// MoveSegment to Folder with ""
	createSegmentPayload = model.SegmentFolderPayload{Name: ""}
	w = segmentFolderReq(t, r, http.MethodPost, createSegmentPayload, project.ID, agent, "account", 0, segmentsData["$domains"][0].Id)
	assert.NotEqual(t, http.StatusAccepted, w.Code)


	// 7. Delete SegmentFolder
	w = segmentFolderReq(t, r, http.MethodDelete, createSegmentPayload, project.ID, agent, "account", jsonData[0].Id, "")
	assert.Equal(t, http.StatusAccepted, w.Code)


	// Validating No. of Rows left
	// It should be 1
	jsonData, errCode = handlerGetAllTestSegmentsFolder(t, r, http.MethodGet, project.ID, agent, "account")
	assert.Equal(t, 1, len(jsonData))
}
