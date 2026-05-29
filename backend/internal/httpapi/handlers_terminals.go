package httpapi

import (
	"net/http"
	"strings"

	"rpo/internal/auth"
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

type createTerminalResponse struct {
	terminalDTO
	APISecret string `json:"api_secret"`
}

// handleTerminalsList godoc
// @Summary      List terminals
// @Tags         terminals
// @Produce      json
// @Param        limit   query  int  false  "Limit"
// @Param        offset  query  int  false  "Offset"
// @Success      200  {array}   SwaggerTerminalDTO
// @Failure      401  {object}  SwaggerErrorResponse
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /terminals [get]
// @Security     BearerAuth
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

// handleTerminalsCreate godoc
// @Summary      Create terminal
// @Tags         terminals
// @Accept       json
// @Produce      json
// @Param        body  body      SwaggerCreateTerminalRequest  true  "Terminal"
// @Success      201   {object}  SwaggerTerminalDTO
// @Failure      400   {object}  SwaggerErrorResponse
// @Failure      401   {object}  SwaggerErrorResponse
// @Failure      500   {object}  SwaggerErrorResponse
// @Router       /terminals [post]
// @Security     BearerAuth
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

	apiSecret, err := auth.GenerateAPISecret()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "api secret generation error")
		return
	}
	hash, err := auth.HashPassword(apiSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "api secret hash error")
		return
	}

	created, err := (store.Terminals{DB: s.DB}).Create(r.Context(), store.CreateTerminalParams{
		SerialNumber:  req.SerialNumber,
		Address:       req.Address,
		Name:          req.Name,
		Extra:         req.Extra,
		APISecretHash: &hash,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	writeJSON(w, http.StatusCreated, createTerminalResponse{
		terminalDTO: terminalToDTO(created),
		APISecret:   apiSecret,
	})
}

// handleTerminalsGet godoc
// @Summary      Get terminal
// @Tags         terminals
// @Produce      json
// @Param        id   path      int  true  "Terminal ID"
// @Success      200  {object}  SwaggerTerminalDTO
// @Failure      400  {object}  SwaggerErrorResponse
// @Failure      401  {object}  SwaggerErrorResponse
// @Failure      404  {object}  SwaggerErrorResponse
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /terminals/{id} [get]
// @Security     BearerAuth
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

// handleTerminalsUpdate godoc
// @Summary      Update terminal
// @Tags         terminals
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Terminal ID"
// @Param        body  body      SwaggerUpdateTerminalRequest  true  "Terminal"
// @Success      200   {object}  SwaggerTerminalDTO
// @Failure      400   {object}  SwaggerErrorResponse
// @Failure      401   {object}  SwaggerErrorResponse
// @Failure      404   {object}  SwaggerErrorResponse
// @Failure      500   {object}  SwaggerErrorResponse
// @Router       /terminals/{id} [put]
// @Security     BearerAuth
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

// handleTerminalsDelete godoc
// @Summary      Delete terminal
// @Tags         terminals
// @Param        id   path  int  true  "Terminal ID"
// @Success      204
// @Failure      400  {object}  SwaggerErrorResponse
// @Failure      401  {object}  SwaggerErrorResponse
// @Failure      404  {object}  SwaggerErrorResponse
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /terminals/{id} [delete]
// @Security     BearerAuth
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
