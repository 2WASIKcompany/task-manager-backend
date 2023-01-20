package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"task-manager-backend/internal/app/models/users"
	"task-manager-backend/internal/app/service/auth"
)

const (
	authHeader       = "Authorization"
	bearerPrefix     = "Bearer"
	userIDContextKey = "uid"
)

type Error struct {
	Err string `json:"err"`
}

type Auth struct {
	Email   users.Email `json:"email"`
	PwdHash string      `json:"pwd_hash"`
}

type Tokens struct {
	Session users.Session `json:"tokens"`
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
// @Failure 400 {object} Error
// @Failure 403 {object} Error
// @Failure 500
// @Router /auth/signup [post]
func (api *Api) SignUp(ctx *gin.Context) {
	var req Auth
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Error{Err: auth.InvalidData.Error()})
		return
	}

	err := api.auth.Register(ctx, string(req.PwdHash), req.Email)
	if err == auth.InvalidData {
		ctx.JSON(http.StatusBadRequest, Error{Err: auth.InvalidData.Error()})
		return
	}
	if err == auth.UserAlreadyExist {
		ctx.JSON(http.StatusForbidden, Error{Err: auth.UserAlreadyExist.Error()})
		return
	}
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

// Logout godoc
// @Summary Выход с аккаунта
// @Schemes
// @Description Инвалидирует сессию для устройства, с которого выполняется выход
// @Tags auth
// @Accept json
// @Param data body Refresh true "Входные параметры"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /auth/logout [post]
func (api *Api) Logout(ctx *gin.Context) {
	var refresh Refresh
	if err := ctx.BindJSON(&refresh); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := api.auth.Logout(ctx, refresh.RefreshToken); err != nil {
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
// @Failure 400 {object} Error
// @Failure 403 {object} Error
// @Failure 500
// @Router /auth/signin [post]
func (api *Api) SignIn(ctx *gin.Context) {
	var req Auth
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Error{Err: auth.InvalidData.Error()})
		return
	}

	session, err := api.auth.Auth(ctx, string(req.PwdHash), req.Email)
	if err == auth.InvalidData {
		ctx.JSON(http.StatusBadRequest, Error{Err: auth.InvalidData.Error()})
		return
	}
	if err == auth.NonConfirmed {
		ctx.JSON(http.StatusForbidden, Error{Err: auth.NonConfirmed.Error()})
		return
	}
	if err == auth.IncorrectCreds {
		ctx.JSON(http.StatusForbidden, Error{Err: auth.IncorrectCreds.Error()})
		return
	}
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, Tokens{Session: session})
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
// @Failure 400
// @Failure 500
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

	ctx.JSON(http.StatusOK, Tokens{Session: session})
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
	Token string `uri:"confirm_token" binding:"required"`
}

// Confirmation godoc
// @Summary Подтверждение регистрации
// @Schemes
// @Description Подтверждает регистрацию пользователя
// @Tags auth
// @Success 200 {object} Tokens
// @Failure 400
// @Failure 500
// @Param confirm_token path string true "token конфирмации"
// @Router /auth/confirm/{confirm_token} [get]
func (api *Api) Confirmation(ctx *gin.Context) {
	var refresh Confirmation
	if err := ctx.ShouldBindUri(&refresh); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := api.auth.ConfirmationUser(ctx, refresh.Token)
	if err == auth.InvalidRefresh {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, Tokens{Session: session})
}

type RestorePasswordEmail struct {
	Email users.Email `json:"email"`
}

// RestorePassword godoc
// @Summary Отправка ссылки для восстановления пароля
// @Schemes
// @Description Отправляет ссылку на страницу с восстановлением пароля
// @Tags auth
// @Accept json
// @Param data body RestorePasswordEmail true "Входные параметры"
// @Success 200
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500
// @Router /auth/restore_password [post]
func (api *Api) RestorePassword(ctx *gin.Context) {
	var restoreEmail RestorePasswordEmail
	if err := ctx.BindJSON(&restoreEmail); err != nil {
		ctx.JSON(http.StatusBadRequest, Error{Err: auth.InvalidData.Error()})
		return
	}

	err := api.auth.SendRestorePasswordMail(ctx, restoreEmail.Email)
	if err == auth.InvalidData {
		ctx.JSON(http.StatusBadRequest, Error{Err: auth.InvalidData.Error()})
		return
	}
	if err == auth.NotFoundEmail {
		ctx.JSON(http.StatusNotFound, Error{Err: auth.NotFoundEmail.Error()})
		return
	} else if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

type ChangePassword struct {
	RestoreRefresh string `json:"restore_refresh"`
	NewPassword    string `json:"new_password"`
}

// NewPassword godoc
// @Summary Восстановление пароля
// @Schemes
// @Description Меняет пароль пользователя на новый
// @Tags auth
// @Accept json
// @Param data body ChangePassword true "Входные параметры"
// @Success 200 {object} Tokens
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500
// @Router /auth/new_password [post]
func (api *Api) NewPassword(ctx *gin.Context) {
	var changePassword ChangePassword
	if err := ctx.BindJSON(&changePassword); err != nil {
		ctx.JSON(http.StatusBadRequest, Error{Err: auth.InvalidData.Error()})
		return
	}

	session, err := api.auth.ChangePassword(ctx, changePassword.RestoreRefresh, changePassword.NewPassword)
	if err == auth.InvalidData {
		ctx.JSON(http.StatusBadRequest, Error{Err: auth.InvalidData.Error()})
		return
	}
	if err == auth.InvalidRefresh {
		ctx.JSON(http.StatusBadRequest, Error{Err: auth.InvalidRefresh.Error()})
		return
	}
	if err == auth.SamePassword {
		ctx.JSON(http.StatusBadRequest, Error{Err: auth.SamePassword.Error()})
		return
	}
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, Tokens{Session: session})
}
