package services

import "errors"

var (
	ErrUserDBUnavailable    = errors.New("系统尚未配置数据库，请先完成初始设置")
	ErrUsernameExists       = errors.New("用户名已存在")
	ErrInvalidCredentials   = errors.New("用户名或密码错误")
	ErrUserNotFound         = errors.New("用户不存在")
	ErrOldPasswordIncorrect = errors.New("原密码错误")
	ErrInvalidUserRole      = errors.New("无效的角色值，必须是admin或user")
	ErrAdminTransferSelf    = errors.New("不能将管理员身份转让给自己")
	ErrTransferTargetAdmin  = errors.New("目标用户已经是管理员")
	ErrAdminAccessDenied    = errors.New("当前用户不是管理员，无法执行该操作")
	ErrPublicRegistrationDisabled = errors.New("公开注册已关闭，请联系管理员创建账户")
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

func IsOldPasswordIncorrect(err error) bool {
	return errors.Is(err, ErrOldPasswordIncorrect)
}

func IsInvalidUserRole(err error) bool {
	return errors.Is(err, ErrInvalidUserRole)
}

func IsAdminTransferSelf(err error) bool {
	return errors.Is(err, ErrAdminTransferSelf)
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
