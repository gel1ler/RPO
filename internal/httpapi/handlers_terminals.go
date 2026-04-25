package httpapi

import (
	"net/http"
	"strings"

	"rpo/internal/store"
)

type terminalDTO struct {
	ID           int64   `json:"id"`
	SerialNumber string  `json:"serial_number"`
	Address      *string `json:"address,omitempty"`
	Name         *string `json:"name,omitempty"`
	Extra        *string `json:"extra,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

func terminalToDTO(t store.Terminal) terminalDTO {
	var address *string
	if t.Address.Valid {
		address = &t.Address.String
	}
	var name *string
	if t.Name.Valid {
		name = &t.Name.String
	}
	var extra *string
	if t.Extra.Valid {
		extra = &t.Extra.String
	}
	return terminalDTO{
		ID:           t.ID,
		SerialNumber: t.SerialNumber,
		Address:      address,
		Name:         name,
		Extra:        extra,
		CreatedAt:    t.CreatedAt,
	}
}

type createTerminalRequest struct {
	SerialNumber string  `json:"serial_number"`
	Address      *string `json:"address"`
	Name         *string `json:"name"`
	Extra        *string `json:"extra"`
}

func (s Server) handleTerminalsList(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := (store.Terminals{DB: s.DB}).List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	out := make([]terminalDTO, 0, len(items))
	for _, it := range items {
		out = append(out, terminalToDTO(it))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s Server) handleTerminalsCreate(w http.ResponseWriter, r *http.Request) {
	var req createTerminalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	req.SerialNumber = strings.TrimSpace(req.SerialNumber)
	if req.SerialNumber == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "serial_number is required")
		return
	}

	created, err := (store.Terminals{DB: s.DB}).Create(r.Context(), store.CreateTerminalParams{
		SerialNumber: req.SerialNumber,
		Address:      req.Address,
		Name:         req.Name,
		Extra:        req.Extra,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	writeJSON(w, http.StatusCreated, terminalToDTO(created))
}

func (s Server) handleTerminalsGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	t, err := (store.Terminals{DB: s.DB}).GetByID(r.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "terminal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	writeJSON(w, http.StatusOK, terminalToDTO(t))
}

type updateTerminalRequest struct {
	Address *string `json:"address"`
	Name    *string `json:"name"`
	Extra   *string `json:"extra"`
}

func (s Server) handleTerminalsUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	var req updateTerminalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}

	updated, err := (store.Terminals{DB: s.DB}).Update(r.Context(), id, store.UpdateTerminalParams{
		Address: req.Address,
		Name:    req.Name,
		Extra:   req.Extra,
	})
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "terminal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	writeJSON(w, http.StatusOK, terminalToDTO(updated))
}

func (s Server) handleTerminalsDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := (store.Terminals{DB: s.DB}).Delete(r.Context(), id); err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "terminal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
