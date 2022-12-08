package feed

import (
	"encoding/json"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/external/version"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/pkg/errors"
)

// ListInspections retrieves the list of inspections from SafetyCulture
func ListInspections(ctx context.Context, params *ListInspectionsParams) (*ListInspectionsResponse, error) {
	var (
		result *ListInspectionsResponse
		errMsg json.RawMessage
	)

	sl := a.sling.New().Get("/audits/search").
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	sl.QueryStruct(params)
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
