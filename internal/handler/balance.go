package handler

import (
	"net/http"
	"tripshare/internal/service"
)

type BalanceHandler struct {
	svc *service.BalanceService
}

func NewBalanceHandler(svc *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{svc: svc}
}

func (h *BalanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	balances, err := h.svc.GetBalances()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to calculate balances")
		return
	}
	writeJSON(w, http.StatusOK, balances)
}
