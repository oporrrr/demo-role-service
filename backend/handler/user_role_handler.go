package handler

import (
	"demo-role-service/service"

	"github.com/gofiber/fiber/v2"
)

type UserRoleHandler struct{ svc *service.RoleService }

func NewUserRoleHandler(svc *service.RoleService) *UserRoleHandler { return &UserRoleHandler{svc: svc} }

func (h *UserRoleHandler) Assign(c *fiber.Ctx) error {
	accountID := c.Params("accountId")
	var req struct {
		SystemCode string `json:"systemCode"`
		RoleID     uint   `json:"roleId"`
	}
	if err := c.BodyParser(&req); err != nil || req.SystemCode == "" || req.RoleID == 0 {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "systemCode and roleId required"})
	}
	if err := h.svc.AssignRole(accountID, req.SystemCode, req.RoleID); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}

func (h *UserRoleHandler) GetRoles(c *fiber.Ctx) error {
	accountID := c.Params("accountId")
	roles, err := h.svc.GetUserRoles(accountID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS", "data": roles})
}

// GET /users?system=CRM  — list all users who have a role in the given system (including removed)
func (h *UserRoleHandler) ListBySystem(c *fiber.Ctx) error {
	systemCode := c.Query("system")
	if systemCode == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "system required"})
	}
	list, err := h.svc.ListUsersBySystem(systemCode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	type row struct {
		ID        uint        `json:"id"`
		AccountID string      `json:"accountId"`
		SystemCode string     `json:"systemCode"`
		Role      interface{} `json:"role"`
		IsActive  bool        `json:"isActive"`
	}
	out := make([]row, len(list))
	for i, ur := range list {
		out[i] = row{
			ID:        ur.ID,
			AccountID: ur.AccountID,
			SystemCode: ur.SystemCode,
			Role:      ur.Role,
			IsActive:  !ur.DeletedAt.Valid,
		}
	}
	return c.JSON(fiber.Map{"code": "SUCCESS", "data": out})
}

func (h *UserRoleHandler) Remove(c *fiber.Ctx) error {
	accountID := c.Params("accountId")
	systemCode := c.Query("system")
	if systemCode == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "system query required"})
	}
	if err := h.svc.RemoveUserRole(accountID, systemCode); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}
