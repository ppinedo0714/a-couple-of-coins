import { QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import { Navbar } from '@/components/layout/Navbar'
import { queryClient } from '@/lib/queryClient'
import { ThemeProvider } from '@/lib/theme'
import { AppRoutes } from '@/routes'

export default function App() {
  return (
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <div className="min-h-screen">
            <Navbar />
            <AppRoutes />
          </div>
        </BrowserRouter>
      </QueryClientProvider>
    </ThemeProvider>
  )
}
