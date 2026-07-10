import type { ReactNode } from "react"
import { Archive, AudioLines, CalendarClock, Database, FileText, MessageSquare, Settings2, Users } from "lucide-react"
import { NavLink } from "react-router-dom"
import { cn } from "@/lib/utils"
import { developerResources, type DeveloperResource } from "@/types/developer"

const resourceIcons: Record<DeveloperResource, typeof Database> = {
  messages: MessageSquare,
  "scheduled-tasks": CalendarClock,
  "user-profile-summaries": FileText,
  "conversation-archives": Archive,
  "user-settings": Settings2,
  users: Users,
  "voice-files": AudioLines,
}

interface DeveloperLayoutProps {
  title: string
  description: string
  actions?: ReactNode
  children: ReactNode
}

export function DeveloperLayout({ title, description, actions, children }: DeveloperLayoutProps) {
  return (
    <div className="min-h-screen bg-background md:grid md:grid-cols-[15rem_minmax(0,1fr)]">
      <aside className="border-b bg-muted/20 md:min-h-screen md:border-r md:border-b-0">
        <div className="flex h-full flex-col px-3 py-4 md:sticky md:top-0 md:h-screen">
          <NavLink to="/developer" end className="mb-4 flex items-center gap-2 px-2 py-2 text-sm font-semibold">
            <Database className="h-5 w-5" /> Developer Data
          </NavLink>
          <nav className="flex gap-1 overflow-x-auto pb-1 md:flex-col md:overflow-visible">
            <NavLink
              to="/developer"
              end
              className={({ isActive }) => cn("flex shrink-0 items-center gap-2 px-2 py-2 text-sm text-muted-foreground hover:bg-muted hover:text-foreground", isActive && "bg-muted text-foreground")}
            >
              <Database className="h-4 w-4" /> Dashboard
            </NavLink>
            {developerResources.map((resource) => {
              const Icon = resourceIcons[resource.slug]
              return (
                <NavLink
                  key={resource.slug}
                  to={`/developer/${resource.slug}`}
                  className={({ isActive }) => cn("flex shrink-0 items-center gap-2 px-2 py-2 text-sm text-muted-foreground hover:bg-muted hover:text-foreground", isActive && "bg-muted text-foreground")}
                >
                  <Icon className="h-4 w-4" /> {resource.label}
                </NavLink>
              )
            })}
          </nav>
        </div>
      </aside>

      <main className="min-w-0 px-4 py-6 md:px-8">
        <div className="mx-auto w-full max-w-[1440px]">
          <header className="flex flex-wrap items-start justify-between gap-4 border-b pb-5">
            <div>
              <h1 className="text-2xl font-semibold">{title}</h1>
              <p className="mt-1 text-sm text-muted-foreground">{description}</p>
            </div>
            {actions ? <div className="flex items-center gap-2">{actions}</div> : null}
          </header>
          {children}
        </div>
      </main>
    </div>
  )
}
