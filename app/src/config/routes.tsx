import { lazy } from "react"
import { Navigate } from "react-router-dom"
import { ProtectedRoute } from "@/components/router/protected-route"
import { InitializationGate } from "@/components/router/initialization-gate"
import { AppLayout } from "@/components/layout/app-layout"

const Chat = lazy(() => import("@/app/chat/page"))
const Setting = lazy(() => import("@/app/setting/page"))
const Profile = lazy(() => import("@/app/profile/page"))
const Vocabulary = lazy(() => import("@/app/vocabulary/page"))
const Usage = lazy(() => import("@/app/usage/page"))
const SignIn = lazy(() => import("@/app/sign-in/page"))
const SignUp = lazy(() => import("@/app/sign-up/page"))
const Initialization = lazy(() => import("@/app/initialize/page"))
const Developer = lazy(() => import("@/app/developer/page"))
const DeveloperResource = lazy(() => import("@/app/developer/resource-page"))

export interface RouteConfig {
  path?: string
  element: React.ReactNode
  children?: RouteConfig[]
}

export const routes: RouteConfig[] = [
  {
    path: "/",
    element: <Navigate to="/chat" replace />,
  },
  {
    element: (
      <ProtectedRoute>
        <InitializationGate>
          <AppLayout />
        </InitializationGate>
      </ProtectedRoute>
    ),
    children: [
      { path: "/chat", element: <Chat /> },
      { path: "/vocabulary", element: <Vocabulary /> },
      { path: "/usage", element: <Usage /> },
      { path: "/setting", element: <Setting /> },
      { path: "/profile", element: <Profile /> },
    ],
  },
  {
    path: "/initialize",
    element: (
      <ProtectedRoute>
        <Initialization />
      </ProtectedRoute>
    ),
  },
  {
    path: "/sign-in",
    element: <SignIn />,
  },
  {
    path: "/sign-up",
    element: <SignUp />,
  },
  {
    path: "/developer",
    element: (
      <ProtectedRoute>
        <Developer />
      </ProtectedRoute>
    ),
  },
  {
    path: "/developer/messages",
    element: (
      <ProtectedRoute>
        <DeveloperResource />
      </ProtectedRoute>
    ),
  },
  {
    path: "/developer/scheduled-tasks",
    element: (
      <ProtectedRoute>
        <DeveloperResource />
      </ProtectedRoute>
    ),
  },
  {
    path: "/developer/user-profile-summaries",
    element: (
      <ProtectedRoute>
        <DeveloperResource />
      </ProtectedRoute>
    ),
  },
  {
    path: "/developer/conversation-archives",
    element: (
      <ProtectedRoute>
        <DeveloperResource />
      </ProtectedRoute>
    ),
  },
  {
    path: "/developer/user-settings",
    element: (
      <ProtectedRoute>
        <DeveloperResource />
      </ProtectedRoute>
    ),
  },
  {
    path: "/developer/users",
    element: (
      <ProtectedRoute>
        <DeveloperResource />
      </ProtectedRoute>
    ),
  },
  {
    path: "/developer/voice-files",
    element: (
      <ProtectedRoute>
        <DeveloperResource />
      </ProtectedRoute>
    ),
  },
]
