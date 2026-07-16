import { FileJson, Plus, Upload } from "lucide-react"
import { Button } from "@/components/ui/button"

export function LoadingState({ label }: { label: string }) {
  return (
    <div className="flex min-h-[40vh] items-center justify-center text-sm text-muted-foreground">
      {label}
    </div>
  )
}

interface ErrorStateProps {
  message: string
  retryLabel: string
  onRetry: () => void
}

export function ErrorState({ message, retryLabel, onRetry }: ErrorStateProps) {
  return (
    <div className="my-8 flex min-h-48 flex-col items-center justify-center gap-3 border border-destructive/25 bg-destructive/5 p-6 text-center">
      <p className="text-sm text-destructive">{message}</p>
      <Button variant="outline" onClick={onRetry}>
        {retryLabel}
      </Button>
    </div>
  )
}

interface EmptyStateProps {
  title: string
  description: string
  actionLabel: string
  actionIcon: "create" | "import"
  onAction: () => void
}

export function EmptyState({
  title,
  description,
  actionLabel,
  actionIcon,
  onAction,
}: EmptyStateProps) {
  return (
    <div className="flex min-h-[48vh] flex-col items-center justify-center px-4 text-center">
      <div className="mb-4 flex size-12 items-center justify-center rounded-lg border border-border bg-muted/40">
        <FileJson className="size-6 text-muted-foreground" />
      </div>
      <h2 className="text-base font-semibold">{title}</h2>
      <p className="mt-1 max-w-sm text-sm text-muted-foreground">
        {description}
      </p>
      <Button className="mt-4" onClick={onAction}>
        {actionIcon === "create" ? <Plus /> : <Upload />}
        {actionLabel}
      </Button>
    </div>
  )
}
