package middleware

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"

	utils "stackforge/fiber-sqlc/Utils"
)

func AuthUser(c *fiber.Ctx) error {
	if accessToken := c.Cookies("accessToken"); accessToken != "" {
		claims, err := utils.VerifyAccessToken(accessToken)
		if err == nil {
			c.Locals("user", claims)
			return c.Next()
		}
	}

	refreshToken := c.Cookies("refreshToken")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication required"})
	}

	claims, err := utils.VerifyRefreshToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	accessToken, err := utils.CreateAccessToken(claims.ID, claims.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "create access token failed"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   os.Getenv("COOKIE_SECURE") == "true",
		SameSite: "Lax",
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
	})
	c.Locals("user", claims)
	return c.Next()
}
