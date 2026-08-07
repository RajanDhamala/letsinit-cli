package routes

import (
	"net/http"

	controller "http-server/internal/controllers"
)

func UserRouter(app *http.ServeMux, ctrl *controller.Controller) {
	app.HandleFunc("GET /login", ctrl.LoginHandler)
}
