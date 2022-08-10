package tests

import (
	"encoding/json"
	// C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestCreateOTPRule(t *testing.T) {

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	t.Run("CreateOTPRule", func(t *testing.T) {
		rule, errCode, str := store.GetStore().CreateOTPRule(
			project.ID,
			&model.OTPRule{
				RuleType:          model.TouchPointRuleTypeSFNormal,
				TouchPointTimeRef: model.LastModifiedTimeRef,
				PropertiesMap: postgres.Jsonb{RawMessage: json.RawMessage(`{
        "$campaign": {
          "ty": "Property",
          "va": "$hubspot_contact_last_marketing_email_replied_name"
        },
        "$channel": {
          "ty": "Constant",
          "va": "Inbound"
        },
        "$source": {
          "ty": "Constant",
          "va": "Marketing Email Reply"
        },
        "$type": {
          "ty": "Constant",
          "va": "Offer"
        }
      }`)},
				Filters: postgres.Jsonb{RawMessage: json.RawMessage(`[
        {
          "lop": "AND",
          "op": "notEqual",
          "pr": "$hubspot_contact_last_marketing_email_replied_name",
          "va": "$none"
        },
        {
          "lop": "AND",
          "op": "equals",
          "pr": "$hubspot_contact_company_channel_new",
          "va": "Inbound"
        }
      ]`)},
				CreatedBy: agent.UUID,
			})
		assert.NotNil(t, rule)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, "", str)
	})
}

func TestDeleteOTPRule(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	rule, errCode, errStr := store.GetStore().CreateOTPRule(
		project.ID,
		&model.OTPRule{
			RuleType:          model.TouchPointRuleTypeSFNormal,
			TouchPointTimeRef: model.LastModifiedTimeRef,
			PropertiesMap: postgres.Jsonb{RawMessage: json.RawMessage(`{
        "$campaign": {
          "ty": "Property",
          "va": "$hubspot_contact_last_marketing_email_replied_name"
        },
        "$channel": {
          "ty": "Constant",
          "va": "Inbound"
        },
        "$source": {
          "ty": "Constant",
          "va": "Marketing Email Reply"
        },
        "$type": {
          "ty": "Constant",
          "va": "Offer"
        }
      }`)},
			Filters: postgres.Jsonb{RawMessage: json.RawMessage(`[
        {
          "lop": "AND",
          "op": "notEqual",
          "pr": "$hubspot_contact_last_marketing_email_replied_name",
          "va": "$none"
        },
        {
          "lop": "AND",
          "op": "equals",
          "pr": "$hubspot_contact_company_channel_new",
          "va": "Inbound"
        }
      ]`)},
			CreatedBy: agent.UUID,
		})
	assert.NotNil(t, rule)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, "", errStr)

	t.Run("DeleteOTPRule", func(t *testing.T) {
		code, _ := store.GetStore().DeleteOTPRule(project.ID, rule.ID)
		assert.Equal(t, http.StatusAccepted, code)

		ruleDeleted, _ := store.GetStore().GetAnyOTPRuleWithRuleId(project.ID, rule.ID)
		assert.NotNil(t, ruleDeleted)
		assert.Equal(t, true, ruleDeleted.IsDeleted)

	})
}

func TestGetOTPRule(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	rule, errCode, str := store.GetStore().CreateOTPRule(
		project.ID,
		&model.OTPRule{
			RuleType:          model.TouchPointRuleTypeSFNormal,
			TouchPointTimeRef: model.LastModifiedTimeRef,
			PropertiesMap: postgres.Jsonb{RawMessage: json.RawMessage(`{
        "$campaign": {
          "ty": "Property",
          "va": "$hubspot_contact_last_marketing_email_replied_name"
        },
        "$channel": {
          "ty": "Constant",
          "va": "Inbound"
        },
        "$source": {
          "ty": "Constant",
          "va": "Marketing Email Reply"
        },
        "$type": {
          "ty": "Constant",
          "va": "Offer"
        }
      }`)},
			Filters: postgres.Jsonb{RawMessage: json.RawMessage(`[
        {
          "lop": "AND",
          "op": "notEqual",
          "pr": "$hubspot_contact_last_marketing_email_replied_name",
          "va": "$none"
        },
        {
          "lop": "AND",
          "op": "equals",
          "pr": "$hubspot_contact_company_channel_new",
          "va": "Inbound"
        }
      ]`)},
			CreatedBy: agent.UUID,
		})
	assert.NotNil(t, rule)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, "", str)

	t.Run("ReadOTPRule", func(t *testing.T) {
		template, errCode := store.GetStore().GetOTPRuleWithRuleId(project.ID, rule.ID)
		assert.NotNil(t, template)
		assert.Equal(t, 302, errCode)
	})
}
