package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	microappErrors "github.com/islax/microapp/errors"
	microlog "github.com/islax/microapp/log"
)

// RespondJSON makes the response with payload as json format
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		microlog.Logger.Error([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

// RespondJSONWithXTotalCount makes the response with payload as json format and adds X-Total-Count header
func RespondJSONWithXTotalCount(w http.ResponseWriter, status int, count int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		microlog.Logger.Error([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Total-Count", strconv.Itoa(count))
	w.WriteHeader(status)
	w.Write([]byte(response))
}

// RespondErrorMessage makes the error response with payload as json format
func RespondErrorMessage(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"error": message})
}

// RespondError returns a validation error else
func RespondError(w http.ResponseWriter, err error) {
	switch err.(type) {
	case microappErrors.ValidationError:
		RespondJSON(w, http.StatusBadRequest, err)
	case microappErrors.HTTPResourceNotFound:
		RespondJSON(w, http.StatusNotFound, err)
	case microappErrors.HTTPError:
		httpError := err.(microappErrors.HTTPError)
		RespondErrorMessage(w, httpError.HTTPStatus, httpError.ErrorKey)
	default:
		RespondErrorMessage(w, http.StatusInternalServerError, "Key_InternalError")
	}
}
