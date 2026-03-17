import { useEffect, useState } from 'react'
import { getSystems, registerSystem, reKeySystem, bootstrapSystem, updateSystemCredentials } from '../api'
import Modal from '../components/Modal'
import type { System } from '../types'

export default function SystemsPage() {
  const [systems, setSystems] = useState<System[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState({ code: '', name: '', description: '', authClientId: '', authClientSecret: '' })
  const [apiKey, setApiKey] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  // credentials state
  const [credTarget, setCredTarget] = useState<System | null>(null)
  const [credForm, setCredForm] = useState({ authClientId: '', authClientSecret: '' })
  const [credSaving, setCredSaving] = useState(false)
  const [credError, setCredError] = useState('')
  const [credDone, setCredDone] = useState(false)

  // re-key state
  const [reKeyTarget, setReKeyTarget] = useState<System | null>(null)
  const [reKeyResult, setReKeyResult] = useState('')
  const [reKeying, setReKeying] = useState(false)

  // bootstrap state
  const [bootstrapTarget, setBootstrapTarget] = useState<System | null>(null)
  const [bootstrapAccountId, setBootstrapAccountId] = useState('')
  const [bootstrapping, setBootstrapping] = useState(false)
  const [bootstrapDone, setBootstrapDone] = useState(false)
  const [bootstrapError, setBootstrapError] = useState('')

  const load = () => {
    setLoading(true)
    getSystems()
      .then((d) => setSystems(d ?? []))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const handleRegister = async () => {
    if (!form.code || !form.name) return
    setSaving(true)
    setError('')
    try {
      const data = await registerSystem(form)
      setApiKey(data.apiKey)
      load()
    } catch (e: any) {
      setError(e.response?.data?.message ?? 'เกิดข้อผิดพลาด')
    } finally {
      setSaving(false)
    }
  }

  const closeModal = () => {
    setShowModal(false)
    setForm({ code: '', name: '', description: '', authClientId: '', authClientSecret: '' })
    setApiKey('')
    setError('')
  }

  const handleUpdateCredentials = async () => {
    if (!credTarget) return
    setCredSaving(true)
    setCredError('')
    try {
      await updateSystemCredentials(credTarget.code, credForm.authClientId, credForm.authClientSecret)
      setCredDone(true)
    } catch (e: any) {
      setCredError(e.response?.data?.message ?? 'เกิดข้อผิดพลาด')
    } finally {
      setCredSaving(false)
    }
  }

  const closeCredentials = () => {
    setCredTarget(null)
    setCredForm({ authClientId: '', authClientSecret: '' })
    setCredSaving(false)
    setCredError('')
    setCredDone(false)
  }

  const handleReKey = async () => {
    if (!reKeyTarget) return
    setReKeying(true)
    try {
      const data = await reKeySystem(reKeyTarget.code)
      setReKeyResult(data.apiKey)
    } finally {
      setReKeying(false)
    }
  }

  const closeReKey = () => {
    setReKeyTarget(null)
    setReKeyResult('')
  }

  const handleBootstrap = async () => {
    if (!bootstrapTarget || !bootstrapAccountId.trim()) return
    setBootstrapping(true)
    setBootstrapError('')
    try {
      await bootstrapSystem(bootstrapTarget.code, bootstrapAccountId.trim())
      setBootstrapDone(true)
    } catch (e: any) {
      setBootstrapError(e.response?.data?.message ?? 'เกิดข้อผิดพลาด')
    } finally {
      setBootstrapping(false)
    }
  }

  const closeBootstrap = () => {
    setBootstrapTarget(null)
    setBootstrapAccountId('')
    setBootstrapping(false)
    setBootstrapDone(false)
    setBootstrapError('')
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Systems</h1>
          <p className="text-sm text-gray-500 mt-0.5">ระบบที่ register ไว้กับ Role Service</p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors"
        >
          + Register System
        </button>
      </div>

      {loading ? (
        <p className="text-gray-400 text-sm">Loading...</p>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {systems.map((s) => (
            <div key={s.id} className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
              <div className="flex items-center justify-between mb-2">
                <span className="bg-indigo-100 text-indigo-700 text-xs font-bold px-2 py-1 rounded-md uppercase">
                  {s.code}
                </span>
                <button
                  onClick={() => setReKeyTarget(s)}
                  className="text-xs text-gray-400 hover:text-orange-500 hover:bg-orange-50 border border-gray-200 hover:border-orange-200 px-2 py-1 rounded-md transition-colors"
                  title="Regenerate API Key"
                >
                  🔁 Re-key
                </button>
              </div>
              <p className="font-semibold text-gray-800">{s.name}</p>
              {s.description && <p className="text-sm text-gray-500 mt-1">{s.description}</p>}
              <div className="mt-3 pt-3 border-t border-gray-100 flex items-center justify-between">
                <button
                  onClick={() => setBootstrapTarget(s)}
                  className="text-xs text-gray-400 hover:text-emerald-600 transition-colors"
                >
                  🚀 Bootstrap first user
                </button>
                <button
                  onClick={() => { setCredTarget(s); setCredForm({ authClientId: '', authClientSecret: '' }) }}
                  className="text-xs text-gray-400 hover:text-indigo-600 transition-colors"
                  title="Set Auth Center credentials"
                >
                  🔑 Credentials
                </button>
              </div>
            </div>
          ))}
          {systems.length === 0 && (
            <div className="col-span-3 text-center py-16 text-gray-400">
              <p className="text-4xl mb-3">🖥️</p>
              <p className="text-sm">ยังไม่มีระบบที่ register</p>
            </div>
          )}
        </div>
      )}

      {/* Credentials Modal */}
      {credTarget && (
        <Modal title={`Auth Credentials: ${credTarget.code}`} onClose={closeCredentials}>
          {credDone ? (
            <div className="space-y-4">
              <div className="bg-emerald-50 border border-emerald-200 rounded-lg p-4">
                <p className="text-sm font-semibold text-emerald-700">✅ บันทึก Credentials สำเร็จ</p>
              </div>
              <button onClick={closeCredentials} className="w-full bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 transition-colors">
                ปิด
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              <p className="text-xs text-gray-500">
                Client credentials สำหรับใช้ขอ token จาก Auth Center ในชื่อของระบบนี้
              </p>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Client ID</label>
                <input
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="Auth Center client_id"
                  value={credForm.authClientId}
                  onChange={(e) => setCredForm({ ...credForm, authClientId: e.target.value })}
                  autoFocus
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Client Secret</label>
                <input
                  type="password"
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="Auth Center client_secret"
                  value={credForm.authClientSecret}
                  onChange={(e) => setCredForm({ ...credForm, authClientSecret: e.target.value })}
                />
              </div>
              {credError && <p className="text-sm text-red-600 bg-red-50 rounded-lg px-3 py-2">{credError}</p>}
              <div className="flex gap-2">
                <button
                  onClick={handleUpdateCredentials}
                  disabled={!credForm.authClientId || !credForm.authClientSecret || credSaving}
                  className="flex-1 bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
                >
                  {credSaving ? 'กำลังบันทึก...' : 'บันทึก'}
                </button>
                <button onClick={closeCredentials} className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm hover:bg-gray-50 transition-colors">
                  ยกเลิก
                </button>
              </div>
            </div>
          )}
        </Modal>
      )}

      {/* Bootstrap Modal */}
      {bootstrapTarget && (
        <Modal title={`Bootstrap: ${bootstrapTarget.code}`} onClose={closeBootstrap}>
          {bootstrapDone ? (
            <div className="space-y-4">
              <div className="bg-emerald-50 border border-emerald-200 rounded-lg p-4">
                <p className="text-sm font-semibold text-emerald-700 mb-1">✅ Bootstrap สำเร็จ!</p>
                <p className="text-xs text-emerald-600">
                  สร้าง role "Super Admin" (permission *:*) และ assign ให้ <strong>{bootstrapAccountId}</strong> แล้ว
                </p>
              </div>
              <p className="text-xs text-gray-500">
                ขั้นตอนต่อไป: ไปที่ Permissions → สร้าง permissions จริงๆ → ไปที่ Roles → สร้าง role ย่อย → ไปที่ Users → ย้าย user ไปใช้ role ที่ถูกต้อง
              </p>
              <button onClick={closeBootstrap} className="w-full bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 transition-colors">
                ปิด
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 text-xs text-blue-700 space-y-1">
                <p className="font-semibold">Bootstrap จะทำสิ่งนี้:</p>
                <p>1. สร้าง permission <code className="bg-blue-100 px-1 rounded">*:*</code> (wildcard = ทำได้ทุกอย่าง)</p>
                <p>2. สร้าง role <strong>Super Admin</strong> พร้อม permission นั้น</p>
                <p>3. Assign role ให้ Account ID ที่ระบุ</p>
                <p className="text-blue-500 pt-1">⚠️ ใช้ได้เฉพาะระบบที่ยังไม่มี role เลย</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Account ID ของ first user <span className="text-red-400">*</span>
                </label>
                <input
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="เช่น user-uuid-123 หรือ admin@company.com"
                  value={bootstrapAccountId}
                  onChange={(e) => setBootstrapAccountId(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleBootstrap()}
                  autoFocus
                />
              </div>
              {bootstrapError && (
                <p className="text-sm text-red-600 bg-red-50 rounded-lg px-3 py-2">{bootstrapError}</p>
              )}
              <div className="flex gap-2">
                <button
                  onClick={handleBootstrap}
                  disabled={!bootstrapAccountId.trim() || bootstrapping}
                  className="flex-1 bg-emerald-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-emerald-700 disabled:opacity-50 transition-colors"
                >
                  {bootstrapping ? 'กำลัง Bootstrap...' : '🚀 Bootstrap'}
                </button>
                <button onClick={closeBootstrap} className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm hover:bg-gray-50 transition-colors">
                  ยกเลิก
                </button>
              </div>
            </div>
          )}
        </Modal>
      )}

      {/* Re-key Modal */}
      {reKeyTarget && (
        <Modal title={`Re-key: ${reKeyTarget.code}`} onClose={closeReKey}>
          {reKeyResult ? (
            <div className="space-y-4">
              <div className="bg-orange-50 border border-orange-200 rounded-lg p-4">
                <p className="text-sm font-semibold text-orange-700 mb-1">⚠️ API Key ถูก rotate แล้ว</p>
                <p className="text-xs text-orange-600">Key เดิมใช้ไม่ได้แล้ว — เก็บ Key ใหม่นี้ไว้ จะไม่แสดงอีก</p>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1 uppercase tracking-wide">New API Key</label>
                <div className="flex gap-2">
                  <code className="flex-1 bg-gray-100 rounded-lg px-3 py-2 text-xs font-mono text-gray-800 break-all">
                    {reKeyResult}
                  </code>
                  <button
                    onClick={() => navigator.clipboard.writeText(reKeyResult)}
                    className="text-xs text-indigo-600 border border-indigo-200 rounded-lg px-3 py-2 hover:bg-indigo-50 whitespace-nowrap transition-colors"
                  >
                    Copy
                  </button>
                </div>
              </div>
              <button
                onClick={closeReKey}
                className="w-full bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 transition-colors"
              >
                ปิด
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                <p className="text-sm font-semibold text-red-700 mb-1">⚠️ ยืนยันการ Re-key</p>
                <p className="text-xs text-red-600">
                  API Key เดิมของ <strong>{reKeyTarget.name}</strong> จะถูกยกเลิกทันที
                  ทุก service ที่ใช้ key เดิมจะหยุดทำงาน จนกว่าจะอัปเดต key ใหม่
                </p>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={handleReKey}
                  disabled={reKeying}
                  className="flex-1 bg-orange-500 text-white rounded-lg py-2 text-sm font-medium hover:bg-orange-600 disabled:opacity-50 transition-colors"
                >
                  {reKeying ? 'กำลัง Generate...' : 'ยืนยัน Re-key'}
                </button>
                <button
                  onClick={closeReKey}
                  className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm hover:bg-gray-50 transition-colors"
                >
                  ยกเลิก
                </button>
              </div>
            </div>
          )}
        </Modal>
      )}

      {showModal && (
        <Modal title="Register System ใหม่" onClose={closeModal}>
          {/* หลังจาก register สำเร็จ แสดง API Key */}
          {apiKey ? (
            <div className="space-y-4">
              <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                <p className="text-sm font-semibold text-green-700 mb-1">✅ Register สำเร็จ!</p>
                <p className="text-xs text-green-600">เก็บ API Key นี้ไว้ จะไม่แสดงอีก</p>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1 uppercase tracking-wide">API Key</label>
                <div className="flex gap-2">
                  <code className="flex-1 bg-gray-100 rounded-lg px-3 py-2 text-xs font-mono text-gray-800 break-all">
                    {apiKey}
                  </code>
                  <button
                    onClick={() => navigator.clipboard.writeText(apiKey)}
                    className="text-xs text-indigo-600 border border-indigo-200 rounded-lg px-3 py-2 hover:bg-indigo-50 whitespace-nowrap transition-colors"
                  >
                    Copy
                  </button>
                </div>
              </div>
              <button
                onClick={closeModal}
                className="w-full bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 transition-colors"
              >
                ปิด
              </button>
            </div>
          ) : (
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  System Code <span className="text-red-400">*</span>
                </label>
                <input
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono uppercase focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="เช่น CRM, ERP, POS"
                  value={form.code}
                  onChange={(e) => setForm({ ...form, code: e.target.value.toUpperCase() })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  ชื่อระบบ <span className="text-red-400">*</span>
                </label>
                <input
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="เช่น ระบบ CRM"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">คำอธิบาย</label>
                <input
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="อธิบายระบบนี้"
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                />
              </div>
              <div className="border-t border-gray-100 pt-3">
                <p className="text-xs text-gray-400 mb-2">Auth Center Credentials (optional)</p>
                <div className="space-y-2">
                  <input
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    placeholder="Client ID"
                    value={form.authClientId}
                    onChange={(e) => setForm({ ...form, authClientId: e.target.value })}
                  />
                  <input
                    type="password"
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    placeholder="Client Secret"
                    value={form.authClientSecret}
                    onChange={(e) => setForm({ ...form, authClientSecret: e.target.value })}
                  />
                </div>
              </div>

              {error && (
                <p className="text-sm text-red-600 bg-red-50 rounded-lg px-3 py-2">{error}</p>
              )}

              <div className="flex gap-2 pt-2">
                <button
                  onClick={handleRegister}
                  disabled={!form.code || !form.name || saving}
                  className="flex-1 bg-indigo-600 text-white rounded-lg py-2 text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
                >
                  {saving ? 'กำลัง Register...' : 'Register'}
                </button>
                <button
                  onClick={closeModal}
                  className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm hover:bg-gray-50 transition-colors"
                >
                  ยกเลิก
                </button>
              </div>
            </div>
          )}
        </Modal>
      )}
    </div>
  )
}
