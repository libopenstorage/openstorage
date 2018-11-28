package volume

// Custom credential operation error codes
const (
	_ = iota + 100
	// ErrInvalidCredential is returned when a credential is invalid
	ErrInvalidCredential
)

// CredentialError error returned for credential operations
type CredentialError struct {
	// Code is one of credential operation error codes
	Code int
	// Msg is human understandable error message
	Msg string
}

func NewCredentialError(code int, msg string) error {
	return &CredentialError{Code: code, Msg: msg}
}

func (e *CredentialError) Error() string {
	return e.Msg
}
