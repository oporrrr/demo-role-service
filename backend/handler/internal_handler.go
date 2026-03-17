package handler

import (
	"demo-role-service/service"

	"github.com/gofiber/fiber/v2"
)

type InternalHandler struct{ svc *service.RoleService }

func NewInternalHandler(svc *service.RoleService) *InternalHandler { return &InternalHandler{svc: svc} }

// POST /internal/check  — called by other services to verify permission
func (h *InternalHandler) Check(c *fiber.Ctx) error {
	var req struct {
		AccountID  string `json:"accountId"`
		SystemCode string `json:"systemCode"`
		Resource   string `json:"resource"`
		Action     string `json:"action"`
	}
	if err := c.BodyParser(&req); err != nil || req.AccountID == "" || req.SystemCode == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	allowed := h.svc.Check(req.AccountID, req.SystemCode, req.Resource, req.Action)
	return c.JSON(fiber.Map{
		"code":    "SUCCESS",
		"allowed": allowed,
	})
}

// POST /internal/validate  — called by consumer services at startup
// Body: { "systemCode": "CRM", "required": ["order:view", "report:view"] }
// Response: { "code": "SUCCESS", "valid": true, "missing": [] }
//        or { "code": "INCOMPLETE", "valid": false, "missing": ["report:view"] }
func (h *InternalHandler) Validate(c *fiber.Ctx) error {
	var req struct {
		SystemCode string   `json:"systemCode"`
		Required   []string `json:"required"`
	}
	if err := c.BodyParser(&req); err != nil || req.SystemCode == "" || len(req.Required) == 0 {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "systemCode and required[] are required"})
	}
	missing, err := h.svc.ValidatePermissions(req.SystemCode, req.Required)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR"})
	}
	if len(missing) > 0 {
		return c.Status(200).JSON(fiber.Map{
			"code":    "INCOMPLETE",
			"valid":   false,
			"missing": missing,
		})
	}
	return c.JSON(fiber.Map{
		"code":    "SUCCESS",
		"valid":   true,
		"missing": []string{},
	})
}

// GET /internal/permissions?accountId=xxx&system=xxx — get full permission list for frontend
func (h *InternalHandler) GetPermissions(c *fiber.Ctx) error {
	accountID := c.Query("accountId")
	systemCode := c.Query("system")
	if accountID == "" || systemCode == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "accountId and system required"})
	}
	perms := h.svc.GetUserPermissions(accountID, systemCode)
	return c.JSON(fiber.Map{"code": "SUCCESS", "data": perms})
}
