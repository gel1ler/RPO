package httpapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	_ "rpo/docs"
	"rpo/internal/auth"
	"rpo/internal/store"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Server struct {
	DB  *sql.DB
	JWT auth.JWT
}

func (s Server) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	swagger := httpSwagger.Handler(httpSwagger.URL("/api/v1/swagger/doc.json"))
	mux.Handle("/api/v1/swagger/", swagger)
	mux.HandleFunc("GET /api/v1/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/v1/swagger/index.html", http.StatusFound)
	})

	mux.Handle("POST /api/v1/auth/login", http.HandlerFunc(s.handleLogin))
	mux.Handle("GET /api/v1/me", requireAuth(s.JWT, http.HandlerFunc(s.handleMe)))

	// terminal API (no auth)
	mux.Handle("POST /api/v1/terminal/authorize", http.HandlerFunc(s.handleTerminalAuthorize))
	mux.Handle("GET /api/v1/terminal/keys", http.HandlerFunc(s.handleTerminalKeys))

	authed := func(h http.HandlerFunc) http.Handler {
		return requireAuth(s.JWT, h)
	}
	adminOnly := func(h http.HandlerFunc) http.Handler {
		return requireAuth(s.JWT, requireAdmin(h))
	}

	// terminals
	mux.Handle("GET /api/v1/terminals", authed(s.handleTerminalsList))
	mux.Handle("POST /api/v1/terminals", authed(s.handleTerminalsCreate))
	mux.Handle("GET /api/v1/terminals/{id}", authed(s.handleTerminalsGet))
	mux.Handle("PUT /api/v1/terminals/{id}", authed(s.handleTerminalsUpdate))
	mux.Handle("DELETE /api/v1/terminals/{id}", authed(s.handleTerminalsDelete))

	// keys (admin)
	mux.Handle("GET /api/v1/keys", adminOnly(s.handleKeysList))
	mux.Handle("POST /api/v1/keys", adminOnly(s.handleKeysCreate))
	mux.Handle("GET /api/v1/keys/{id}", adminOnly(s.handleKeysGet))
	mux.Handle("PUT /api/v1/keys/{id}", adminOnly(s.handleKeysUpdate))
	mux.Handle("DELETE /api/v1/keys/{id}", adminOnly(s.handleKeysDelete))

	// cards
	mux.Handle("GET /api/v1/cards", authed(s.handleCardsList))
	mux.Handle("POST /api/v1/cards", authed(s.handleCardsCreate))
	mux.Handle("GET /api/v1/cards/{id}", authed(s.handleCardsGet))
	mux.Handle("PUT /api/v1/cards/{id}", authed(s.handleCardsUpdate))
	mux.Handle("DELETE /api/v1/cards/{id}", authed(s.handleCardsDelete))

	// transactions
	mux.Handle("GET /api/v1/transactions", authed(s.handleTransactionsList))
	mux.Handle("POST /api/v1/transactions", authed(s.handleTransactionsCreate))
	mux.Handle("GET /api/v1/transactions/{id}", authed(s.handleTransactionsGet))
	mux.Handle("PUT /api/v1/transactions/{id}", authed(s.handleTransactionsUpdate))
	mux.Handle("DELETE /api/v1/transactions/{id}", authed(s.handleTransactionsDelete))

	// users
	mux.Handle("GET /api/v1/users", adminOnly(s.handleUsersList))
	mux.Handle("POST /api/v1/users", adminOnly(s.handleUsersCreate))
	mux.Handle("GET /api/v1/users/{id}", authed(s.handleUsersGet))
	mux.Handle("PUT /api/v1/users/{id}", authed(s.handleUsersUpdate))
	mux.Handle("DELETE /api/v1/users/{id}", adminOnly(s.handleUsersDelete))

	return mux
}

func parseID(r *http.Request) (int64, bool) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func parseLimitOffset(r *http.Request) (int64, int64) {
	q := r.URL.Query()
	limit, _ := strconv.ParseInt(q.Get("limit"), 10, 64)
	offset, _ := strconv.ParseInt(q.Get("offset"), 10, 64)
	return limit, offset
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// handleLogin godoc
// @Summary      Login with password
// @Description  Returns JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      SwaggerAuthLoginRequest  true  "Credentials"
// @Success      200   {object}  SwaggerAuthLoginResponse
// @Failure      400   {object}  SwaggerErrorResponse
// @Failure      401   {object}  SwaggerErrorResponse
// @Failure      500   {object}  SwaggerErrorResponse
// @Router       /auth/login [post]
func (s Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if req.Login == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "login and password are required")
		return
	}

	u, err := store.Users{DB: s.DB}.GetByLogin(r.Context(), req.Login)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	if !auth.CheckPassword(u.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}

	token, err := s.JWT.Sign(u.ID, u.IsAdmin, time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "token sign error")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

type meResponse struct {
	ID          int64   `json:"id"`
	Login       string  `json:"login"`
	DisplayName *string `json:"display_name,omitempty"`
	IsAdmin     bool    `json:"is_admin"`
	CreatedAt   string  `json:"created_at"`
}

// handleMe godoc
// @Summary      Current user
// @Tags         auth
// @Produce      json
// @Success      200  {object}  SwaggerMeResponse
// @Failure      401  {object}  SwaggerErrorResponse
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /me [get]
// @Security     BearerAuth
func (s Server) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing auth context")
		return
	}

	u, err := store.Users{DB: s.DB}.GetByID(r.Context(), userID)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusUnauthorized, "unauthorized", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	var displayName *string
	if u.DisplayName.Valid {
		displayName = &u.DisplayName.String
	}

	writeJSON(w, http.StatusOK, meResponse{
		ID:          u.ID,
		Login:       u.Login,
		DisplayName: displayName,
		IsAdmin:     u.IsAdmin,
		CreatedAt:   u.CreatedAt,
	})
}
