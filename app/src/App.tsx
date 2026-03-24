import { BrowserRouter as Router } from 'react-router-dom'
import { ThemeProvider } from '@/components/theme/theme-provider'
import { AuthProvider } from '@/contexts/auth-context'
import { AppRouter } from '@/components/router/app-router'
import { Toaster } from 'sonner'
import { useEffect } from 'react'
import { initGTM } from '@/utils/analytics'
import './App.css'

const basename = import.meta.env.VITE_BASENAME || ''

function App() {
  useEffect(() => {
    initGTM();
  }, []);

  return (
    <div className="font-sans antialiased" style={{ fontFamily: 'var(--font-inter)' }}>
      <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
        <AuthProvider>
          <Router basename={basename}>
            <AppRouter />
          </Router>
        </AuthProvider>
        <Toaster richColors position="top-right" />
      </ThemeProvider>
    </div>
  )
}

export default App
