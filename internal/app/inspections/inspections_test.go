package inspections_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	exportermock "github.com/SafetyCulture/iauditor-exporter/internal/app/exporter/mocks"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/inspections"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/h2non/gock.v1"
)

func initMockInspections(httpClient *http.Client) {
	gock.InterceptClient(httpClient)

	gock.New("http://localhost:9999").
		Get("/audits/search").
		Reply(200).
		File("mocks/inspections.json")

	gock.New("http://localhost:9999").
		Get("/audits/audit_d7e2f55b95094bd48fac601850e1db63").
		Reply(200).
		File("mocks/audits/audit_d7e2f55b95094bd48fac601850e1db63.json")

	gock.New("http://localhost:9999").
		Get("/audits/audit_872d26e479924353b957b832cd98a7f6").
		Reply(200).
		File("mocks/audits/audit_872d26e479924353b957b832cd98a7f6.json")

	gock.New("http://localhost:9999").
		Get("/audits/audit_7f1ecd2e66474418b1e062657aeb1389").
		Reply(200).
		File("mocks/audits/audit_7f1ecd2e66474418b1e062657aeb1389.json")
}

func TestInspectionsExport(t *testing.T) {

	viperConfig := viper.New()
	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	initMockInspections(apiClient.HTTPClient())

	exporterMock := new(exportermock.Exporter)
	exporterMock.On("WriteRow", mock.Anything, mock.Anything)
	exporterMock.On("SetLastModifiedAt", mock.Anything)
	exporterMock.On("GetLastModifiedAt", mock.Anything).Return(nil)

	inspectionClient := inspections.NewInspectionClient(
		viperConfig,
		apiClient,
		exporterMock,
	)
	err := inspectionClient.Export(context.Background())
	assert.Nil(t, err)
}
