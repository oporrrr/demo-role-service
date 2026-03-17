import { useEffect, useState } from 'react'
import { getRoles, getRole, getPermissions, getMenus, createRole, updateRole, deleteRole, setRolePermissions, setDefaultRole } from '../api'
import { useSystem } from '../hooks/useSystem'
import SystemSelect from '../components/SystemSelect'
import Modal from '../components/Modal'
import type { Role, Permission, Menu } from '../types'

export default function RolesPage() {
  const { systems, selected, setSelected } = useSystem()
  const [roles, setRoles] = useState<Role[]>([])
  const [loading, setLoading] = useState(false)

  // create / edit
  const [showForm, setShowForm] = useState(false)
  const [editRole, setEditRole] = useState<Role | null>(null)
  const [form, setForm] = useState({ name: '', description: '' })

  // permission assignment modal
  const [permModal, setPermModal] = useState<Role | null>(null)
  const [allPerms, setAllPerms] = useState<Permission[]>([])
  const [allMenus, setAllMenus] = useState<Menu[]>([])
  const [checkedIds, setCheckedIds] = useState<Set<number>>(new Set())

  const load = (code: string) => {
    setLoading(true)
    getRoles(code)
      .then((d) => setRoles(d ?? []))
      .finally(() => setLoading(false))
  }

  useEffect(() => { if (selected) load(selected) }, [selected])

  const openCreate = () => {
    setEditRole(null)
    setForm({ name: '', description: '' })
    setShowForm(true)
  }

  const openEdit = (r: Role) => {
    setEditRole(r)
    setForm({ name: r.name, description: r.description })
    setShowForm(true)
  }

  const handleSave = async () => {
    if (editRole) {
      await updateRole(editRole.id, form)
    } else {
      await createRole({ ...form, systemCode: selected })
    }
    setShowForm(false)
    load(selected)
  }

  const handleDelete = async (id: number) => {
    if (!confirm('ลบ role นี้?')) return
    await deleteRole(id)
    load(selected)
  }

  const handleSetDefault = async (id: number) => {
    await setDefaultRole(id)
    load(selected)
  }

  const openPermModal = async (r: Role) => {
    const [full, perms, menus] = await Promise.all([getRole(r.id), getPermissions(selected), getMenus(selected)])
    setAllPerms(perms ?? [])
    setAllMenus(menus ?? [])
    setCheckedIds(new Set((full.permissions ?? []).map((p: Permission) => p.id)))
    setPermModal(r)
  }

  const togglePerm = (id: number) => {
    setCheckedIds((prev) => {
      const s = new Set(prev)
      s.has(id) ? s.delete(id) : s.add(id)
      return s
    })
  }

  const savePerms = async () => {
    if (!permModal) return
    await setRolePermissions(permModal.id, [...checkedIds])
    setPermModal(null)
    load(selected)
  }

  // group perms by resource
  const grouped = allPerms.reduce<Record<string, Permission[]>>((acc, p) => {
    if (!acc[p.resource]) acc[p.resource] = []
    acc[p.resource].push(p)
    return acc
  }, {})

  // build hierarchical menu list for rendering order
  const menuByCode = allMenus.reduce<Record<string, Menu>>((acc, m) => { acc[m.code] = m; return acc }, {})
  const topMenus = allMenus.filter((m) => !m.parentId)
  const childMenus = allMenus.filter((m) => m.parentId)
  // ordered list: [parent, ...children, ...] for resources that have perms
  const orderedGroups: { resource: string; indent: boolean }[] = []
  const seen = new Set<string>()
  for (const top of topMenus) {
    if (grouped[top.code]) { orderedGroups.push({ resource: top.code, indent: false }); seen.add(top.code) }
    for (const child of childMenus.filter((c) => c.parentId === top.id)) {
      if (grouped[child.code]) { orderedGroups.push({ resource: child.code, indent: true }); seen.add(child.code) }
    }
  }
  // append any resources not matched to a menu (e.g. manually created)
  for (const resource of Object.keys(grouped)) {
    if (!seen.has(resource)) orderedGroups.push({ resource, indent: false })
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Roles</h1>
          <p className="text-sm text-gray-500 mt-0.5">จัดการ roles และ permissions ของแต่ละระบบ</p>
        </div>
        <div className="flex items-center gap-3">
          <SystemSelect systems={systems} value={selected} onChange={setSelected} />
          <button
            onClick={openCreate}
            className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors"
          >
            + New Role
          </button>
        </div>
      </div>

      {loading ? (
        <p className="text-gray-400 text-sm">Loading...</p>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {roles.map((r) => (
            <div key={r.id} className={`bg-white border rounded-xl p-5 shadow-sm ${r.isDefault ? 'border-emerald-300 ring-1 ring-emerald-200' : 'border-gray-200'}`}>
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="bg-indigo-100 text-indigo-700 text-xs font-bold px-2 py-1 rounded-md uppercase">
                    {r.name}
                  </span>
                  {r.isDefault && (
                    <span className="bg-emerald-100 text-emerald-700 text-xs font-semibold px-2 py-0.5 rounded-full">
                      ✦ default
                    </span>
                  )}
                  {r.description && <p className="text-sm text-gray-500 w-full mt-1">{r.description}</p>}
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <button
                    onClick={() => openEdit(r)}
                    className="text-xs text-gray-500 hover:text-indigo-600 border border-gray-200 rounded px-2 py-1 transition-colors"
                  >
                    แก้ไข
                  </button>
                  <button
                    onClick={() => handleDelete(r.id)}
                    className="text-xs text-red-400 hover:text-red-600 border border-red-100 rounded px-2 py-1 transition-colors"
                  >
                    ลบ
                  </button>
                </div>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => openPermModal(r)}
                  className="flex-1 text-center text-xs text-indigo-600 border border-indigo-200 rounded-lg py-2 hover:bg-indigo-50 transition-colors"
                >
                  🔑 จัดการ Permissions
                </button>
                {!r.isDefault && (
                  <button
                    onClick={() => handleSetDefault(r.id)}
                    className="text-xs text-gray-400 hover:text-emerald-600 border border-gray-200 hover:border-emerald-300 rounded-lg px-3 py-2 transition-colors"
                    title="ตั้งเป็น default role สำหรับ user ใหม่"
                  >
                    ✦ Set default
                  </button>
                )}
              </div>
            </div>
          ))}
          {roles.length === 0 && !loading && (
            <p className="text-gray-400 text-sm col-span-2">ยังไม่มี roles สำหรับ {selected}</p>
          )}
        </div>
      )}

      {/* Create / Edit Role Modal */}
      {showForm && (
        <Modal title={editRole ? 'แก้ไข Role' : 'สร้าง Role ใหม่'} onClose={() => setShowForm(false)}>
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Role Name</label>
              <input
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 uppercase"
                placeholder="เช่น ADMIN, MANAGER, VIEWER"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value.toUpperCase() })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <input
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="อธิบาย role นี้"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
              />
            </div>
            <div className="flex gap-2 pt-2">
              <button
                onClick={handleSave}
                disabled={!form.name}
                className="flex-1 bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
              >
                บันทึก
              </button>
              <button
                onClick={() => setShowForm(false)}
                className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm hover:bg-gray-50 transition-colors"
              >
                ยกเลิก
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* Permission Assignment Modal */}
      {permModal && (
        <Modal title={`Permissions — ${permModal.name}`} onClose={() => setPermModal(null)}>
          <div className="max-h-96 overflow-y-auto space-y-1.5 pr-1">
            {orderedGroups.map(({ resource, indent }) => {
              const items = grouped[resource]
              const menu = menuByCode[resource]
              return (
                <div key={resource} className={`border border-gray-100 rounded-lg overflow-hidden ${indent ? 'ml-5' : ''}`}>
                  <div className="px-3 py-2 bg-gray-50 flex items-center gap-2">
                    {indent && <span className="text-gray-300 text-xs">└</span>}
                    {menu?.icon && <span className="text-sm">{menu.icon}</span>}
                    <span className="text-xs font-medium text-gray-700">{menu?.name ?? resource}</span>
                    <span className="text-xs font-mono text-indigo-400 bg-indigo-50 px-1.5 py-0.5 rounded ml-auto">{resource}</span>
                  </div>
                  <div className="divide-y divide-gray-50">
                    {items.map((p) => (
                      <label key={p.id} className="flex items-center gap-3 px-3 py-2 hover:bg-gray-50 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={checkedIds.has(p.id)}
                          onChange={() => togglePerm(p.id)}
                          className="accent-indigo-600"
                        />
                        <span className="bg-green-100 text-green-700 text-xs font-mono px-1.5 py-0.5 rounded">
                          {p.action}
                        </span>
                        {p.description && <span className="text-xs text-gray-400">{p.description}</span>}
                      </label>
                    ))}
                  </div>
                </div>
              )
            })}
            {allPerms.length === 0 && (
              <p className="text-xs text-gray-400 py-4 text-center">
                ไม่มี permissions สำหรับระบบนี้ — เพิ่มที่หน้า Permissions ก่อน
              </p>
            )}
          </div>
          <div className="flex gap-2 pt-4 border-t border-gray-100 mt-4">
            <button
              onClick={savePerms}
              className="flex-1 bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 transition-colors"
            >
              บันทึก
            </button>
            <button
              onClick={() => setPermModal(null)}
              className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm hover:bg-gray-50 transition-colors"
            >
              ยกเลิก
            </button>
          </div>
        </Modal>
      )}
    </div>
  )
}
