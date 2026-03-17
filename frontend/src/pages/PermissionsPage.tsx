import { useEffect, useState, useMemo } from 'react'
import { getMenus, updateMenu, getPermissions, createPermission, deletePermission } from '../api'
import { useSystem } from '../hooks/useSystem'
import SystemSelect from '../components/SystemSelect'
import Modal from '../components/Modal'
import type { Menu, Permission } from '../types'

const DEFAULT_ACTIONS = ['view', 'add', 'edit', 'delete', 'byId']

export default function PermissionsPage() {
  const { systems, selected, setSelected } = useSystem()
  const [menus, setMenus] = useState<Menu[]>([])
  const [perms, setPerms] = useState<Permission[]>([])
  const [loading, setLoading] = useState(false)

  // local draft: "resource:action" → checked bool
  const [draft, setDraft] = useState<Record<string, boolean>>({})
  const [saving, setSaving] = useState(false)

  // toggle menu is_active
  const [toggling, setToggling] = useState<number | null>(null)

  // custom action modal
  const [customModal, setCustomModal] = useState<Menu | null>(null)
  const [customAction, setCustomAction] = useState('')

  const load = (code: string) => {
    setLoading(true)
    Promise.all([getMenus(code), getPermissions(code)])
      .then(([m, p]) => {
        setMenus(m ?? [])
        setPerms(p ?? [])
        // init draft from existing permissions
        const d: Record<string, boolean> = {}
        for (const perm of (p ?? [])) d[`${perm.resource}:${perm.action}`] = true
        setDraft(d)
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => { if (selected) load(selected) }, [selected])

  const topLevel = menus.filter((m) => !m.parentId)
  const childMap = menus.reduce<Record<number, Menu[]>>((acc, m) => {
    if (m.parentId) { if (!acc[m.parentId]) acc[m.parentId] = []; acc[m.parentId].push(m) }
    return acc
  }, {})

  // existing permission lookup: "resource:action" → Permission
  const permMap = useMemo(() =>
    perms.reduce<Record<string, Permission>>((acc, p) => { acc[`${p.resource}:${p.action}`] = p; return acc }, {}),
    [perms]
  )

  // resource for a menu = menu.code
  const getResource = (m: Menu) => m.code?.trim() || ''

  // all extra actions used by this resource (beyond DEFAULT_ACTIONS)
  const extraActions = (resource: string) =>
    [...new Set(perms.filter((p) => p.resource === resource && !DEFAULT_ACTIONS.includes(p.action)).map((p) => p.action))]

  const toggle = (resource: string, action: string) => {
    const key = `${resource}:${action}`
    setDraft((prev) => ({ ...prev, [key]: !prev[key] }))
  }

  // diff: what needs to be created or deleted
  const toAdd = useMemo(() => Object.entries(draft).filter(([k, v]) => v && !permMap[k]).map(([k]) => k), [draft, permMap])
  const toRemove = useMemo(() => Object.entries(draft).filter(([k, v]) => !v && permMap[k]).map(([k]) => k), [draft, permMap])
  const hasChanges = toAdd.length > 0 || toRemove.length > 0

  const handleSave = async () => {
    setSaving(true)
    try {
      await Promise.all([
        ...toAdd.map((k) => {
          const [resource, action] = k.split(':')
          return createPermission({ systemCode: selected, resource, action })
        }),
        ...toRemove.map((k) => deletePermission(permMap[k].id)),
      ])
      load(selected)
    } finally {
      setSaving(false)
    }
  }

  const handleDiscard = () => {
    const d: Record<string, boolean> = {}
    for (const p of perms) d[`${p.resource}:${p.action}`] = true
    setDraft(d)
  }

  const toggleActive = async (m: Menu) => {
    if (toggling !== null) return
    setToggling(m.id)
    try {
      await updateMenu(m.id, { ...m, isActive: !m.isActive })
      load(selected)
    } finally {
      setToggling(null)
    }
  }

  const addCustomAction = async () => {
    const action = customAction.trim().toLowerCase()
    if (!action || !customModal) return
    const resource = getResource(customModal)
    if (!resource) return
    setDraft((prev) => ({ ...prev, [`${resource}:${action}`]: true }))
    setCustomModal(null)
    setCustomAction('')
  }

  const MenuPermRow = ({ m, indent = false }: { m: Menu; indent?: boolean }) => {
    const resource = getResource(m)
    const actions = [...DEFAULT_ACTIONS, ...extraActions(resource)]

    return (
      <tr className={`transition-colors ${m.isActive ? 'hover:bg-gray-50' : 'opacity-40'}`}>
        {/* Menu name */}
        <td className={`px-4 py-3 ${indent ? 'pl-10' : ''}`}>
          <div className="flex items-center gap-2">
            {indent && <span className="text-gray-300 text-xs">└</span>}
            <span>{m.icon || '📄'}</span>
            <div>
              <span className="font-medium text-gray-800 text-sm">{m.name}</span>
              {resource && (
                <span className="ml-2 font-mono text-xs text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">{resource}</span>
              )}
            </div>
          </div>
        </td>

        {/* Action checkboxes */}
        <td className="px-4 py-3">
          {resource ? (
            <div className="flex items-center flex-wrap gap-2">
              {actions.map((action) => {
                const key = `${resource}:${action}`
                const checked = !!draft[key]
                const wasChecked = !!permMap[key]
                const changed = checked !== wasChecked
                return (
                  <label key={action} className="flex items-center gap-1 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={() => toggle(resource, action)}
                      className="accent-indigo-600"
                    />
                    <span className={`text-xs font-mono px-1.5 py-0.5 rounded ${
                      checked
                        ? changed ? 'bg-emerald-100 text-emerald-700' : 'bg-indigo-50 text-indigo-700'
                        : changed ? 'bg-red-50 text-red-400 line-through' : 'text-gray-400'
                    }`}>
                      {action}
                    </span>
                  </label>
                )
              })}
              <button
                onClick={() => { setCustomModal(m); setCustomAction('') }}
                className="text-xs text-gray-400 hover:text-indigo-500 border border-dashed border-gray-300 hover:border-indigo-300 rounded px-2 py-0.5 transition-colors"
              >
                + custom
              </button>
            </div>
          ) : (
            <span className="text-xs text-gray-300">— public (ไม่มี resource)</span>
          )}
        </td>

        {/* is_active toggle */}
        <td className="px-4 py-3 text-center w-20">
          <button
            onClick={() => toggleActive(m)}
            disabled={toggling !== null}
            className={`relative inline-flex h-5 w-9 rounded-full transition-colors focus:outline-none disabled:opacity-50 ${
              m.isActive ? 'bg-indigo-500' : 'bg-gray-200'
            }`}
          >
            <span className={`inline-block h-4 w-4 mt-0.5 rounded-full bg-white shadow transition-transform ${
              m.isActive ? 'translate-x-4' : 'translate-x-0.5'
            }`} />
          </button>
        </td>
      </tr>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Permissions</h1>
          <p className="text-sm text-gray-500 mt-0.5">กำหนด actions ต่อ menu — resource = menu code</p>
        </div>
        <SystemSelect systems={systems} value={selected} onChange={setSelected} />
      </div>

      {loading ? (
        <p className="text-gray-400 text-sm">Loading...</p>
      ) : !selected ? (
        <div className="text-center py-20 text-gray-400">
          <p className="text-4xl mb-3">🔑</p>
          <p className="text-sm">เลือก System ก่อน</p>
        </div>
      ) : menus.length === 0 ? (
        <div className="text-center py-20 text-gray-400">
          <p className="text-4xl mb-3">🔑</p>
          <p className="text-sm">ยังไม่มี menu — สร้าง menu แล้วใส่ permission field เพื่อกำหนด resource</p>
        </div>
      ) : (
        <div className="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-gray-50 border-b border-gray-200">
                  <th className="text-left px-4 py-3 font-semibold text-gray-600 w-56">Menu</th>
                  <th className="text-left px-4 py-3 font-semibold text-gray-600">Actions</th>
                  <th className="text-center px-4 py-3 font-semibold text-gray-600 w-20">Active</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {topLevel.map((parent) => (
                  <>
                    <MenuPermRow key={parent.id} m={parent} />
                    {(childMap[parent.id] ?? []).map((child) => (
                      <MenuPermRow key={child.id} m={child} indent />
                    ))}
                  </>
                ))}
              </tbody>
            </table>
          </div>

          {/* Footer save bar */}
          <div className="px-5 py-3 border-t border-gray-100 bg-gray-50 flex items-center justify-between">
            <div className="flex items-center gap-4 text-xs text-gray-500">
              <span className="flex items-center gap-1.5">
                <span className="w-3 h-3 rounded bg-indigo-500 inline-block" /> มีสิทธิ์
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-3 h-3 rounded bg-emerald-400 inline-block" /> จะเพิ่ม
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-3 h-3 rounded border border-red-300 bg-red-50 inline-block" /> จะลบ
              </span>
            </div>

            {hasChanges ? (
              <div className="flex items-center gap-2">
                <span className="text-xs text-gray-500">
                  {toAdd.length > 0 && <span className="text-emerald-600">+{toAdd.length} </span>}
                  {toRemove.length > 0 && <span className="text-red-500">-{toRemove.length} </span>}
                  รายการ
                </span>
                <button
                  onClick={handleDiscard}
                  className="text-xs text-gray-500 border border-gray-300 rounded-lg px-3 py-1.5 hover:bg-white transition-colors"
                >
                  ยกเลิก
                </button>
                <button
                  onClick={handleSave}
                  disabled={saving}
                  className="text-xs bg-indigo-600 text-white rounded-lg px-4 py-1.5 font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
                >
                  {saving ? 'กำลังบันทึก...' : 'บันทึก'}
                </button>
              </div>
            ) : (
              <span className="text-xs text-gray-400">resource = menu code — tick actions แล้วกดบันทึก</span>
            )}
          </div>
        </div>
      )}

      {/* Custom action modal */}
      {customModal && (
        <Modal title={`+ Custom Action — ${customModal.name}`} onClose={() => setCustomModal(null)}>
          <div className="space-y-3">
            <p className="text-xs text-gray-500">
              resource: <span className="font-mono bg-gray-100 px-1.5 py-0.5 rounded">{getResource(customModal)}</span>
            </p>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Action name</label>
              <input
                autoFocus
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="เช่น approve, export, byId"
                value={customAction}
                onChange={(e) => setCustomAction(e.target.value.toLowerCase())}
                onKeyDown={(e) => e.key === 'Enter' && addCustomAction()}
              />
            </div>
            <div className="flex gap-2 pt-1">
              <button
                onClick={addCustomAction}
                disabled={!customAction.trim()}
                className="flex-1 bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
              >
                เพิ่ม
              </button>
              <button
                onClick={() => setCustomModal(null)}
                className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm hover:bg-gray-50 transition-colors"
              >
                ยกเลิก
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
