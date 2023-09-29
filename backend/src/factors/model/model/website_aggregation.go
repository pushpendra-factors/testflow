package model


type WebsiteAggregation struct {
	ProjectID                 int64
	TimestampAtDay            int64
	EventName                 string
	EventType                 string
	Source                    string
	Medium                    string
	Campaign                  string
	ReferrerUrl               string
	LandingPageUrl            string
	Country                   string
	Region                    string
	City                      string
	Browser                   string
	BrowserVersion            string
	Os                        string
	OsVersion                 string
	Device                    string
	SixSignalIndustry         string `gorm:"column:6signal_industry"`
	SixSignalEmployeeRange    string `gorm:"column:6signal_employee_range"`
	SixSignalRevenueRange     string `gorm:"column:6signal_revenue_range"`
	SixSignalNaicsDescription string `gorm:"column:6signal_naics_description"`
	SixSignalSicDescription   string `gorm:"column:6signal_sic_description"`
	CountOfRecords            int64
	SpentTime                 float64
}

func (WebsiteAggregation) TableName() string {
	return "website_aggregation"
}
