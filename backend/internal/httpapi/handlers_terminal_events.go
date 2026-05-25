package httpapi

import (
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
	TripsDelta     int64   `json:"trips_delta"`
	Approved       *bool   `json:"approved,omitempty"`
	Reason         *string `json:"reason,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

type terminalEventCreateRequest struct {
	TerminalSerialNumber string  `json:"terminal_serial_number"`
	CardNumber           string  `json:"card_number"`
	Operation            string  `json:"operation"`
	Amount               int64   `json:"amount"`
	TripsDelta           int64   `json:"trips_delta"`
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
		TripsDelta:     e.TripsDelta,
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

// handleTerminalEventPost — события с терминала (пополнение и т.п.), без JWT.
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
	switch req.Operation {
	case "credit_balance", "credit_trips", "debit_card":
	default:
		writeError(w, http.StatusBadRequest, "bad_request", "unknown operation")
		return
	}
	if req.Operation == "credit_trips" && req.TripsDelta <= 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "trips_delta must be > 0 for credit_trips")
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

	ev, err := (store.TerminalEvents{DB: s.DB}).Create(r.Context(), store.CreateTerminalEventParams{
		TerminalSerial: req.TerminalSerialNumber,
		CardNumber:     req.CardNumber,
		Operation:      req.Operation,
		Amount:         req.Amount,
		TripsDelta:     req.TripsDelta,
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
