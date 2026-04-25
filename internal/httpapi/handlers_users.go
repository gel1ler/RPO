package httpapi

import (
	"net/http"
	"strings"

	"rpo/internal/auth"
	"rpo/internal/store"
)

type userDTO struct {
	ID          int64   `json:"id"`
	Login       string  `json:"login"`
	DisplayName *string `json:"display_name,omitempty"`
	IsAdmin     bool    `json:"is_admin"`
	CreatedAt   string  `json:"created_at"`
}

func userToDTO(u store.User) userDTO {
	var displayName *string
	if u.DisplayName.Valid {
		displayName = &u.DisplayName.String
	}
	return userDTO{
		ID:          u.ID,
		Login:       u.Login,
		DisplayName: displayName,
		IsAdmin:     u.IsAdmin,
		CreatedAt:   u.CreatedAt,
	}
}

func (s Server) handleUsersList(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := (store.Users{DB: s.DB}).List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	out := make([]userDTO, 0, len(items))
	for _, it := range items {
		out = append(out, userToDTO(it))
	}
	writeJSON(w, http.StatusOK, out)
}

type createUserRequest struct {
	Login       string  `json:"login"`
	DisplayName *string `json:"display_name"`
	Password    string  `json:"password"`
	IsAdmin     bool    `json:"is_admin"`
}

func (s Server) handleUsersCreate(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "login and password are required")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "password hash error")
		return
	}

	created, err := (store.Users{DB: s.DB}).Create(r.Context(), store.CreateUserParams{
		Login:        req.Login,
		DisplayName:  req.DisplayName,
		PasswordHash: hash,
		IsAdmin:      req.IsAdmin,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	writeJSON(w, http.StatusCreated, userToDTO(created))
}

func (s Server) handleUsersGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	currentID, ok := currentUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing auth context")
		return
	}
	if !currentIsAdmin(r.Context()) && currentID != id {
		writeError(w, http.StatusForbidden, "forbidden", "can only access own user")
		return
	}

	u, err := (store.Users{DB: s.DB}).GetByID(r.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusOK, userToDTO(u))
}

type updateUserRequest struct {
	DisplayName *string `json:"display_name"`
	Password    *string `json:"password"`
	IsAdmin     *bool   `json:"is_admin"`
	Login       *string `json:"login"`
}

func (s Server) handleUsersUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	currentID, ok := currentUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing auth context")
		return
	}
	isAdmin := currentIsAdmin(r.Context())
	if !isAdmin && currentID != id {
		writeError(w, http.StatusForbidden, "forbidden", "can only update own user")
		return
	}

	var req updateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}

	if req.Login != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "login cannot be updated")
		return
	}

	var passwordHash *string
	if req.Password != nil {
		p := strings.TrimSpace(*req.Password)
		if p == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "password cannot be empty")
			return
		}
		hash, err := auth.HashPassword(p)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal", "password hash error")
			return
		}
		passwordHash = &hash
	}

	var isAdminUpdate *bool
	if req.IsAdmin != nil {
		if !isAdmin {
			writeError(w, http.StatusForbidden, "forbidden", "cannot change is_admin")
			return
		}
		isAdminUpdate = req.IsAdmin
	}

	updated, err := (store.Users{DB: s.DB}).Update(r.Context(), id, store.UpdateUserParams{
		DisplayName:  req.DisplayName,
		PasswordHash: passwordHash,
		IsAdmin:      isAdminUpdate,
	})
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusOK, userToDTO(updated))
}

func (s Server) handleUsersDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := (store.Users{DB: s.DB}).Delete(r.Context(), id); err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
