import * as React from "react"
import { Upload } from "lucide-react"
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
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { getErrorMessage } from "@/lib/error"
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
  const [error, setError] = React.useState("")

  const reset = () => {
    setFile(null)
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

          <div className="grid gap-2">
            <Label htmlFor="vocabulary-file">
              {t("vocabulary.jsonFile", "JSON file")}
            </Label>
            <Input
              id="vocabulary-file"
              type="file"
              accept=".json,application/json"
              disabled={importing}
              onChange={(event) => {
                setFile(event.target.files?.[0] ?? null)
                setError("")
              }}
            />
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
