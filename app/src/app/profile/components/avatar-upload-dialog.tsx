import * as React from "react"
import { ImagePlus, Upload, X } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { cn } from "@/lib/utils"

const MAX_AVATAR_FILE_SIZE = 5 * 1024 * 1024

interface AvatarUploadDialogProps {
  open: boolean
  uploading: boolean
  onOpenChange: (open: boolean) => void
  onUpload: (file: File) => Promise<boolean>
}

export function AvatarUploadDialog({
  open,
  uploading,
  onOpenChange,
  onUpload,
}: AvatarUploadDialogProps) {
  const { t } = useTranslation()
  const [file, setFile] = React.useState<File | null>(null)
  const [previewURL, setPreviewURL] = React.useState("")
  const [dragActive, setDragActive] = React.useState(false)
  const [error, setError] = React.useState("")
  const inputRef = React.useRef<HTMLInputElement>(null)

  React.useEffect(() => {
    return () => {
      if (previewURL) URL.revokeObjectURL(previewURL)
    }
  }, [previewURL])

  const clearFile = () => {
    setFile(null)
    setPreviewURL("")
    if (inputRef.current) inputRef.current.value = ""
  }

  const reset = () => {
    clearFile()
    setDragActive(false)
    setError("")
  }

  const selectFile = (nextFile: File | null) => {
    if (!nextFile) return
    if (!nextFile.type.startsWith("image/")) {
      clearFile()
      setError(t("profile.avatarInvalidType"))
      return
    }
    if (nextFile.size > MAX_AVATAR_FILE_SIZE) {
      clearFile()
      setError(t("profile.avatarTooLarge"))
      return
    }

    setFile(nextFile)
    setPreviewURL(URL.createObjectURL(nextFile))
    setError("")
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && uploading) return
    onOpenChange(nextOpen)
    if (!nextOpen) reset()
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!file || uploading) return
    if (await onUpload(file)) handleOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleSubmit} className="grid gap-4">
          <DialogHeader>
            <DialogTitle>{t("profile.avatarDialogTitle")}</DialogTitle>
            <DialogDescription>
              {t("profile.avatarDialogDescription")}
            </DialogDescription>
          </DialogHeader>

          <div
            className={cn(
              "relative overflow-hidden rounded-lg border border-dashed border-input bg-muted/20 transition-colors",
              dragActive && "border-primary bg-primary/5",
              uploading && "pointer-events-none opacity-60"
            )}
            onDragEnter={(event) => {
              event.preventDefault()
              if (!uploading) setDragActive(true)
            }}
            onDragOver={(event) => {
              event.preventDefault()
              event.dataTransfer.dropEffect = "copy"
            }}
            onDragLeave={(event) => {
              if (
                !event.relatedTarget ||
                !event.currentTarget.contains(event.relatedTarget as Node)
              ) {
                setDragActive(false)
              }
            }}
            onDrop={(event) => {
              event.preventDefault()
              setDragActive(false)
              if (!uploading) selectFile(event.dataTransfer.files[0] ?? null)
            }}
          >
            <input
              ref={inputRef}
              id="avatar-file"
              type="file"
              accept="image/*"
              disabled={uploading}
              className="sr-only"
              onChange={(event) => selectFile(event.target.files?.[0] ?? null)}
            />
            <label
              htmlFor="avatar-file"
              className="flex min-h-52 cursor-pointer flex-col items-center justify-center px-6 py-8 text-center"
            >
              {file && previewURL ? (
                <>
                  <img
                    src={previewURL}
                    alt=""
                    className="size-24 rounded-full object-cover ring-1 ring-border"
                  />
                  <span className="mt-3 max-w-full truncate text-sm font-medium">
                    {file.name}
                  </span>
                  <span className="mt-1 text-xs text-muted-foreground">
                    {formatFileSize(file.size)}
                  </span>
                </>
              ) : (
                <>
                  <span className="mb-3 flex size-11 items-center justify-center">
                    <ImagePlus className="size-8" />
                  </span>
                  <span className="text-sm font-medium">
                    {t("profile.avatarDropTitle")}
                  </span>
                  <span className="mt-1 text-xs text-muted-foreground">
                    {t("profile.avatarTip")}
                  </span>
                </>
              )}
            </label>
            {file ? (
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="absolute top-2 right-2"
                disabled={uploading}
                title={t("profile.removeAvatarFile")}
                onClick={() => {
                  clearFile()
                  setError("")
                }}
              >
                <X />
              </Button>
            ) : null}
          </div>

          {error ? (
            <div className="border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          ) : null}

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" disabled={uploading}>
                {t("common.cancel")}
              </Button>
            </DialogClose>
            <Button type="submit" disabled={!file || uploading}>
              <Upload />
              {uploading ? t("common.saving") : t("profile.updateAvatar")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function formatFileSize(size: number) {
  const megabytes = size / (1024 * 1024)
  if (megabytes >= 1) return `${megabytes.toFixed(2)} MB`
  return `${Math.max(1, Math.round(size / 1024))} KB`
}
