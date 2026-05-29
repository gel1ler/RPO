package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"rpo/internal/store"
)

type terminalEventDTO struct {
	ID             int64   `json:"id"`
	TerminalSerial string  `json:"terminal_serial"`
	CardNumber     string  `json:"card_number"`
	Operation      string  `json:"operation"`
	Amount         int64   `json:"amount"`
	Approved       *bool   `json:"approved,omitempty"`
	Reason         *string `json:"reason,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

type terminalEventCreateRequest struct {
	TerminalSerialNumber string  `json:"terminal_serial_number"`
	CardNumber           string  `json:"card_number"`
	Operation            string  `json:"operation"`
	Amount               int64   `json:"amount"`
	Approved             *bool   `json:"approved"`
	Reason               *string `json:"reason"`
}

func terminalEventToDTO(e store.TerminalEvent) terminalEventDTO {
	d := terminalEventDTO{
		ID:             e.ID,
		TerminalSerial: e.TerminalSerial,
		CardNumber:     e.CardNumber,
		Operation:      e.Operation,
		Amount:         e.Amount,
		CreatedAt:      e.CreatedAt,
	}
	if e.Approved.Valid {
		v := e.Approved.Int64 == 1
		d.Approved = &v
	}
	if e.Reason.Valid {
		r := e.Reason.String
		d.Reason = &r
	}
	return d
}

// handleTerminalEventPost — события с терминала (пополнение и т.п.), Bearer terminal JWT.
func (s Server) handleTerminalEventPost(w http.ResponseWriter, r *http.Request) {
	var req terminalEventCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	req.TerminalSerialNumber = strings.TrimSpace(req.TerminalSerialNumber)
	req.CardNumber = strings.TrimSpace(req.CardNumber)
	req.Operation = strings.TrimSpace(req.Operation)
	if req.TerminalSerialNumber == "" || req.CardNumber == "" || req.Operation == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "terminal_serial_number, card_number and operation are required")
		return
	}
	if !terminalSerialMatches(r.Context(), req.TerminalSerialNumber) {
		writeTerminalSerialMismatch(w)
		return
	}
	switch req.Operation {
	case "credit_balance", "debit_card", "register_card":
	default:
		writeError(w, http.StatusBadRequest, "bad_request", "unknown operation")
		return
	}
	if req.Operation == "credit_balance" && req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "amount must be > 0 for credit_balance")
		return
	}
	if req.Operation == "debit_card" && req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "amount must be > 0 for debit_card")
		return
	}

	if req.Operation == "debit_card" || req.Operation == "credit_balance" {
		terminal, err := (store.Terminals{DB: s.DB}).GetBySerialNumber(r.Context(), req.TerminalSerialNumber)
		if err != nil {
			if err == store.ErrNotFound {
				writeError(w, http.StatusBadRequest, "bad_request", "terminal_not_found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal", "database error")
			return
		}
		card, err := (store.Cards{DB: s.DB}).GetByCardNumber(r.Context(), req.CardNumber)
		if err != nil {
			if err == store.ErrNotFound {
				writeError(w, http.StatusBadRequest, "bad_request", "card_not_found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal", "database error")
			return
		}
		txnStore := store.Transactions{DB: s.DB}
		if req.Operation == "debit_card" {
			_, err = txnStore.ApplyDebit(r.Context(), card.ID, terminal.ID, req.Amount)
		} else {
			_, err = txnStore.ApplyCredit(r.Context(), card.ID, terminal.ID, req.Amount)
		}
		if err != nil {
			if errors.Is(err, store.ErrInsufficientFunds) {
				writeError(w, http.StatusConflict, "insufficient_funds", "insufficient funds")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal", "database error")
			return
		}
	}

	ev, err := (store.TerminalEvents{DB: s.DB}).Create(r.Context(), store.CreateTerminalEventParams{
		TerminalSerial: req.TerminalSerialNumber,
		CardNumber:     req.CardNumber,
		Operation:      req.Operation,
		Amount:         req.Amount,
		Approved:       req.Approved,
		Reason:         req.Reason,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusCreated, terminalEventToDTO(ev))
}

// handleTerminalEventsSince — опрос журнала для SPA (Bearer JWT).
func (s Server) handleTerminalEventsSince(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	since, _ := strconv.ParseInt(q.Get("since"), 10, 64)
	if since < 0 {
		since = 0
	}
	limit, _ := strconv.ParseInt(q.Get("limit"), 10, 64)
	items, err := (store.TerminalEvents{DB: s.DB}).ListSince(r.Context(), since, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	out := make([]terminalEventDTO, 0, len(items))
	for _, it := range items {
		out = append(out, terminalEventToDTO(it))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}
