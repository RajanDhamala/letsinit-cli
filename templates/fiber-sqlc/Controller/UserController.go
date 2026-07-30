package controller

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	utils "stackforge/fiber-sqlc/Utils"
	sqlc "stackforge/fiber-sqlc/db/sqlc"
)

type Controller struct {
	queries sqlc.Querier
}

func NewController(queries sqlc.Querier) *Controller {
	return &Controller{queries: queries}
}

type registerInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ctrl *Controller) Register(c *fiber.Ctx) error {
	var input registerInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}

	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Username == "" || input.Email == "" || len(input.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username, email, and a password of at least 8 characters are required",
		})
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "hash password failed"})
	}

	user, err := ctrl.queries.CreateUser(c.UserContext(), sqlc.CreateUserParams{
		Username: input.Username,
		Email:    input.Email,
		Password: hashedPassword,
	})
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "create user failed"})
	}

	accessToken, refreshToken, err := utils.CreateTokens(user.ID, user.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "create tokens failed"})
	}
	setAuthCookies(c, accessToken, refreshToken, os.Getenv("COOKIE_SECURE") == "true")

	return c.Status(fiber.StatusCreated).JSON(toUserResponse(user))
}

func (ctrl *Controller) Login(c *fiber.Ctx) error {
	var input loginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}

	user, err := ctrl.queries.GetUserByEmail(c.UserContext(), strings.ToLower(strings.TrimSpace(input.Email)))
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && utils.ComparePassword(input.Password, user.Password) != nil) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "login failed"})
	}

	accessToken, refreshToken, err := utils.CreateTokens(user.ID, user.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "create tokens failed"})
	}
	setAuthCookies(c, accessToken, refreshToken, os.Getenv("COOKIE_SECURE") == "true")

	return c.JSON(toUserResponse(user))
}

func (ctrl *Controller) Me(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(*utils.UserJWT)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	user, err := ctrl.queries.GetUserByID(c.UserContext(), claims.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "get user failed"})
	}

	return c.JSON(toUserResponse(user))
}

func (ctrl *Controller) Logout(c *fiber.Ctx) error {
	expires := time.Now().Add(-time.Hour)
	secure := os.Getenv("COOKIE_SECURE") == "true"
	c.Cookie(&fiber.Cookie{Name: "accessToken", Value: "", HTTPOnly: true, Secure: secure, SameSite: "Lax", Path: "/", Expires: expires})
	c.Cookie(&fiber.Cookie{Name: "refreshToken", Value: "", HTTPOnly: true, Secure: secure, SameSite: "Lax", Path: "/", Expires: expires})
	return c.SendStatus(fiber.StatusNoContent)
}

func setAuthCookies(c *fiber.Ctx, accessToken, refreshToken string, secure bool) {
	c.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func toUserResponse(user sqlc.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}
}
