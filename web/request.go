package web

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	microLog "github.com/islax/microapp/log"
)

// UnmarshalJSON checks for empty body and then parses JSON into the target
func UnmarshalJSON(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return errors.New("Key_EmptyBody")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		microLog.Formatted().Errorf("%#v", err)
		return errors.New("Key_InternalError")
	}

	if len(body) == 0 {
		return errors.New("Key_EmptyBody")
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		microLog.Formatted().Errorf("%#v", err)
		return errors.New("Key_InternalError")
	}

	return nil
}
