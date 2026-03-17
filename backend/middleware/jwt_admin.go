package middleware

import (
	"strings"

	"demo-role-service/config"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// RequireAdmin accepts either:
//   - Legacy ADMIN_SECRET bearer token (backwards compat)
//   - JWT issued by POST /auth/login
func RequireAdmin(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return c.Status(401).JSON(fiber.Map{"code": "UNAUTHORIZED", "message": "Bearer token required"})
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")

	// legacy shared secret — still works
	if tokenStr == config.Cfg.AdminSecret {
		return c.Next()
	}

	// verify JWT
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(config.Cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"code": "UNAUTHORIZED", "message": "invalid token"})
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		c.Locals("adminUser", claims["sub"])
	}
	return c.Next()
}
