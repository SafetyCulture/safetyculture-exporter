package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/external/version"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/pkg/errors"
)

// GetInspection retrieves the inspection of the given id.
func GetInspection(ctx context.Context, a *httpapi.Client, id string) (*json.RawMessage, error) {
	var (
		result *json.RawMessage
		errMsg json.RawMessage
	)

	sl := a.Sling.New().Get(fmt.Sprintf("/audits/%s", id)).
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
