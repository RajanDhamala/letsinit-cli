package route

import (
	"github.com/gofiber/fiber/v2"

	controller "letsinit/fiber-sqlc/Controller"
	middleware "letsinit/fiber-sqlc/Middleware"
)

func UserRouter(app *fiber.App, ctrl *controller.Controller) {
	users := app.Group("/users")
	users.Post("/register", ctrl.Register)
	users.Post("/login", ctrl.Login)
	users.Get("/me", middleware.AuthUser, ctrl.Me)
	users.Post("/logout", ctrl.Logout)
}
