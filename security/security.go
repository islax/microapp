package security

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/islax/microapp/config"
	"github.com/islax/microapp/web"

	jwt "github.com/dgrijalva/jwt-go"
)

// Protect authenticates and makes sure that caller is authorized to make the call before
// before invoking actual handler
func Protect(config *config.Config, handlerFunc func(w http.ResponseWriter, r *http.Request, token *JwtToken), allowedScopes []string, requireAdmin bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")
		token, err := validateTokenHeader(config, tokenHeader)

		if err != nil {
			web.RespondErrorMessage(w, http.StatusUnauthorized, err.Error())
			return
		}

		if requireAdmin && token.Admin != true {
			web.RespondErrorMessage(w, http.StatusForbidden, "Key_InsufficientCredentials")
			return
		}

		if !token.isValidForScope(allowedScopes) {
			web.RespondErrorMessage(w, http.StatusForbidden, "Key_Unauthorized")
			return
		}

		token.Raw = tokenHeader
		handlerFunc(w, r, token)
	}
}

func validateTokenHeader(config *config.Config, tokenHeader string) (*JwtToken, error) {
	if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
		return nil, errors.New("Key_MissingAuthToken")
	}
	splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
	if len(splitted) != 2 {
		return nil, errors.New("Key_InvalidAuthToken")
	}

	tokenPart := splitted[1] //Grab the token part, what we are truly interested in
	tk := &JwtToken{}

	pubkeyBytes, _ := ioutil.ReadFile(os.Getenv("ISLA_JWT_PUBLIC_KEY_PATH"))
	jwtPublicKey, _ := jwt.ParseRSAPublicKeyFromPEM(pubkeyBytes)

	token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
		return jwtPublicKey, nil
	})

	if err != nil { //Malformed token, returns with http code 403 as usual
		return nil, errors.New("Key_InvalidAuthToken")
	}

	if !token.Valid { //Token is invalid, maybe not signed on this server
		return nil, errors.New("Key_InvalidAuthToken")
	}

	return tk, nil
}
