package handler

import (
	"demo-role-service/service"

	"github.com/gofiber/fiber/v2"
)

type PermissionHandler struct{ svc *service.RoleService }

func NewPermissionHandler(svc *service.RoleService) *PermissionHandler {
	return &PermissionHandler{svc: svc}
}

func (h *PermissionHandler) BulkRegister(c *fiber.Ctx) error {
	systemCode := c.Params("code")
	var req struct {
		Permissions []service.PermissionInput `json:"permissions"`
	}
	if err := c.BodyParser(&req); err != nil || len(req.Permissions) == 0 {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "permissions array required"})
	}
	perms, err := h.svc.BulkRegisterPermissions(systemCode, req.Permissions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"code": "SUCCESS", "data": perms})
}

func (h *PermissionHandler) List(c *fiber.Ctx) error {
	systemCode := c.Query("system")
	perms, err := h.svc.ListPermissions(systemCode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS", "data": perms})
}

func (h *PermissionHandler) Create(c *fiber.Ctx) error {
	var req struct {
		SystemCode  string `json:"systemCode"`
		Resource    string `json:"resource"`
		Action      string `json:"action"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil || req.SystemCode == "" || req.Resource == "" || req.Action == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "systemCode, resource, action required"})
	}
	p, err := h.svc.CreatePermission(req.SystemCode, req.Resource, req.Action, req.Description)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"code": "SUCCESS", "data": p})
}

func (h *PermissionHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	if err := h.svc.DeletePermission(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}
