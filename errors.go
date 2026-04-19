package amapclient

import "fmt"

// ErrMissingAPIKey indicates a missing API key.
var ErrMissingAPIKey = fmt.Errorf("amapclient: missing api key")

// ValidationError describes an invalid request payload.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("amapclient: invalid %s: %s", e.Field, e.Message)
}

type APIError struct {
	InfoCode string
	Info     string
}

func (e APIError) Error() string {
	return fmt.Sprintf("amapclient: api error (%s): %s", e.InfoCode, e.Info)
}
