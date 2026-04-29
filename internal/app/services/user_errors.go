package services

import "errors"

var (
	ErrUserDBUnavailable          = errors.New("database is not configured; complete initial setup first")
	ErrInvalidUsername            = errors.New("username is required")
	ErrUsernameExists             = errors.New("username already exists")
	ErrInvalidCredentials         = errors.New("username or password is incorrect")
	ErrUserNotFound               = errors.New("user not found")
	ErrOldPasswordIncorrect       = errors.New("current password is incorrect")
	ErrInvalidUserRole            = errors.New("invalid role; must be admin or user")
	ErrAdminTransferSelf          = errors.New("cannot transfer administrator role to yourself")
	ErrAdminResetSelf             = errors.New("cannot reset your own password; use the change password flow")
	ErrAdminDeleteSelf            = errors.New("cannot delete yourself; transfer administrator role first or use another administrator account")
	ErrAdminDeleteBlocked         = errors.New("cannot delete the current administrator account; transfer administrator role first")
	ErrTransferTargetAdmin        = errors.New("target user is already an administrator")
	ErrAdminAccessDenied          = errors.New("current user is not an administrator")
	ErrPublicRegistrationDisabled = errors.New("public registration is disabled; contact an administrator to create an account")
	ErrSessionNotFound            = errors.New("session not found")
	ErrSessionAccessDenied        = errors.New("session access denied")
	ErrSessionRevoked             = errors.New("session is no longer valid; sign in again")
	ErrSessionExpired             = errors.New("session expired; sign in again")
)

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

func IsOldPasswordIncorrect(err error) bool {
	return errors.Is(err, ErrOldPasswordIncorrect)
}

func IsInvalidUserRole(err error) bool {
	return errors.Is(err, ErrInvalidUserRole)
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
