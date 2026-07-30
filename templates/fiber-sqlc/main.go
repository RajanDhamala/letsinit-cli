package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	controller "stackforge/fiber-sqlc/Controller"
	route "stackforge/fiber-sqlc/Route"
	sqlc "stackforge/fiber-sqlc/db/sqlc"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal("load .env: ", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required in .env or the process environment")
	}
	if os.Getenv("ACCESS_TOKEN_SECRET") == "" {
		log.Fatal("ACCESS_TOKEN_SECRET is required in .env or the process environment")
	}
	if os.Getenv("REFRESH_TOKEN_SECRET") == "" {
		log.Fatal("REFRESH_TOKEN_SECRET is required in .env or the process environment")
	}

	dbPool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("create database pool: ", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatal("connect to database: ", err)
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	app := fiber.New()
	app.Use(recover.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Fiber SQLC API is running"})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		if err := dbPool.Ping(c.UserContext()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "database unavailable",
			})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	ctrl := controller.NewController(sqlc.New(dbPool))
	route.UserRouter(app, ctrl)

	log.Printf("server listening on http://%s:%s", host, port)
	log.Fatal(app.Listen(host + ":" + port))
}
