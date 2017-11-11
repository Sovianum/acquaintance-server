package mylog

import (
	golog "github.com/op/go-logging"
	"io"
	"net/http"
)

const (
	requestStartLogTemplate   = `Started handling request to url %v with method %v`
	requestSuccessLogTemplate = `Request to url %v with method %v handled successfully`
	requestBodyTemplate = "Request to url %v with method %v has body %v"
	requestErrorTemplate      = `Failed on URL %v with error \"%v\"`
)

func NewLogger(writer io.Writer) *Logger {
	var format = golog.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	backend := golog.NewLogBackend(writer, "", 0)
	backendFormatter := golog.NewBackendFormatter(backend, format)

	backendLeveled := golog.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(golog.INFO, "")

	var logger = golog.MustGetLogger("main")

	logger.SetBackend(backendLeveled)

	return &Logger{*logger}
}

type Logger struct {
	golog.Logger
}

func (logger *Logger) LogRequestBody(r *http.Request, body string) {
	logger.Infof(requestBodyTemplate, r.URL.Path, r.Method, body)
}

func (logger *Logger) LogRequestStart(r *http.Request) {
	logger.Infof(requestStartLogTemplate, r.URL.Path, r.Method)
}

func (logger *Logger) LogRequestSuccess(r *http.Request) {
	logger.Infof(requestSuccessLogTemplate, r.URL.Path, r.Method)
}

func (logger *Logger) LogRequestError(r *http.Request, err error) {
	logger.Errorf(requestErrorTemplate, r.URL.Path, err.Error())
}
