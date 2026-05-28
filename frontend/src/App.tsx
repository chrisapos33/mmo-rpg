import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Landing } from './pages/landing/Landing'
import { Register } from './pages/auth/Register'
import { Login } from './pages/auth/Login'
import { Upload } from './pages/onboarding/Upload'
import { Reveal } from './pages/onboarding/Reveal'
import { Hub } from './pages/hub/Hub'
import { PublicProfile } from './pages/profile/PublicProfile'
import { ProtectedRoute } from './components/ProtectedRoute'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/"           element={<Landing />} />
        <Route path="/join"       element={<Register />} />
        <Route path="/login"      element={<Login />} />
        <Route path="/p/:userId"  element={<PublicProfile />} />
        <Route
          path="/onboarding/upload"
          element={<ProtectedRoute><Upload /></ProtectedRoute>}
        />
        <Route
          path="/onboarding/reveal"
          element={<ProtectedRoute><Reveal /></ProtectedRoute>}
        />
        <Route
          path="/hub"
          element={<ProtectedRoute><Hub /></ProtectedRoute>}
        />
      </Routes>
    </BrowserRouter>
  )
}
