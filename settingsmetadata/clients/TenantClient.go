package clients

import (
	"fmt"
	"net/http"

	apiclients "github.com/islax/microapp/clients"
	microappCtx "github.com/islax/microapp/context"
)

// TenantClient is used to interface with Security microservice
type TenantClient interface {
	GetTenant(context microappCtx.ExecutionContext, rawToken string, tenantID string) (map[string]interface{}, error)
	GetAllTenants(context microappCtx.ExecutionContext, rawToken string) ([]map[string]interface{}, error)
}

// NewTenantClient returns a new instance of ServerManagerClient
func NewTenantClient(appName, url string) TenantClient {
	client := &tenantClientImpl{}
	client.HTTPClient = &http.Client{}
	client.BaseURL = url
	client.AppName = appName

	return client
}

type tenantClientImpl struct {
	apiclients.APIClient
}

func (tenantClient *tenantClientImpl) GetTenant(context microappCtx.ExecutionContext, rawToken string, tenantID string) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("/api/tenants/%v", tenantID)
	tenantResult, err := tenantClient.DoGet(context, apiURL, rawToken)
	if err != nil {
		return nil, err
	}
	return tenantResult, nil
}

func (tenantClient *tenantClientImpl) GetAllTenants(context microappCtx.ExecutionContext, rawToken string) ([]map[string]interface{}, error) {
	apiURL := "/api/tenants"
	tenantResult, err := tenantClient.DoGetList(context, apiURL, rawToken)
	if err != nil {
		return nil, err
	}
	return tenantResult, nil
}
