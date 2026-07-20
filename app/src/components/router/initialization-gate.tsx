import { useEffect, useState, type ReactNode } from "react"
import { Navigate } from "react-router-dom"
import { getSettings } from "@/api/settings"
import type { Settings } from "@/types/settings"
import { LoadingSpinner } from "@/components/ui/loading-spinner"

function needsInitialization(settings: Settings) {
  return !settings.timezone || !settings.native_language || !settings.target_language
}

export function InitializationGate({ children }: { children: ReactNode }) {
  const [state, setState] = useState<
    { status: "loading" } | { status: "ready"; needsSetup: boolean } | { status: "error"; message: string }
  >({ status: "loading" })

  useEffect(() => {
    let active = true
    getSettings()
      .then((settings) => {
        if (active) setState({ status: "ready", needsSetup: needsInitialization(settings) })
      })
      .catch(() => {
        if (active) setState({ status: "error", message: "Unable to load your settings." })
      })
    return () => {
      active = false
    }
  }, [])

  if (state.status === "loading") {
    return <div className="flex h-screen items-center justify-center"><LoadingSpinner /></div>
  }
  if (state.status === "error") {
    return <div className="flex h-screen items-center justify-center p-6 text-sm text-destructive">{state.message}</div>
  }
  if (state.needsSetup) return <Navigate to="/initialize" replace />
  return <>{children}</>
}
