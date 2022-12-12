package feed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/version"
	"github.com/pkg/errors"
)

type GetSheqsyCompanyResponse struct {
	CompanyID   int         `json:"companyId"`
	CompanyName string      `json:"companyName"`
	Name        interface{} `json:"name"`
	CompanyUID  string      `json:"companyUId"`
}

// GetSheqsyCompany returns the details for the selected company
func GetSheqsyCompany(ctx context.Context, a *httpapi.Client, companyID string) (*GetSheqsyCompanyResponse, error) {
	var (
		result *GetSheqsyCompanyResponse
		errMsg json.RawMessage
	)

	sl := a.Sling.New().Get(fmt.Sprintf("/SheqsyIntegrationApi/api/v3/companies/%s", companyID)).
		Set(string(httpapi.Authorization), a.AuthorizationHeader).
		Set(string(httpapi.IntegrationID), "safetyculture-exporter").
		Set(string(httpapi.IntegrationVersion), version.GetVersion()).
		Set(string(httpapi.XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.Do(&util.SlingHTTPDoer{
		Sl:       sl,
		Req:      req,
		SuccessV: &result,
		FailureV: &errMsg,
	})
	if err != nil {
		return nil, errors.Wrap(err, "api request")
	}

	return result, nil
}
