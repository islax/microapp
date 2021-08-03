package eventhandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/islax/microapp"
	microappCtx "github.com/islax/microapp/context"
	"github.com/islax/microapp/event/monitor"
	microappLog "github.com/islax/microapp/log"
	microappRepo "github.com/islax/microapp/repository"
	microappSecurity "github.com/islax/microapp/security"
	tenantModel "github.com/islax/microapp/settingsmetadata/model"
	uuid "github.com/satori/go.uuid"
)

//EventHandler handles events
type EventHandler struct {
	app          *microapp.App
	repository   microappRepo.Repository
	eventChannel chan *monitor.EventInfo
}

// NewEventHandler creates new instance of TenantActionEventHandler
func NewEventHandler(app *microapp.App, repository microappRepo.Repository, eventChannel chan *monitor.EventInfo) *EventHandler {
	return &EventHandler{app, repository, eventChannel}
}

// Start will start listening to channel for events
func (handler *EventHandler) Start() {
	for eventPayload := range handler.eventChannel {
		switch eventPayload.Name {
		case "tenant.added":
			handler.processTenantAdd(eventPayload)
		case "tenant.deleted":
			handler.processTenantDelete(eventPayload)
		}
	}
}

func (handler *EventHandler) processTenantAdd(eventPayload *monitor.EventInfo) {
	token, _ := microappSecurity.GetTokenFromRawAuthHeader(handler.app.Config, eventPayload.RawToken)
	context := handler.app.NewExecutionContext(token, eventPayload.CorelationID, "tenantsettings.add", true, false)
	uow := context.GetUOW()
	defer uow.Complete()
	handler.createTenantSettingsMetadata(uow, context, eventPayload)
	uow.Commit()
	context.Logger(microappLog.EventTypeSuccess, microappLog.EventCodeActionComplete).Info().Msg("Finished adding new tenant settings")
}

func (handler *EventHandler) createTenantSettingsMetadata(uow *microappRepo.UnitOfWork, context microappCtx.ExecutionContext, eventPayload *monitor.EventInfo) {
	eventData := make(map[string]interface{})
	var settingsmetadata []tenantModel.SettingsMetaData
	var tenantID uuid.UUID

	jsonFile, err := os.Open(handler.app.Config.GetString("SETTINGS_METADATA_PATH"))
	if err != nil {
		context.LogError(err, fmt.Sprintf(microappLog.MessageGenericErrorTemplate, "opening settings-metadata config file."))
		return
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		context.LogError(err, fmt.Sprintf(microappLog.MessageGenericErrorTemplate, "reading tenant role config file."))
		return
	}
	json.Unmarshal(byteValue, &settingsmetadata)

	json.Unmarshal([]byte(eventPayload.Payload), &eventData)
	tenantID, _ = uuid.FromString(eventData["id"].(string))
	tenant, err := tenantModel.NewTenant(context, tenantID, settingsmetadata)
	if err != nil {
		context.LogError(err, "Unable to add new tenant.")
		return
	}
	if err := handler.repository.Add(uow, tenant); err != nil {
		context.LogError(err, "Unable to add tenant settings.")
		return
	}
}

func (handler *EventHandler) processTenantDelete(eventPayload *monitor.EventInfo) {
	token, _ := microappSecurity.GetTokenFromRawAuthHeader(handler.app.Config, eventPayload.RawToken)
	context := handler.app.NewExecutionContext(token, eventPayload.CorelationID, "tenantsettings.delete", true, false)
	uow := context.GetUOW()
	defer uow.Complete()

	eventData := make(map[string]string)
	json.Unmarshal([]byte(eventPayload.Payload), &eventData)

	tenantID, _ := uuid.FromString(eventData["id"])
	if err := handler.repository.Delete(uow, tenantModel.TenantSettings{}, tenantID); err != nil {
		context.Logger(microappLog.EventTypeServiceDataReplication, "Key_TenantDataReplication").Error().Err(err).Str("forTenant", tenantID.String()).Msg("Unable to delete tenant.")
		return
	}

	uow.Commit()
	context.LoggerEventActionCompletion().Msg("Tenant deleted.")
}
