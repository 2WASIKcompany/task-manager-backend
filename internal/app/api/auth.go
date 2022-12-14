package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

type Auth struct {
	Email   string `json:"email"`
	PwdHash string `json:"pwd_hash"`
}

type Tokens struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// SignUp godoc
// @Summary Регистрация
// @Schemes
// @Description Прямая регистрация нового пользователя в системе
// @Tags auth
// @Accept json
// @Produce json
// @Param data body Auth true "Входные параметры"
// @Success 200
// @Router /auth/signup [post]
func (api *Api) SignUp(ctx *gin.Context) {
	var req Auth
	if err := ctx.BindJSON(&req); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := api.auth.Register(ctx, string(req.PwdHash), string(req.Email))
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

// SignIn godoc
// @Summary Вход в систему
// @Schemes
// @Description Вход в систему по логину и хешу-пароля
// @Tags auth
// @Accept json
// @Produce json
// @Param data body Auth true "Входные параметры"
// @Success 200 {object} Tokens
// @Router /auth/signin [post]
func (api *Api) SignIn(ctx *gin.Context) {
	var req Auth
	if err := ctx.BindJSON(&req); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := api.auth.Auth(ctx, string(req.PwdHash), string(req.Email))
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, Tokens{Token: session.Token, RefreshToken: session.RefreshToken})
}

type Refresh struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh godoc
// @Summary Обновить JWT
// @Schemes
// @Description Обновляет JWT по refresh токену
// @Description Для того что бы обновить токен надо быть
// @Description аунтифицированным
// @Tags auth
// @Accept json
// @Produce json
// @Param data body Refresh true "Входные параметры"
// @Success 200 {object} Tokens
// @Router /auth/refresh [post]
func (api *Api) Refresh(ctx *gin.Context) {
	var token Refresh
	if err := ctx.BindJSON(&token); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := api.auth.RefreshToken(ctx, token.RefreshToken)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, Tokens{Token: session.Token, RefreshToken: session.RefreshToken})
}

func (api *Api) AuthMW() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := extractAuthToken(ctx)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID, err := api.auth.UnmarshalToken(token)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		putUserIDtoContext(ctx, userID)
	}
}

func putUserIDtoContext(ctx *gin.Context, userID users.ID) {
	ctx.Set(userIDContextKey, userID)
}

func popUserIDfromContext(ctx *gin.Context) (users.ID, error) {
	userID, ok := ctx.Get(userIDContextKey)
	if !ok {
		return users.ID(0), errors.New("no user id in ctx")
	}

	return userID.(users.ID), nil
}

func extractAuthToken(ctx *gin.Context) (string, error) {
	header := ctx.GetHeader(authHeader)
	if header == "" {
		return "", errors.New("request has no auth header")
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 {
		return "", errors.New("request has no auth header")
	}

	if parts[0] != bearerPrefix {
		return "", errors.New("request has no auth header")
	}

	return parts[1], nil
}

type Confirmation struct {
	UID users.ConfirmationUID `uri:"confirm_uid" binding:"required"`
}

// Confirmation godoc
// @Summary Подтверждение регистрации
// @Schemes
// @Description Подтверждает регистрацию пользователя
// @Tags auth
// @Success 200
// @Param confirm_uid path string true "uid конфирмации"
// @Router /auth/confirm/{confirm_uid} [get]
func (api *Api) Confirmation(ctx *gin.Context) {
	var confirmationUID Confirmation
	if err := ctx.ShouldBindUri(&confirmationUID); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := api.auth.ConfirmationUser(ctx, confirmationUID.UID)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Redirect(http.StatusFound, "https://cartips.ru/static/email_ok.html")
}

type RestorePasswordEmail struct {
	Email string `json:"email"`
}

// RestorePassword godoc
// @Summary Отправка ссылки для восстановления пароля
// @Schemes
// @Description Отправляет ссылку на страницу с восстановлением пароля
// @Tags auth
// @Accept json
// @Param data body RestorePasswordEmail true "Входные параметры"
// @Success 200
// @Router /auth/restore_password [post]
func (api *Api) RestorePassword(ctx *gin.Context) {
	var restoreEmail RestorePasswordEmail
	if err := ctx.BindJSON(&restoreEmail); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := api.auth.SendRestorePasswordMail(ctx, restoreEmail.Email)
	if err == auth.NotFoundEmailErr {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

type ChangePassword struct {
	RestoreUID  string `json:"restore_uid"`
	NewPassword string `json:"new_password"`
}

// NewPassword godoc
// @Summary Восстановление пароля
// @Schemes
// @Description Меняет пароль пользователя на новый
// @Tags auth
// @Accept json
// @Param data body ChangePassword true "Входные параметры"
// @Success 200
// @Router /auth/new_password [post]
func (api *Api) NewPassword(ctx *gin.Context) {
	var changePassword ChangePassword
	if err := ctx.BindJSON(&changePassword); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := api.auth.ChangePassword(ctx, changePassword.RestoreUID, changePassword.NewPassword)
	if err == auth.NotFoundRestoreUIDErr {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}
