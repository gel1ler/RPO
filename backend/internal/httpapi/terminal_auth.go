package httpapi

import (
	"context"
	"net/http"
	"strings"
)

// terminalSerialMatches проверяет, что serial в теле запроса совпадает с JWT терминала.
func terminalSerialMatches(ctx context.Context, reqSerial string) bool {
	authSerial, ok := currentTerminalSerial(ctx)
	if !ok {
		return false
	}
	return strings.TrimSpace(reqSerial) == authSerial
}

func writeTerminalSerialMismatch(w http.ResponseWriter) {
	writeError(w, http.StatusForbidden, "forbidden", "terminal_serial_number does not match token")
}
