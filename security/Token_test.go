package security

import (
	"testing"
)

var scopeCombinations = []struct {
	allowedScopes []string
	scopesToTest  []string
	result        bool
}{
	{[]string{"*"}, []string{"trustedGroup:read"}, true},
	{[]string{"trustedGroup:*"}, []string{"trustedGroup:read"}, true},
	{[]string{"appliance:*", "trustedGroup:*"}, []string{"trustedGroup:read"}, true},
	{[]string{"trustedGroup:read"}, []string{"trustedGroup:read"}, true},
	{[]string{"trustedGroup:write"}, []string{"trustedGroup:read"}, false},
}

func TestScopes(t *testing.T) {
	for _, combination := range scopeCombinations {
		t.Run("Match", func(t *testing.T) {
			token := &JwtToken{Scopes: combination.allowedScopes}
			result := token.isValidForScope(combination.scopesToTest)
			if result != combination.result {
				t.Errorf("got %v, want %v", result, combination.result)
			}
		})
	}
}
