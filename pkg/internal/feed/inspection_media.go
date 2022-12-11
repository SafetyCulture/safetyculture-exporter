package feed

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/version"
)

// GetMedia fetches the media object from SafetyCulture.
func GetMedia(ctx context.Context, a *httpapi.Client, request *GetMediaRequest) (*GetMediaResponse, error) {
	baseURL := strings.TrimPrefix(request.URL, a.BaseURL)

	// The mediaURL will be in the following format:
	// https://api.eu.safetyculture.com/audits/audit_xxx/media/4c83fcf2-180b-4d3e-958f-389f7ac49777
	// The string that is after the word "media/" is the ID of it.
	mediaIDURL := strings.Split(request.URL, "/")
	mediaID := mediaIDURL[len(mediaIDURL)-1]

	sl := a.Sling.New().Get(baseURL).
		Set(string(httpapi.Authorization), a.AuthorizationHeader).
		Set(string(httpapi.IntegrationID), "safetyculture-exporter").
		Set(string(httpapi.IntegrationVersion), version.GetVersion()).
		Set(string(httpapi.XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	result, err := a.Do(&util.DefaultHTTPDoer{
		Req:        req,
		HttpClient: a.HttpClient,
	})
	if err != nil {
		// Ignore forbidden errors for media objects.
		if result != nil && result.StatusCode == http.StatusForbidden {
			return nil, nil
		}
		return nil, err
	}
	defer result.Body.Close()

	if result.StatusCode == 204 {
		return nil, nil
	}

	contentType := result.Header.Get("Content-Type")
	if contentType == "" {
		return nil, fmt.Errorf("failed to get content-type of media")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)

	resp := &GetMediaResponse{
		ContentType: contentType,
		Body:        buf.Bytes(),
		MediaID:     mediaID,
	}
	return resp, nil
}

// GetMediaRequest has all the data needed to make a request to get a media
type GetMediaRequest struct {
	URL     string
	AuditID string
}

// GetMediaResponse is a representation of the data returned when fetching media
type GetMediaResponse struct {
	ContentType string
	Body        []byte
	MediaID     string
}
