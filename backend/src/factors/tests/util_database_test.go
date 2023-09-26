package tests

import (
	"errors"
	"net/http"
	"testing"

	"factors/model/model"
	"factors/model/store"
	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestIsPostgresUniqueIndexViolationError(t *testing.T) {
	assert.True(t, U.IsPostgresUniqueIndexViolationError("column_unique_index",
		errors.New("pq: duplicate key value violates unique constraint \"column_unique_index\"")))
}

func TestIsPostgresIntegrityViolationError(t *testing.T) {
	assert.True(t, U.IsPostgresIntegrityViolationError(
		errors.New("pq: duplicate key value violates unique constraint \"column_unique_index\"")))
}

func TestCleanupPostgresJsonStringBytes(t *testing.T) {
	assert.Equal(
		t,
		string(U.CleanupUnsupportedCharOnStringBytes([]byte("🌎💧🍃🌾🏭🔬🚽🚿🇯🇵  Environmental Bioengineering Lab. led by Akihiko Terada and Shohei Riya at Dep. Chem. Eng. in Tokyo Univ. Agri. & Tech., "))),
		"  Environmental Bioengineering Lab. led by Akihiko Terada and Shohei Riya at Dep. Chem. Eng. in Tokyo Univ. Agri. & Tech., ",
	)
}

func TestMemsqlSpecialCharactersCleanup(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	// Unicode characters with value upto U+FFFF is supported by Singlestore
	// ￿ (U+FFFF) - Supported
	// 𐀁 (U+10001) - Not Supported
	CustomerUserId := make(map[string]string)
	CustomerUserId["🌎💧🍃🔬🚽🚿🇯🇵 Sample Test."] = " Sample Test."
	CustomerUserId["विक्रमसिंह7698@gmail.com"] = "विक्रमसिंह7698@gmail.com"
	CustomerUserId["ßƒ©ðœ@gmail.com"] = "ßƒ©ðœ@gmail.com"
	CustomerUserId["+@gmail.com"] = "+@gmail.com"
	CustomerUserId["∆@gmail.com"] = "∆@gmail.com"
	CustomerUserId["Paupulapravallika9@gmail.cok -"] = "Paupulapravallika9@gmail.cok -"
	CustomerUserId["Angélicajitetic@gmail.com"] = "Angélicajitetic@gmail.com"
	CustomerUserId["sandeep.roy123@outlook.com"] = "sandeep.roy123@outlook.com"
	CustomerUserId["παράγοντεςε😁ίναιστηΜπανγκαλό😎ρ@gmail.com"] = "παράγοντεςείναιστηΜπανγκαλόρ@gmail.com"
	CustomerUserId["🀄因素在🈳🈴🈵🈶🈸🈹🈺班加罗尔㊟@gmail.com"] = "因素在班加罗尔㊟@gmail.com"
	CustomerUserId["☹️♈️♥️⚽️￿𐀁@gmail.com"] = "☹️♈️♥️⚽️￿@gmail.com"

	// Create Users without Properties and CustomerUserID as emojiText.
	for actualCustomerUserId, expectedCustomerUserId := range CustomerUserId {
		createdUserID, errCode := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
			Group1ID:       "1",
			Group2ID:       "2",
			CustomerUserId: actualCustomerUserId,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, expectedCustomerUserId, user.CustomerUserId)
	}
}

func TestDiffPostgresJsonb(t *testing.T) {

	existingProps := postgres.Jsonb{RawMessage: []byte(`{"$property1": 10,"$property3": "value3", "$property_unchanged": "value", "name": "john"}`)}
	newProps := postgres.Jsonb{RawMessage: []byte(`{"$property1": 20,"$property2": 10, "$property3": "value31", "$property4": "value4", "name": "johan" }`)}

	pMap := U.DiffPostgresJsonb(1, &existingProps, &newProps, "TEST")
	// Updated properties.
	assert.Equal(t, float64(20), (*pMap)["$property1"])
	assert.Equal(t, "value31", (*pMap)["$property3"])
	assert.Equal(t, "johan", (*pMap)["name"]) // custom property.

	// New properites.
	assert.Equal(t, float64(10), (*pMap)["$property2"])
	assert.Equal(t, "value4", (*pMap)["$property4"])

	// Unchanged property.
	assert.Nil(t, (*pMap)["property_unchanged"])
}
