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
	microappCtx "github.com/islax/microapp/context"
	microappLog "github.com/islax/microapp/log"
	"github.com/islax/microapp/repository"
	microappRepo "github.com/islax/microapp/repository"
	microappSecurity "github.com/islax/microapp/security"
	tenantService "github.com/islax/microapp/service"
	tenantModel "github.com/islax/microapp/settingsmetadata/model"
	microappWeb "github.com/islax/microapp/web"
	uuid "github.com/satori/go.uuid"
)

// NewSettingsMetadataController creates a new setting metadata controller
func NewSettingsMetadataController(app *microapp.App, repository microappRepo.Repository) *SettingsMetadataController {
	controller := &SettingsMetadataController{app: app, repository: repository}
	return controller

}

//SettingsMetadataController
type SettingsMetadataController struct {
	app                     *microapp.App
	repository              microappRepo.Repository
	settingsMetadatas       []tenantModel.SettingsMetaData
	globalsettingsMetadatas []tenantModel.SettingsMetaData
}

// RegisterRoutes implements interface RouteSpecifier
func (controller *SettingsMetadataController) RegisterRoutes(muxRouter *mux.Router) {
	apiRouter := muxRouter.PathPrefix("/api").Subrouter()
	policySettingsRouter := apiRouter.PathPrefix(fmt.Sprintf("/%s", strings.ToLower(controller.app.Name))).Subrouter()

	settingsMetadataRouter := policySettingsRouter.PathPrefix("/settings-metadata").Subrouter()
	settingsMetadataRouter.HandleFunc("", microappSecurity.Protect(controller.app.Config, controller.getSettingsMetadata, []string{"settingsmetadata:read"}, false)).Methods("GET")

	globalSettingsMetadataRouter := policySettingsRouter.PathPrefix("/global-settings-metadata").Subrouter()
	globalSettingsMetadataRouter.HandleFunc("", microappSecurity.Protect(controller.app.Config, controller.getGlobalSettingsMetadata, []string{"settingsmetadata:read"}, false)).Methods("GET")

	tenantSettingsRouter := policySettingsRouter.PathPrefix("/tenants/{id}/settings").Subrouter()
	tenantSettingsRouter.HandleFunc("", microappSecurity.Protect(controller.app.Config, controller.get, []string{"tenantsettings:read"}, false)).Methods("GET")
	tenantSettingsRouter.HandleFunc("", microappSecurity.Protect(controller.app.Config, controller.update, []string{"tenantsettings:write"}, false)).Methods("PUT")
	tenantSettingsRouter.HandleFunc("/{settingName}", microappSecurity.Protect(controller.app.Config, controller.getByName, []string{"tenantsettings:read"}, false)).Methods("GET")

}

func (controller *SettingsMetadataController) getGlobalSettingsMetadata(w http.ResponseWriter, r *http.Request, token *microappSecurity.JwtToken) {
	context := controller.app.NewExecutionContext(token, microapp.GetCorrelationIDFromRequest(r), "globalsettingsmetadata.get", false, false)
	if err := controller.checkAndInitializeSettingsMetadata(); err != nil {
		context.LogError(err, fmt.Sprintf(microappLog.MessageGenericErrorTemplate, "initializing settings-metadata"))
		microappWeb.RespondError(w, err)
		return
	}
	microappWeb.RespondJSON(w, http.StatusOK, controller.globalsettingsMetadatas)
}

func (controller *SettingsMetadataController) getSettingsMetadata(w http.ResponseWriter, r *http.Request, token *microappSecurity.JwtToken) {
	context := controller.app.NewExecutionContext(token, microapp.GetCorrelationIDFromRequest(r), "settingsmetadata.get", false, false)
	if err := controller.checkAndInitializeSettingsMetadata(); err != nil {
		context.LogError(err, fmt.Sprintf(microappLog.MessageGenericErrorTemplate, "initializing settings-metadata"))
		microappWeb.RespondError(w, err)
		return
	}
	microappWeb.RespondJSON(w, http.StatusOK, controller.settingsMetadatas)
}

func (controller *SettingsMetadataController) get(w http.ResponseWriter, r *http.Request, token *microappSecurity.JwtToken) {
	context := controller.app.NewExecutionContext(token, microapp.GetCorrelationIDFromRequest(r), "tenantsettings.get", true, true)
	uow := context.GetUOW()
	defer uow.Complete()
	params := mux.Vars(r)
	stringTenantID := params["id"]

	tenantID, err := tenantService.GetTenantIDFromToken().GetTenantIDAsUUID(mux.Vars(r), token, stringTenantID)
	if err != nil {
		context.LogError(err, microappLog.MessageUnableToFindURLResource)
		microappWeb.RespondError(w, err)
		return
	}

	if err := controller.checkAndInitializeSettingsMetadata(); err != nil {
		context.LogError(err, fmt.Sprintf(microappLog.MessageGenericErrorTemplate, "initializing settings-metadata"))
		microappWeb.RespondError(w, err)
		return
	}

	tenant, err := controller.getTenant(context, uow, controller.repository, tenantID)
	if err != nil {
		context.LogError(err, fmt.Sprintf(microappLog.MessageGenericErrorTemplate, "getting tenant from database"))
		microappWeb.RespondError(w, err)
		return
	}

	err = tenant.GetTenantSettings(controller.settingsMetadatas)
	if err != nil {
		context.LogError(err, fmt.Sprintf(microappLog.MessageGenericErrorTemplate, "getting tenant settings from database"))
		microappWeb.RespondError(w, err)
		return
	}

	microappWeb.RespondJSON(w, http.StatusOK, toDTO(tenant))
}

func (controller *SettingsMetadataController) update(w http.ResponseWriter, r *http.Request, token *microappSecurity.JwtToken) {
	context := controller.app.NewExecutionContext(token, microapp.GetCorrelationIDFromRequest(r), "tenantsettings.update", true, false)
	uow := context.GetUOW()
	defer uow.Complete()
	params := mux.Vars(r)
	stringTenantID := params["id"]
	var reqDTO tenantDTO
	if err := microappWeb.UnmarshalJSON(r, &reqDTO); err != nil {
		context.LogJSONParseError(err)
		microappWeb.RespondError(w, err)
		return
	}

	tenantID, err := tenantService.GetTenantIDFromToken().GetTenantIDAsUUID(mux.Vars(r), token, stringTenantID)
	if err != nil {
		context.LogError(err, microappLog.MessageUnableToFindURLResource)
		microappWeb.RespondError(w, err)
		return
	}

	if err := controller.checkAndInitializeSettingsMetadata(); err != nil {
		microappWeb.RespondError(w, err)
		return
	}

	settingsmetadata := controller.settingsMetadatas
	if tenantID.String() == "00000000-0000-0000-0000-000000000000" {
		settingsmetadata = controller.globalsettingsMetadatas
	}

	tenant, err := controller.getTenant(context, uow, controller.repository, tenantID)
	if err != nil {
		microappWeb.RespondError(w, err)
		return
	}

	if err = tenant.Update(reqDTO.Settings, settingsmetadata); err != nil {
		context.LogError(err, microappLog.MessageNewEntityError)
		microappWeb.RespondError(w, err)
		return
	}

	if tenant.Settings != "{}" {
		queryProcessor := []repository.QueryProcessor{repository.Filter("id = ?", tenantID)}
		err = controller.repository.Upsert(uow, &tenant, queryProcessor)
		if err != nil {
			context.LogError(err, microappLog.MessageUpdateEntityError)
			microappWeb.RespondError(w, err)
			return
		}
	} else {
		err = controller.repository.DeletePermanent(uow, tenantModel.TenantSettings{}, tenantID)
		if err != nil {
			context.LogError(err, microappLog.MessageUpdateEntityError)
			microappWeb.RespondError(w, err)
			return
		}
	}

	uow.Commit()
	responseDTO := toDTO(tenant)
	context.LoggerEventActionCompletion().Str("TenantId", responseDTO.ID.String()).Msg("Tenant settings updated")
	microappWeb.RespondJSON(w, http.StatusOK, responseDTO)
}

func (controller *SettingsMetadataController) getByName(w http.ResponseWriter, r *http.Request, token *microappSecurity.JwtToken) {
	context := controller.app.NewExecutionContext(token, microapp.GetCorrelationIDFromRequest(r), "tenantsettings.get", true, true)
	uow := context.GetUOW()
	defer uow.Complete()

	params := mux.Vars(r)
	stringTenantID := params["id"]
	//var settingsName interface{}
	tenantID, err := tenantService.GetTenantIDFromToken().GetTenantIDAsUUID(params, token, stringTenantID)
	if err != nil {
		context.LogError(err, microappLog.MessageUnableToFindURLResource)
		microappWeb.RespondError(w, err)
		return
	}

	tenant, err := controller.getTenant(context, uow, controller.repository, tenantID)
	if err != nil {
		microappWeb.RespondError(w, err)
		return
	}

	settings, err := tenant.GetSettings()
	if err != nil {
		context.LogError(err, microappLog.MessageGetEntityError)
		microappWeb.RespondError(w, err)
		return
	}

	settingsParameter := map[string]interface{}{params["settingName"]: settings[params["settingName"]]}
	microappWeb.RespondJSON(w, http.StatusOK, settingsParameter)
}

func (controller *SettingsMetadataController) getTenant(context microappCtx.ExecutionContext, uow *microappRepo.UnitOfWork, repository microappRepo.Repository, tenantID uuid.UUID) (*tenantModel.TenantSettings, error) {
	tenant := &tenantModel.TenantSettings{}
	tenant.ID = tenantID
	queryProcessor := []microappRepo.QueryProcessor{}
	queryProcessor = append(queryProcessor, microappRepo.Filter("id = ?", tenantID))
	if err := repository.GetFirst(uow, tenant, queryProcessor); err != nil {
		if !err.IsRecordNotFoundError() {
			context.LogError(err, fmt.Sprintf(microappLog.MessageGenericErrorTemplate, "getting tenant from database"))
			return nil, err
		}
	}
	return tenant, nil
}

func (controller *SettingsMetadataController) checkAndInitializeSettingsMetadata() error {
	if len(controller.settingsMetadatas) == 0 && controller.app.Config.IsSet(config.EvSuffixForSettingsMetadataPath) {
		settingMetadata, err := controller.initSettingsMetaData(config.EvSuffixForSettingsMetadataPath)
		if err != nil {
			return err
		}
		controller.settingsMetadatas = settingMetadata
	}
	if len(controller.globalsettingsMetadatas) == 0 && controller.app.Config.IsSet(config.EvSuffixForGlobalSettingsMetadataPath) {
		globalsettingMetadata, err := controller.initSettingsMetaData(config.EvSuffixForGlobalSettingsMetadataPath)
		if err != nil {
			return err
		}
		controller.globalsettingsMetadatas = globalsettingMetadata
	}
	return nil
}

func (controller *SettingsMetadataController) initSettingsMetaData(filePath string) ([]tenantModel.SettingsMetaData, error) {
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

func toDTO(tenant *tenantModel.TenantSettings) tenantDTO {
	settings := make(map[string]interface{})
	json.Unmarshal([]byte(tenant.Settings), &settings)
	return tenantDTO{
		ID:       tenant.ID,
		Settings: settings,
	}
}

type tenantDTO struct {
	ID       uuid.UUID              `json:"id"`
	Settings map[string]interface{} `json:"settings"`
}
