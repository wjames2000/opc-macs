package platform

import "fmt"

type ErrorCode string

const (
	ErrAuthFailed           ErrorCode = "auth_failed"
	ErrTokenExpired         ErrorCode = "token_expired"
	ErrRateLimited          ErrorCode = "rate_limited"
	ErrInvalidRequest       ErrorCode = "invalid_request"
	ErrMediaTooLarge        ErrorCode = "media_too_large"
	ErrMediaTypeUnsupported ErrorCode = "media_type_unsupported"
	ErrPublishFailed        ErrorCode = "publish_failed"
	ErrPlatformError        ErrorCode = "platform_error"
	ErrNetworkError         ErrorCode = "network_error"
	ErrAccountNotFound      ErrorCode = "account_not_found"
	ErrNotImplemented       ErrorCode = "not_implemented"
)

type PlatformError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *PlatformError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *PlatformError) Unwrap() error {
	return e.Err
}

func NewPlatformError(code ErrorCode, message string, err error) *PlatformError {
	return &PlatformError{Code: code, Message: message, Err: err}
}

// IsRateLimited checks if an error is a rate limit error.
func IsRateLimited(err error) bool {
	var pe *PlatformError
	return asError(err, &pe) && pe.Code == ErrRateLimited
}

// IsAuthFailed checks if an error is an authentication failure.
func IsAuthFailed(err error) bool {
	var pe *PlatformError
	return asError(err, &pe) && pe.Code == ErrAuthFailed
}

func asError(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	if t, ok := target.(*error); ok {
		*t = err
		return true
	}
	return false
}
