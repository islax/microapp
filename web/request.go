package web

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// UnmarshalJSON checks for empty body and then parses JSON into the target
func UnmarshalJSON(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return errors.New("Key_EmptyBody")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// microLog.Formatted().Errorf("%#v", err)
		return errors.New("Key_InternalError")
	}

	if len(body) == 0 {
		return errors.New("Key_EmptyBody")
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		// microLog.Formatted().Errorf("%#v", err)
		// microLog.Formatted().Printf("error decoding request: %v", err)
		// if e, ok := err.(*json.SyntaxError); ok {
		// 	log.Printf("syntax error at byte offset %d", e.Offset)
		// }
		// microLog.Formatted().Printf("request: %q", body)
		return errors.New("Key_InternalError")
	}
	return nil
}
