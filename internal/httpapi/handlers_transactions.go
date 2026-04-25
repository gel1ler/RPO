package httpapi

import (
	"net/http"

	"rpo/internal/store"
)

type transactionDTO struct {
	ID         int64  `json:"id"`
	Amount     int64  `json:"amount"`
	CardID     int64  `json:"card_id"`
	TerminalID int64  `json:"terminal_id"`
	CreatedAt  string `json:"created_at"`
}

func transactionToDTO(t store.Transaction) transactionDTO {
	return transactionDTO{
		ID:         t.ID,
		Amount:     t.Amount,
		CardID:     t.CardID,
		TerminalID: t.TerminalID,
		CreatedAt:  t.CreatedAt,
	}
}

func (s Server) handleTransactionsList(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := (store.Transactions{DB: s.DB}).List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	out := make([]transactionDTO, 0, len(items))
	for _, it := range items {
		out = append(out, transactionToDTO(it))
	}
	writeJSON(w, http.StatusOK, out)
}

type createTransactionRequest struct {
	Amount     int64 `json:"amount"`
	CardID     int64 `json:"card_id"`
	TerminalID int64 `json:"terminal_id"`
}

func (s Server) handleTransactionsCreate(w http.ResponseWriter, r *http.Request) {
	var req createTransactionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "amount must be > 0")
		return
	}
	if req.CardID <= 0 || req.TerminalID <= 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "card_id and terminal_id are required")
		return
	}

	if _, err := (store.Cards{DB: s.DB}).GetByID(r.Context(), req.CardID); err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusBadRequest, "bad_request", "card_id not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	if _, err := (store.Terminals{DB: s.DB}).GetByID(r.Context(), req.TerminalID); err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusBadRequest, "bad_request", "terminal_id not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	created, err := (store.Transactions{DB: s.DB}).Create(r.Context(), store.CreateTransactionParams{
		Amount:     req.Amount,
		CardID:     req.CardID,
		TerminalID: req.TerminalID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusCreated, transactionToDTO(created))
}

func (s Server) handleTransactionsGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	t, err := (store.Transactions{DB: s.DB}).GetByID(r.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "transaction not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusOK, transactionToDTO(t))
}

func (s Server) handleTransactionsDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := (store.Transactions{DB: s.DB}).Delete(r.Context(), id); err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "transaction not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
