package handler

import (
	"demo-role-service/service"

	"github.com/gofiber/fiber/v2"
)

type SystemHandler struct{ svc *service.RoleService }

func NewSystemHandler(svc *service.RoleService) *SystemHandler { return &SystemHandler{svc: svc} }

func (h *SystemHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Code             string `json:"code"`
		Name             string `json:"name"`
		Description      string `json:"description"`
		AuthClientID     string `json:"authClientId"`
		AuthClientSecret string `json:"authClientSecret"`
	}
	if err := c.BodyParser(&req); err != nil || req.Code == "" || req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "code and name are required"})
	}
	apiKey, err := h.svc.RegisterSystem(req.Code, req.Name, req.Description, req.AuthClientID, req.AuthClientSecret)
	if err != nil {
		return c.Status(409).JSON(fiber.Map{"code": "CONFLICT", "message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{
		"code":    "SUCCESS",
		"message": "system registered — save the apiKey, it won't be shown again",
		"data":    fiber.Map{"code": req.Code, "apiKey": apiKey},
	})
}

// POST /systems/:code/bootstrap  — first-time setup: creates Super Admin role + assigns to accountId
// Fails if the system already has roles (prevents accidental overwrite)
func (h *SystemHandler) Bootstrap(c *fiber.Ctx) error {
	code := c.Params("code")
	var req struct {
		AccountID string `json:"accountId"`
	}
	if err := c.BodyParser(&req); err != nil || code == "" || req.AccountID == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "accountId required"})
	}
	if err := h.svc.BootstrapSystem(code, req.AccountID); err != nil {
		return c.Status(409).JSON(fiber.Map{"code": "CONFLICT", "message": err.Error()})
	}
	return c.JSON(fiber.Map{
		"code":    "SUCCESS",
		"message": "Super Admin role created and assigned — go to Role Manager to create proper roles",
	})
}

// PUT /systems/:code/credentials  — update Auth Center client credentials for a system
func (h *SystemHandler) UpdateCredentials(c *fiber.Ctx) error {
	code := c.Params("code")
	var req struct {
		AuthClientID     string `json:"authClientId"`
		AuthClientSecret string `json:"authClientSecret"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "invalid body"})
	}
	if err := h.svc.UpdateSystemCredentials(code, req.AuthClientID, req.AuthClientSecret); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}

// POST /systems/:code/rekey  — generate a new API key for a system (old key is revoked immediately)
func (h *SystemHandler) ReKey(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	apiKey, err := h.svc.ReKeySystem(code)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{
		"code":    "SUCCESS",
		"message": "API key rotated — save the new apiKey, it won't be shown again",
		"data":    fiber.Map{"code": code, "apiKey": apiKey},
	})
}

func (h *SystemHandler) List(c *fiber.Ctx) error {
	list, err := h.svc.ListSystems()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	// strip apiKey from response
	type safeSystem struct {
		ID          uint   `json:"id"`
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	var result []safeSystem
	for _, s := range list {
		result = append(result, safeSystem{ID: s.ID, Code: s.Code, Name: s.Name, Description: s.Description})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS", "data": result})
}
