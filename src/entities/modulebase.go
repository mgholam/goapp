package entities

import (
	"encoding/json"
	"net/http"
)

type ModuleBase struct {
}

// write `data` as json to ResponseWriter
func (m *ModuleBase) JSON(w http.ResponseWriter, data interface{}) {
	b, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// write data as string json to ResponseWriter
func (m *ModuleBase) JSONString(w http.ResponseWriter, data string) {
	// b, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

// send str to ResponseWriter
func (m *ModuleBase) SendString(w http.ResponseWriter, str string) {
	// w.Header().Add("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(str))
}

// write status code to ResponseWriter
func (m *ModuleBase) Status(w http.ResponseWriter, stat int) *ModuleBase {
	w.WriteHeader(stat)
	return m
}

// parse body json as `obj`
func (m *ModuleBase) BodyParser(r *http.Request, obj interface{}) error {
	err := json.NewDecoder(r.Body).Decode(&obj)
	return err
}
