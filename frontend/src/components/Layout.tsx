import { NavLink, useNavigate } from 'react-router-dom'

const nav = [
  { to: '/systems',     label: 'Systems',     icon: '🖥️', step: '1' },
  { to: '/permissions', label: 'Permissions', icon: '🔑', step: '2' },
  { to: '/roles',       label: 'Roles',       icon: '🎭', step: '3' },
  { to: '/users',       label: 'Users',       icon: '👤', step: '4' },
  { to: '/menus',       label: 'Menus',       icon: '📋', step: '5' },
]

export default function Layout({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate()

  const logout = () => {
    localStorage.removeItem('token')
    navigate('/login')
  }

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <aside className="w-56 bg-gray-900 text-white flex flex-col">
        <div className="px-5 py-5 border-b border-gray-700">
          <p className="text-xs text-gray-400 uppercase tracking-widest">Role Manager</p>
          <p className="text-lg font-semibold mt-0.5">Admin</p>
        </div>
        <nav className="flex-1 py-4">
          {nav.map((n) => (
            <NavLink
              key={n.to}
              to={n.to}
              className={({ isActive }) =>
                `flex items-center gap-3 px-5 py-2.5 text-sm transition-colors ${
                  isActive
                    ? 'bg-indigo-600 text-white'
                    : 'text-gray-300 hover:bg-gray-800'
                }`
              }
            >
              <span className="text-gray-500 text-xs w-4 text-right">{n.step}.</span>
              <span>{n.icon}</span>
              <span>{n.label}</span>
            </NavLink>
          ))}
        </nav>
        <button
          onClick={logout}
          className="px-5 py-4 text-sm text-gray-400 hover:text-white hover:bg-gray-800 text-left border-t border-gray-700 transition-colors"
        >
          🚪 Logout
        </button>
      </aside>

      {/* Content */}
      <main className="flex-1 overflow-auto p-8">{children}</main>
    </div>
  )
}
