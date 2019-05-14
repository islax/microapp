package security

import (
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
)

// JwtToken represents the parsed Token from Authentication Header
type JwtToken struct {
	// UserID is id of user matchimg the token
	UserID   uuid.UUID `json:"user,omitempty"`
	UserName string    `json:"name,omitempty"`
	TenantID uuid.UUID `json:"tenant,omitempty"`
	Scopes   []string  `json:"scope,omitempty"`
	Admin    bool      `json:"admin,omitempty"`
	jwt.StandardClaims
}

func (token *JwtToken) isValidForScope(allowedScopes []string) bool {
	if ok, _ := inArray("*", token.Scopes); ok {
		return true
	}
	allScopesMatched := true
	for _, allowedScope := range allowedScopes {
		if ok, _ := inArray(allowedScope, token.Scopes); !ok {
			scopeParts := strings.Split(allowedScope, ":")
			if ok, _ := inArray(scopeParts[0]+":*", token.Scopes); !ok {
				allScopesMatched = false
			}
		}
	}
	return allScopesMatched
}

// GetTenantID returns the tenantId associated with the Token
func (token *JwtToken) GetTenantID() uuid.UUID {
	return token.TenantID
}

// GetUserID returns the userId associated with the Token
func (token *JwtToken) GetUserID() uuid.UUID {
	return token.UserID
}

// GetUserName returns the userName associated with the Token
func (token *JwtToken) GetUserName() string {
	return token.UserName
}

// GetScopes returns the scopes associated with the Token
func (token *JwtToken) GetScopes() []string {
	return token.Scopes
}

func inArray(val string, array []string) (ok bool, i int) {
	for i = range array {
		if ok = array[i] == val; ok {
			return
		}
	}
	return
}
