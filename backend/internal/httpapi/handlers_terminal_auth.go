package httpapi

import (
	"net/http"
	"strings"
	"time"

	"rpo/internal/auth"
	"rpo/internal/store"
)

type terminalLoginRequest struct {
	SerialNumber string `json:"serial_number"`
	APISecret    string `json:"api_secret"`
}

type terminalLoginResponse struct {
	Token        string `json:"token"`
	TerminalID   int64  `json:"terminal_id"`
	SerialNumber string `json:"serial_number"`
	ExpiresAt    string `json:"expires_at"`
}

// handleTerminalLogin godoc
// @Summary      Terminal device login
// @Tags         terminal
// @Accept       json
// @Produce      json
// @Param        body  body      terminalLoginRequest  true  "Terminal credentials"
// @Success      200   {object}  terminalLoginResponse
// @Failure      400   {object}  SwaggerErrorResponse
// @Failure      401   {object}  SwaggerErrorResponse
// @Failure      500   {object}  SwaggerErrorResponse
// @Router       /terminal/auth/login [post]
func (s Server) handleTerminalLogin(w http.ResponseWriter, r *http.Request) {
	var req terminalLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	req.SerialNumber = strings.TrimSpace(req.SerialNumber)
	req.APISecret = strings.TrimSpace(req.APISecret)
	if req.SerialNumber == "" || req.APISecret == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "serial_number and api_secret are required")
		return
	}

	t, err := (store.Terminals{DB: s.DB}).GetBySerialNumber(r.Context(), req.SerialNumber)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	if !t.APISecretHash.Valid || t.APISecretHash.String == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "terminal api secret is not configured")
		return
	}
	if !auth.CheckPassword(t.APISecretHash.String, req.APISecret) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}

	now := time.Now()
	token, err := s.JWT.SignTerminal(t.ID, t.SerialNumber, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "token sign error")
		return
	}

	writeJSON(w, http.StatusOK, terminalLoginResponse{
		Token:        token,
		TerminalID:   t.ID,
		SerialNumber: t.SerialNumber,
		ExpiresAt:    now.Add(s.JWT.TTL).UTC().Format(time.RFC3339),
	})
}
