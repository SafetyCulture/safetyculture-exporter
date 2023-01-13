package httpapi

// WhoAmIResponse represents the response of  WhoAmI
type WhoAmIResponse struct {
	UserID         string `json:"user_id"`
	OrganisationID string `json:"organisation_id"`
	Firstname      string `json:"firstname"`
	Lastname       string `json:"lastname"`
}

type GetSheqsyCompanyResponse struct {
	CompanyID   int         `json:"companyId"`
	CompanyName string      `json:"companyName"`
	Name        interface{} `json:"name"`
	CompanyUID  string      `json:"companyUId"`
}
