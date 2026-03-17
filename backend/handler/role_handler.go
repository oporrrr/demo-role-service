package handler

import (
	"demo-role-service/service"

	"github.com/gofiber/fiber/v2"
)

type RoleHandler struct{ svc *service.RoleService }

func NewRoleHandler(svc *service.RoleService) *RoleHandler { return &RoleHandler{svc: svc} }

func (h *RoleHandler) Create(c *fiber.Ctx) error {
	var req struct {
		SystemCode  string `json:"systemCode"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil || req.SystemCode == "" || req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "systemCode and name required"})
	}
	role, err := h.svc.CreateRole(req.SystemCode, req.Name, req.Description)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"code": "SUCCESS", "data": role})
}

func (h *RoleHandler) List(c *fiber.Ctx) error {
	systemCode := c.Query("system")
	roles, err := h.svc.ListRoles(systemCode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS", "data": roles})
}

func (h *RoleHandler) Get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	role, err := h.svc.GetRole(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": "NOT_FOUND"})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS", "data": role})
}

func (h *RoleHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	if err := h.svc.UpdateRole(uint(id), req.Name, req.Description); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}

func (h *RoleHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	if err := h.svc.DeleteRole(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}

// PUT /roles/:id/default  — mark this role as the default for new users in its system
func (h *RoleHandler) SetDefault(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	role, err := h.svc.GetRole(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": "NOT_FOUND"})
	}
	if err := h.svc.SetDefaultRole(uint(id), role.SystemCode); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}

func (h *RoleHandler) SetPermissions(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	var req struct {
		PermissionIDs []uint `json:"permissionIds"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	if err := h.svc.SetRolePermissions(uint(id), req.PermissionIDs); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}

func (h *RoleHandler) AddPermissions(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	var req struct {
		PermissionIDs []uint `json:"permissionIds"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	if err := h.svc.AddRolePermissions(uint(id), req.PermissionIDs); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}

func (h *RoleHandler) RemovePermission(c *fiber.Ctx) error {
	roleID, err1 := c.ParamsInt("id")
	permID, err2 := c.ParamsInt("pid")
	if err1 != nil || err2 != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	if err := h.svc.RemoveRolePermission(uint(roleID), uint(permID)); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}
