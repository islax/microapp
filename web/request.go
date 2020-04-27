package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	microAppErrors "github.com/islax/microapp/errors"
)

// UnmarshalJSON checks for empty body and then parses JSON into the target
func UnmarshalJSON(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return microAppErrors.NewValidationError("Key_InvalidPayload", map[string]string{"payload": "Key_EmptyBody"})
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return microAppErrors.NewDataReadWriteError(err)
	}

	if len(body) == 0 {
		return microAppErrors.NewValidationError("Key_InvalidPayload", map[string]string{"payload": "Key_EmptyBody"})
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		return microAppErrors.NewValidationError("Key_InvalidPayload", map[string]string{"payload": "Key_InvalidJSON"})
	}
	return nil
}
