package routes

import (
	"net/http"

	controller "http-server/internal/controllers"
	"http-server/internal/middlewares"
)

func UserRouter(app *http.ServeMux, ctrl *controller.Controller) {
	app.HandleFunc("POST /users/register", ctrl.RegisterHandler)
	app.HandleFunc("POST /users/login", ctrl.LoginHandler)
	app.HandleFunc("GET /users/me", middleware.Auth(ctrl.MeHandler))
	app.HandleFunc("POST /users/logout", ctrl.LogoutHandler)
}
