package context

import (
	microappError "github.com/islax/microapp/error"
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
	var executionCtxLogger zerolog.Logger
	if token != nil {
		executionCtxLogger = logger.With().
			Str("actionByTenantId", token.TenantID.String()).
			Str("actionByUserId", token.UserID.String()).
			Str("actionByUsername", token.UserName).
			Str("action", action).
			Str("corelationId", cid.String()).Logger()

	} else {
		executionCtxLogger = logger.With().
			Str("action", action).
			Str("corelationId", cid.String()).Logger()
	}

	return ExecutionContext{CorelationID: cid, UOW: uow, Token: token, Action: action, logger: executionCtxLogger}
}

// AddLoggerStrFields adds given string fields to the context logger
func (context *ExecutionContext) AddLoggerStrFields(strFields map[string]string) {
	loggerWith := context.logger.With()
	for k, v := range strFields {
		loggerWith = loggerWith.Str(k, v)
	}
	context.logger = loggerWith.Logger()
}

// Logger creates a logger with eventType and eventCode
func (context *ExecutionContext) Logger(eventType, eventCode string) *zerolog.Logger {
	logger := context.logger.With().Str("eventType", eventType).Str("eventCode", eventCode).Logger()
	return &logger
}

// LogError log error
func (context *ExecutionContext) LogError(err error, validationMessage, errorMessage string) {
	switch err.(type) {
	case microappError.ValidationError:
		context.Logger(log.EventTypeValidationErr, log.EventCodeInvalidData).Debug().Err(err).Msg(validationMessage)
	case microappError.HTTPResourceNotFound:
		resourceNotFoundErr := err.(microappError.HTTPResourceNotFound)
		context.Logger(log.EventTypeUnexpectedErr, resourceNotFoundErr.ErrorKey).Debug().Err(err).Str("resourceName", resourceNotFoundErr.ResourceName).Str("resourceValue", resourceNotFoundErr.ResourceValue).Msg(errorMessage)
	case microappError.UnexpectedError:
		context.Logger(log.EventTypeUnexpectedErr, err.(microappError.UnexpectedError).GetErrorCode()).Error().Err(err).Str("stack", err.(microappError.UnexpectedError).GetStackTrace()).Msg(errorMessage)
	default:
		context.Logger(log.EventTypeUnexpectedErr, log.EventCodeUnknown).Error().Err(err).Msg(errorMessage)
	}
}

// LogJSONParseError log JSON payload parsing error
func (context *ExecutionContext) LogJSONParseError(err error) {
	context.LogError(err, microappError.MessageInvalidPayload, microappError.MessageUnexpectedErrWhileRequetPayloadParsing)
}

// LoggerEventActionCompletion logger event with eventType success and eventCode action complete
func (context *ExecutionContext) LoggerEventActionCompletion() *zerolog.Event {
	logger := context.logger.Info().Str("eventType", log.EventTypeSuccess).Str("eventCode", log.EventCodeActionComplete)
	return logger
}
