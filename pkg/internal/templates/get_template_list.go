package templates

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/pkg/errors"
)

// Client to be used with inspections
type Client struct {
	apiClient *httpapi.Client
}

func NewTemplatesClient(apiClient *httpapi.Client) *Client {
	return &Client{
		apiClient: apiClient,
	}
}

func (c *Client) GetTemplateList(ctx context.Context, pageSize int) ([]TemplateResponseItem, error) {

	var items []TemplateResponseItem
	callback := func(resp *listTemplatesResponse) {
		items = append(items, resp.Templates...)
	}

	if err := c.drainTemplates(ctx, pageSize, callback); err != nil {
		return nil, err
	}
	return items, nil
}

// DrainTemplates will process a paginated response and collect all responses before returning the response
func (c *Client) drainTemplates(ctx context.Context, pageSize int, callbackFn func(response *listTemplatesResponse)) error {
	nextToken := ""

	for {
		resp, err := c.getTemplateList(ctx, &templateSearchRequest{
			PageSize: pageSize,
			Token:    nextToken,
		})

		if err != nil {
			return err
		}

		callbackFn(resp)

		if resp.NextPageToken == "" {
			// there is no another page
			break
		}
		nextToken = resp.NextPageToken
	}
	return nil
}

// getTemplateList will return a simplified list of customer's templates
func (c *Client) getTemplateList(ctx context.Context, params *templateSearchRequest) (*listTemplatesResponse, error) {

	sl := c.apiClient.Sling.New().Get("/templates/v1/templates/search").
		Set(string(httpapi.Authorization), c.apiClient.AuthorizationHeader).
		Set(string(httpapi.IntegrationID), c.apiClient.IntegrationID).
		Set(string(httpapi.IntegrationVersion), c.apiClient.IntegrationVersion).
		Set(string(httpapi.XRequestID), util.RequestIDFromContext(ctx))

	sl.QueryStruct(params)
	req, _ := sl.Request()
	req = req.WithContext(ctx)

	var result *listTemplatesResponse
	var errMsg json.RawMessage

	_, err := c.apiClient.Do(&util.SlingHTTPDoer{
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

// listTemplatesResponse list of templates
type listTemplatesResponse struct {
	NextPageToken string                 `json:"next_page_token"`
	Templates     []TemplateResponseItem `json:"items"`
}

// TemplateResponseItem simple representation of a template date
type TemplateResponseItem struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
}

// templateSearchRequest contains parameters for calling api-templates template search
type templateSearchRequest struct {
	PageSize int    `url:"page_size"`
	Token    string `url:"page_token,omitempty"`
}
