package auth

const (
	confimEmailMsg = "To: %s\r\nFrom: alexeyshish92@gmail.com\r\nSubject: Подверждение почты\r\n\r\nДля подверждения регистрации перейдите по ссылке: http://localhost:5173/auth/confirm/%s"
	changePassMsg  = "To: %s\r\nFrom: alexeyshish92@gmail.com\r\nSubject: Смена пароля\r\n\r\nДля смены пароля перейдите по ссылке: http://localhost:5173/auth/restore_password/%s"
)
