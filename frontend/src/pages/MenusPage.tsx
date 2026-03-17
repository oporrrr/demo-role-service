import { useEffect, useState } from 'react'
import { getMenus, createMenu, updateMenu, deleteMenu } from '../api'
import { useSystem } from '../hooks/useSystem'
import SystemSelect from '../components/SystemSelect'
import Modal from '../components/Modal'
import type { Menu } from '../types'

const emptyForm = (): Partial<Menu> => ({
  name: '', code: '', icon: '', path: '',
  parentId: null, sortOrder: 0, isActive: true,
})

export default function MenusPage() {
  const { systems, selected, setSelected } = useSystem()
  const [menus, setMenus] = useState<Menu[]>([])
  const [loading, setLoading] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [editMenu, setEditMenu] = useState<Menu | null>(null)
  const [form, setForm] = useState<Partial<Menu>>(emptyForm())

  const load = (code: string) => {
    setLoading(true)
    getMenus(code)
      .then((m) => setMenus(m ?? []))
      .finally(() => setLoading(false))
  }

  useEffect(() => { if (selected) load(selected) }, [selected])

  const openCreate = () => {
    setEditMenu(null)
    setForm({ ...emptyForm(), systemCode: selected })
    setShowForm(true)
  }

  const openEdit = (m: Menu) => {
    setEditMenu(m)
    setForm({ ...m })
    setShowForm(true)
  }

  const handleSave = async () => {
    if (editMenu) {
      await updateMenu(editMenu.id, form)
    } else {
      await createMenu({ ...form, systemCode: selected })
    }
    setShowForm(false)
    load(selected)
  }

  const handleDelete = async (id: number) => {
    if (!confirm('ลบ menu นี้?')) return
    await deleteMenu(id)
    load(selected)
  }

  const toggleActive = async (m: Menu) => {
    await updateMenu(m.id, { isActive: !m.isActive })
    load(selected)
  }

  const topLevel = menus.filter((m) => !m.parentId)
  const children = menus.filter((m) => m.parentId)

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Menus</h1>
          <p className="text-sm text-gray-500 mt-0.5">จัดการ menu — กำหนด actions ได้ที่หน้า Permissions</p>
        </div>
        <div className="flex items-center gap-3">
          <SystemSelect systems={systems} value={selected} onChange={setSelected} />
          <button
            onClick={openCreate}
            className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors"
          >
            + New Menu
          </button>
        </div>
      </div>

      {loading ? (
        <p className="text-gray-400 text-sm">Loading...</p>
      ) : (
        <div className="space-y-3">
          {topLevel.map((m) => (
            <div key={m.id} className="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
              <MenuRow m={m} onEdit={openEdit} onDelete={handleDelete} onToggle={toggleActive} />
              {/* Sub-menus */}
              {children.filter((c) => c.parentId === m.id).map((c) => (
                <div key={c.id} className="border-t border-gray-100 bg-gray-50">
                  <MenuRow m={c} onEdit={openEdit} onDelete={handleDelete} onToggle={toggleActive} indent />
                </div>
              ))}
            </div>
          ))}
          {menus.length === 0 && !loading && (
            <p className="text-gray-400 text-sm">ยังไม่มี menus สำหรับ {selected}</p>
          )}
        </div>
      )}

      {showForm && (
        <Modal title={editMenu ? 'แก้ไข Menu' : 'สร้าง Menu ใหม่'} onClose={() => setShowForm(false)}>
          <div className="space-y-3 max-h-[70vh] overflow-y-auto pr-1">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">ชื่อ Menu</label>
                <input
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="เช่น Dashboard"
                  value={form.name ?? ''}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Code</label>
                <input
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="เช่น menu_dashboard"
                  value={form.code ?? ''}
                  onChange={(e) => setForm({ ...form, code: e.target.value })}
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Icon</label>
                <input
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="เช่น 📊 หรือ ชื่อ icon"
                  value={form.icon ?? ''}
                  onChange={(e) => setForm({ ...form, icon: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Path</label>
                <input
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="เช่น /dashboard"
                  value={form.path ?? ''}
                  onChange={(e) => setForm({ ...form, path: e.target.value })}
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Parent Menu</label>
                <select
                  value={form.parentId ?? ''}
                  onChange={(e) => setForm({ ...form, parentId: e.target.value ? Number(e.target.value) : null })}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                >
                  <option value="">— ไม่มี (top level) —</option>
                  {menus.filter((m) => !m.parentId && m.id !== editMenu?.id).map((m) => (
                    <option key={m.id} value={m.id}>{m.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Sort Order</label>
                <input
                  type="number"
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={form.sortOrder ?? 0}
                  onChange={(e) => setForm({ ...form, sortOrder: Number(e.target.value) })}
                />
              </div>
            </div>
          </div>
          <div className="flex gap-2 pt-4 border-t border-gray-100 mt-2">
            <button
              onClick={handleSave}
              disabled={!form.name || !form.code}
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
        </Modal>
      )}
    </div>
  )
}

function MenuRow({
  m, onEdit, onDelete, onToggle, indent = false,
}: {
  m: Menu
  onEdit: (m: Menu) => void
  onDelete: (id: number) => void
  onToggle: (m: Menu) => void
  indent?: boolean
}) {
  return (
    <div className={`flex items-center justify-between px-5 py-3 ${indent ? 'pl-10' : ''}`}>
      <div className="flex items-center gap-3">
        {indent && <span className="text-gray-300 text-sm">└</span>}
        <span className="text-lg">{m.icon || '📄'}</span>
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium text-gray-800 text-sm">{m.name}</span>
            {!m.isActive && (
              <span className="text-xs bg-gray-100 text-gray-400 px-1.5 py-0.5 rounded">inactive</span>
            )}
          </div>
          <div className="flex items-center gap-2 mt-0.5">
            <span className="text-xs font-mono text-gray-400">{m.path || '-'}</span>
            <span className="text-xs font-mono text-indigo-400 bg-indigo-50 px-1.5 py-0.5 rounded">{m.code}</span>
          </div>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <button
          onClick={() => onToggle(m)}
          className={`text-xs px-2 py-1 rounded border transition-colors ${
            m.isActive
              ? 'text-green-600 border-green-200 hover:bg-green-50'
              : 'text-gray-400 border-gray-200 hover:bg-gray-50'
          }`}
        >
          {m.isActive ? 'ON' : 'OFF'}
        </button>
        <button
          onClick={() => onEdit(m)}
          className="text-xs text-gray-500 hover:text-indigo-600 border border-gray-200 rounded px-2 py-1 transition-colors"
        >
          แก้ไข
        </button>
        <button
          onClick={() => onDelete(m.id)}
          className="text-xs text-red-400 hover:text-red-600 border border-red-100 rounded px-2 py-1 transition-colors"
        >
          ลบ
        </button>
      </div>
    </div>
  )
}
