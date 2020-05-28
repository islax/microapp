package microapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Used
	"github.com/rs/zerolog"
)

// TestApp Provides convinience methods for test
type TestApp struct {
	application             *App
	controllerRouteProvider func(*App) []RouteSpecifier
	dbInitializer           func(db *gorm.DB)
}

// NewTestApp returns new instance of TestApp
func NewTestApp(appName string, controllerRouteProvider func(*App) []RouteSpecifier, dbInitializer func(db *gorm.DB), verbose bool) *TestApp {
	dbFile := "./test_islax.db"
	db, err := gorm.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}

	db.LogMode(verbose)

	logger := zerolog.New(os.Stdout)
	application := New(appName, map[string]interface{}{}, logger, db, nil)

	return &TestApp{application: application, controllerRouteProvider: controllerRouteProvider, dbInitializer: dbInitializer}
}

// Initialize prepares the app for testing
func (testApp *TestApp) Initialize() {
	testApp.application.Initialize(testApp.controllerRouteProvider(testApp.application))
	testApp.PrepareEmptyTables()
	go testApp.application.Start()
}

// Stop the app
func (testApp *TestApp) Stop() {
	testApp.application.Stop()
	testApp.application.DB.Close()

	os.Remove("./test_islax.db")
}

// PrepareEmptyTables clears all table of data
func (testApp *TestApp) PrepareEmptyTables() {
	testApp.dbInitializer(testApp.application.DB)
}

// ExecuteRequest executes the http request
func (testApp *TestApp) ExecuteRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	testApp.application.Router.ServeHTTP(rr, req)

	return rr
}

// AddAssociations adds associations to the given entity
func (testApp *TestApp) AddAssociations(entity interface{}, associationName string, associations ...interface{}) error {
	return testApp.application.DB.Model(entity).Association(associationName).Append(associations...).Error
}

// AssertEqualWithFieldsToIgnore asserts whether two objects are equal
func (testApp *TestApp) AssertEqualWithFieldsToIgnore(t *testing.T, expected interface{}, actual interface{}, fieldsToIgnore []string, mapOfExpectedToActualField map[string]string) {
	expectedElems := reflect.ValueOf(expected).Elem()
	actualElem := reflect.ValueOf(actual).Elem()
	typeOfExpectedElems := reflect.ValueOf(expected).Elem().Type()

	mapOfIgnoreFields := make(map[string]bool)
	for _, ignoreField := range fieldsToIgnore {
		mapOfIgnoreFields[ignoreField] = true
	}

	for i := 0; i < expectedElems.NumField(); i++ {
		expectedFieldName := typeOfExpectedElems.Field(i).Name
		// t.Logf("Expected assert field: %v", expectedFieldName)
		if !mapOfIgnoreFields[expectedFieldName] {
			expectedField := expectedElems.Field(i)
			expectedFieldMetadata := typeOfExpectedElems.Field(i).Type

			if expectedFieldMetadata.Kind() == reflect.Struct {
				for j := 0; j < expectedFieldMetadata.NumField(); j++ {
					if !mapOfIgnoreFields[expectedFieldMetadata.Field(j).Name] {
						expectedValue := testApp.getReflectFieldValueAsString(expectedField.Field(j), expectedFieldMetadata.Field(j).Type)

						actualFieldName := expectedFieldMetadata.Field(j).Name
						// t.Logf("Expected assert field (nested): %v", actualFieldName)
						if v, ok := mapOfExpectedToActualField[actualFieldName]; ok {
							actualFieldName = v
						}
						actualField := actualElem.FieldByName(actualFieldName)
						if actualField.Kind() == reflect.Ptr {
							actualField = actualField.Elem()
						}
						actualValue := fmt.Sprintf("%v", actualField.Interface())

						if expectedValue != actualValue {
							t.Errorf("Expected %v [%v], Actual [%v]!", actualFieldName, expectedValue, actualValue)
						}
					}
				}
			} else {
				expectedValue := testApp.getReflectFieldValueAsString(expectedField, expectedFieldMetadata)

				actualFieldName := expectedFieldName
				if v, ok := mapOfExpectedToActualField[actualFieldName]; ok {
					actualFieldName = v
				}
				// t.Logf("Actual assert field: %v", actualFieldName)
				actualField := actualElem.FieldByName(actualFieldName)
				if actualField.Kind() == reflect.Ptr {
					actualField = actualField.Elem()
				}
				actualValue := fmt.Sprintf("%v", actualField.Interface())

				if expectedValue != actualValue {
					t.Errorf("Expected %v [%v], Actual [%v]!", actualFieldName, expectedValue, actualValue)
				}
			}
		}
	}
}

// AssertEqualWithFieldsToCheck asserts whether two objects are equal
func (testApp *TestApp) AssertEqualWithFieldsToCheck(t *testing.T, expected interface{}, actual interface{}, fieldsToChk []string, mapOfExpectedToActualField map[string]string) {
	expectedElems := reflect.ValueOf(expected).Elem()
	actualElems := reflect.ValueOf(actual).Elem()
	typeOfExpectedElems := reflect.ValueOf(expected).Elem().Type()

	for _, attrToChk := range fieldsToChk {
		expectedFieldMetadata, _ := typeOfExpectedElems.FieldByName(attrToChk)
		expectedField := expectedElems.FieldByName(attrToChk)
		if expectedField.Kind() == reflect.Ptr {
			expectedField = expectedField.Elem()
		}
		var expectedValue string
		if fmt.Sprintf("%v", expectedFieldMetadata.Type) == "time.Time" {
			expectedValue = expectedField.Interface().(time.Time).Format(time.RFC3339)
		} else {
			expectedValue = fmt.Sprintf("%v", expectedField.Interface())
		}

		actualAttrName := attrToChk
		if v, ok := mapOfExpectedToActualField[attrToChk]; ok {
			actualAttrName = v
		}
		actualField := actualElems.FieldByName(actualAttrName)
		if actualField.Kind() == reflect.Ptr {
			actualField = actualField.Elem()
		}
		actualValue := fmt.Sprintf("%v", actualField.Interface())

		if expectedValue != actualValue {
			t.Errorf("Expected %v [%v], Actual [%v]!", attrToChk, expectedValue, actualValue)
		}
	}
}

// AssertErrorResponse checks if the http response contains expected errorKey, errorField and errorMessage
func (testApp *TestApp) AssertErrorResponse(t *testing.T, response *httptest.ResponseRecorder, expectedErrorKey string, expectedErrorField string, expectedError string) {
	testApp.CheckResponseCode(t, http.StatusBadRequest, response.Code)
	var errData map[string]interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &errData); err != nil {
		t.Errorf("Unable to parse response: %v", err)
	}
	if errData["errorKey"] != expectedErrorKey {
		t.Errorf("Expected errorKey [%v], Got [%v]!", expectedErrorKey, errData["errorKey"])
	}
	errors := errData["errors"].(map[string]interface{})
	if fmt.Sprintf("%v", errors[expectedErrorField]) != expectedError {
		t.Errorf("Expected error [%v], Got [%v]!", expectedError, errors[expectedErrorField])
	}
}

// AssertXTotalCount checks if the http response header contains expected x-total-count
func (testApp *TestApp) AssertXTotalCount(t *testing.T, response *httptest.ResponseRecorder, expectedXTotalCount int) {
	xTotalCount := response.Header().Get("X-Total-Count")
	if xTotalCount == "" {
		t.Fatalf("Expected X-Total-Count [%v], not found!", expectedXTotalCount)
	}

	xTotalCountInt, err := strconv.Atoi(xTotalCount)
	if err != nil {
		t.Fatalf("Expected X-Total-Count [%v], Actual [%v]!", expectedXTotalCount, xTotalCount)
	} else if expectedXTotalCount != xTotalCountInt {
		t.Fatalf("Expected X-Total-Count [%v], Actual [%v]!", expectedXTotalCount, xTotalCountInt)
	}
}

// CallAPI invokes http API
func (testApp *TestApp) CallAPI(httpMethod string, apiURL string, token string, req interface{}) *httptest.ResponseRecorder {

	var payload io.Reader
	if req != nil {
		reqJSON, _ := json.Marshal(req)
		payload = bytes.NewBuffer(reqJSON)
	}

	httpReq, _ := http.NewRequest(httpMethod, apiURL, payload)

	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	return testApp.ExecuteRequest(httpReq)
}

// CheckResponseCode checks if the http response is as expected
func (testApp *TestApp) CheckResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Fatalf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// Check checks whether the expected and actual value matches
func (testApp *TestApp) Check(t *testing.T, name string, expected, actual interface{}, failNow bool) {
	if !reflect.DeepEqual(expected, actual) {
		if failNow {
			t.Fatalf("Expected %v [%v]. Got [%v]\n", name, expected, actual)
		} else {
			t.Errorf("Expected %v [%v]. Got [%v]\n", name, expected, actual)
		}
	}
}

// GetAdminToken returns a test token
func (testApp *TestApp) GetAdminToken(tenantID string, userID string, scope []string) string {
	return testApp.generateToken(tenantID, userID, "", "", uuid.UUID{}.String(), "", scope, true)
}

// GetAll gets all from DB
func (testApp *TestApp) GetAll(out interface{}, preloads []string, whereClause string, whereParams []interface{}, orderBy string) error {
	db := testApp.application.DB
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if strings.TrimSpace(whereClause) != "" {
		db = db.Where(whereClause, whereParams...)
	}
	if strings.TrimSpace(orderBy) != "" {
		db = db.Order(orderBy, true)
	}
	return db.Find(out).Error
}

// GetByID gets entity by ids
func (testApp *TestApp) GetByID(out interface{}, preloads []string, id string) error {
	db := testApp.application.DB
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	return db.Where("id = ?", id).Find(out).Error
}

// GetFullAdminToken returns a test token with all the fields along with different external IDs for types such as Appliance, Session, User. These external IDs are used with REST api is invoked from another REST API service as opposed to the getting hit from UI by the user.
func (testApp *TestApp) GetFullAdminToken(tenantID string, userID string, username string, name string, externalID string, externalIDType string, scope []string) string {
	return testApp.generateToken(tenantID, userID, username, name, externalID, externalIDType, scope, true)
}

// GetStandardAdminToken returns a test admin token with all standard fields.
func (testApp *TestApp) GetStandardAdminToken(tenantID string, userID string, username string, name string, scope []string) string {
	return testApp.generateToken(tenantID, userID, username, name, uuid.Nil.String(), "", scope, true)
}

// GetFullToken returns a test token with all the fields along with different external IDs for types such as Appliance, Session, User. These external IDs are used with REST api is invoked from another REST API service as opposed to the getting hit from UI by the user.
func (testApp *TestApp) GetFullToken(tenantID string, userID string, username string, name string, externalID string, externalIDType string, scope []string) string {
	return testApp.generateToken(tenantID, userID, username, name, externalID, externalIDType, scope, false)
}

// GetStandardToken returns a test token with all standard fields
func (testApp *TestApp) GetStandardToken(tenantID string, userID string, username string, name string, scope []string) string {
	return testApp.generateToken(tenantID, userID, username, name, uuid.Nil.String(), "", scope, false)
}

// GetToken gets a token to connect to API
func (testApp *TestApp) GetToken(tenantID string, userID string, scope []string) string {
	return testApp.generateToken(tenantID, userID, userID, userID, uuid.Nil.String(), "", scope, false)
}

// SaveToDB saves the entity to database
func (testApp *TestApp) SaveToDB(entity interface{}) error {
	return testApp.application.DB.Create(entity).Error
}

// SetControllerRouteProviderAndInitialize sets the controllerRouteProvider and initializes application
func (testApp *TestApp) SetControllerRouteProviderAndInitialize(controllerRouteProvider func(*App) []RouteSpecifier) {
	testApp.controllerRouteProvider = controllerRouteProvider
	testApp.PrepareEmptyTables()
	testApp.application.Initialize(testApp.controllerRouteProvider(testApp.application))

	go testApp.application.Start()
}

// generateToken generates and return token
func (testApp *TestApp) generateToken(tenantID string, userID string, username string, name string, externalID string, externalIDType string, scope []string, admin bool) string {
	hmacSampleSecret := []byte(testApp.application.Config.GetString("JWT_SECRET"))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":              "http://isla.cyberinc.com",
		"aud":              "http://isla.cyberinc.com",
		"iat":              time.Now().Unix(),
		"exp":              time.Now().Add(time.Minute * 60).Unix(), // Expires in 1 hour
		"tenant":           tenantID,
		"user":             userID,
		"admin":            admin,
		"name":             username,
		"displayName":      name,
		"scope":            scope,
		"externalId":       externalID,
		"externalIdType":   externalIDType,
		"identityProvider": "",
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(hmacSampleSecret)

	if err != nil {
		panic(err)
	}

	return tokenString
}

func (testApp *TestApp) getReflectFieldValueAsString(fieldElem reflect.Value, fieldType reflect.Type) string {
	var strValue string
	if fieldElem.Kind() == reflect.Ptr {
		fieldElem = fieldElem.Elem()
	}

	if fmt.Sprintf("%v", fieldType) == "time.Time" {
		strValue = fieldElem.Interface().(time.Time).Format(time.RFC3339)
	} else {
		strValue = fmt.Sprintf("%v", fieldElem.Interface())
	}
	return strValue
}
