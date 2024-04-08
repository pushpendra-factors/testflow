package model

import (
	U "factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type RequestSegmentKPI struct {
	From     int64  `json:"fr"`
	To       int64  `json:"to"`
	Timezone string `json:"tz"`
}

type WidgetGroup struct {
	ProjectID       int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID              string          `gorm:"primary_key:true;type:varchar(255)" json:"wid_g_id"`
	DisplayName     string          `gorm:"display_name" json:"wid_g_d_name"`
	Name            string          `gorm:"name" json:"name"`
	IsNonComparable bool            `json:"non_comp"`
	Widgets         *postgres.Jsonb `json:"wids"`
	DecodedWidgets  []Widget        `json:"de_wids" gorm:"-"`
	WidgetsAdded    bool            `json:"-" gorm:"-"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (widgetGroup *WidgetGroup) CreateWidgetJsonWithNoElements() {
	widgets := make([]Widget, 0)
	encodedWidgets, _ := U.EncodeStructTypeToPostgresJsonb(widgets)
	widgetGroup.Widgets = encodedWidgets
}

func (widgetGroup *WidgetGroup) DecodeWidgetsAndSetDecodedWidgets() {

	var widgets []Widget
	U.DecodePostgresJsonbToStructType(widgetGroup.Widgets, &widgets)
	widgetGroup.DecodedWidgets = widgets
}

func (widgetGroup *WidgetGroup) GetWidget(ID string) (Widget, int) {
	for _, presentWidget := range widgetGroup.DecodedWidgets {
		if presentWidget.ID == ID {
			return presentWidget, http.StatusFound
		}
	}
	return Widget{}, http.StatusNotFound
}

func (widgetGroup *WidgetGroup) UpdateWidget(inputWidget Widget) {
	for index, presentWidget := range widgetGroup.DecodedWidgets {
		if presentWidget.ID == inputWidget.ID {
			widgetGroup.DecodedWidgets[index] = inputWidget
		}
	}
}

func (widgetGroup *WidgetGroup) DeleteWidget(inputWidgetID string) {
	result := make([]Widget, 0)
	for _, presentWidget := range widgetGroup.DecodedWidgets {
		if presentWidget.ID != inputWidgetID {
			result = append(result, presentWidget)
		}
	}
	widgetGroup.DecodedWidgets = result
}

type Widget struct {
	ID              string    `json:"id"`
	DisplayName     string    `json:"d_name"`
	QueryType       string    `json:"q_ty"`
	QueryMetric     string    `json:"q_me"`
	QueryMetricType string    `json:"q_me_ty"`
	IsNonEditable   bool      `json:"non_edit"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (widget *Widget) IsValid() (bool, string) {
	if widget.DisplayName == "" || widget.QueryType == "" || widget.QueryMetric == "" {
		return false, "Required Fields are not provided."
	}

	if !U.ContainsStringInArray([]string{QueryClassKPI}, widget.QueryType) {
		return false, "Wrong QueryType are not provided."
	}

	return true, ""
}

func (widget *Widget) ValidateConstraints(widgetGroup WidgetGroup) (bool, string) {
	isNameValid, nameErrMsg := widget.ValidateWidgetName(widgetGroup)
	if !isNameValid {
		return isNameValid, nameErrMsg
	}
	isMetricValid, metricErrMsg := widget.ValidateWidgetQueryMetric(widgetGroup)
	if !isMetricValid {
		return isMetricValid, metricErrMsg
	}
	return true, ""
}

func (widget *Widget) ValidateUpdatedConstraints(widgetGroup WidgetGroup, displayNameChanged bool, queryMetricChanged bool) (bool, string) {
	if displayNameChanged {
		isNameValid, nameErrMsg := widget.ValidateWidgetName(widgetGroup)
		if !isNameValid {
			return isNameValid, nameErrMsg
		}
	}
	if queryMetricChanged {
		isMetricValid, metricErrMsg := widget.ValidateWidgetQueryMetric(widgetGroup)
		if !isMetricValid {
			return isMetricValid, metricErrMsg
		}
	}
	return true, ""
}

func (widget *Widget) ValidateWidgetName(widgetGroup WidgetGroup) (bool, string) {
	for _, alreadyPresentWidget := range widgetGroup.DecodedWidgets {
		if widget.DisplayName == alreadyPresentWidget.DisplayName {
			return false, "Duplicate widget name is passed"
		}
	}
	return true, ""
}

func (widget *Widget) ValidateWidgetQueryMetric(widgetGroup WidgetGroup) (bool, string) {
	for _, alreadyPresentWidget := range widgetGroup.DecodedWidgets {
		if widget.QueryMetric == alreadyPresentWidget.QueryMetric {
			return false, "Duplicate widget query metric is passed"
		}
	}
	return true, ""
}
