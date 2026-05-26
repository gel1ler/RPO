package httpapi

import (
	"context"
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

func (s Server) persistTerminalAuthorize(ctx context.Context, req terminalAuthorizeRequest, resp terminalAuthorizeResponse) {
	approved := resp.Approved
	reason := resp.Reason
	_, _ = (store.TerminalEvents{DB: s.DB}).Create(ctx, store.CreateTerminalEventParams{
		TerminalSerial: req.TerminalSerialNumber,
		CardNumber:     req.CardNumber,
		Operation:      "authorize",
		Amount:         req.Amount,
		Approved:       &approved,
		Reason:         &reason,
	})
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

	resp := terminalAuthorizeResponse{Approved: false, Reason: "internal"}

	if _, err := (store.Terminals{DB: s.DB}).GetBySerialNumber(r.Context(), req.TerminalSerialNumber); err != nil {
		if err == store.ErrNotFound {
			resp = terminalAuthorizeResponse{Approved: false, Reason: "terminal_not_found"}
			s.persistTerminalAuthorize(r.Context(), req, resp)
			writeJSON(w, http.StatusOK, resp)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	card, err := (store.Cards{DB: s.DB}).GetByCardNumber(r.Context(), req.CardNumber)
	if err != nil {
		if err == store.ErrNotFound {
			resp = terminalAuthorizeResponse{Approved: false, Reason: "card_not_found"}
			s.persistTerminalAuthorize(r.Context(), req, resp)
			writeJSON(w, http.StatusOK, resp)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	if card.IsBlocked {
		resp = terminalAuthorizeResponse{Approved: false, Reason: "card_blocked"}
	} else if card.Balance < req.Amount {
		resp = terminalAuthorizeResponse{Approved: false, Reason: "insufficient_funds"}
	} else {
		resp = terminalAuthorizeResponse{Approved: true, Reason: "ok"}
	}

	s.persistTerminalAuthorize(r.Context(), req, resp)
	writeJSON(w, http.StatusOK, resp)
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

type terminalRegisterCardRequest struct {
	TerminalSerialNumber string `json:"terminal_serial_number"`
	CardNumber           string `json:"card_number"`
	Balance              int64  `json:"balance"`
	KeyID                int64  `json:"key_id"`
}

type terminalRegisterCardResponse struct {
	Card    cardDTO `json:"card"`
	Created bool    `json:"created"`
}

// handleTerminalRegisterCard — регистрация карты в БД с терминала (без JWT).
func (s Server) handleTerminalRegisterCard(w http.ResponseWriter, r *http.Request) {
	var req terminalRegisterCardRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}

	req.TerminalSerialNumber = strings.TrimSpace(req.TerminalSerialNumber)
	req.CardNumber = strings.TrimSpace(req.CardNumber)
	if req.TerminalSerialNumber == "" || req.CardNumber == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "terminal_serial_number and card_number are required")
		return
	}
	if req.Balance < 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "balance must be >= 0")
		return
	}
	if req.KeyID <= 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "key_id is required")
		return
	}

	if _, err := (store.Terminals{DB: s.DB}).GetBySerialNumber(r.Context(), req.TerminalSerialNumber); err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusBadRequest, "bad_request", "terminal_not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	if _, err := (store.Keys{DB: s.DB}).GetByID(r.Context(), req.KeyID); err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusBadRequest, "bad_request", "key_id not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	cards := store.Cards{DB: s.DB}
	existing, err := cards.GetByCardNumber(r.Context(), req.CardNumber)
	created := false
	var card store.Card

	if err != nil {
		if err != store.ErrNotFound {
			writeError(w, http.StatusInternalServerError, "internal", "database error")
			return
		}
		card, err = cards.Create(r.Context(), store.CreateCardParams{
			CardNumber: req.CardNumber,
			Balance:    req.Balance,
			IsBlocked:  false,
			KeyID:      req.KeyID,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal", "database error")
			return
		}
		created = true
	} else {
		bal := req.Balance
		keyID := req.KeyID
		card, err = cards.Update(r.Context(), existing.ID, store.UpdateCardParams{
			Balance: &bal,
			KeyID:   &keyID,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal", "database error")
			return
		}
	}

	_, _ = (store.TerminalEvents{DB: s.DB}).Create(r.Context(), store.CreateTerminalEventParams{
		TerminalSerial: req.TerminalSerialNumber,
		CardNumber:     req.CardNumber,
		Operation:      "register_card",
		Amount:         req.Balance,
	})

	writeJSON(w, http.StatusOK, terminalRegisterCardResponse{
		Card:    cardToDTO(card),
		Created: created,
	})
}
