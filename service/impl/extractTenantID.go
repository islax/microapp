package impl

import (
	"net/http"

	microappError "github.com/islax/microapp/error"
	microappSecurity "github.com/islax/microapp/security"
	"github.com/islax/microapp/service"
	uuid "github.com/satori/go.uuid"
)

type extractTenantID struct {
}

// NewExtractTenantID gets the Tenant ID from token/current
func NewExtractTenantID() service.ExtractTenantID {
	return &extractTenantID{}
}

func (service *extractTenantID) GetTenantIDAsUUID(params map[string]string, token *microappSecurity.JwtToken, tenantID string) (uuid.UUID, error) {
	if token.Admin == false {
		if tenantID == "current" {
			tenantID = token.TenantID.String()
		} else if tenantID != token.TenantID.String() {
			return uuid.NewV4(), microappError.NewHTTPError("Key_Unauthorized", http.StatusUnauthorized) //TODO: add Key_Unauthorized in contants.go file of the microapp
		}
	}

	tenantIDAsUUID, err := uuid.FromString(tenantID)
	if err != nil {
		return uuid.NewV4(), microappError.NewHTTPResourceNotFound("tenant", tenantID)
	}
	return tenantIDAsUUID, nil
}

func (service *extractTenantID) GetTenantIDAsString(params map[string]string, token *microappSecurity.JwtToken) (string, error) {
	tenantID := params["tenantId"]
	if token.Admin == false {
		if tenantID == "current" {
			tenantID = token.TenantID.String()
		} else if tenantID != token.TenantID.String() {
			return "", microappError.NewHTTPError("Key_Unauthorized", http.StatusUnauthorized) //TODO: add Key_Unauthorized in contants.go file of the microapp
		}
	}
	_, err := uuid.FromString(tenantID)
	if err != nil {
		return "", microappError.NewHTTPResourceNotFound("tenant", tenantID)
	}
	return tenantID, nil
}
