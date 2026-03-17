package handler

import (
	"demo-role-service/service"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct{ svc *service.AdminService }

func NewAuthHandler(svc *service.AdminService) *AuthHandler { return &AuthHandler{svc: svc} }

// POST /auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil || req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "username and password required"})
	}
	token, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"code": "UNAUTHORIZED", "message": "invalid credentials"})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS", "data": fiber.Map{"token": token}})
}
