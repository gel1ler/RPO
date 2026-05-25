package httpapi

import (
	"net/http"
	"strings"

	"rpo/internal/store"
)

type keyDTO struct {
	ID        int64   `json:"id"`
	Label     *string `json:"label,omitempty"`
	KeyValue  string  `json:"key_value"`
	CreatedAt string  `json:"created_at"`
}

func keyToDTO(k store.Key) keyDTO {
	var label *string
	if k.Label.Valid {
		label = &k.Label.String
	}
	return keyDTO{
		ID:        k.ID,
		Label:     label,
		KeyValue:  k.KeyValue,
		CreatedAt: k.CreatedAt,
	}
}

// handleKeysList godoc
// @Summary      List keys
// @Tags         keys
// @Produce      json
// @Param        limit   query  int  false  "Limit"
// @Param        offset  query  int  false  "Offset"
// @Success      200  {array}   SwaggerKeyDTO
// @Failure      401  {object}  SwaggerErrorResponse
// @Failure      403  {object}  SwaggerErrorResponse
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /keys [get]
// @Security     BearerAuth
func (s Server) handleKeysList(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := (store.Keys{DB: s.DB}).List(r.Context(), limit, offset)
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

type createKeyRequest struct {
	Label    *string `json:"label"`
	KeyValue string  `json:"key_value"`
}

// handleKeysCreate godoc
// @Summary      Create key
// @Tags         keys
// @Accept       json
// @Produce      json
// @Param        body  body      SwaggerCreateKeyRequest  true  "Key"
// @Success      201   {object}  SwaggerKeyDTO
// @Failure      400   {object}  SwaggerErrorResponse
// @Failure      401   {object}  SwaggerErrorResponse
// @Failure      403   {object}  SwaggerErrorResponse
// @Failure      500   {object}  SwaggerErrorResponse
// @Router       /keys [post]
// @Security     BearerAuth
func (s Server) handleKeysCreate(w http.ResponseWriter, r *http.Request) {
	var req createKeyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	req.KeyValue = strings.TrimSpace(req.KeyValue)
	if req.KeyValue == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "key_value is required")
		return
	}

	created, err := (store.Keys{DB: s.DB}).Create(r.Context(), store.CreateKeyParams{
		Label:    req.Label,
		KeyValue: req.KeyValue,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusCreated, keyToDTO(created))
}

// handleKeysGet godoc
// @Summary      Get key
// @Tags         keys
// @Produce      json
// @Param        id   path      int  true  "Key ID"
// @Success      200  {object}  SwaggerKeyDTO
// @Failure      400  {object}  SwaggerErrorResponse
// @Failure      401  {object}  SwaggerErrorResponse
// @Failure      403  {object}  SwaggerErrorResponse
// @Failure      404  {object}  SwaggerErrorResponse
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /keys/{id} [get]
// @Security     BearerAuth
func (s Server) handleKeysGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	k, err := (store.Keys{DB: s.DB}).GetByID(r.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "key not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusOK, keyToDTO(k))
}

type updateKeyRequest struct {
	Label    *string `json:"label"`
	KeyValue *string `json:"key_value"`
}

// handleKeysUpdate godoc
// @Summary      Update key
// @Tags         keys
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Key ID"
// @Param        body  body      SwaggerUpdateKeyRequest  true  "Key"
// @Success      200   {object}  SwaggerKeyDTO
// @Failure      400   {object}  SwaggerErrorResponse
// @Failure      401   {object}  SwaggerErrorResponse
// @Failure      403   {object}  SwaggerErrorResponse
// @Failure      404   {object}  SwaggerErrorResponse
// @Failure      500   {object}  SwaggerErrorResponse
// @Router       /keys/{id} [put]
// @Security     BearerAuth
func (s Server) handleKeysUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var req updateKeyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if req.KeyValue != nil {
		v := strings.TrimSpace(*req.KeyValue)
		req.KeyValue = &v
		if v == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "key_value cannot be empty")
			return
		}
	}

	updated, err := (store.Keys{DB: s.DB}).Update(r.Context(), id, store.UpdateKeyParams{
		Label:    req.Label,
		KeyValue: req.KeyValue,
	})
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "key not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusOK, keyToDTO(updated))
}

// handleKeysDelete godoc
// @Summary      Delete key
// @Tags         keys
// @Param        id   path  int  true  "Key ID"
// @Success      204
// @Failure      400  {object}  SwaggerErrorResponse
// @Failure      401  {object}  SwaggerErrorResponse
// @Failure      403  {object}  SwaggerErrorResponse
// @Failure      404  {object}  SwaggerErrorResponse
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /keys/{id} [delete]
// @Security     BearerAuth
func (s Server) handleKeysDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := (store.Keys{DB: s.DB}).Delete(r.Context(), id); err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "key not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
