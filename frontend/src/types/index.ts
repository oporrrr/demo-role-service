export interface System {
  id: number
  code: string
  name: string
  description: string
}

export interface Permission {
  id: number
  systemCode: string
  resource: string
  action: string
  description: string
}

export interface Role {
  id: number
  systemCode: string
  name: string
  description: string
  isDefault: boolean
  permissions?: Permission[]
}

export interface UserRole {
  id: number
  accountId: string
  systemCode: string
  role: Role
  isActive: boolean
}

export interface Menu {
  id: number
  systemCode: string
  name: string
  code: string
  icon: string
  path: string
  parentId: number | null
  sortOrder: number
  isActive: boolean
  children?: Menu[]
}
