package main

import (
	"log"

	"demo-role-service/config"
	"demo-role-service/database"
	"demo-role-service/handler"
	"demo-role-service/middleware"
	"demo-role-service/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	config.Load()
	database.Connect()
	database.ConnectRedis()

	app := fiber.New(fiber.Config{AppName: "Demo Role Service"})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	svc := service.NewRoleService(database.DB, database.Redis)

	adminSvc := service.NewAdminService(database.DB)
	// seed first admin user if ADMIN_INITIAL_PASSWORD is set and no admin exists yet
	if config.Cfg.AdminInitialPassword != "" {
		if err := adminSvc.Bootstrap(config.Cfg.AdminInitialPassword); err != nil {
			log.Printf("admin bootstrap: %v", err)
		}
	}

	systemH := handler.NewSystemHandler(svc)
	roleH := handler.NewRoleHandler(svc)
	permH := handler.NewPermissionHandler(svc)
	userRoleH := handler.NewUserRoleHandler(svc)
	internalH := handler.NewInternalHandler(svc)
	menuH := handler.NewMenuHandler(svc)
	authH := handler.NewAuthHandler(adminSvc)

	// ── Auth (management frontend login) ───────────────
	app.Post("/auth/login", authH.Login)

	// ── System Registration ────────────────────────────
	app.Post("/systems/register", middleware.RequireAdmin, systemH.Register)
	app.Post("/systems/:code/bootstrap", middleware.RequireAdmin, systemH.Bootstrap)
	app.Post("/systems/:code/rekey", middleware.RequireAdmin, systemH.ReKey)
	app.Get("/systems", middleware.RequireAdmin, systemH.List)

	// ── Permissions ────────────────────────────────────
	app.Post("/systems/:code/permissions", middleware.ExtractAPIKey, permH.BulkRegister)
	app.Get("/permissions", middleware.RequireAdmin, permH.List)
	app.Post("/permissions", middleware.RequireAdmin, permH.Create)
	app.Delete("/permissions/:id", middleware.RequireAdmin, permH.Delete)

	// ── Roles ──────────────────────────────────────────
	app.Get("/roles", middleware.RequireAdmin, roleH.List)
	app.Post("/roles", middleware.RequireAdmin, roleH.Create)
	app.Get("/roles/:id", middleware.RequireAdmin, roleH.Get)
	app.Put("/roles/:id", middleware.RequireAdmin, roleH.Update)
	app.Delete("/roles/:id", middleware.RequireAdmin, roleH.Delete)

	app.Put("/roles/:id/default", middleware.RequireAdmin, roleH.SetDefault)
	app.Put("/roles/:id/permissions", middleware.RequireAdmin, roleH.SetPermissions)
	app.Post("/roles/:id/permissions", middleware.RequireAdmin, roleH.AddPermissions)
	app.Delete("/roles/:id/permissions/:pid", middleware.RequireAdmin, roleH.RemovePermission)

	// ── User Role Assignment ───────────────────────────
	app.Get("/users", middleware.RequireAdmin, userRoleH.ListBySystem)
	app.Get("/users/:accountId/roles", middleware.RequireAdmin, userRoleH.GetRoles)
	app.Put("/users/:accountId/role", middleware.RequireAdmin, userRoleH.Assign)
	app.Delete("/users/:accountId/role", middleware.RequireAdmin, userRoleH.Remove)

	// ── Menu Management (admin) ────────────────────────
	app.Get("/menus", middleware.RequireAdmin, menuH.List)
	app.Post("/menus", middleware.RequireAdmin, menuH.Create)
	app.Put("/menus/:id", middleware.RequireAdmin, menuH.Update)
	app.Delete("/menus/:id", middleware.RequireAdmin, menuH.Delete)

	// ── Internal (service-to-service via API key) ──────
	app.Post("/internal/check", middleware.ExtractAPIKey, internalH.Check)
	app.Post("/internal/validate", middleware.ExtractAPIKey, internalH.Validate)
	app.Get("/internal/permissions", middleware.ExtractAPIKey, internalH.GetPermissions)
	app.Get("/internal/menus", middleware.ExtractAPIKey, menuH.UserMenus)

	log.Printf("role-service running on port %s", config.Cfg.AppPort)
	log.Fatal(app.Listen(":" + config.Cfg.AppPort))
}
