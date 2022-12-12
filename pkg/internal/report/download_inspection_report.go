package report

import (
	"context"
	"io"
	"net/http"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/version"
)

// DownloadInspectionReportFile downloads the report file of the inspection.
func DownloadInspectionReportFile(ctx context.Context, apiClient *httpapi.Client, url string) (io.ReadCloser, error) {
	var res *http.Response

	sl := apiClient.Sling.New().Get(url).
		Set(string(httpapi.Authorization), apiClient.AuthorizationHeader).
		Set(string(httpapi.IntegrationID), "safetyculture-exporter").
		Set(string(httpapi.IntegrationVersion), version.GetVersion()).
		Set(string(httpapi.XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	res, err := apiClient.Do(&util.DefaultHTTPDoer{
		Req:        req,
		HttpClient: apiClient.HttpClient,
	})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
