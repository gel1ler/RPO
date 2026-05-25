package httpapi

import (
	"net/http"
	"strings"

	"rpo/internal/store"
)

type cardDTO struct {
	ID         int64   `json:"id"`
	CardNumber string  `json:"card_number"`
	Balance    int64   `json:"balance"`
	IsBlocked  bool    `json:"is_blocked"`
	OwnerName  *string `json:"owner_name,omitempty"`
	Extra      *string `json:"extra,omitempty"`
	KeyID      int64   `json:"key_id"`
	CreatedAt  string  `json:"created_at"`
}

func cardToDTO(c store.Card) cardDTO {
	var ownerName *string
	if c.OwnerName.Valid {
		ownerName = &c.OwnerName.String
	}
	var extra *string
	if c.Extra.Valid {
		extra = &c.Extra.String
	}
	return cardDTO{
		ID:         c.ID,
		CardNumber: c.CardNumber,
		Balance:    c.Balance,
		IsBlocked:  c.IsBlocked,
		OwnerName:  ownerName,
		Extra:      extra,
		KeyID:      c.KeyID,
		CreatedAt:  c.CreatedAt,
	}
}

// handleCardsList godoc
// @Summary      List cards
// @Tags         cards
// @Produce      json
// @Param        limit   query  int  false  "Limit"
// @Param        offset  query  int  false  "Offset"
// @Success      200  {array}   SwaggerCardDTO
// @Failure      401  {object}  SwaggerErrorResponse
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /cards [get]
// @Security     BearerAuth
func (s Server) handleCardsList(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := (store.Cards{DB: s.DB}).List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	out := make([]cardDTO, 0, len(items))
	for _, it := range items {
		out = append(out, cardToDTO(it))
	}
	writeJSON(w, http.StatusOK, out)
}

type createCardRequest struct {
	CardNumber string  `json:"card_number"`
	Balance    int64   `json:"balance"`
	IsBlocked  bool    `json:"is_blocked"`
	OwnerName  *string `json:"owner_name"`
	Extra      *string `json:"extra"`
	KeyID      int64   `json:"key_id"`
}

// handleCardsCreate godoc
// @Summary      Create card
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        body  body      SwaggerCreateCardRequest  true  "Card"
// @Success      201   {object}  SwaggerCardDTO
// @Failure      400   {object}  SwaggerErrorResponse
// @Failure      401   {object}  SwaggerErrorResponse
// @Failure      500   {object}  SwaggerErrorResponse
// @Router       /cards [post]
// @Security     BearerAuth
func (s Server) handleCardsCreate(w http.ResponseWriter, r *http.Request) {
	var req createCardRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	req.CardNumber = strings.TrimSpace(req.CardNumber)
	if req.CardNumber == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "card_number is required")
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

	// Validate key exists.
	if _, err := (store.Keys{DB: s.DB}).GetByID(r.Context(), req.KeyID); err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusBadRequest, "bad_request", "key_id not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}

	created, err := (store.Cards{DB: s.DB}).Create(r.Context(), store.CreateCardParams{
		CardNumber: req.CardNumber,
		Balance:    req.Balance,
		IsBlocked:  req.IsBlocked,
		OwnerName:  req.OwnerName,
		Extra:      req.Extra,
		KeyID:      req.KeyID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusCreated, cardToDTO(created))
}

// handleCardsGet godoc
// @Summary      Get card
// @Tags         cards
// @Produce      json
// @Param        id   path      int  true  "Card ID"
// @Success      200  {object}  SwaggerCardDTO
// @Failure      400  {object}  SwaggerErrorResponse
// @Failure      401  {object}  SwaggerErrorResponse
// @Failure      404  {object}  SwaggerErrorResponse
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /cards/{id} [get]
// @Security     BearerAuth
func (s Server) handleCardsGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	c, err := (store.Cards{DB: s.DB}).GetByID(r.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusOK, cardToDTO(c))
}

type updateCardRequest struct {
	Balance   *int64  `json:"balance"`
	IsBlocked *bool   `json:"is_blocked"`
	OwnerName *string `json:"owner_name"`
	Extra     *string `json:"extra"`
	KeyID     *int64  `json:"key_id"`
}

// handleCardsUpdate godoc
// @Summary      Update card
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Card ID"
// @Param        body  body      SwaggerUpdateCardRequest  true  "Card"
// @Success      200   {object}  SwaggerCardDTO
// @Failure      400   {object}  SwaggerErrorResponse
// @Failure      401   {object}  SwaggerErrorResponse
// @Failure      404   {object}  SwaggerErrorResponse
// @Failure      500   {object}  SwaggerErrorResponse
// @Router       /cards/{id} [put]
// @Security     BearerAuth
func (s Server) handleCardsUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var req updateCardRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if req.Balance != nil && *req.Balance < 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "balance must be >= 0")
		return
	}
	if req.KeyID != nil {
		if *req.KeyID <= 0 {
			writeError(w, http.StatusBadRequest, "bad_request", "key_id must be > 0")
			return
		}
		if _, err := (store.Keys{DB: s.DB}).GetByID(r.Context(), *req.KeyID); err != nil {
			if err == store.ErrNotFound {
				writeError(w, http.StatusBadRequest, "bad_request", "key_id not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal", "database error")
			return
		}
	}

	updated, err := (store.Cards{DB: s.DB}).Update(r.Context(), id, store.UpdateCardParams{
		Balance:   req.Balance,
		IsBlocked: req.IsBlocked,
		OwnerName: req.OwnerName,
		Extra:     req.Extra,
		KeyID:     req.KeyID,
	})
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	writeJSON(w, http.StatusOK, cardToDTO(updated))
}

// handleCardsDelete godoc
// @Summary      Delete card
// @Tags         cards
// @Param        id   path  int  true  "Card ID"
// @Success      204
// @Failure      400  {object}  SwaggerErrorResponse
// @Failure      401  {object}  SwaggerErrorResponse
// @Failure      404  {object}  SwaggerErrorResponse
// @Failure      500  {object}  SwaggerErrorResponse
// @Router       /cards/{id} [delete]
// @Security     BearerAuth
func (s Server) handleCardsDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := (store.Cards{DB: s.DB}).Delete(r.Context(), id); err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
