import client from './client'
import type { Menu, Permission, Role } from '../types'

// ── Auth ────────────────────────────────────────────
export const login = (username: string, password: string) =>
  client.post('/auth/login', { username, password }).then((r) => r.data.data.token as string)

// ── Systems ─────────────────────────────────────────
export const getSystems = () => client.get('/systems').then((r) => r.data.data)

export const registerSystem = (data: {
  code: string
  name: string
  description: string
  authClientId?: string
  authClientSecret?: string
}) => client.post('/systems/register', data).then((r) => r.data.data)

export const updateSystemCredentials = (code: string, authClientId: string, authClientSecret: string) =>
  client.put(`/systems/${code}/credentials`, { authClientId, authClientSecret }).then((r) => r.data)

export const reKeySystem = (code: string) =>
  client.post(`/systems/${code}/rekey`).then((r) => r.data.data)

export const bootstrapSystem = (code: string, accountId: string) =>
  client.post(`/systems/${code}/bootstrap`, { accountId }).then((r) => r.data)

// ── Permissions ─────────────────────────────────────
export const getPermissions = (system: string) =>
  client.get('/permissions', { params: { system } }).then((r) => r.data.data)

export const createPermission = (data: Partial<Permission>) =>
  client.post('/permissions', data).then((r) => r.data.data)

export const deletePermission = (id: number) => client.delete(`/permissions/${id}`)

// ── Roles ───────────────────────────────────────────
export const getRoles = (system: string) =>
  client.get('/roles', { params: { system } }).then((r) => r.data.data)

export const getRole = (id: number) => client.get(`/roles/${id}`).then((r) => r.data.data)

export const createRole = (data: Partial<Role>) =>
  client.post('/roles', data).then((r) => r.data.data)

export const updateRole = (id: number, data: Partial<Role>) =>
  client.put(`/roles/${id}`, data).then((r) => r.data)

export const deleteRole = (id: number) => client.delete(`/roles/${id}`)

export const setRolePermissions = (roleId: number, permissionIds: number[]) =>
  client.put(`/roles/${roleId}/permissions`, { permissionIds })

export const setDefaultRole = (roleId: number) =>
  client.put(`/roles/${roleId}/default`)

// ── Users ───────────────────────────────────────────
export const getUsersBySystem = (system: string) =>
  client.get('/users', { params: { system } }).then((r) => r.data.data)

export const getUserRoles = (accountId: string) =>
  client.get(`/users/${accountId}/roles`).then((r) => r.data.data)

export const assignRole = (accountId: string, systemCode: string, roleId: number) =>
  client.put(`/users/${accountId}/role`, { systemCode, roleId })

export const removeUserRole = (accountId: string, system: string) =>
  client.delete(`/users/${accountId}/role`, { params: { system } })

// ── Menus ───────────────────────────────────────────
export const getMenus = (system: string) =>
  client.get('/menus', { params: { system } }).then((r) => r.data.data)

export const createMenu = (data: Partial<Menu>) =>
  client.post('/menus', data).then((r) => r.data.data)

export const updateMenu = (id: number, data: Partial<Menu>) =>
  client.put(`/menus/${id}`, data).then((r) => r.data)

export const deleteMenu = (id: number) => client.delete(`/menus/${id}`)
