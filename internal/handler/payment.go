package handler

import (
	"net/http"
	"tripshare/internal/models"
	"tripshare/internal/service"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	tripID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid trip ID")
		return
	}
	var req models.CreatePaymentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	payment, err := h.svc.Create(tripID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, payment)
}

func (h *PaymentHandler) Update(w http.ResponseWriter, r *http.Request) {
	tripID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid trip ID")
		return
	}
	paymentID, err := parseID(r, "paymentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment ID")
		return
	}
	var req models.UpdatePaymentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	payment, err := h.svc.Update(tripID, paymentID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, payment)
}

func (h *PaymentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tripID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid trip ID")
		return
	}
	paymentID, err := parseID(r, "paymentId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment ID")
		return
	}
	if err := h.svc.Delete(tripID, paymentID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
