package handler

import (
	"demo-role-service/entity"
	"demo-role-service/service"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct{ svc *service.AdminService }

func NewAuthHandler(svc *service.AdminService) *AuthHandler { return &AuthHandler{svc: svc} }

// POST /auth/login — admin management login
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

// ── User Auth (proxy to Auth Center) ─────────────────

type UserAuthHandler struct {
	svc     *service.UserService
	roleSvc *service.RoleService
}

func NewUserAuthHandler(svc *service.UserService, roleSvc *service.RoleService) *UserAuthHandler {
	return &UserAuthHandler{svc: svc, roleSvc: roleSvc}
}

// POST /api/v1/auth/login?system=<systemCode>
// Forwards login to Auth Center using the system's stored client credentials,
// then enriches the response with accountInformation, permissions, and menus.
func (h *UserAuthHandler) Login(c *fiber.Ctx) error {
	systemCode := c.Query("system")
	if systemCode == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "system query parameter is required"})
	}

	sys, err := h.roleSvc.GetSystem(systemCode)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": "NOT_FOUND", "message": "system not found"})
	}

	outcome, err := h.svc.Login(c.Body(), sys.AuthClientID, sys.AuthClientSecret)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "INTERNAL_ERROR", "message": "auth center unavailable"})
	}

	// pass-through for non-SUCCESS (wrong password, upstream errors, etc.)
	if outcome.Code != "SUCCESS" {
		c.Set("Content-Type", "application/json")
		return c.Status(outcome.StatusCode).Send(outcome.RawBody)
	}

	permissions := h.roleSvc.GetUserPermissions(outcome.Profile.ID, systemCode)
	if permissions == nil {
		permissions = []string{}
	}
	menus, _ := h.roleSvc.GetUserMenus(outcome.Profile.ID, systemCode)
	if menus == nil {
		menus = []entity.Menu{}
	}

	return c.Status(200).JSON(fiber.Map{
		"statusCode": 200,
		"code":       "SUCCESS",
		"data": fiber.Map{
			"accessToken":        outcome.AccessToken,
			"refreshToken":       outcome.RefreshToken,
			"expiresIn":          outcome.ExpiresIn,
			"refreshExpiresIn":   outcome.RefreshExpiresIn,
			"accountInformation": outcome.Profile,
			"permissions":        permissions,
			"menus":              menus,
		},
	})
}

// POST /api/v1/auth/logout — protected, pass-through to Auth Center
func (h *UserAuthHandler) Logout(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "refreshToken required"})
	}
	userToken := c.Locals("userToken").(string)
	body, statusCode, err := h.svc.Logout(req.RefreshToken, userToken)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "INTERNAL_ERROR", "message": "auth center unavailable"})
	}
	c.Set("Content-Type", "application/json")
	return c.Status(statusCode).Send(body)
}

// POST /api/v1/auth/register?system=<systemCode>
func (h *UserAuthHandler) Register(c *fiber.Ctx) error {
	systemCode := c.Query("system")
	if systemCode == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "system query parameter is required"})
	}

	sys, err := h.roleSvc.GetSystem(systemCode)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": "NOT_FOUND", "message": "system not found"})
	}

	var req service.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "invalid request body"})
	}
	if req.Email == "" && req.PhoneNumber == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "email or phoneNumber is required"})
	}
	if req.PhoneNumber != "" && req.CountryCode == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "countryCode is required when using phoneNumber"})
	}

	body, statusCode, err := h.svc.Register(req, sys.AuthClientID, sys.AuthClientSecret)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()})
	}
	c.Set("Content-Type", "application/json")
	return c.Status(statusCode).Send(body)
}
