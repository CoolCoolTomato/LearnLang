import { useCallback, useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Archive, AudioLines, CalendarCheck2, CalendarClock, FileText, MessageSquare, RefreshCw, UserRound } from "lucide-react"
import { toast } from "sonner"
import { getDeveloperDashboard } from "@/api/developer"
import { Button } from "@/components/ui/button"
import { getErrorMessage } from "@/lib/error"
import type { DeveloperDashboard } from "@/types/developer"
import { DeveloperLayout } from "./developer-layout"

function formatBytes(bytes: number) {
  if (bytes === 0) return "0 B"
  const units = ["B", "KB", "MB", "GB", "TB"]
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / 1024 ** index).toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

export default function DeveloperPage() {
  const { t } = useTranslation()
  const [dashboard, setDashboard] = useState<DeveloperDashboard | null>(null)
  const [loading, setLoading] = useState(true)

  const loadDashboard = useCallback(async () => {
    try {
      setLoading(true)
      setDashboard(await getDeveloperDashboard())
    } catch (error) {
      toast.error(getErrorMessage(error, t("developer.loadDashboardFailed")))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadDashboard()
  }, [loadDashboard])

  const stats = dashboard ? [
    { label: "Messages", value: dashboard.messages.toLocaleString(), icon: MessageSquare },
    { label: "Completed tasks", value: dashboard.completed_tasks.toLocaleString(), icon: CalendarCheck2 },
    { label: "Waiting tasks", value: dashboard.waiting_tasks.toLocaleString(), icon: CalendarClock },
    { label: "User summaries", value: dashboard.user_profile_summaries.toLocaleString(), icon: FileText },
    { label: "Conversation archives", value: dashboard.conversation_archives.toLocaleString(), icon: Archive },
    { label: "Voice files", value: dashboard.voice_files.toLocaleString(), icon: AudioLines },
    { label: "Voice storage", value: formatBytes(dashboard.voice_file_bytes), icon: AudioLines },
  ] : []

  return (
    <DeveloperLayout
      title="Developer Dashboard"
      description="Live overview of persisted application data."
      actions={<Button variant="outline" size="icon" onClick={loadDashboard} disabled={loading} title="Refresh dashboard"><RefreshCw className="h-4 w-4" /></Button>}
    >
      <div className="mt-6 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {loading ? Array.from({ length: 8 }, (_, index) => <div key={index} className="h-28 animate-pulse border bg-muted/40" />) : <>
          <div className="border p-4 sm:col-span-2">
            <UserRound className="h-5 w-5 text-muted-foreground" />
            <div className="mt-5 flex flex-wrap items-baseline gap-x-3 gap-y-1">
              <p className="text-2xl font-semibold">{dashboard?.current_user.username}</p>
              <p className="text-sm text-muted-foreground">ID {dashboard?.current_user.id} · {dashboard?.current_user.role}</p>
            </div>
            <p className="mt-1 truncate text-sm text-muted-foreground">{dashboard?.current_user.email || dashboard?.current_user.phone || "No contact information"}</p>
          </div>
          {stats.map((stat) => {
          const Icon = stat.icon
          return <div key={stat.label} className="border p-4">
            <Icon className="h-5 w-5 text-muted-foreground" />
            <p className="mt-5 text-2xl font-semibold tabular-nums">{stat.value}</p>
            <p className="mt-1 text-sm text-muted-foreground">{stat.label}</p>
          </div>
          })}
        </>}
      </div>
    </DeveloperLayout>
  )
}
