import { Navigate, Route, Routes } from 'react-router-dom'
import { ProtectedRoute, UnprotectedRoute } from '@/components/layout/ProtectedRoute'
import WelcomePage from '@/pages/Welcome'
import LoginPage from '@/pages/Login'
import RegisterPage from '@/pages/Register'
import OnboardingPage from '@/pages/Onboarding'
import AccountsPage from '@/pages/Accounts'
import TransactionsPage from '@/pages/Transactions'
import ImportPage from '@/pages/Import'
import SettingsPage from '@/pages/Settings'
import NotFoundPage from '@/pages/NotFound'

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<WelcomePage />} />

      <Route element={<UnprotectedRoute />}>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
      </Route>

      <Route element={<ProtectedRoute />}>
        <Route path="/onboarding" element={<OnboardingPage />} />
        <Route path="/dashboard" element={<Navigate to="/accounts" replace />} />
        <Route path="/accounts" element={<AccountsPage />} />
        <Route path="/transactions" element={<TransactionsPage />} />
        <Route path="/import" element={<ImportPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Route>

      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}
