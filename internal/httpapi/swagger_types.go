package httpapi

// Exported swagger-only types.
// These types exist to let swag generate schemas for the existing handlers
// without changing the runtime request/response DTOs.

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

type SwaggerKeyDTO struct {
	ID        int64   `json:"id"`
	Label     *string `json:"label,omitempty"`
	KeyValue  string  `json:"key_value"`
	CreatedAt string  `json:"created_at"`
}

