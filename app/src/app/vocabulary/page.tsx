import * as React from "react"
import {
  BookOpenText,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  FileJson,
  Languages,
  Trash2,
  Upload,
  Volume2,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import {
  clearVocabulary,
  getVocabulary,
  importVocabulary,
} from "@/api/vocabulary"
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
import { cn } from "@/lib/utils"
import type {
  VocabularyEntry,
  VocabularyImportResult,
  VocabularyPage as VocabularyPageData,
} from "@/types/vocabulary"

const PAGE_SIZE = 20
const MAX_IMPORT_FILE_SIZE = 20 * 1024 * 1024

export default function VocabularyPage() {
  const { t } = useTranslation()
  const [page, setPage] = React.useState(1)
  const [data, setData] = React.useState<VocabularyPageData | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [importOpen, setImportOpen] = React.useState(false)
  const [clearOpen, setClearOpen] = React.useState(false)
  const [clearing, setClearing] = React.useState(false)
  const [expanded, setExpanded] = React.useState<Set<number>>(new Set())

  const loadVocabulary = React.useCallback(async () => {
    try {
      setLoading(true)
      const result = await getVocabulary(page, PAGE_SIZE)
      setData(result)
      setError("")
    } catch (loadError: unknown) {
      setError(
        getErrorMessage(loadError, t("vocabulary.loadFailed", "Load failed"))
      )
    } finally {
      setLoading(false)
    }
  }, [page, t])

  React.useEffect(() => {
    void loadVocabulary()
  }, [loadVocabulary])

  React.useEffect(() => {
    setExpanded(new Set())
  }, [page])

  const handleImported = async (result: VocabularyImportResult) => {
    setImportOpen(false)
    toast.success(
      t("vocabulary.importSuccess", {
        created: result.entries_created,
        updated: result.entries_updated,
      })
    )
    if (page !== 1) {
      setPage(1)
      return
    }
    await loadVocabulary()
  }

  const handleClear = async () => {
    try {
      setClearing(true)
      const result = await clearVocabulary()
      setClearOpen(false)
      setPage(1)
      setExpanded(new Set())
      toast.success(t("vocabulary.clearSuccess", { count: result.deleted }))
      if (page === 1) {
        await loadVocabulary()
      }
    } catch (clearError: unknown) {
      toast.error(
        getErrorMessage(
          clearError,
          t("vocabulary.clearFailed", "Failed to clear vocabulary")
        )
      )
    } finally {
      setClearing(false)
    }
  }

  const toggleExpanded = (entryID: number) => {
    setExpanded((current) => {
      const next = new Set(current)
      if (next.has(entryID)) {
        next.delete(entryID)
      } else {
        next.add(entryID)
      }
      return next
    })
  }

  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <div className="min-h-full bg-background">
      <div className="mx-auto flex w-full max-w-6xl flex-col px-4 py-5 md:px-6 md:py-7">
        <section className="flex flex-col gap-4 border-b border-border/70 pb-5 sm:flex-row sm:items-end sm:justify-between">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <BookOpenText className="size-5 text-muted-foreground" />
              <h2 className="truncate text-xl font-semibold">
                {data?.vocabulary?.name || t("vocabulary.title", "Vocabulary")}
              </h2>
            </div>
            <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted-foreground">
              {data?.vocabulary ? (
                <span className="inline-flex items-center gap-1.5">
                  <Languages className="size-3.5" />
                  {data.vocabulary.target_language}
                  <span aria-hidden="true">→</span>
                  {data.vocabulary.native_language}
                </span>
              ) : null}
              <span>{t("vocabulary.entryCount", { count: total })}</span>
            </div>
          </div>

          <div className="flex shrink-0 flex-wrap items-center gap-2">
            <Button
              variant="outline"
              onClick={() => setClearOpen(true)}
              disabled={total === 0}
            >
              <Trash2 />
              {t("vocabulary.clear", "Clear")}
            </Button>
            <Button onClick={() => setImportOpen(true)}>
              <Upload />
              {t("vocabulary.import", "Import")}
            </Button>
          </div>
        </section>

        {loading && !data ? (
          <LoadingState label={t("common.loading", "Loading...")} />
        ) : error ? (
          <ErrorState
            message={error}
            retryLabel={t("vocabulary.retry", "Retry")}
            onRetry={() => void loadVocabulary()}
          />
        ) : total === 0 ? (
          <EmptyState
            title={t("vocabulary.emptyTitle", "No vocabulary yet")}
            description={t(
              "vocabulary.emptyDescription",
              "Import a JSON vocabulary to get started."
            )}
            importLabel={t("vocabulary.import", "Import")}
            onImport={() => setImportOpen(true)}
          />
        ) : (
          <>
            <div className="mt-5 overflow-hidden rounded-lg border border-border/70">
              <div className="hidden h-9 grid-cols-[minmax(180px,1.2fr)_minmax(220px,1.6fr)_120px_40px] items-center gap-4 border-b bg-muted/40 px-4 text-xs font-medium text-muted-foreground md:grid">
                <span>{t("vocabulary.targetText", "Target language")}</span>
                <span>{t("vocabulary.meaning", "Native meaning")}</span>
                <span>{t("vocabulary.status", "Status")}</span>
                <span className="sr-only">
                  {t("vocabulary.details", "Details")}
                </span>
              </div>
              <div className="divide-y divide-border/70">
                {data?.data.map((entry) => (
                  <VocabularyRow
                    key={entry.id}
                    entry={entry}
                    expanded={expanded.has(entry.id)}
                    onToggle={() => toggleExpanded(entry.id)}
                  />
                ))}
              </div>
            </div>

            <Pagination
              page={page}
              totalPages={totalPages}
              total={total}
              onPrevious={() => setPage((current) => Math.max(1, current - 1))}
              onNext={() =>
                setPage((current) => Math.min(totalPages, current + 1))
              }
            />
          </>
        )}
      </div>

      <ImportDialog
        open={importOpen}
        onOpenChange={setImportOpen}
        onImported={handleImported}
      />

      <Dialog open={clearOpen} onOpenChange={setClearOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("vocabulary.clearTitle")}</DialogTitle>
            <DialogDescription>
              {t("vocabulary.clearDescription", { count: total })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline" disabled={clearing}>
                {t("common.cancel", "Cancel")}
              </Button>
            </DialogClose>
            <Button
              variant="destructive"
              onClick={() => void handleClear()}
              disabled={clearing}
            >
              <Trash2 />
              {clearing
                ? t("vocabulary.clearing", "Clearing...")
                : t("vocabulary.confirmClear", "Clear vocabulary")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function VocabularyRow({
  entry,
  expanded,
  onToggle,
}: {
  entry: VocabularyEntry
  expanded: boolean
  onToggle: () => void
}) {
  const { t } = useTranslation()
  const pronunciations = entry.pronunciations ?? []
  const meanings = entry.meanings ?? []
  const examples = entry.examples ?? []
  const relations = entry.relations ?? []
  const tags = entry.tags ?? []
  const hasDetails =
    examples.length > 0 ||
    relations.length > 0 ||
    tags.length > 0 ||
    Boolean(entry.notes)

  const playAudio = (audioURL: string) => {
    void new Audio(audioURL).play().catch(() => {
      toast.error(t("vocabulary.audioFailed", "Unable to play audio"))
    })
  }

  return (
    <article className="bg-background">
      <div className="relative grid min-h-20 grid-cols-[minmax(0,1fr)_36px] gap-3 px-4 py-3 md:grid-cols-[minmax(180px,1.2fr)_minmax(220px,1.6fr)_120px_40px] md:items-center md:gap-4">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-base font-semibold break-words">
              {entry.target_text}
            </span>
            {entry.entry_type === "phrase" ? (
              <span className="rounded border border-border bg-muted/50 px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground uppercase">
                {t("vocabulary.phrase", "Phrase")}
              </span>
            ) : null}
          </div>
          {pronunciations.length > 0 ? (
            <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
              {pronunciations.map((item) => (
                <span key={item.id} className="inline-flex items-center gap-1">
                  {item.region ? (
                    <span className="font-medium uppercase">{item.region}</span>
                  ) : null}
                  <span>/{item.pronunciation}/</span>
                  {item.audio_url ? (
                    <button
                      type="button"
                      className="inline-flex size-5 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
                      onClick={() => playAudio(item.audio_url!)}
                      title={t("vocabulary.playAudio", "Play audio")}
                      aria-label={t("vocabulary.playAudio", "Play audio")}
                    >
                      <Volume2 className="size-3" />
                    </button>
                  ) : null}
                </span>
              ))}
            </div>
          ) : null}
        </div>

        <div className="col-span-2 min-w-0 md:col-span-1">
          {meanings.length > 0 ? (
            <div className="space-y-1">
              {meanings.map((meaning) => (
                <div key={meaning.id} className="flex min-w-0 gap-2 text-sm">
                  {meaning.part_of_speech ? (
                    <span className="shrink-0 font-mono text-xs text-muted-foreground">
                      {meaning.part_of_speech}
                    </span>
                  ) : null}
                  <span className="break-words">{meaning.native_text}</span>
                </div>
              ))}
            </div>
          ) : (
            <span className="text-sm text-muted-foreground">—</span>
          )}
        </div>

        <div className="col-span-2 flex items-center gap-2 md:col-span-1">
          <span
            className={cn(
              "inline-flex rounded border px-1.5 py-0.5 text-xs font-medium",
              entry.encountered
                ? "border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400"
                : "border-border bg-muted/40 text-muted-foreground"
            )}
          >
            {entry.encountered
              ? t("vocabulary.encountered", "Encountered")
              : t("vocabulary.notEncountered", "Not encountered")}
          </span>
          {entry.encounter_count > 0 ? (
            <span className="text-xs text-muted-foreground">
              ×{entry.encounter_count}
            </span>
          ) : null}
        </div>

        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="absolute right-3 mt-0 md:static"
          onClick={onToggle}
          disabled={!hasDetails}
          aria-expanded={expanded}
          title={t("vocabulary.details", "Details")}
        >
          <ChevronDown
            className={cn("transition-transform", expanded && "rotate-180")}
          />
        </Button>
      </div>

      {expanded && hasDetails ? (
        <div className="grid gap-5 border-t border-dashed border-border/70 bg-muted/20 px-4 py-4 md:grid-cols-2">
          {relations.length > 0 ? (
            <DetailSection title={t("vocabulary.relatedPhrases")}>
              <div className="flex flex-wrap gap-2">
                {relations.map((relation) => (
                  <span
                    key={relation.id}
                    className="inline-flex max-w-full items-baseline gap-1.5 rounded border border-border bg-background px-2 py-1 text-sm"
                  >
                    <span className="font-medium break-words">
                      {relation.related_entry?.target_text}
                    </span>
                    <span className="break-words text-muted-foreground">
                      {relation.related_entry?.meanings?.[0]?.native_text}
                    </span>
                  </span>
                ))}
              </div>
            </DetailSection>
          ) : null}

          {examples.length > 0 ? (
            <DetailSection title={t("vocabulary.examples", "Examples")}>
              <div className="space-y-3">
                {examples.map((example) => (
                  <div key={example.id} className="text-sm">
                    <p className="break-words text-foreground">
                      {example.target_text}
                    </p>
                    {example.native_text ? (
                      <p className="mt-0.5 break-words text-muted-foreground">
                        {example.native_text}
                      </p>
                    ) : null}
                  </div>
                ))}
              </div>
            </DetailSection>
          ) : null}

          {tags.length > 0 ? (
            <DetailSection title={t("vocabulary.tags", "Tags")}>
              <div className="flex flex-wrap gap-1.5">
                {tags.map((tag) => (
                  <span
                    key={tag}
                    className="rounded bg-muted px-2 py-1 text-xs text-muted-foreground"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </DetailSection>
          ) : null}

          {entry.notes ? (
            <DetailSection title={t("vocabulary.notes", "Notes")}>
              <p className="text-sm break-words whitespace-pre-wrap text-muted-foreground">
                {entry.notes}
              </p>
            </DetailSection>
          ) : null}
        </div>
      ) : null}
    </article>
  )
}

function DetailSection({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (
    <section className="min-w-0">
      <h3 className="mb-2 text-xs font-semibold text-muted-foreground uppercase">
        {title}
      </h3>
      {children}
    </section>
  )
}

function Pagination({
  page,
  totalPages,
  total,
  onPrevious,
  onNext,
}: {
  page: number
  totalPages: number
  total: number
  onPrevious: () => void
  onNext: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="flex items-center justify-between gap-3 py-4">
      <span className="text-xs text-muted-foreground">
        {t("vocabulary.pagination", { page, totalPages, total })}
      </span>
      <div className="flex items-center gap-1">
        <Button
          variant="outline"
          size="icon-sm"
          onClick={onPrevious}
          disabled={page <= 1}
          title={t("vocabulary.previous", "Previous page")}
        >
          <ChevronLeft />
        </Button>
        <Button
          variant="outline"
          size="icon-sm"
          onClick={onNext}
          disabled={page >= totalPages}
          title={t("vocabulary.next", "Next page")}
        >
          <ChevronRight />
        </Button>
      </div>
    </div>
  )
}

function ImportDialog({
  open,
  onOpenChange,
  onImported,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onImported: (result: VocabularyImportResult) => Promise<void>
}) {
  const { t } = useTranslation()
  const [file, setFile] = React.useState<File | null>(null)
  const [name, setName] = React.useState("")
  const [targetLanguage, setTargetLanguage] = React.useState("")
  const [nativeLanguage, setNativeLanguage] = React.useState("")
  const [importing, setImporting] = React.useState(false)
  const [error, setError] = React.useState("")

  const reset = () => {
    setFile(null)
    setName("")
    setTargetLanguage("")
    setNativeLanguage("")
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
    if (file.size > MAX_IMPORT_FILE_SIZE) {
      setError(t("vocabulary.fileTooLarge", "The file exceeds 20 MB"))
      return
    }

    try {
      setImporting(true)
      setError("")
      const parsed: unknown = JSON.parse(await file.text())
      const payload = buildImportPayload(
        parsed,
        name,
        targetLanguage,
        nativeLanguage
      )
      const result = await importVocabulary(payload)
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
              {t("vocabulary.importDescription")}
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

          <div className="grid gap-2">
            <Label htmlFor="vocabulary-name">
              {t("vocabulary.name", "Vocabulary name")}
            </Label>
            <Input
              id="vocabulary-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder={t("vocabulary.defaultName", "Default Vocabulary")}
              disabled={importing}
            />
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <div className="grid gap-2">
              <Label htmlFor="vocabulary-target-language">
                {t("settings.targetLanguage", "Target Language")}
              </Label>
              <Input
                id="vocabulary-target-language"
                value={targetLanguage}
                onChange={(event) => setTargetLanguage(event.target.value)}
                placeholder="en-US"
                disabled={importing}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="vocabulary-native-language">
                {t("settings.nativeLanguage", "Native Language")}
              </Label>
              <Input
                id="vocabulary-native-language"
                value={nativeLanguage}
                onChange={(event) => setNativeLanguage(event.target.value)}
                placeholder="zh-CN"
                disabled={importing}
              />
            </div>
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

function buildImportPayload(
  parsed: unknown,
  name: string,
  targetLanguage: string,
  nativeLanguage: string
) {
  const overrides = {
    ...(name.trim() ? { name: name.trim() } : {}),
    ...(targetLanguage.trim()
      ? { target_language: targetLanguage.trim() }
      : {}),
    ...(nativeLanguage.trim()
      ? { native_language: nativeLanguage.trim() }
      : {}),
  }

  if (Array.isArray(parsed)) {
    return { ...overrides, entries: parsed }
  }
  if (typeof parsed !== "object" || parsed === null) {
    throw new Error("JSON must contain a vocabulary object or array")
  }
  if ("entries" in parsed && Array.isArray(parsed.entries)) {
    return { ...parsed, ...overrides }
  }
  return { ...overrides, entries: [parsed] }
}

function LoadingState({ label }: { label: string }) {
  return (
    <div className="flex min-h-[40vh] items-center justify-center text-sm text-muted-foreground">
      {label}
    </div>
  )
}

function ErrorState({
  message,
  retryLabel,
  onRetry,
}: {
  message: string
  retryLabel: string
  onRetry: () => void
}) {
  return (
    <div className="my-8 flex min-h-48 flex-col items-center justify-center gap-3 border border-destructive/25 bg-destructive/5 p-6 text-center">
      <p className="text-sm text-destructive">{message}</p>
      <Button variant="outline" onClick={onRetry}>
        {retryLabel}
      </Button>
    </div>
  )
}

function EmptyState({
  title,
  description,
  importLabel,
  onImport,
}: {
  title: string
  description: string
  importLabel: string
  onImport: () => void
}) {
  return (
    <div className="flex min-h-[48vh] flex-col items-center justify-center px-4 text-center">
      <div className="mb-4 flex size-12 items-center justify-center rounded-lg border border-border bg-muted/40">
        <FileJson className="size-6 text-muted-foreground" />
      </div>
      <h2 className="text-base font-semibold">{title}</h2>
      <p className="mt-1 max-w-sm text-sm text-muted-foreground">
        {description}
      </p>
      <Button className="mt-4" onClick={onImport}>
        <Upload />
        {importLabel}
      </Button>
    </div>
  )
}
