package httpapi

import (
	"net/http"
	"strings"

	"rpo/internal/store"
)

type terminalAuthorizeRequest struct {
	TerminalSerialNumber string `json:"terminal_serial_number"`
	CardNumber           string `json:"card_number"`
	Amount               int64  `json:"amount"`
}

type terminalAuthorizeResponse struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason"`
}

// handleTerminalAuthorize godoc
// @Summary      Authorize payment transaction (decision only)
// @Tags         terminal
// @Accept       json
// @Produce      json
// @Param        body  body      SwaggerTerminalAuthorizeRequest  true  "Authorization request"
// @Success      200   {object}  SwaggerTerminalAuthorizeResponse
// @Failure      400   {object}  SwaggerErrorResponse
// @Failure      500   {object}  SwaggerErrorResponse
// @Router       /terminal/authorize [post]
func (s Server) handleTerminalAuthorize(w http.ResponseWriter, r *http.Request) {
	var req terminalAuthorizeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}

	req.TerminalSerialNumber = strings.TrimSpace(req.TerminalSerialNumber)
	req.CardNumber = strings.TrimSpace(req.CardNumber)
	if req.TerminalSerialNumber == "" || req.CardNumber == "" || req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "terminal_serial_number, card_number and amount are required")
		return
	}

	if _, err := (store.Terminals{DB: s.DB}).GetBySerialNumber(r.Context(), req.TerminalSerialNumber); err != nil {
		if err == store.ErrNotFound {
			writeJSON(w, http.StatusOK, terminalAuthorizeResponse{Approved: false, Reason: "terminal_not_found"})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	card, err := (store.Cards{DB: s.DB}).GetByCardNumber(r.Context(), req.CardNumber)
	if err != nil {
		if err == store.ErrNotFound {
			writeJSON(w, http.StatusOK, terminalAuthorizeResponse{Approved: false, Reason: "card_not_found"})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	if card.IsBlocked {
		writeJSON(w, http.StatusOK, terminalAuthorizeResponse{Approved: false, Reason: "card_blocked"})
		return
	}
	if card.Balance < req.Amount {
		writeJSON(w, http.StatusOK, terminalAuthorizeResponse{Approved: false, Reason: "insufficient_funds"})
		return
	}

	writeJSON(w, http.StatusOK, terminalAuthorizeResponse{Approved: true, Reason: "ok"})
}

// handleTerminalKeys godoc
// @Summary      Load keys for terminals
// @Tags         terminal
// @Produce      json
// @Success      200  {array}   SwaggerKeyDTO
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /terminal/keys [get]
func (s Server) handleTerminalKeys(w http.ResponseWriter, r *http.Request) {
	items, err := (store.Keys{DB: s.DB}).List(r.Context(), 500, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	out := make([]keyDTO, 0, len(items))
	for _, it := range items {
		out = append(out, keyToDTO(it))
	}
	writeJSON(w, http.StatusOK, out)
}

