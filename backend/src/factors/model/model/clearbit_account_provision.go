package model

type ClearbitProvisionAPIResponse struct {
	Id     string `json:"id"`
	Domain string `json:"domain"`
	Email  string `json:"email"`
	Plans  struct {
		FactorsPbcCustomer []Factor `json:"factors-pbc-customer"`
	} `json:"plans"`
	Keys Keys `json:"keys"`
}

type Factor struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Limit int    `json:"limit"`
}

type Keys struct {
	Public string `json:"public"`
	Secret string `json:"secret"`
}
