package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/external/version"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/pkg/errors"
)

type GetSheqsyCompanyResponse struct {
	CompanyID   int         `json:"companyId"`
	CompanyName string      `json:"companyName"`
	Name        interface{} `json:"name"`
	CompanyUID  string      `json:"companyUId"`
}

// GetSheqsyCompany returns the details for the selected company
func (a *Client) GetSheqsyCompany(ctx context.Context, companyID string) (*GetSheqsyCompanyResponse, error) {
	var (
		result *GetSheqsyCompanyResponse
		errMsg json.RawMessage
	)

	sl := a.sling.New().Get(fmt.Sprintf("/SheqsyIntegrationApi/api/v3/companies/%s", companyID)).
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.do(&slingHTTPDoer{
		sl:       sl,
		req:      req,
		successV: &result,
		failureV: &errMsg,
	})
	if err != nil {
		return nil, errors.Wrap(err, "api request")
	}

	return result, nil
}
