package handler

import (
	"demo-role-service/entity"
	"demo-role-service/service"

	"github.com/gofiber/fiber/v2"
)

type MenuHandler struct{ svc *service.RoleService }

func NewMenuHandler(svc *service.RoleService) *MenuHandler { return &MenuHandler{svc: svc} }

// GET /menus?system=CRM  — admin: list all menus (flat, for management UI)
func (h *MenuHandler) List(c *fiber.Ctx) error {
	system := c.Query("system")
	if system == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "system required"})
	}
	menus, err := h.svc.ListMenus(system)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR"})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS", "data": menus})
}

// POST /menus  — admin: create menu item
func (h *MenuHandler) Create(c *fiber.Ctx) error {
	var m entity.Menu
	if err := c.BodyParser(&m); err != nil || m.SystemCode == "" || m.Name == "" || m.Code == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "systemCode, name and code required"})
	}
	if err := h.svc.CreateMenu(&m); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"code": "SUCCESS", "data": m})
}

// PUT /menus/:id  — admin: update menu item
func (h *MenuHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	var req struct {
		Name      string `json:"name"`
		Code      string `json:"code"`
		Icon      string `json:"icon"`
		Path      string `json:"path"`
		ParentID  *uint  `json:"parentId"`
		SortOrder int    `json:"sortOrder"`
		IsActive  bool   `json:"isActive"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	updates := map[string]interface{}{
		"name":       req.Name,
		"code":       req.Code,
		"icon":       req.Icon,
		"path":       req.Path,
		"parent_id":  req.ParentID,
		"sort_order": req.SortOrder,
		"is_active":  req.IsActive,
	}
	if err := h.svc.UpdateMenu(uint(id), updates); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}

// DELETE /menus/:id  — admin: delete menu item
func (h *MenuHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST"})
	}
	if err := h.svc.DeleteMenu(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR"})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS"})
}

// GET /internal/menus?system=CRM&accountId=user123
// Called by frontend after auth — returns tree filtered by user's permissions
func (h *MenuHandler) UserMenus(c *fiber.Ctx) error {
	accountID := c.Query("accountId")
	system := c.Query("system")
	if accountID == "" || system == "" {
		return c.Status(400).JSON(fiber.Map{"code": "BAD_REQUEST", "message": "accountId and system required"})
	}
	menus, err := h.svc.GetUserMenus(accountID, system)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": "ERROR"})
	}
	return c.JSON(fiber.Map{"code": "SUCCESS", "data": menus})
}
