import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Landing } from './pages/landing/Landing'
import { Register } from './pages/auth/Register'
import { Login } from './pages/auth/Login'
import { Hub } from './pages/hub/Hub'
import { ProtectedRoute } from './components/ProtectedRoute'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/"      element={<Landing />} />
        <Route path="/join"  element={<Register />} />
        <Route path="/login" element={<Login />} />
        <Route
          path="/hub"
          element={
            <ProtectedRoute>
              <Hub />
            </ProtectedRoute>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}
