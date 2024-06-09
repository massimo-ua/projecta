package exceptions

import "errors"

type ErrorCode = string

const (
    NotFound         ErrorCode = "NOT_FOUND"
    ValidationFailed ErrorCode = "VALIDATION_FAILED"
    Internal         ErrorCode = "INTERNAL"
    Unauthorized     ErrorCode = "UNAUTHORIZED"
)

var OptimisticLockError = errors.New("optimistic lock error")
var NotFoundError = errors.New("not found error")

type unwrappable interface {
    Unwrap() error
}

type Exception struct {
    Message string    `json:"message"`
    Code    ErrorCode `json:"code"`
    err     error
}

func (e Exception) Error() string {
    return e.Message
}

func (e Exception) Unwrap() error {
    if _, ok := e.err.(unwrappable); ok {
        return errors.Unwrap(e.err)
    }

    return e.err
}

func (e Exception) Is(target error) bool {
    var exception Exception
    if errors.As(target, &exception) {
        return false
    }

    return errors.Is(e.err, target)
}

func (e Exception) As(target any) bool {
    t, ok := target.(*Exception)

    if !ok {
        return errors.As(e.err, &target)
    }

    *t = e
    return true
}

func NewApplicationError(message string, code ErrorCode, err error) Exception {
    if code == "" {
        code = Internal
    }

    if message == "" {
        message = err.Error()
    }

    return Exception{
        Message: message,
        Code:    code,
        err:     err,
    }
}

func NewNotFoundException(message string, e error) Exception {
    return NewApplicationError(message, NotFound, errors.Join(NotFoundError, e))
}

func NewInternalException(message string, e error) Exception {
    return NewApplicationError(message, Internal, e)
}

func NewValidationException(message string, e error) Exception {
    return NewApplicationError(message, ValidationFailed, e)
}

func NewUnauthorizedException(message string, e error) Exception {
    return NewApplicationError(message, Unauthorized, e)
}
