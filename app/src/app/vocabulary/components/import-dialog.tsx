import * as React from "react"
import { FileJson, Upload, X } from "lucide-react"
import { useTranslation } from "react-i18next"
import { importVocabulary } from "@/api/vocabulary"
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
import { getErrorMessage } from "@/lib/error"
import { cn } from "@/lib/utils"
import type {
  VocabularyImportResult,
  VocabularySummary,
} from "@/types/vocabulary"

const MAX_IMPORT_FILE_SIZE = 20 * 1024 * 1024

interface ImportDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  vocabulary: VocabularySummary | null
  onImported: (result: VocabularyImportResult) => Promise<void>
}

export function ImportDialog({
  open,
  onOpenChange,
  vocabulary,
  onImported,
}: ImportDialogProps) {
  const { t } = useTranslation()
  const [file, setFile] = React.useState<File | null>(null)
  const [importing, setImporting] = React.useState(false)
  const [dragActive, setDragActive] = React.useState(false)
  const [error, setError] = React.useState("")
  const inputRef = React.useRef<HTMLInputElement>(null)

  const clearFile = () => {
    setFile(null)
    if (inputRef.current) inputRef.current.value = ""
  }

  const reset = () => {
    clearFile()
    setDragActive(false)
    setError("")
  }

  const selectFile = (nextFile: File | null) => {
    if (!nextFile) return
    const isJSON =
      nextFile.type === "application/json" ||
      nextFile.name.toLowerCase().endsWith(".json")
    if (!isJSON) {
      clearFile()
      setError(t("vocabulary.fileInvalidType"))
      return
    }
    if (nextFile.size > MAX_IMPORT_FILE_SIZE) {
      clearFile()
      setError(t("vocabulary.fileTooLarge", "The file exceeds 20 MB"))
      return
    }
    setFile(nextFile)
    setError("")
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && importing) return
    onOpenChange(nextOpen)
    if (!nextOpen) reset()
  }

  const handleImport = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!file) {
      setError(t("vocabulary.fileRequired", "Choose a JSON file"))
      return
    }
    if (!vocabulary) {
      setError(t("vocabulary.createBeforeImport"))
      return
    }
    if (file.size > MAX_IMPORT_FILE_SIZE) {
      setError(t("vocabulary.fileTooLarge", "The file exceeds 20 MB"))
      return
    }

    try {
      setImporting(true)
      setError("")
      const parsed: unknown = JSON.parse(await file.text())
      const payload = buildImportPayload(parsed)
      const result = await importVocabulary(vocabulary.id, payload)
      reset()
      await onImported(result)
    } catch (importError: unknown) {
      setError(
        getErrorMessage(
          importError,
          t("vocabulary.importFailed", "Import failed")
        )
      )
    } finally {
      setImporting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleImport} className="grid gap-4">
          <DialogHeader>
            <DialogTitle>{t("vocabulary.importTitle")}</DialogTitle>
            <DialogDescription>
              {t("vocabulary.importInto", { name: vocabulary?.name })}
            </DialogDescription>
          </DialogHeader>

          <div
            className={cn(
              "relative overflow-hidden rounded-lg border border-dashed border-input bg-muted/20 transition-colors",
              dragActive && "border-primary bg-primary/5",
              importing && "pointer-events-none opacity-60"
            )}
            onDragEnter={(event) => {
              event.preventDefault()
              if (!importing) setDragActive(true)
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
              if (!importing) selectFile(event.dataTransfer.files[0] ?? null)
            }}
          >
            <input
              ref={inputRef}
              id="vocabulary-file"
              type="file"
              accept=".json,application/json"
              disabled={importing}
              className="sr-only"
              onChange={(event) => {
                selectFile(event.target.files?.[0] ?? null)
              }}
            />
            <label
              htmlFor="vocabulary-file"
              className="flex min-h-40 cursor-pointer flex-col items-center justify-center px-6 py-8 text-center"
            >
              <span className="mb-3 flex size-11 items-center justify-center rounded-md border border-border bg-background text-muted-foreground shadow-sm">
                <FileJson className="size-5" />
              </span>
              {file ? (
                <>
                  <span className="max-w-full truncate text-sm font-medium">
                    {file.name}
                  </span>
                  <span className="mt-1 text-xs text-muted-foreground">
                    {formatFileSize(file.size)}
                  </span>
                </>
              ) : (
                <>
                  <span className="text-sm font-medium">
                    {t("vocabulary.dropFileTitle")}
                  </span>
                  <span className="mt-1 text-xs text-muted-foreground">
                    {t("vocabulary.dropFileDescription")}
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
                disabled={importing}
                title={t("vocabulary.removeFile")}
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
              <Button type="button" variant="outline" disabled={importing}>
                {t("common.cancel", "Cancel")}
              </Button>
            </DialogClose>
            <Button type="submit" disabled={importing || !file}>
              <Upload />
              {importing
                ? t("vocabulary.importing", "Importing...")
                : t("vocabulary.import", "Import")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function buildImportPayload(parsed: unknown) {
  if (Array.isArray(parsed)) {
    return { entries: parsed }
  }
  if (typeof parsed !== "object" || parsed === null) {
    throw new Error("JSON must contain a vocabulary object or array")
  }
  if ("entries" in parsed && Array.isArray(parsed.entries)) {
    return { entries: parsed.entries }
  }
  return { entries: [parsed] }
}

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}
