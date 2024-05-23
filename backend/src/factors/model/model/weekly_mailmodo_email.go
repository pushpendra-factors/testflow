package model

type WeeklyMailmodoEmailMetrics struct {
	ProjectName                   string
	SessionUniqueUserCount        int64
	FormSubmittedUniqueUserCount  int64
	IdentifiedCompaniesCount      int64
	TotalIdentifiedCompaniesCount int64
	Industry                      TopValueAndCountObject
	EmployeeRange                 TopValueAndCountObject
	Country                       TopValueAndCountObject
	RevenueRange                  TopValueAndCountObject
	SessionCount                  int64
	TopChannel                    string
	TopSource                     string
	TopCampaign                   string
	Company1                      ActiveCompanyProp
	Company2                      ActiveCompanyProp
	Company3                      ActiveCompanyProp
	Company4                      ActiveCompanyProp
	Company5                      ActiveCompanyProp
}

type TopValueAndCountObject struct {
	TopValue                string
	TopValuePercent         int64
	TopValueLastWeek        string
	TopValuePercentLastWeek int64
}

type ActiveCompanyProp struct {
	Domain   string `json:"domain"`
	Industry string `json:"industry"`
	Country  string `json:"country"`
	Revenue  string `json:"revenue"`
}
