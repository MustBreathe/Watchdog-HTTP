package api

import (
	"net/http"
	"watchdog/main.go/internal/monitor"
)

type MonitorHandler struct {
	service *monitor.MonitorService
}

func NewMonitorHandler(service *monitor.MonitorService) *MonitorHandler {
	return &MonitorHandler{
		service: service,
	}
}

func (handler *MonitorHandler) Create(w http.ResponseWriter, r *http.Request) {
	panic("not implemented exception")
}
func (handler *MonitorHandler) List(w http.ResponseWriter, r *http.Request) {
	panic("not implemented exception")
}
func (handler *MonitorHandler) Get(w http.ResponseWriter, r *http.Request) {
	panic("not implemented exception")
}
func (handler *MonitorHandler) Update(w http.ResponseWriter, r *http.Request) {
	panic("not implemented exception")
}
func (handler *MonitorHandler) Enable(w http.ResponseWriter, r *http.Request) {
	panic("not implemented exception")
}
func (handler *MonitorHandler) Disable(w http.ResponseWriter, r *http.Request) {
	panic("not implemented exception")
}
func (handler *MonitorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	panic("not implemented exception")
}
