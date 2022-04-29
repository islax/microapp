package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	microappCtx "github.com/islax/microapp/context"
	microappError "github.com/islax/microapp/error"
)

// APIClient represents the actual client calling microservice
type APIClient struct {
	AppName    string
	BaseURL    string
	HTTPClient *http.Client
}

func (apiClient *APIClient) getJSONRequestBody(payload interface{}) (io.Reader, error) {
	var reader io.Reader
	if payload != nil {
		payloadByteArray, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewBuffer(payloadByteArray)
	}
	return reader, nil
}

// DoRequestBasic ...
func (apiClient *APIClient) DoRequestBasic(context microappCtx.ExecutionContext, uri string, requestMethod string, rawToken string, payload interface{}) (*http.Response, microappError.APIClientError) {
	apiURL := apiClient.BaseURL + uri

	payloadAsIOReader, err := apiClient.getJSONRequestBody(payload)
	if err != nil {
		return nil, microappError.NewAPIClientError(apiURL, nil, nil, fmt.Errorf("unable to encode payload: %w", err))
	}

	// Not checking for error here, as request and apiURL are internal values and body is already checked for err above.
	request, _ := http.NewRequest(requestMethod, url.QueryEscape(apiURL), payloadAsIOReader)

	// Set Authorization header
	if rawToken != "" {
		if strings.HasPrefix(rawToken, "Bearer") {
			request.Header.Set("Authorization", rawToken)
		} else {
			request.Header.Set("Authorization", "Bearer "+rawToken)
		}
	}

	// Set other headers
	request.Header.Set("X-Client", apiClient.AppName)
	request.Header.Set("X-Correlation-ID", context.GetCorrelationID())
	request.Header.Set("Content-Type", "application/json")

	response, err := apiClient.HTTPClient.Do(request)
	if err != nil {
		return nil, microappError.NewAPIClientError(apiURL, nil, nil, fmt.Errorf("unable to invoke API: %w", err))
	}

	return response, nil
}

// DoRequestProxy do request with response param
func (apiClient *APIClient) DoRequestProxy(context microappCtx.ExecutionContext, r *http.Request, url string, rawToken string) (*http.Response, microappError.APIClientError) {
	apiURL := apiClient.BaseURL
	if url != "" {
		apiURL = apiURL + url
	} else {
		apiURL = apiURL + r.URL.Path
	}

	request, err := http.NewRequest(r.Method, apiURL, r.Body)
	if err != nil {
		return nil, microappError.NewAPIClientError(apiURL, nil, nil, fmt.Errorf("unable to create HTTP request: %w", err))
	}

	request.Header.Set("X-Client", apiClient.AppName)
	request.Header.Set("X-Correlation-ID", context.GetCorrelationID())
	request.Header.Set("Content-Type", "application/json")

	if rawToken != "" {
		if strings.HasPrefix(rawToken, "Bearer") {
			request.Header.Set("Authorization", rawToken)
		} else {
			request.Header.Set("Authorization", "Bearer "+rawToken)
		}
	} else if r.Header.Get("Authorization") != "" {
		request.Header.Set("Authorization", r.Header.Get("Authorization"))
	}

	response, err := apiClient.HTTPClient.Do(request)
	if err != nil {
		return nil, microappError.NewAPIClientError(apiURL, nil, nil, fmt.Errorf("unable to invoke API: %w", err))
	}
	return response, nil
}

// DoRequestWithResponseParam do request with response param
func (apiClient *APIClient) DoRequestWithResponseParam(context microappCtx.ExecutionContext, url string, requestMethod string, rawToken string, payload interface{}, out interface{}) microappError.APIClientError {
	apiURL := apiClient.BaseURL + url

	response, apiClientErr := apiClient.DoRequestBasic(context, url, requestMethod, rawToken, payload)
	if apiClientErr != nil {
		return apiClientErr
	}
	defer response.Body.Close()
	if response.StatusCode > 300 { // All 3xx, 4xx, 5xx are considered errors
		responseBodyString := ""
		if responseBodyBytes, err := ioutil.ReadAll(response.Body); err == nil {
			responseBodyString = string(responseBodyBytes)
		}
		return microappError.NewAPIClientError(apiURL, &response.StatusCode, &responseBodyString, fmt.Errorf("received non-success code: %v", response.StatusCode))
	}

	if out != nil {
		if err := json.NewDecoder(response.Body).Decode(out); err != nil {
			return microappError.NewAPIClientError(apiURL, &response.StatusCode, nil, fmt.Errorf("unable parse response payload: %w", err))
		}
	}
	return nil
}

func (apiClient *APIClient) doRequest(context microappCtx.ExecutionContext, url string, requestMethod string, rawToken string, payload map[string]interface{}) (interface{}, error) {
	apiURL := apiClient.BaseURL + url

	response, apiClientErr := apiClient.DoRequestBasic(context, url, requestMethod, rawToken, payload)
	if apiClientErr != nil {
		return nil, apiClientErr
	}

	defer response.Body.Close()
	if response.StatusCode > 300 { // All 3xx, 4xx, 5xx are considered errors
		responseBodyString := ""
		if responseBodyBytes, err := ioutil.ReadAll(response.Body); err == nil {
			responseBodyString = string(responseBodyBytes)
		}
		return nil, microappError.NewAPIClientError(apiURL, &response.StatusCode, &responseBodyString, fmt.Errorf("received non-success code: %v", response.StatusCode))
	}

	var mapResponse interface{}
	if err := json.NewDecoder(response.Body).Decode(&mapResponse); err != nil {
		return nil, microappError.NewAPIClientError(apiURL, &response.StatusCode, nil, fmt.Errorf("unable parse response payload: %w", err))
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
		return nil, errors.New("could not parse Json to map")
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
		return nil, errors.New("could not parse Json to map")
	}
	var sliceOfMapObjects []map[string]interface{}
	for _, obj := range sliceOfGenericObjects {
		mapObject, ok := obj.(map[string]interface{})
		if ok {
			sliceOfMapObjects = append(sliceOfMapObjects, mapObject)
		} else {
			return nil, errors.New("could not parse Json to map")
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
		return nil, errors.New("could not parse Json to map")
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
