import { createContext, useContext, useState, type ReactNode } from 'react'
import type { User } from '@/types/auth'
import { getCurrentUser, isAuthenticated as checkAuth } from '@/api/auth'
import { USER_KEY } from '@/api/config'

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  setUser: (user: User | null) => void
  clearAuth: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUserState] = useState<User | null>(() => getCurrentUser())

  const setUser = (nextUser: User | null) => {
    setUserState(nextUser)
    if (nextUser) {
      localStorage.setItem(USER_KEY, JSON.stringify(nextUser))
    } else {
      localStorage.removeItem(USER_KEY)
    }
  }

  const clearAuth = () => {
    setUser(null)
  }

  const value: AuthContextType = {
    user,
    isAuthenticated: checkAuth(),
    isLoading: false,
    setUser,
    clearAuth,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
