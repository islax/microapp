package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// APIClient represents the actual client calling microservice
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (apiClient *APIClient) doRequest(url string, requestMethod string, rawToken string, payload map[string]interface{}) (interface{}, error) {
	var body io.Reader
	if payload != nil {
		bytePayload, err := json.Marshal(payload)
		stringPayload := string(bytePayload)
		log.Info(stringPayload)

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
	request.Header.Set("Content-Type", "application/json")

	if err != nil {
		log.Error(err)
		return nil, err
	}

	response, err := apiClient.HTTPClient.Do(request)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer response.Body.Close()
	if response.StatusCode > 300 { // All 3xx, 4xx, 5xx are considered errors
		errorBytes, _ := ioutil.ReadAll(response.Body)
		errorString := string(errorBytes)
		log.Error(errorString)
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
func (apiClient *APIClient) DoGet(requestString string, rawToken string) (map[string]interface{}, error) {
	response, err := apiClient.doRequest(requestString, http.MethodGet, rawToken, nil)
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
func (apiClient *APIClient) DoGetList(requestString string, rawToken string) ([]map[string]interface{}, error) {
	response, err := apiClient.doRequest(requestString, http.MethodGet, rawToken, nil)
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
func (apiClient *APIClient) DoPost(requestString string, rawToken string, payload map[string]interface{}) (map[string]interface{}, error) {
	response, err := apiClient.doRequest(requestString, http.MethodPost, rawToken, payload)
	if err != nil {
		return nil, err
	}

	mapResponse, ok := response.(map[string]interface{})
	if !ok {
		return nil, errors.New("Could not parse Json to map")
	}
	return mapResponse, nil
}
