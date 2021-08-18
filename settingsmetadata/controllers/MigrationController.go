package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/islax/microapp"
	"github.com/islax/microapp/config"
	microappLog "github.com/islax/microapp/log"
	microappRepo "github.com/islax/microapp/repository"
	microappSecurity "github.com/islax/microapp/security"
	"github.com/islax/microapp/settingsmetadata/clients"
	tenantModel "github.com/islax/microapp/settingsmetadata/model"
	microappWeb "github.com/islax/microapp/web"
	uuid "github.com/satori/go.uuid"
)

// NewSettingsMetadataController creates a new setting metadata controller
func NewSettingsMetadataMigrationController(app *microapp.App, repository microappRepo.Repository, tenantClient clients.TenantClient) *SettingsMetadataMigrationController {
	controller := &SettingsMetadataMigrationController{app: app, repository: repository, tenantClient: tenantClient}
	return controller

}

//SettingsMetadataMigrationController
type SettingsMetadataMigrationController struct {
	app               *microapp.App
	repository        microappRepo.Repository
	settingsMetadatas []tenantModel.SettingsMetaData
	tenantClient      clients.TenantClient
}

// RegisterRoutes implements interface RouteSpecifier
func (controller *SettingsMetadataMigrationController) RegisterRoutes(muxRouter *mux.Router) {
	apiRouter := muxRouter.PathPrefix("/api").Subrouter()
	tenantSettingsRouter := apiRouter.PathPrefix(fmt.Sprintf("/%s", strings.ToLower(controller.app.Name))).Subrouter()

	migrationsRouter := tenantSettingsRouter.PathPrefix("/tenants/migrate/").Subrouter()
	migrationsRouter.HandleFunc("", microappSecurity.Protect(controller.app.Config, controller.migratetenants, []string{"settingsmetadata:read"}, false)).Methods("PUT")
	migrationsRouter.HandleFunc("{id}", microappSecurity.Protect(controller.app.Config, controller.migratetenant, []string{"settingsmetadata:read"}, false)).Methods("PUT")
}

func (controller *SettingsMetadataMigrationController) migratetenants(w http.ResponseWriter, r *http.Request, token *microappSecurity.JwtToken) {
	context := controller.app.NewExecutionContext(token, microapp.GetCorrelationIDFromRequest(r), "tenantsettings.migrate", true, false)
	uow := context.GetUOW()
	defer uow.Complete()

	if err := controller.checkAndInitializeSettingsMetadata(); err != nil {
		context.LogError(err, fmt.Sprintf(microappLog.MessageGenericErrorTemplate, "initializing settings-metadata"))
		microappWeb.RespondError(w, err)
		return
	}

	tenantSettings, err := controller.tenantClient.GetAllTenants(context, token.Raw)
	if err != nil {
		context.LogError(err, microappLog.MessageUnableToFindURLResource)
		microappWeb.RespondError(w, err)
		return
	}

	successTenants := make([]string, 0)
	failureTenants := make([]string, 0)
	for _, tenantMap := range tenantSettings {
		tenantIDStr := tenantMap["id"].(string)
		tenantID, err := uuid.FromString(tenantMap["id"].(string))
		if err != nil {
			context.LogError(err, "Unable to get tenant id")
			failureTenants = append(failureTenants, tenantIDStr)
			continue
		}
		settings := tenantMap["settings"].(map[string]interface{})
		tenant, err := tenantModel.NewTenant(context, tenantID, settings, controller.settingsMetadatas)
		if err != nil {
			context.LogError(err, "Unable to add new tenant.")
			failureTenants = append(failureTenants, tenantIDStr)
			continue
		}
		if tenant.Settings != "{}" {
			if err := controller.repository.Add(uow, tenant); err != nil {
				context.LogError(err, "Unable to add tenant settings.")
				failureTenants = append(failureTenants, tenantIDStr)
				continue
			}
		}
		successTenants = append(successTenants, tenantIDStr)
	}

	microappWeb.RespondJSON(w, http.StatusOK, map[string]interface{}{"successTenants": successTenants, "failureTenants": failureTenants})
}

func (controller *SettingsMetadataMigrationController) migratetenant(w http.ResponseWriter, r *http.Request, token *microappSecurity.JwtToken) {
	context := controller.app.NewExecutionContext(token, microapp.GetCorrelationIDFromRequest(r), "tenantsettings.migrate", true, false)
	uow := context.GetUOW()
	defer uow.Complete()

	params := mux.Vars(r)
	stringTenantID := params["id"]

	if err := controller.checkAndInitializeSettingsMetadata(); err != nil {
		context.LogError(err, fmt.Sprintf(microappLog.MessageGenericErrorTemplate, "initializing settings-metadata"))
		microappWeb.RespondError(w, err)
		return
	}

	tenantMap, err := controller.tenantClient.GetTenant(context, token.Raw, stringTenantID)
	if err != nil {
		context.LogError(err, microappLog.MessageUnableToFindURLResource)
		microappWeb.RespondError(w, err)
		return
	}

	tenantID, err := uuid.FromString(tenantMap["id"].(string))
	if err != nil {
		context.LogError(err, "Unable to get tenant id")
		microappWeb.RespondError(w, err)
		return
	}
	settings := tenantMap["settings"].(map[string]interface{})
	tenant, err := tenantModel.NewTenant(context, tenantID, settings, controller.settingsMetadatas)
	if err != nil {
		context.LogError(err, "Unable to add new tenant.")
		microappWeb.RespondError(w, err)
		return
	}
	if tenant.Settings != "{}" {
		if err := controller.repository.Add(uow, tenant); err != nil {
			context.LogError(err, "Unable to add tenant settings.")
			microappWeb.RespondError(w, err)
			return
		}
	}

	microappWeb.RespondJSON(w, http.StatusOK, "")
}

func (controller *SettingsMetadataMigrationController) checkAndInitializeSettingsMetadata() error {
	if len(controller.settingsMetadatas) == 0 && controller.app.Config.IsSet(config.EvSuffixForSettingsMetadataPath) {
		settingMetadata, err := controller.initSettingsMetaData(config.EvSuffixForSettingsMetadataPath)
		if err != nil {
			return err
		}
		controller.settingsMetadatas = settingMetadata
	}
	return nil
}

func (controller *SettingsMetadataMigrationController) initSettingsMetaData(filePath string) ([]tenantModel.SettingsMetaData, error) {
	var settingsmetadata []tenantModel.SettingsMetaData
	jsonFile, err := os.Open(controller.app.Config.GetString(filePath))
	if err != nil {
		return settingsmetadata, err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return settingsmetadata, err
	}
	json.Unmarshal(byteValue, &settingsmetadata)
	return settingsmetadata, nil
}
