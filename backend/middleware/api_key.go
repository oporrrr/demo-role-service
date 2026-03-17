package middleware

import (
	"github.com/gofiber/fiber/v2"
)

const APIKeyHeader = "X-API-Key"

func ExtractAPIKey(c *fiber.Ctx) error {
	key := c.Get(APIKeyHeader)
	if key == "" {
		return c.Status(401).JSON(fiber.Map{"code": "UNAUTHORIZED", "message": "X-API-Key header required"})
	}
	c.Locals("rawAPIKey", key)
	return c.Next()
}
