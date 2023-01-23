package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// Bit positions for display categories
var sectionBitMapping = map[string]int{
	WebsiteSessionDisplayCategory:  1,
	FormSubmissionsDisplayCategory: 2,

	AllChannelsDisplayCategory:   3,
	GoogleAdsDisplayCategory:     4,
	FacebookDisplayCategory:      5,
	LinkedinDisplayCategory:      6,
	BingAdsDisplayCategory:       7,
	GoogleOrganicDisplayCategory: 8,

	HubspotContactsDisplayCategory:  9,
	HubspotCompaniesDisplayCategory: 10,
	HubspotDealsDisplayCategory:     11,

	SalesforceUsersDisplayCategory:         12,
	SalesforceAccountsDisplayCategory:      13,
	SalesforceOpportunitiesDisplayCategory: 14,

	EventsBasedDisplayCategory:  15,
	MarketoLeadsDisplayCategory: 16,
	PageViewsDisplayCategory:    17,
}

type Property struct {
	Category        string `json:"ca"`
	DisplayCategory string `json:"dc"`
	ObjectType      string `json:"obj_ty"`
	Name            string `json:"name"`
	DataType        string `json:"da_ty"`
	Entity          string `json:"en"`
	GroupByType     string `json:"gb_ty"`
}

type PropertyMapping struct {
	ID            string          `gorm:"primary_key:true;type:varchar(255)" json:"id"`
	ProjectID     int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	Name          string          `json:"name"`
	DisplayName   string          `json:"display_name"`
	SectionBitMap int64           `json:"-"`
	DataType      string          `json:"data_type"`
	Properties    *postgres.Jsonb `json:"properties"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	IsDeleted     bool            `json:"is_deleted"`
}

func (propertyMapping *PropertyMapping) IsValid(properties []Property) (bool, string) {

	// Validating project_id and name
	if propertyMapping.ProjectID == 0 {
		return false, "Invalid project ID for property mapping"
	}
	// TODO: more validation to be added, $ _ validation
	if propertyMapping.DisplayName == "" {
		return false, "Invalid display name for property mapping"
	}

	// Validating properties
	if len(properties) < 2 {
		return false, "At least two properties requiered for property_mapping - property_mapping handler."
	}

	dataType := properties[0].DataType
	if dataType != "categorical" && dataType != "numerical" {
		return false, "Invalid data type for property_mapping - property_mapping handler."
	}

	displayCategorySet := make(map[string]struct{})
	for _, property := range properties {
		if !property.IsValid() {
			return false, "Error with values passed in properties - property_mapping handler"
		}
		if property.DataType != dataType {
			return false, "All properties should have same data type - property_mapping handler"
		}
		if _, present := displayCategorySet[property.DisplayCategory]; present {
			return false, "Duplicate display category - property_mapping handler"
		}
		displayCategorySet[property.DisplayCategory] = struct{}{}
	}
	return true, ""
}

func (properties *Property) IsValid() bool {

	if properties.Category == "" || properties.DisplayCategory == "" ||
		properties.Name == "" || properties.DataType == "" {
		return false
	}

	return true
}

// Returns a array of string containing display_category from properties json.
func GenerateSectionBitMapFromProperties(properties []Property) (int64, string) {

	displayCategories := make([]string, 0)
	for _, property := range properties {
		displayCategories = append(displayCategories, property.DisplayCategory)
	}

	return GenerateSectionBitMap(displayCategories)
}

// Takes list of display category
// Returns sectionBitMap 
// Binary bits are marked based on display_category from properties from left to right.
func GenerateSectionBitMap(displayCategories []string) (int64, string) {

	sectionBitMap := int64(0)
	for _, displayCategory := range displayCategories {
		bitPosition, present := sectionBitMapping[displayCategory]
		if !present {
			return 0, "Invalid object type for property mapping"
		}
		// Mark the bit as per position of display category
		sectionBitMap = sectionBitMap | (1 << (bitPosition - 1))
	}

	return sectionBitMap, ""
}
