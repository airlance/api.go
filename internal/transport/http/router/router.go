package router

import (
	"net/http"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	"github.com/resoul/api/internal/config"
	"github.com/resoul/api/internal/domain"
	"github.com/resoul/api/internal/middleware"
	"github.com/resoul/api/internal/transport/http/handlers"
	"github.com/resoul/api/internal/transport/http/utils"
	"gorm.io/gorm"
)

func New(cfg *config.Config, db *gorm.DB, accountSvc domain.AccountService) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://dashboard.studio.localhost"},
		AllowMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	accountHandler := handlers.NewAccountHandler(accountSvc)

	api := r.Group("/api/v1")
	{
		// Public routes — no auth required.
		api.GET("/health", func(c *gin.Context) {
			utils.RespondOK(c, gin.H{"status": "ok"})
		})

		// Authenticated routes.
		authed := api.Group("/", middleware.Auth(cfg.Auth.JWTSecret))
		{
			authed.GET("/user/me", accountHandler.GetMe)
			authed.PATCH("/user/account", accountHandler.UpdateAccount)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		utils.RespondError(c, http.StatusNotFound, "not_found", "the requested resource does not exist")
	})

	return r
}
