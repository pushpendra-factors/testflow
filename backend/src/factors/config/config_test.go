package config

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectsFromListWithAllProjectSupport(t *testing.T) {
	type args struct {
		projectIdsList           string
		disallowedProjectIdsList string
	}
	tests := []struct {
		name              string
		args              args
		wantAllProjects   bool
		wantAllowedMap    map[uint64]bool
		wantDisallowedMap map[uint64]bool
	}{
		{"test1", args{"*", ""},
			true, map[uint64]bool{}, map[uint64]bool{}},
		{"test2", args{"*", "2,3"},
			true, map[uint64]bool{}, map[uint64]bool{2: true, 3: true}},
		{"test3", args{"1,2,3", ""},
			false, map[uint64]bool{1: true, 2: true, 3: true}, map[uint64]bool{}},
		{"test4", args{"", "1,2,3"},
			false, map[uint64]bool{}, map[uint64]bool{1: true, 2: true, 3: true}},
		{"test5", args{"4,5,6", "1,2,3"},
			false, map[uint64]bool{4: true, 5: true, 6: true}, map[uint64]bool{1: true, 2: true, 3: true}},
		//Prioritizing the skip list over project list!
		{"test6", args{"1,2,3,4,5,6", "1,2,3"},
			false, map[uint64]bool{4: true, 5: true, 6: true}, map[uint64]bool{1: true, 2: true, 3: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllProjects, gotAllowedMap, gotDisallowedMap := GetProjectsFromListWithAllProjectSupport(tt.args.projectIdsList, tt.args.disallowedProjectIdsList)
			if gotAllProjects != tt.wantAllProjects {
				t.Errorf("GetProjectsFromListWithAllProjectSupport() gotAllProjects = %v, want %v", gotAllProjects, tt.wantAllProjects)
			}
			if !reflect.DeepEqual(gotAllowedMap, tt.wantAllowedMap) {
				t.Errorf("GetProjectsFromListWithAllProjectSupport() gotAllowedMap = %v, want %v", gotAllowedMap, tt.wantAllowedMap)
			}
			if !reflect.DeepEqual(gotDisallowedMap, tt.wantDisallowedMap) {
				t.Errorf("GetProjectsFromListWithAllProjectSupport() gotDisallowedMap = %v, want %v", gotDisallowedMap, tt.wantDisallowedMap)
			}
		})
	}
}

func TestJSSDKSize(t *testing.T) {
	res, err := http.Get(SDKAssetsURL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	sdkBody, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.True(t, len(sdkBody) > 20000)
	assert.Equal(t, string(sdkBody[0:12]), "var factors=")

	// Non existent URL. StatusCode is still 200 as it returns empty page.
	res, err = http.Get("https://app.factors.ai/assets/non_existent.js")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	sdkBody, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.True(t, len(sdkBody) < 20000)
}
