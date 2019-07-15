package security

import (
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
)

// ExternalTypes enumeration holds values for several types an Id can have
type ExternalTypes uint

// state values
const (
	None      = 0
	Appliance = 1 + iota
	Session
	User
)

// JwtToken represents the parsed Token from Authentication Header
type JwtToken struct {
	// UserID is id of user matchimg the token
	UserID       uuid.UUID `json:"user,omitempty"`
	UserName     string    `json:"name,omitempty"`
	TenantID     uuid.UUID `json:"tenant,omitempty"`
	ExternalID   uuid.UUID `json:"externalId,omitempty"`
	ExternalType uint      `json:"externalType,omitempty"`
	Scopes       []string  `json:"scope,omitempty"`
	Admin        bool      `json:"admin,omitempty"`
	Raw          string    `json:"-"`
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

func inArray(val string, array []string) (ok bool, i int) {
	for i = range array {
		if ok = array[i] == val; ok {
			return
		}
	}
	return
}
