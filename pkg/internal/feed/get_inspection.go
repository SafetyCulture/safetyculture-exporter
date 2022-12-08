package feed

import (
	"encoding/json"
	"fmt"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/external/version"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/pkg/errors"
)

// GetInspection retrieves the inspection of the given id.
func GetInspection(ctx context.Context, id string) (*json.RawMessage, error) {
	var (
		result *json.RawMessage
		errMsg json.RawMessage
	)

	sl := a.sling.New().Get(fmt.Sprintf("/audits/%s", id)).
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
