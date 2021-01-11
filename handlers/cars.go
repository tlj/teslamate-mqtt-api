package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"net/http"
	"teslamate-mqtt-api/config"
	"teslamate-mqtt-api/store"
)

type CarsHandler struct {
	store *store.Store
	cfg   *config.Config
}

func NewCarsHandler(store *store.Store, cfg *config.Config) *CarsHandler {
	return &CarsHandler{
		store: store,
		cfg:   cfg,
	}
}

func (ch *CarsHandler) Cars(w http.ResponseWriter, r *http.Request) {
	j, _ := json.Marshal(ch.store)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func (ch *CarsHandler) Car(w http.ResponseWriter, r *http.Request) {
	carID := chi.URLParam(r, "id")
	if _, ok := (*ch.store)[carID]; !ok {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"response":"invalid id"}`))
		return
	}

	resp := make(map[string]interface{})
	resp["response"] = nil
	for k, v := range (*ch.store)[carID] {
		resp[k] = v
	}

	j, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}
