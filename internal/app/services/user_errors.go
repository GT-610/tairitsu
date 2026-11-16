package services

import (
	"errors"
	"time"
)

var (
	ErrUserDBUnavailable          = errors.New("database is not configured; complete initial setup first")
	ErrInvalidUsername            = errors.New("username is required")
	ErrUsernameTooLong            = errors.New("username is too long")
	ErrPasswordTooShort           = errors.New("password must be at least 6 characters")
	ErrPasswordTooLong            = errors.New("password must be at most 32 characters")
	ErrUsernameExists             = errors.New("username already exists")
	ErrInvalidCredentials         = errors.New("username or password is incorrect")
	ErrUserNotFound               = errors.New("user not found")
	ErrOldPasswordIncorrect       = errors.New("current password is incorrect")
	ErrAdminTransferSelf          = errors.New("cannot transfer administrator role to yourself")
	ErrAdminResetSelf             = errors.New("cannot reset your own password; use the change password flow")
	ErrAdminDeleteSelf            = errors.New("cannot delete yourself; transfer administrator role first or use another administrator account")
	ErrAdminDeleteBlocked         = errors.New("cannot delete the current administrator account; transfer administrator role first")
	ErrTransferTargetAdmin        = errors.New("target user is already an administrator")
	ErrAdminAccessDenied          = errors.New("current user is not an administrator")
	ErrAdminAlreadyExists         = errors.New("an administrator account already exists")
	ErrPublicRegistrationDisabled = errors.New("public registration is disabled; contact an administrator to create an account")
	ErrSessionNotFound            = errors.New("session not found")
	ErrSessionAccessDenied        = errors.New("session access denied")
	ErrSessionRevoked             = errors.New("session is no longer valid; sign in again")
	ErrSessionExpired             = errors.New("session expired; sign in again")
	ErrLoginThrottled             = errors.New("too many sign-in attempts; try again later")
)

// LoginThrottledError reports that sign-in was refused without the password
// being examined, and how long the caller should wait.
type LoginThrottledError struct {
	RetryAfter time.Duration
}

func (e *LoginThrottledError) Error() string { return ErrLoginThrottled.Error() }

// Unwrap makes errors.Is(err, ErrLoginThrottled) hold for the wrapped form.
func (e *LoginThrottledError) Unwrap() error { return ErrLoginThrottled }

// LoginRetryAfter returns how long a throttled sign-in asks the caller to wait,
// or zero when err is not a throttling refusal.
func LoginRetryAfter(err error) time.Duration {
	var throttled *LoginThrottledError
	if errors.As(err, &throttled) {
		return throttled.RetryAfter
	}
	return 0
}

func IsUserDBUnavailable(err error) bool {
	return errors.Is(err, ErrUserDBUnavailable)
}

func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}

func IsInvalidCredentials(err error) bool {
	return errors.Is(err, ErrInvalidCredentials)
}

func IsUsernameExists(err error) bool {
	return errors.Is(err, ErrUsernameExists)
}

func IsInvalidUsername(err error) bool {
	return errors.Is(err, ErrInvalidUsername)
}

func IsUsernameTooLong(err error) bool {
	return errors.Is(err, ErrUsernameTooLong)
}

func IsPasswordTooShort(err error) bool {
	return errors.Is(err, ErrPasswordTooShort)
}

func IsPasswordTooLong(err error) bool {
	return errors.Is(err, ErrPasswordTooLong)
}

func IsOldPasswordIncorrect(err error) bool {
	return errors.Is(err, ErrOldPasswordIncorrect)
}

func IsAdminTransferSelf(err error) bool {
	return errors.Is(err, ErrAdminTransferSelf)
}

func IsAdminResetSelf(err error) bool {
	return errors.Is(err, ErrAdminResetSelf)
}

func IsAdminDeleteSelf(err error) bool {
	return errors.Is(err, ErrAdminDeleteSelf)
}

func IsAdminDeleteBlocked(err error) bool {
	return errors.Is(err, ErrAdminDeleteBlocked)
}

func IsTransferTargetAdmin(err error) bool {
	return errors.Is(err, ErrTransferTargetAdmin)
}

func IsAdminAccessDenied(err error) bool {
	return errors.Is(err, ErrAdminAccessDenied)
}

func IsAdminAlreadyExists(err error) bool {
	return errors.Is(err, ErrAdminAlreadyExists)
}

func IsPublicRegistrationDisabled(err error) bool {
	return errors.Is(err, ErrPublicRegistrationDisabled)
}

func IsSessionNotFound(err error) bool {
	return errors.Is(err, ErrSessionNotFound)
}

func IsSessionAccessDenied(err error) bool {
	return errors.Is(err, ErrSessionAccessDenied)
}

func IsSessionRevoked(err error) bool {
	return errors.Is(err, ErrSessionRevoked)
}

func IsSessionExpired(err error) bool {
	return errors.Is(err, ErrSessionExpired)
}

func IsLoginThrottled(err error) bool {
	return errors.Is(err, ErrLoginThrottled)
}
