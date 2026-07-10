import { Database, List, Settings2, UserRound, Volume2 } from "lucide-react"
import { Link } from "react-router-dom"
import { developerResources } from "@/types/developer"

const resourceIcons = {
  messages: List,
  "scheduled-tasks": Settings2,
  "user-profile-summaries": UserRound,
  "conversation-archives": Database,
  "user-settings": Settings2,
  users: UserRound,
  "voice-files": Volume2,
}

export default function DeveloperPage() {
  return (
    <main className="min-h-screen bg-background px-4 py-6 md:px-8">
      <div className="mx-auto w-full max-w-6xl">
        <div className="mb-8 border-b pb-5">
          <div className="flex items-center gap-3">
            <Database className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-2xl font-semibold">Developer Data</h1>
          </div>
          <p className="mt-2 text-sm text-muted-foreground">Inspect and manage persisted application records.</p>
        </div>

        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {developerResources.map((resource) => {
            const Icon = resourceIcons[resource.slug]
            return (
              <Link
                key={resource.slug}
                to={`/developer/${resource.slug}`}
                className="group border p-4 transition-colors hover:bg-muted/50"
              >
                <Icon className="h-5 w-5 text-muted-foreground group-hover:text-foreground" />
                <h2 className="mt-4 font-medium">{resource.label}</h2>
                <p className="mt-1 text-sm text-muted-foreground">{resource.description}</p>
              </Link>
            )
          })}
        </div>
      </div>
    </main>
  )
}
