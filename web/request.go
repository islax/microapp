package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	microappError "github.com/islax/microapp/error"
)

// UnmarshalJSON checks for empty body and then parses JSON into the target
func UnmarshalJSON(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return microappError.NewInvalidRequestPayloadError(microappError.ErrorCodeEmptyRequestBody)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return microappError.NewDataReadWriteError(err)
	}

	if len(body) == 0 {
		return microappError.NewInvalidRequestPayloadError(microappError.ErrorCodeEmptyRequestBody)
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		return microappError.NewInvalidRequestPayloadError(microappError.ErrorCodeInvalidJSON)
	}
	return nil
}
