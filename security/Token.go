package security

import (
	"strings"

	"github.com/golang-jwt/jwt"
	uuid "github.com/satori/go.uuid"
)

const (
	// ApplianceExternalIdType indicates Appliance ExternalID Type
	ApplianceExternalIdType = "Appliance"
	// SessionExternalIdType indicates Session ExternalID Type
	SessionExternalIdType = "Session"
	// UserExternalIdType indicates User ExternalID Type
	UserExternalIdType = "User"
	// PartnerExternalIdType indicates Partner ExternalID Type
	PartnerExternalIdType = "Partner"
)

// JwtToken represents the parsed Token from Authentication Header
type JwtToken struct {
	// UserID is id of user matching the token
	UserID         uuid.UUID   `json:"user,omitempty"`
	UserName       string      `json:"name,omitempty"`
	DisplayName    string      `json:"displayName,omitempty"`
	UserGroupIDs   []uuid.UUID `json:"usergroupIds,omitempty"`
	UserGroupNames []string    `json:"usergroupNames,omitempty"`
	TenantID       uuid.UUID   `json:"tenant,omitempty"`
	TenantName     string      `json:"tenantName,omitempty"`
	ExternalID     string      `json:"externalId,omitempty"`
	ExternalIDType string      `json:"externalIdType,omitempty"`
	Scopes         []string    `json:"scope,omitempty"`
	Admin          bool        `json:"admin,omitempty"`
	PolicyID       uuid.UUID   `json:"policyId,omitempty"`  // PolicyID is id of policy matching the token
	PartnerID      uuid.UUID   `json:"partnerId,omitempty"` // PartnerID is id of partner matching the token
	Raw            string      `json:"-"`
	jwt.StandardClaims
}

func (token *JwtToken) isValidForScope(allowedScopes []string) bool {
	permissiveTokenScopes := []string{}
	nonPermissiveTokenScopes := []string{}

	for _, tokenScope := range token.Scopes {
		if strings.HasPrefix(tokenScope, "-") {
			nonPermissiveTokenScopes = append(nonPermissiveTokenScopes, tokenScope[1:])
		} else {
			permissiveTokenScopes = append(permissiveTokenScopes, tokenScope)
		}
	}

	if len(nonPermissiveTokenScopes) > 0 && len(allowedScopes) > 0 {
		if isScopePresent(nonPermissiveTokenScopes, allowedScopes) {
			return false
		}
	}
	return isScopePresent(permissiveTokenScopes, allowedScopes)
}

func isScopePresent(scopes []string, scopeToCheck []string) bool {
	if ok, _ := inArray("*", scopes); ok {
		return true
	}
	allScopesMatched := true
	for _, allowedScope := range scopeToCheck {
		if ok, _ := inArray(allowedScope, scopes); !ok {
			scopeParts := strings.Split(allowedScope, ":")
			// If the api scope only has 1 value with no separator, e.g []string{"*"}, then we need to check if both scopes are same
			if len(scopeParts) == 1 {
				if ok, _ := inArray(scopeParts[0], scopes); !ok {
					allScopesMatched = false
				}
				continue
			}
			if ok, _ := inArray(scopeParts[0]+":*", scopes); !ok {
				if ok, _ := inArray("*:"+scopeParts[1], scopes); !ok {
					allScopesMatched = false
				}
			}
		}
	}
	return allScopesMatched
}

func inArray(val string, array []string) (ok bool, i int) {
	for i = range array {
		if ok = array[i] == val; ok {
			return
		}
	}
	return
}
