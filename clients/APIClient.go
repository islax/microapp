package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	microappCtx "github.com/islax/microapp/context"
)

// APIClient represents the actual client calling microservice
type APIClient struct {
	AppName    string
	BaseURL    string
	HTTPClient *http.Client
}

func (apiClient *APIClient) doRequest(context microappCtx.ExecutionContext, url string, requestMethod string, rawToken string, payload map[string]interface{}) (interface{}, error) {
	var body io.Reader
	if payload != nil {
		bytePayload, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(bytePayload)
	}

	request, err := http.NewRequest(requestMethod, apiClient.BaseURL+url, body)
	if err != nil {
		return nil, err
	}

	if rawToken != "" {
		if strings.HasPrefix(rawToken, "Bearer") {
			request.Header.Add("Authorization", rawToken)
		} else {
			request.Header.Add("Authorization", "Bearer "+rawToken)
		}
	}
	request.Header.Set("X-Client", apiClient.AppName)
	request.Header.Set("X-Correlation-ID", context.GetCorrelationID())
	request.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	response, err := apiClient.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if response.StatusCode > 300 { // All 3xx, 4xx, 5xx are considered errors
		return nil, errors.New("Received Status Code " + strconv.Itoa(response.StatusCode))
	}

	var mapResponse interface{}
	err = json.NewDecoder(response.Body).Decode(&mapResponse)
	if err != nil {
		return nil, err
	}

	return mapResponse, nil
}

// DoGet is a generic method to carry out RESTful calls to the other external microservices in ISLA
func (apiClient *APIClient) DoGet(context microappCtx.ExecutionContext, requestString string, rawToken string) (map[string]interface{}, error) {
	response, err := apiClient.doRequest(context, requestString, http.MethodGet, rawToken, nil)
	if err != nil {
		return nil, err
	}

	mapResponse, ok := response.(map[string]interface{})
	if !ok {
		return nil, errors.New("Could not parse Json to map")
	}
	return mapResponse, nil
}

// DoGetList is a generic method to carry out RESTful calls to the other external microservices in ISLA
func (apiClient *APIClient) DoGetList(context microappCtx.ExecutionContext, requestString string, rawToken string) ([]map[string]interface{}, error) {
	response, err := apiClient.doRequest(context, requestString, http.MethodGet, rawToken, nil)
	if err != nil {
		return nil, err
	}
	sliceOfGenericObjects, ok := response.([]interface{})
	if !ok {
		return nil, errors.New("Could not parse Json to map")
	}
	var sliceOfMapObjects []map[string]interface{}
	for _, obj := range sliceOfGenericObjects {
		mapObject, ok := obj.(map[string]interface{})
		if ok {
			sliceOfMapObjects = append(sliceOfMapObjects, mapObject)
		} else {
			return nil, errors.New("Could not parse Json to map")
		}
	}
	return sliceOfMapObjects, nil
}

// DoPost is a generic method to carry out RESTful calls to the other external microservices in ISLA
func (apiClient *APIClient) DoPost(context microappCtx.ExecutionContext, requestString string, rawToken string, payload map[string]interface{}) (map[string]interface{}, error) {
	response, err := apiClient.doRequest(context, requestString, http.MethodPost, rawToken, payload)
	if err != nil {
		return nil, err
	}

	mapResponse, ok := response.(map[string]interface{})
	if !ok {
		return nil, errors.New("Could not parse Json to map")
	}
	return mapResponse, nil
}

// DoDelete is a generic method to carry out RESTful calls to the other external microservices in ISLA
func (apiClient *APIClient) DoDelete(context microappCtx.ExecutionContext, requestString string, rawToken string, payload map[string]interface{}) error {
	_, err := apiClient.doRequest(context, requestString, http.MethodDelete, rawToken, payload)
	if err != nil {
		return err
	}
	return nil
}
