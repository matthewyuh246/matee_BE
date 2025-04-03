package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/matthewyuh246/Matee/internal/controller"
)

func NewRouter(uc controller.IUserController) *echo.Echo {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowCredentials: true,
	}))

	e.GET("/auth/github", uc.GitHubLogin)
	e.GET("/auth/github/callback", uc.GitHubCallback)
	e.GET("/logout", uc.Logout)

	return e
}
