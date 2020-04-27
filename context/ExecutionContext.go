package context

import (
	microappErrors "github.com/islax/microapp/errors"
	"github.com/islax/microapp/log"
	"github.com/islax/microapp/repository"
	"github.com/islax/microapp/security"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
)

// ExecutionContext execution context
type ExecutionContext struct {
	CorelationID uuid.UUID
	UOW          *repository.UnitOfWork
	Token        *security.JwtToken
	Action       string
	logger       zerolog.Logger
}

// NewExecutionContext creates new execution context
func NewExecutionContext(uow *repository.UnitOfWork, token *security.JwtToken, action string, logger zerolog.Logger) ExecutionContext {
	cid := uuid.NewV4()
	executionCtxLogger := logger.With().
		Str("tenantId", token.TenantID.String()).
		Str("actionByUserId", token.UserID.String()).
		Str("actionByUsername", token.UserName).
		Str("action", action).
		Str("corelationId", cid.String()).Logger()
	return ExecutionContext{CorelationID: cid, UOW: uow, Token: token, Action: action, logger: executionCtxLogger}
}

// Logger creates a logger with eventType and eventCode
func (context *ExecutionContext) Logger(eventType, eventCode string) *zerolog.Logger {
	logger := context.logger.With().Str("eventType", eventType).Str("eventCode", eventCode).Logger()
	return &logger
}

// LogError log error
func (context *ExecutionContext) LogError(err error, validationMessage, errorMessage string) {
	switch err.(type) {
	case microappErrors.ValidationError:
		context.Logger(log.EventTypeValidationErr, log.EventCodeInvalidData).Debug().Err(err).Msg(validationMessage)
	case microappErrors.HTTPResourceNotFound:
		resourceNotFoundErr := err.(microappErrors.HTTPResourceNotFound)
		context.Logger(log.EventTypeUnexpectedErr, resourceNotFoundErr.ErrorKey).Debug().Err(err).Str("resourceName", resourceNotFoundErr.ResourceName).Str("resourceName", resourceNotFoundErr.ResourceValue).Msg(errorMessage)
	case microappErrors.UnexpectedError:
		context.Logger(log.EventTypeUnexpectedErr, err.(microappErrors.UnexpectedError).GetErrorCode()).Error().Err(err).Str("stack", err.(microappErrors.UnexpectedError).GetStackTrace()).Msg(errorMessage)
	default:
		context.Logger(log.EventTypeUnexpectedErr, log.EventCodeUnknown).Error().Err(err).Msg(errorMessage)
	}
}

// AddLoggerStrFields adds given string fields to the context logger
func (context *ExecutionContext) AddLoggerStrFields(strFields map[string]string) {
	loggerWith := context.logger.With()
	for k, v := range strFields {
		loggerWith = loggerWith.Str(k, v)
	}
	context.logger = loggerWith.Logger()
}
