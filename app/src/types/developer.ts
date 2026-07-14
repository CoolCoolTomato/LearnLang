export const developerResources = [
  { slug: "messages", label: "Messages", description: "Chat messages and translations" },
  { slug: "scheduled-tasks", label: "Scheduled Tasks", description: "Queued agent work" },
  { slug: "user-profile-summaries", label: "User Profile Summaries", description: "Long-lived learner context" },
  { slug: "conversation-archives", label: "Conversation Archives", description: "Archived conversation segments" },
  { slug: "user-settings", label: "User Settings", description: "Model and language configuration" },
  { slug: "users", label: "Users", description: "Application accounts" },
  { slug: "voice-files", label: "Voice Files", description: "Stored speech assets" },
] as const

export type DeveloperResource = (typeof developerResources)[number]["slug"]

export interface DeveloperPage<T = Record<string, unknown>> {
  data: T[]
  total: number
  page: number
  size: number
}

export interface DeveloperDashboard {
  messages: number
  completed_tasks: number
  waiting_tasks: number
  current_user: {
    id: number
    username: string
    email: string | null
    phone: string | null
    role: string
    last_active_at: string | null
    created_at: string
    updated_at: string
  }
  user_profile_summaries: number
  conversation_archives: number
  voice_files: number
  voice_file_bytes: number
}

export interface DeveloperArchiveSearchResult {
  embedding_id: string
  archive_id: number
  score: number
  summary: string
  message_ids: number[]
  messages: Record<string, unknown>[]
  created_at: string
}
