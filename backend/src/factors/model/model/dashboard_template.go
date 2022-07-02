package model

import (
	"github.com/jinzhu/gorm/dialects/postgres"
)

type UnitInfo struct {
	Title		 string 		  `json:"title"`
	ID 			 int 			  `json:"type"`
	Description	 string			  `json:"description"`
	Query        postgres.Jsonb  `json:"query"`
}

type DashboardTemplate struct {	
	// Template ID
	ID 						string 		   `gorm:"primary_key:true" json:"-"`
	Title         			string         `gorm:"not null" json:"title"`
	Description   			string 		   `gorm:"not null" json:"description"`
	Dashboard 	  			postgres.Jsonb `gorm:"not null" json:"dashboard"`
	Units 		  			postgres.Jsonb `gorm:"not null" json:"units"`
	IsDeleted     			bool           `gorm:"not null;default:false" json:"is_deleted"`
	SimilarTemplateIds		postgres.Jsonb `gorm:"not null" json:"similar_template_ids"`
	Tags					postgres.Jsonb `gorm:"not null" json:"tags"`
}
