package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	sqlc "http-server/db/sqlc"
	"http-server/internal/controllers"
	"http-server/internal/routes"
	"http-server/internal/utils"
)

func main() {
	app := http.NewServeMux()
	err := godotenv.Load()
	if err != nil {
		fmt.Println("failed to load env")
	}

	port := os.Getenv("PORT")
	host := os.Getenv("HOST")
	dbUrl := os.Getenv("DATABASE_URL")

	db, error := utils.ConnectDB(dbUrl)
	if error != nil {
		fmt.Println("failed to connect to databse")
		panic("close")
	}
	queries := sqlc.New(db)

	ctrl := controller.NewController(queries)

	if host == "" || port == "" {
		panic("no host name env found")
	}

	routes.UserRouter(app, ctrl)

	fmt.Println("server running on port", port)

	if err := http.ListenAndServe(":"+port, app); err != nil {
		fmt.Println("server error:", err)
	}
}
