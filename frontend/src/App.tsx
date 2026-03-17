import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import LoginPage from './pages/LoginPage'
import SystemsPage from './pages/SystemsPage'
import PermissionsPage from './pages/PermissionsPage'
import RolesPage from './pages/RolesPage'
import UsersPage from './pages/UsersPage'
import MenusPage from './pages/MenusPage'

function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token')
  if (!token) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/*"
          element={
            <RequireAuth>
              <Layout>
                <Routes>
                  <Route path="/" element={<Navigate to="/roles" replace />} />
                  <Route path="/systems" element={<SystemsPage />} />
                  <Route path="/permissions" element={<PermissionsPage />} />
                  <Route path="/roles" element={<RolesPage />} />
                  <Route path="/users" element={<UsersPage />} />
                  <Route path="/menus" element={<MenusPage />} />
                </Routes>
              </Layout>
            </RequireAuth>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}
