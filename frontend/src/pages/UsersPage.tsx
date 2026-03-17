import { useEffect, useState } from 'react'
import { getUsersBySystem, assignRole, removeUserRole, getRoles } from '../api'
import { useSystem } from '../hooks/useSystem'
import SystemSelect from '../components/SystemSelect'
import Modal from '../components/Modal'
import type { UserRole, Role } from '../types'

export default function UsersPage() {
  const { systems, selected, setSelected } = useSystem()
  const [users, setUsers] = useState<UserRole[]>([])
  const [loading, setLoading] = useState(false)
  const [roles, setRoles] = useState<Role[]>([])

  // assign modal (new user)
  const [showAssign, setShowAssign] = useState(false)
  const [assignAccountId, setAssignAccountId] = useState('')
  const [selectedRole, setSelectedRole] = useState<number>(0)
  const [assigning, setAssigning] = useState(false)

  // edit modal (existing user)
  const [editUser, setEditUser] = useState<UserRole | null>(null)
  const [editRoleId, setEditRoleId] = useState<number>(0)
  const [editing, setEditing] = useState(false)

  const load = (system: string) => {
    if (!system) return
    setLoading(true)
    getUsersBySystem(system)
      .then((d) => setUsers(d ?? []))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load(selected) }, [selected])

  const loadRoles = async () => {
    const data = await getRoles(selected)
    setRoles(data ?? [])
    return data ?? []
  }

  const openAssign = async () => {
    const data = await loadRoles()
    setSelectedRole(data?.[0]?.id ?? 0)
    setAssignAccountId('')
    setShowAssign(true)
  }

  const handleAssign = async () => {
    if (!selectedRole || !assignAccountId.trim()) return
    setAssigning(true)
    try {
      await assignRole(assignAccountId.trim(), selected, selectedRole)
      setShowAssign(false)
      load(selected)
    } finally {
      setAssigning(false)
    }
  }

  const openEdit = async (ur: UserRole) => {
    const data = await loadRoles()
    setEditRoleId(ur.role?.id ?? data?.[0]?.id ?? 0)
    setEditUser(ur)
  }

  const handleEdit = async () => {
    if (!editUser || !editRoleId) return
    setEditing(true)
    try {
      await assignRole(editUser.accountId, selected, editRoleId)
      setEditUser(null)
      load(selected)
    } finally {
      setEditing(false)
    }
  }

  const handleRemove = async (accountId: string) => {
    if (!confirm(`ถอน role ของ ${accountId} ใน ${selected}?`)) return
    await removeUserRole(accountId, selected)
    load(selected)
  }

  const handleRestore = async (ur: UserRole) => {
    await assignRole(ur.accountId, selected, ur.role?.id)
    load(selected)
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Users</h1>
          <p className="text-sm text-gray-500 mt-0.5">จัดการ role ของ user ในแต่ละระบบ</p>
        </div>
        <div className="flex items-center gap-3">
          <SystemSelect systems={systems} value={selected} onChange={setSelected} />
          <button
            onClick={openAssign}
            disabled={!selected}
            className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
          >
            + Assign Role
          </button>
        </div>
      </div>

      {loading ? (
        <p className="text-gray-400 text-sm">Loading...</p>
      ) : !selected ? (
        <div className="text-center py-20 text-gray-400">
          <p className="text-4xl mb-3">👤</p>
          <p className="text-sm">เลือก System ก่อน</p>
        </div>
      ) : users.length === 0 ? (
        <div className="text-center py-20 text-gray-400">
          <p className="text-4xl mb-3">👤</p>
          <p className="text-sm">ยังไม่มี user ที่มี role ใน {selected}</p>
        </div>
      ) : (
        <div className="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="text-left px-5 py-3 font-semibold text-gray-600">#</th>
                <th className="text-left px-5 py-3 font-semibold text-gray-600">Account ID</th>
                <th className="text-left px-5 py-3 font-semibold text-gray-600">Role</th>
                <th className="px-5 py-3"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {users.map((ur, i) => (
                <tr key={ur.id} className={`transition-colors ${ur.isActive ? 'hover:bg-gray-50' : 'bg-gray-50 opacity-50'}`}>
                  <td className="px-5 py-3 text-gray-400 text-xs">{i + 1}</td>
                  <td className="px-5 py-3 font-mono text-gray-800">
                    <span className={ur.isActive ? '' : 'line-through text-gray-400'}>{ur.accountId}</span>
                  </td>
                  <td className="px-5 py-3">
                    {ur.isActive ? (
                      <span className="bg-indigo-50 text-indigo-700 text-xs font-semibold px-2.5 py-1 rounded-full">
                        {ur.role?.name ?? '-'}
                      </span>
                    ) : (
                      <span className="bg-gray-100 text-gray-400 text-xs px-2.5 py-1 rounded-full">ถูกถอน</span>
                    )}
                  </td>
                  <td className="px-5 py-3 text-right">
                    {ur.isActive ? (
                      <div className="flex items-center justify-end gap-3">
                        <button
                          onClick={() => openEdit(ur)}
                          className="text-xs text-gray-500 hover:text-indigo-600 transition-colors"
                        >
                          แก้ไข
                        </button>
                        <button
                          onClick={() => handleRemove(ur.accountId)}
                          className="text-xs text-red-400 hover:text-red-600 transition-colors"
                        >
                          ถอน role
                        </button>
                      </div>
                    ) : (
                      <button
                        onClick={() => handleRestore(ur)}
                        className="text-xs text-emerald-500 hover:text-emerald-700 transition-colors"
                      >
                        กู้คืน
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="px-5 py-3 border-t border-gray-100 bg-gray-50 text-xs text-gray-400">
            {users.length} user ใน {selected}
          </div>
        </div>
      )}

      {/* Assign Role Modal (new user) */}
      {showAssign && (
        <Modal title={`Assign Role — ${selected}`} onClose={() => setShowAssign(false)}>
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Account ID <span className="text-red-400">*</span>
              </label>
              <input
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="เช่น user-uuid-123"
                value={assignAccountId}
                onChange={(e) => setAssignAccountId(e.target.value)}
                autoFocus
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Role</label>
              <select
                value={selectedRole}
                onChange={(e) => setSelectedRole(Number(e.target.value))}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                {roles.map((r) => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
            </div>
            <div className="flex gap-2 pt-2">
              <button
                onClick={handleAssign}
                disabled={!assignAccountId.trim() || !selectedRole || assigning}
                className="flex-1 bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
              >
                {assigning ? 'กำลัง Assign...' : 'Assign'}
              </button>
              <button
                onClick={() => setShowAssign(false)}
                className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm hover:bg-gray-50 transition-colors"
              >
                ยกเลิก
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* Edit Role Modal (existing user) */}
      {editUser && (
        <Modal title={`แก้ไข Role — ${editUser.accountId}`} onClose={() => setEditUser(null)}>
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Account ID</label>
              <p className="font-mono text-sm text-gray-600 bg-gray-50 border border-gray-200 rounded-lg px-3 py-2">
                {editUser.accountId}
              </p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Role</label>
              <select
                value={editRoleId}
                onChange={(e) => setEditRoleId(Number(e.target.value))}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                {roles.map((r) => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
            </div>
            <div className="flex gap-2 pt-2">
              <button
                onClick={handleEdit}
                disabled={!editRoleId || editing}
                className="flex-1 bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
              >
                {editing ? 'กำลังบันทึก...' : 'บันทึก'}
              </button>
              <button
                onClick={() => setEditUser(null)}
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
