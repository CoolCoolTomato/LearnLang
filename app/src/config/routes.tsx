import { lazy } from 'react'
import { Navigate } from 'react-router-dom'
import { ProtectedRoute } from '@/components/router/protected-route'

const Chat = lazy(() => import('@/app/chat/page'))
const Setting = lazy(() => import('@/app/setting/page'))
const Profile = lazy(() => import('@/app/profile/page'))
const SignIn = lazy(() => import('@/app/sign-in/page'))
const SignUp = lazy(() => import('@/app/sign-up/page'))

export interface RouteConfig {
  path: string
  element: React.ReactNode
  children?: RouteConfig[]
}

export const routes: RouteConfig[] = [
  {
    path: "/",
    element: <Navigate to="/chat" replace />
  },
  {
    path: "/chat",
    element: (
      <ProtectedRoute>
        <Chat />
      </ProtectedRoute>
    )
  },
  {
    path: "/setting",
    element: (
      <ProtectedRoute>
        <Setting />
      </ProtectedRoute>
    )
  },
  {
    path: "/profile",
    element: (
      <ProtectedRoute>
        <Profile />
      </ProtectedRoute>
    )
  },
  {
    path: "/sign-in",
    element: <SignIn />
  },
  {
    path: "/sign-up",
    element: <SignUp />
  },
]
