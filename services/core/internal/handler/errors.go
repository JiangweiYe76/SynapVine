package handler

// errorResponse is a helper for building standard error JSON.
func errorResponse(code, message string) map[string]string {
	return map[string]string{
		"error":   code,
		"message": message,
	}
}
