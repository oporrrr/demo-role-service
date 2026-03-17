package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"demo-role-service/entity"
	"demo-role-service/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var rctx = context.Background()

type cachedSystem struct {
	system *entity.System
	expiry time.Time
}

type RoleService struct {
	systemRepo     *repository.SystemRepository
	roleRepo       *repository.RoleRepository
	permissionRepo *repository.PermissionRepository
	userRoleRepo   *repository.UserRoleRepository
	menuRepo       *repository.MenuRepository
	redis          *redis.Client // nil = cache disabled
	systemCache    sync.Map      // key: systemCode → *cachedSystem
}

func NewRoleService(db *gorm.DB, rdb *redis.Client) *RoleService {
	return &RoleService{
		systemRepo:     repository.NewSystemRepository(db),
		roleRepo:       repository.NewRoleRepository(db),
		permissionRepo: repository.NewPermissionRepository(db),
		userRoleRepo:   repository.NewUserRoleRepository(db),
		menuRepo:       repository.NewMenuRepository(db),
		redis:          rdb,
	}
}

// ── System ────────────────────────────────────────────

func (s *RoleService) RegisterSystem(code, name, description, authClientID, authClientSecret string) (string, error) {
	raw, err := generateAPIKey()
	if err != nil {
		return "", err
	}
	sys := &entity.System{
		Code:             code,
		Name:             name,
		Description:      description,
		APIKey:           repository.HashAPIKey(raw),
		AuthClientID:     authClientID,
		AuthClientSecret: authClientSecret,
	}
	if err := s.systemRepo.Create(sys); err != nil {
		return "", fmt.Errorf("system already exists or DB error: %w", err)
	}
	return raw, nil
}

const systemCacheTTL = 5 * time.Minute

func (s *RoleService) GetSystem(code string) (*entity.System, error) {
	if v, ok := s.systemCache.Load(code); ok {
		if cs := v.(*cachedSystem); time.Now().Before(cs.expiry) {
			return cs.system, nil
		}
		s.systemCache.Delete(code)
	}
	sys, err := s.systemRepo.FindByCode(code)
	if err != nil {
		return nil, err
	}
	s.systemCache.Store(code, &cachedSystem{system: sys, expiry: time.Now().Add(systemCacheTTL)})
	return sys, nil
}

func (s *RoleService) ListSystems() ([]entity.System, error) {
	return s.systemRepo.List()
}

// BootstrapSystem creates a "Super Admin" role with wildcard permission (*:*)
// and assigns it to accountID. Only works when the system has no roles yet.
func (s *RoleService) BootstrapSystem(systemCode, accountID string) error {
	existing, err := s.roleRepo.List(systemCode)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return fmt.Errorf("system %q already has roles — assign manually via Role Manager", systemCode)
	}

	// 1. wildcard permission *:*
	wildcard := &entity.Permission{
		SystemCode:  systemCode,
		Resource:    "*",
		Action:      "*",
		Description: "Super Admin — full access (auto-created by bootstrap)",
	}
	if err := s.permissionRepo.Create(wildcard); err != nil {
		return fmt.Errorf("create wildcard permission: %w", err)
	}

	// 2. Super Admin role
	role := &entity.Role{
		SystemCode:  systemCode,
		Name:        "Super Admin",
		Description: "Full access — created by bootstrap",
	}
	if err := s.roleRepo.Create(role); err != nil {
		return fmt.Errorf("create role: %w", err)
	}

	// 3. assign permission → role
	if err := s.roleRepo.SetPermissions(role.ID, []uint{wildcard.ID}); err != nil {
		return fmt.Errorf("assign permission: %w", err)
	}

	// 4. assign role → user
	return s.userRoleRepo.Set(accountID, systemCode, role.ID)
}

func (s *RoleService) UpdateSystemCredentials(code, clientID, clientSecret string) error {
	if err := s.systemRepo.UpdateCredentials(code, clientID, clientSecret); err != nil {
		return err
	}
	s.systemCache.Delete(code)
	return nil
}

func (s *RoleService) ReKeySystem(code string) (string, error) {
	raw, err := generateAPIKey()
	if err != nil {
		return "", err
	}
	if err := s.systemRepo.UpdateAPIKey(code, repository.HashAPIKey(raw)); err != nil {
		return "", err
	}
	return raw, nil
}

func (s *RoleService) ValidateAPIKey(raw string) (*entity.System, error) {
	hashed := repository.HashAPIKey(raw)
	sys, err := s.systemRepo.FindByAPIKey(hashed)
	if err != nil {
		return nil, errors.New("invalid API key")
	}
	return sys, nil
}

// ── Permissions ───────────────────────────────────────

type PermissionInput struct {
	Resource    string
	Action      string
	Description string
}

func (s *RoleService) BulkRegisterPermissions(systemCode string, inputs []PermissionInput) ([]entity.Permission, error) {
	var perms []entity.Permission
	for _, in := range inputs {
		perms = append(perms, entity.Permission{
			SystemCode:  systemCode,
			Resource:    in.Resource,
			Action:      in.Action,
			Description: in.Description,
		})
	}
	if err := s.permissionRepo.BulkUpsert(perms); err != nil {
		return nil, err
	}
	return perms, nil
}

func (s *RoleService) ListPermissions(systemCode string) ([]entity.Permission, error) {
	return s.permissionRepo.List(systemCode)
}

func (s *RoleService) CreatePermission(systemCode, resource, action, description string) (*entity.Permission, error) {
	p := &entity.Permission{SystemCode: systemCode, Resource: resource, Action: action, Description: description}
	return p, s.permissionRepo.Create(p)
}

func (s *RoleService) DeletePermission(id uint) error {
	return s.permissionRepo.Delete(id)
}

// ── Roles ─────────────────────────────────────────────

func (s *RoleService) CreateRole(systemCode, name, description string) (*entity.Role, error) {
	r := &entity.Role{SystemCode: systemCode, Name: name, Description: description}
	return r, s.roleRepo.Create(r)
}

func (s *RoleService) ListRoles(systemCode string) ([]entity.Role, error) {
	return s.roleRepo.List(systemCode)
}

func (s *RoleService) GetRole(id uint) (*entity.Role, error) {
	return s.roleRepo.FindByID(id)
}

func (s *RoleService) UpdateRole(id uint, name, description string) error {
	return s.roleRepo.Update(id, name, description)
}

func (s *RoleService) DeleteRole(id uint) error {
	return s.roleRepo.Delete(id)
}

func (s *RoleService) SetDefaultRole(roleID uint, systemCode string) error {
	return s.roleRepo.SetDefault(roleID, systemCode)
}

func (s *RoleService) SetRolePermissions(roleID uint, permissionIDs []uint) error {
	if err := s.roleRepo.SetPermissions(roleID, permissionIDs); err != nil {
		return err
	}
	if role, err := s.roleRepo.FindByID(roleID); err == nil {
		s.invalidatePermCacheBySystem(role.SystemCode)
	}
	return nil
}

func (s *RoleService) AddRolePermissions(roleID uint, permissionIDs []uint) error {
	if err := s.roleRepo.AddPermissions(roleID, permissionIDs); err != nil {
		return err
	}
	if role, err := s.roleRepo.FindByID(roleID); err == nil {
		s.invalidatePermCacheBySystem(role.SystemCode)
	}
	return nil
}

func (s *RoleService) RemoveRolePermission(roleID, permissionID uint) error {
	if err := s.roleRepo.RemovePermission(roleID, permissionID); err != nil {
		return err
	}
	if role, err := s.roleRepo.FindByID(roleID); err == nil {
		s.invalidatePermCacheBySystem(role.SystemCode)
	}
	return nil
}

// ── User Role ─────────────────────────────────────────

func (s *RoleService) AssignRole(accountID, systemCode string, roleID uint) error {
	err := s.userRoleRepo.Set(accountID, systemCode, roleID)
	if err == nil {
		s.invalidatePermCache(accountID, systemCode)
	}
	return err
}

func (s *RoleService) GetUserRoles(accountID string) ([]entity.UserRole, error) {
	return s.userRoleRepo.GetRoles(accountID)
}

func (s *RoleService) ListUsersBySystem(systemCode string) ([]entity.UserRole, error) {
	return s.userRoleRepo.ListBySystem(systemCode)
}

func (s *RoleService) RemoveUserRole(accountID, systemCode string) error {
	err := s.userRoleRepo.Remove(accountID, systemCode)
	if err == nil {
		s.invalidatePermCache(accountID, systemCode)
	}
	return err
}

// ── Permission Cache ──────────────────────────────────

func (s *RoleService) permCacheKey(accountID, systemCode string) string {
	return fmt.Sprintf("perms:%s:%s", accountID, systemCode)
}

func (s *RoleService) invalidatePermCache(accountID, systemCode string) {
	if s.redis == nil {
		return
	}
	s.redis.Del(rctx, s.permCacheKey(accountID, systemCode))
}

// invalidatePermCacheBySystem removes all cached permission entries for every
// user in the given system (pattern: perms:*:<systemCode>).
// Called when a role's permission set changes — affects all users holding that role.
func (s *RoleService) invalidatePermCacheBySystem(systemCode string) {
	if s.redis == nil {
		return
	}
	pattern := fmt.Sprintf("perms:*:%s", systemCode)
	var cursor uint64
	for {
		keys, next, err := s.redis.Scan(rctx, cursor, pattern, 100).Result()
		if err != nil {
			break
		}
		if len(keys) > 0 {
			s.redis.Del(rctx, keys...)
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
}

// ── Permission Check ──────────────────────────────────

// GetUserRoleName returns the role name assigned to accountID in systemCode.
// Falls back to the default role name if the user has no explicit assignment.
// Returns "" if no role is found.
func (s *RoleService) GetUserRoleName(accountID, systemCode string) string {
	if ur, err := s.userRoleRepo.GetRoleForSystem(accountID, systemCode); err == nil {
		return ur.Role.Name
	}
	if def, err := s.roleRepo.FindDefault(systemCode); err == nil {
		return def.Name
	}
	return ""
}

// GetUserPermissionsStrict is like GetUserPermissions but returns an error when
// the user has no role and no default role is configured.
// Used at login time — callers should reject the request on error.
func (s *RoleService) GetUserPermissionsStrict(accountID, systemCode string) ([]string, error) {
	if s.userRoleRepo.IsRemoved(accountID, systemCode) {
		return nil, errors.New("role has been revoked")
	}

	_, err := s.userRoleRepo.GetRoleForSystem(accountID, systemCode)
	if err != nil {
		// never had a role — try default
		if systemCode != "*" {
			if _, defErr := s.roleRepo.FindDefault(systemCode); defErr != nil {
				return nil, errors.New("no role assigned and no default role configured for this system")
			}
		}
	}

	return s.GetUserPermissions(accountID, systemCode), nil
}

// GetUserPermissions returns ["resource:action", ...] for a user in a system.
// Results are cached in Redis for 5 minutes.
func (s *RoleService) GetUserPermissions(accountID, systemCode string) []string {
	if s.redis != nil {
		key := s.permCacheKey(accountID, systemCode)
		if cached, err := s.redis.Get(rctx, key).Result(); err == nil {
			var perms []string
			if json.Unmarshal([]byte(cached), &perms) == nil {
				return perms
			}
		}
	}

	ur, err := s.userRoleRepo.GetRoleForSystem(accountID, systemCode)
	if err != nil {
		// role was explicitly revoked — do NOT assign default, return empty
		if s.userRoleRepo.IsRemoved(accountID, systemCode) {
			return nil
		}
		// never had a role — auto-assign default if one exists
		if systemCode != "*" {
			if def, defErr := s.roleRepo.FindDefault(systemCode); defErr == nil {
				_ = s.userRoleRepo.Set(accountID, systemCode, def.ID)
				ur = &entity.UserRole{Role: *def}
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
	var result []string
	seen := map[string]bool{}
	for _, p := range ur.Role.Permissions {
		k := p.Resource + ":" + p.Action
		if !seen[k] {
			result = append(result, k)
			seen[k] = true
		}
	}

	if s.redis != nil {
		if data, err := json.Marshal(result); err == nil {
			s.redis.Set(rctx, s.permCacheKey(accountID, systemCode), data, 5*time.Minute)
		}
	}
	return result
}

// Check returns true if the user has the given resource:action permission.
// Uses cached GetUserPermissions; falls back to global role ("*").
func (s *RoleService) Check(accountID, systemCode, resource, action string) bool {
	if matchPerms(s.GetUserPermissions(accountID, systemCode), resource, action) {
		return true
	}
	return matchPerms(s.GetUserPermissions(accountID, "*"), resource, action)
}

func matchPerms(perms []string, resource, action string) bool {
	for _, p := range perms {
		parts := strings.SplitN(p, ":", 2)
		if len(parts) != 2 {
			continue
		}
		r, a := parts[0], parts[1]
		if (r == resource || r == "*") && (a == action || a == "*") {
			return true
		}
	}
	return false
}

// ── Validate ──────────────────────────────────────────

// ValidatePermissions checks which of the required "resource:action" strings
// are missing from the Role Service for the given systemCode.
// Used by consumer services at startup via POST /internal/validate.
func (s *RoleService) ValidatePermissions(systemCode string, required []string) ([]string, error) {
	perms, err := s.permissionRepo.List(systemCode)
	if err != nil {
		return nil, err
	}
	existing := make(map[string]bool, len(perms))
	for _, p := range perms {
		existing[p.Resource+":"+p.Action] = true
	}
	var missing []string
	for _, req := range required {
		if !existing[req] {
			missing = append(missing, req)
		}
	}
	return missing, nil
}

// ── Menu ──────────────────────────────────────────────

func (s *RoleService) ListMenus(systemCode string) ([]entity.Menu, error) {
	return s.menuRepo.List(systemCode)
}

func (s *RoleService) CreateMenu(m *entity.Menu) error {
	return s.menuRepo.Create(m)
}

func (s *RoleService) UpdateMenu(id uint, updates map[string]interface{}) error {
	delete(updates, "id")
	delete(updates, "system_code")
	return s.menuRepo.Update(id, updates)
}

func (s *RoleService) DeleteMenu(id uint) error {
	return s.menuRepo.Delete(id)
}

// GetUserMenus returns the menu tree filtered by the user's permissions.
func (s *RoleService) GetUserMenus(accountID, systemCode string) ([]entity.Menu, error) {
	menus, err := s.menuRepo.List(systemCode)
	if err != nil {
		return nil, err
	}

	permSet := map[string]bool{}
	for _, p := range s.GetUserPermissions(accountID, systemCode) {
		permSet[p] = true
	}
	for _, p := range s.GetUserPermissions(accountID, "*") {
		permSet[p] = true
	}
	hasWildcard := permSet["*:*"]

	// build set of resources the user has any permission on
	resourceSet := map[string]bool{}
	for p := range permSet {
		if parts := strings.SplitN(p, ":", 2); len(parts) == 2 {
			resourceSet[parts[0]] = true
		}
	}

	var allowed []entity.Menu
	for _, m := range menus {
		if !m.IsActive {
			continue
		}
		// menu is visible if role has ANY action on this resource (or wildcard *:*)
		if hasWildcard || resourceSet[m.Code] {
			allowed = append(allowed, m)
		}
	}
	return buildMenuTree(allowed), nil
}

func buildMenuTree(menus []entity.Menu) []entity.Menu {
	type node struct {
		menu     entity.Menu
		children []*node
	}

	nodes := make(map[uint]*node, len(menus))
	for _, m := range menus {
		m.Children = nil
		nodes[m.ID] = &node{menu: m}
	}

	var roots []*node
	for _, n := range nodes {
		if n.menu.ParentID == nil {
			roots = append(roots, n)
		} else if parent, ok := nodes[*n.menu.ParentID]; ok {
			parent.children = append(parent.children, n)
		}
	}

	// recursively convert node tree → []entity.Menu (works for any depth)
	var convert func(n *node) entity.Menu
	convert = func(n *node) entity.Menu {
		m := n.menu
		for _, child := range n.children {
			m.Children = append(m.Children, convert(child))
		}
		return m
	}

	result := make([]entity.Menu, 0, len(roots))
	for _, r := range roots {
		result = append(result, convert(r))
	}
	return result
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
