package service

import (
	"github.com/golobby/container"
	microappSecurity "github.com/islax/microapp/security"
	uuid "github.com/satori/go.uuid"
)

// ExtractTenantID service to get tenant ID from token/current
type ExtractTenantID interface {
	GetTenantIDAsUUID(params map[string]string, token *microappSecurity.JwtToken, tenantID string) (uuid.UUID, error)
	GetTenantIDAsString(params map[string]string, token *microappSecurity.JwtToken) (string, error)
}

// GetTenantIDFromToken extracts tenantId from params and validates it with token
func GetTenantIDFromToken() ExtractTenantID {
	var service ExtractTenantID
	container.Make(&service)

	return service
}
