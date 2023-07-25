package api

import (
	"text2phenotype.com/fdl/logger"
	"github.com/rs/zerolog"
	"net/http"
)

var defaultLogger = logger.NewLogger("API")

type endpointLoggerFields struct {
	Method string `json:"method"`
	Url    string `json:"url"`
}

const RequestInfoFieldsKey = "request_info"

func makeRequestLogger(request *http.Request) zerolog.Logger {
	fields := endpointLoggerFields{
		Method: request.Method,
		Url:    request.URL.String(),
	}
	return defaultLogger.
		With().Interface(RequestInfoFieldsKey, fields).Logger()
}
