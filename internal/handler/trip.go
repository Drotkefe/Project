package handler

import (
	"net/http"
	"tripshare/internal/models"
	"tripshare/internal/service"
)

type TripHandler struct {
	svc *service.TripService
}

func NewTripHandler(svc *service.TripService) *TripHandler {
	return &TripHandler{svc: svc}
}

func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTripRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	trip, err := h.svc.Create(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, trip)
}

func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	trips, err := h.svc.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list trips")
		return
	}
	writeJSON(w, http.StatusOK, trips)
}

func (h *TripHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid trip ID")
		return
	}
	detail, err := h.svc.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid trip ID")
		return
	}
	var req models.UpdateTripRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	trip, err := h.svc.Update(id, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, trip)
}

func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid trip ID")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
