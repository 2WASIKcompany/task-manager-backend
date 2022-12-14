package api

import (
	"context"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
	"go.uber.org/fx"
	"task-manager-backend/internal/app/config"
)

const (
	Title    = "Balance handler API"
	BasePath = "/api/v1/"
)

type Api struct {
	router *gin.Engine
}

func (api *Api) Run() {
	cfg := config.Load()
	docs.SwaggerInfo.BasePath = BasePath

	api.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.DefaultModelsExpandDepth(-1)))

	api.router.Use(CORSMiddleware())
	api.router.Run(cfg.Api.GetAddr())
}

func NewApi(
	router *gin.Engine,
) *Api {
	svc := &Api{
		router: router,
	}
	svc.registerRoutes()
	return svc
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func StartHook(lifecycle fx.Lifecycle, api *Api) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go api.Run()
				return nil
			},
		})
}

func (api *Api) registerRoutes() {
	base := api.router.Group(BasePath)
	baseWithAuth := base.Group("/")
	baseWithAuth.Use(api.AuthMW())

	auth := base.Group("/auth")
	auth.POST("/signup", api.SignUp)
	auth.POST("/signin", api.SignIn)
	auth.POST("/refresh", api.Refresh)
	auth.GET("/confirm/:confirm_uid", api.Confirmation)
	auth.POST("/restore_password", api.RestorePassword)
	auth.POST("/new_password", api.NewPassword)
}
