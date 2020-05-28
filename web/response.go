package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	microappError "github.com/islax/microapp/error"
)

// RespondJSON makes the response with payload as json format
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
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
	case microappError.ValidationError:
		RespondJSON(w, http.StatusBadRequest, err)
	case microappError.HTTPResourceNotFound:
		RespondJSON(w, http.StatusNotFound, err)
	case microappError.HTTPError:
		httpError := err.(microappError.HTTPError)
		RespondErrorMessage(w, httpError.HTTPStatus, httpError.ErrorKey)
	default:
		RespondErrorMessage(w, http.StatusInternalServerError, microappError.ErrorCodeInternalError)
	}
}
