package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ExtractBearerToken(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return c.Status(401).JSON(fiber.Map{"code": "UNAUTHORIZED", "message": "Bearer token required"})
	}
	c.Locals("userToken", strings.TrimPrefix(auth, "Bearer "))
	return c.Next()
}

func RequireAdminSecret(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		if token != secret {
			return c.Status(403).JSON(fiber.Map{"code": "FORBIDDEN", "message": "admin access required"})
		}
		return c.Next()
	}
}
