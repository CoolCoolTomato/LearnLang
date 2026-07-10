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
