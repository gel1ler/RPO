package httpapi

// Swagger-only types for generated documentation.

type SwaggerAuthLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type SwaggerAuthLoginResponse struct {
	Token string `json:"token"`
}

type SwaggerErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SwaggerMeResponse struct {
	ID          int64   `json:"id"`
	Login       string  `json:"login"`
	DisplayName *string `json:"display_name,omitempty"`
	IsAdmin     bool    `json:"is_admin"`
	CreatedAt   string  `json:"created_at"`
}

type SwaggerTerminalAuthorizeRequest struct {
	TerminalSerialNumber string `json:"terminal_serial_number"`
	CardNumber           string `json:"card_number"`
	Amount               int64  `json:"amount"`
}

type SwaggerTerminalAuthorizeResponse struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason"`
}

type SwaggerTerminalDTO struct {
	ID           int64   `json:"id"`
	SerialNumber string  `json:"serial_number"`
	Address      *string `json:"address,omitempty"`
	Name         *string `json:"name,omitempty"`
	Extra        *string `json:"extra,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

type SwaggerCreateTerminalRequest struct {
	SerialNumber string  `json:"serial_number"`
	Address      *string `json:"address"`
	Name         *string `json:"name"`
	Extra        *string `json:"extra"`
}

type SwaggerUpdateTerminalRequest struct {
	Address *string `json:"address"`
	Name    *string `json:"name"`
	Extra   *string `json:"extra"`
}

type SwaggerKeyDTO struct {
	ID        int64   `json:"id"`
	Label     *string `json:"label,omitempty"`
	KeyValue  string  `json:"key_value"`
	CreatedAt string  `json:"created_at"`
}

type SwaggerCreateKeyRequest struct {
	Label    *string `json:"label"`
	KeyValue string  `json:"key_value"`
}

type SwaggerUpdateKeyRequest struct {
	Label    *string `json:"label"`
	KeyValue *string `json:"key_value"`
}

type SwaggerCardDTO struct {
	ID         int64   `json:"id"`
	CardNumber string  `json:"card_number"`
	Balance    int64   `json:"balance"`
	IsBlocked  bool    `json:"is_blocked"`
	OwnerName  *string `json:"owner_name,omitempty"`
	Extra      *string `json:"extra,omitempty"`
	KeyID      int64   `json:"key_id"`
	CreatedAt  string  `json:"created_at"`
}

type SwaggerCreateCardRequest struct {
	CardNumber string  `json:"card_number"`
	Balance    int64   `json:"balance"`
	IsBlocked  bool    `json:"is_blocked"`
	OwnerName  *string `json:"owner_name"`
	Extra      *string `json:"extra"`
	KeyID      int64   `json:"key_id"`
}

type SwaggerUpdateCardRequest struct {
	Balance   *int64  `json:"balance"`
	IsBlocked *bool   `json:"is_blocked"`
	OwnerName *string `json:"owner_name"`
	Extra     *string `json:"extra"`
	KeyID     *int64  `json:"key_id"`
}

type SwaggerTransactionDTO struct {
	ID         int64  `json:"id"`
	Amount     int64  `json:"amount"`
	CardID     int64  `json:"card_id"`
	TerminalID int64  `json:"terminal_id"`
	CreatedAt  string `json:"created_at"`
}

type SwaggerCreateTransactionRequest struct {
	Amount     int64 `json:"amount"`
	CardID     int64 `json:"card_id"`
	TerminalID int64 `json:"terminal_id"`
}

type SwaggerUpdateTransactionRequest struct {
	Amount     *int64 `json:"amount"`
	CardID     *int64 `json:"card_id"`
	TerminalID *int64 `json:"terminal_id"`
}

type SwaggerUserDTO struct {
	ID          int64   `json:"id"`
	Login       string  `json:"login"`
	DisplayName *string `json:"display_name,omitempty"`
	IsAdmin     bool    `json:"is_admin"`
	CreatedAt   string  `json:"created_at"`
}

type SwaggerCreateUserRequest struct {
	Login       string  `json:"login"`
	DisplayName *string `json:"display_name"`
	Password    string  `json:"password"`
	IsAdmin     bool    `json:"is_admin"`
}

type SwaggerUpdateUserRequest struct {
	DisplayName *string `json:"display_name"`
	Password    *string `json:"password"`
	IsAdmin     *bool   `json:"is_admin"`
}
