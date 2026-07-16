import * as React from "react"
import {
  ArrowRight,
  BookOpenText,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Eraser,
  FileJson,
  LibraryBig,
  Languages,
  Pencil,
  Plus,
  Star,
  Trash2,
  Upload,
  Volume2,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import {
  clearVocabulary,
  createVocabulary,
  deleteVocabulary,
  getVocabularyEntries,
  importVocabulary,
  listVocabularies,
  updateVocabulary,
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
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { getErrorMessage } from "@/lib/error"
import { cn } from "@/lib/utils"
import type {
  Vocabulary,
  VocabularyEntry,
  VocabularyImportResult,
  VocabularyPage as VocabularyPageData,
  VocabularySummary,
} from "@/types/vocabulary"

const DEFAULT_PAGE_SIZE = 5
const PAGE_SIZE_OPTIONS = [5, 10, 20, 50]
const MAX_IMPORT_FILE_SIZE = 20 * 1024 * 1024

export default function VocabularyPage() {
  const { t } = useTranslation()
  const [vocabularies, setVocabularies] = React.useState<VocabularySummary[]>(
    []
  )
  const [selectedID, setSelectedID] = React.useState<number | null>(null)
  const [page, setPage] = React.useState(1)
  const [pageSize, setPageSize] = React.useState(DEFAULT_PAGE_SIZE)
  const [data, setData] = React.useState<VocabularyPageData | null>(null)
  const [listLoading, setListLoading] = React.useState(true)
  const [entriesLoading, setEntriesLoading] = React.useState(false)
  const [error, setError] = React.useState("")
  const [createOpen, setCreateOpen] = React.useState(false)
  const [editOpen, setEditOpen] = React.useState(false)
  const [importOpen, setImportOpen] = React.useState(false)
  const [clearOpen, setClearOpen] = React.useState(false)
  const [deleteOpen, setDeleteOpen] = React.useState(false)
  const [clearing, setClearing] = React.useState(false)
  const [deleting, setDeleting] = React.useState(false)
  const [expanded, setExpanded] = React.useState<Set<number>>(new Set())
  const [entriesVersion, setEntriesVersion] = React.useState(0)
  const entriesRequest = React.useRef(0)

  const loadVocabularies = React.useCallback(
    async (preferredID?: number) => {
      try {
        setListLoading(true)
        const result = await listVocabularies()
        const items = result.data ?? []
        setVocabularies(items)
        setSelectedID((current) => {
          const requested = preferredID ?? current
          if (requested && items.some((item) => item.id === requested)) {
            return requested
          }
          return (
            items.find((item) => item.is_default)?.id ?? items[0]?.id ?? null
          )
        })
        setError("")
      } catch (listError: unknown) {
        setError(
          getErrorMessage(listError, t("vocabulary.loadFailed", "Load failed"))
        )
      } finally {
        setListLoading(false)
      }
    },
    [t]
  )

  React.useEffect(() => {
    void loadVocabularies()
  }, [loadVocabularies])

  const loadEntries = React.useCallback(async () => {
    const request = ++entriesRequest.current
    if (!selectedID) {
      setData(null)
      setEntriesLoading(false)
      return
    }
    try {
      setEntriesLoading(true)
      const result = await getVocabularyEntries(selectedID, page, pageSize)
      if (request !== entriesRequest.current) return
      setData(result)
      setError("")
    } catch (loadError: unknown) {
      if (request !== entriesRequest.current) return
      setError(
        getErrorMessage(loadError, t("vocabulary.loadFailed", "Load failed"))
      )
    } finally {
      if (request === entriesRequest.current) {
        setEntriesLoading(false)
      }
    }
  }, [page, pageSize, selectedID, t])

  React.useEffect(() => {
    void entriesVersion
    void loadEntries()
  }, [entriesVersion, loadEntries])

  React.useEffect(() => {
    setExpanded(new Set())
    setPage(1)
    setData(null)
  }, [selectedID])

  React.useEffect(() => {
    setExpanded(new Set())
  }, [page])

  const selected = vocabularies.find((item) => item.id === selectedID) ?? null

  const handleVocabularySaved = async (vocabulary: Vocabulary) => {
    setCreateOpen(false)
    setEditOpen(false)
    setPage(1)
    await loadVocabularies(vocabulary.id)
  }

  const handleImported = async (result: VocabularyImportResult) => {
    setImportOpen(false)
    toast.success(
      t("vocabulary.importSuccess", {
        created: result.entries_created,
        updated: result.entries_updated,
      })
    )
    setPage(1)
    setData(null)
    setEntriesVersion((current) => current + 1)
    await loadVocabularies(result.vocabulary.id)
  }

  const handleClear = async () => {
    if (!selectedID) return
    try {
      setClearing(true)
      const result = await clearVocabulary(selectedID)
      setClearOpen(false)
      setPage(1)
      setData(null)
      setEntriesVersion((current) => current + 1)
      setExpanded(new Set())
      toast.success(t("vocabulary.clearSuccess", { count: result.deleted }))
      await loadVocabularies(selectedID)
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

  const handleDelete = async () => {
    if (!selectedID) return
    try {
      setDeleting(true)
      await deleteVocabulary(selectedID)
      setDeleteOpen(false)
      setData(null)
      setPage(1)
      await loadVocabularies()
      toast.success(t("vocabulary.deleteSuccess"))
    } catch (deleteError: unknown) {
      toast.error(getErrorMessage(deleteError, t("vocabulary.deleteFailed")))
    } finally {
      setDeleting(false)
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
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return (
    <div className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] bg-background lg:grid-cols-[minmax(0,1fr)_280px] lg:grid-rows-1">
      <main className="order-last min-h-0 min-w-0 lg:order-first">
        <ScrollArea className="h-full">
          <div className="mx-auto flex w-full max-w-6xl flex-col px-4 py-5 md:px-6 md:py-7">
            {selected ? (
              <section className="flex flex-col gap-4 border-b border-border/70 pb-5 sm:flex-row sm:items-end sm:justify-between">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <BookOpenText className="size-5 text-muted-foreground" />
                    <h2 className="truncate text-xl font-semibold">
                      {selected.name}
                    </h2>
                    {selected.is_default ? (
                      <Star className="size-4 fill-current text-amber-500" />
                    ) : null}
                  </div>
                  <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted-foreground">
                    <span className="inline-flex items-center gap-1.5">
                      <Languages className="size-3.5" />
                      {selected.target_language}
                      <span aria-hidden="true">→</span>
                      {selected.native_language}
                    </span>
                    <span>{t("vocabulary.entryCount", { count: total })}</span>
                  </div>
                </div>

                <div className="flex shrink-0 flex-wrap items-center gap-1.5">
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setEditOpen(true)}
                    title={t("vocabulary.edit")}
                  >
                    <Pencil />
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setClearOpen(true)}
                    disabled={total === 0}
                    title={t("vocabulary.clear")}
                  >
                    <Eraser />
                  </Button>
                  <Button
                    variant="destructive"
                    size="icon"
                    onClick={() => setDeleteOpen(true)}
                    title={t("vocabulary.delete")}
                  >
                    <Trash2 />
                  </Button>
                  <Button onClick={() => setImportOpen(true)}>
                    <Upload />
                    {t("vocabulary.import", "Import")}
                  </Button>
                </div>
              </section>
            ) : null}

            {listLoading ? (
              <LoadingState label={t("common.loading", "Loading...")} />
            ) : error ? (
              <ErrorState
                message={error}
                retryLabel={t("vocabulary.retry", "Retry")}
                onRetry={() => {
                  setEntriesVersion((current) => current + 1)
                  void loadVocabularies()
                }}
              />
            ) : !selected ? (
              <EmptyState
                title={t("vocabulary.noLibrariesTitle")}
                description={t("vocabulary.noLibrariesDescription")}
                actionLabel={t("vocabulary.create")}
                actionIcon="create"
                onAction={() => setCreateOpen(true)}
              />
            ) : entriesLoading && !data ? (
              <LoadingState label={t("common.loading", "Loading...")} />
            ) : total === 0 ? (
              <EmptyState
                title={t("vocabulary.emptyTitle", "No vocabulary yet")}
                description={t("vocabulary.emptyDescription")}
                actionLabel={t("vocabulary.import", "Import")}
                actionIcon="import"
                onAction={() => setImportOpen(true)}
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
                  pageSize={pageSize}
                  totalPages={totalPages}
                  total={total}
                  onPageChange={setPage}
                  onPageSizeChange={(nextPageSize) => {
                    setPageSize(nextPageSize)
                    setPage(1)
                    setData(null)
                  }}
                />
              </>
            )}
          </div>
        </ScrollArea>
      </main>

      <aside className="order-first min-h-0 border-b border-border/70 bg-muted/15 lg:order-last lg:border-b-0 lg:border-l">
        <div className="flex items-center justify-between gap-2 px-4 py-3 lg:h-14 lg:border-b lg:border-border/70">
          <div className="flex items-center gap-2 text-sm font-semibold">
            <LibraryBig className="size-4 text-muted-foreground" />
            {t("vocabulary.libraries")}
          </div>
          <Button
            size="icon-sm"
            variant="ghost"
            onClick={() => setCreateOpen(true)}
            title={t("vocabulary.create")}
          >
            <Plus />
          </Button>
        </div>
        <ScrollArea className="h-24 lg:h-[calc(100%-3.5rem)]" scrollbars="both">
          <div className="flex w-max gap-2 px-3 pb-3 lg:w-auto lg:flex-col lg:p-2">
            {vocabularies.map((vocabulary) => (
              <button
                key={vocabulary.id}
                type="button"
                className={cn(
                  "w-48 rounded-md border px-3 py-2 text-left transition-colors lg:w-auto",
                  vocabulary.id === selectedID
                    ? "border-primary/30 bg-background shadow-sm"
                    : "border-transparent hover:bg-muted"
                )}
                onClick={() => setSelectedID(vocabulary.id)}
              >
                <div className="flex items-center gap-1.5">
                  <span className="min-w-0 flex-1 truncate text-sm font-medium">
                    {vocabulary.name}
                  </span>
                  {vocabulary.is_default ? (
                    <Star className="size-3.5 fill-current text-amber-500" />
                  ) : null}
                </div>
                <div className="mt-1 flex items-center justify-between gap-2 text-xs text-muted-foreground">
                  <span>
                    {vocabulary.target_language} → {vocabulary.native_language}
                  </span>
                  <span>{vocabulary.entry_count}</span>
                </div>
              </button>
            ))}
          </div>
        </ScrollArea>
      </aside>

      <ImportDialog
        open={importOpen}
        onOpenChange={setImportOpen}
        vocabulary={selected}
        onImported={handleImported}
      />

      <VocabularyFormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        vocabulary={null}
        onSaved={handleVocabularySaved}
      />
      <VocabularyFormDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        vocabulary={selected}
        onSaved={handleVocabularySaved}
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

      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("vocabulary.deleteTitle")}</DialogTitle>
            <DialogDescription>
              {t("vocabulary.deleteDescription", { name: selected?.name })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline" disabled={deleting}>
                {t("common.cancel")}
              </Button>
            </DialogClose>
            <Button
              variant="destructive"
              onClick={() => void handleDelete()}
              disabled={deleting}
            >
              <Trash2 />
              {deleting
                ? t("vocabulary.deleting")
                : t("vocabulary.confirmDelete")}
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
  pageSize,
  totalPages,
  total,
  onPageChange,
  onPageSizeChange,
}: {
  page: number
  pageSize: number
  totalPages: number
  total: number
  onPageChange: (page: number) => void
  onPageSizeChange: (pageSize: number) => void
}) {
  const { t } = useTranslation()
  const [jumpPage, setJumpPage] = React.useState(String(page))

  React.useEffect(() => {
    setJumpPage(String(page))
  }, [page])

  const submitJump = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const requested = Number.parseInt(jumpPage, 10)
    if (Number.isNaN(requested)) {
      setJumpPage(String(page))
      return
    }
    const nextPage = Math.min(totalPages, Math.max(1, requested))
    setJumpPage(String(nextPage))
    onPageChange(nextPage)
  }

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 py-4">
      <span className="text-xs text-muted-foreground">
        {t("vocabulary.pagination", { page, totalPages, total })}
      </span>
      <div className="flex flex-wrap items-center justify-end gap-2">
        <div className="flex items-center gap-1.5">
          <span className="text-xs text-muted-foreground">
            {t("vocabulary.pageSize")}
          </span>
          <Select
            value={String(pageSize)}
            onValueChange={(value) => onPageSizeChange(Number(value))}
          >
            <SelectTrigger size="sm" className="w-16">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {PAGE_SIZE_OPTIONS.map((option) => (
                <SelectItem key={option} value={String(option)}>
                  {option}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <Button
          variant="outline"
          size="icon-sm"
          onClick={() => onPageChange(Math.max(1, page - 1))}
          disabled={page <= 1}
          title={t("vocabulary.previous", "Previous page")}
        >
          <ChevronLeft />
        </Button>
        <Button
          variant="outline"
          size="icon-sm"
          onClick={() => onPageChange(Math.min(totalPages, page + 1))}
          disabled={page >= totalPages}
          title={t("vocabulary.next", "Next page")}
        >
          <ChevronRight />
        </Button>

        <form className="flex items-center gap-1.5" onSubmit={submitJump}>
          <Label
            htmlFor="vocabulary-jump-page"
            className="text-xs text-muted-foreground"
          >
            {t("vocabulary.jumpTo")}
          </Label>
          <Input
            id="vocabulary-jump-page"
            type="number"
            min={1}
            max={totalPages}
            value={jumpPage}
            onChange={(event) => setJumpPage(event.target.value)}
            className="h-7 w-16 text-center"
          />
          <Button
            type="submit"
            variant="outline"
            size="icon-sm"
            title={t("vocabulary.goToPage")}
          >
            <ArrowRight />
          </Button>
        </form>
      </div>
    </div>
  )
}

function VocabularyFormDialog({
  open,
  onOpenChange,
  vocabulary,
  onSaved,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  vocabulary: VocabularySummary | null
  onSaved: (vocabulary: Vocabulary) => Promise<void>
}) {
  const { t } = useTranslation()
  const [name, setName] = React.useState("")
  const [targetLanguage, setTargetLanguage] = React.useState("")
  const [nativeLanguage, setNativeLanguage] = React.useState("")
  const [isDefault, setIsDefault] = React.useState(false)
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState("")
  const editing = vocabulary !== null

  React.useEffect(() => {
    if (!open) return
    setName(vocabulary?.name ?? "")
    setTargetLanguage(vocabulary?.target_language ?? "")
    setNativeLanguage(vocabulary?.native_language ?? "")
    setIsDefault(vocabulary?.is_default ?? false)
    setError("")
  }, [open, vocabulary])

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!name.trim()) {
      setError(t("vocabulary.nameRequired"))
      return
    }
    try {
      setSaving(true)
      setError("")
      const payload = {
        name: name.trim(),
        target_language: targetLanguage.trim(),
        native_language: nativeLanguage.trim(),
        is_default: isDefault,
      }
      const result = vocabulary
        ? await updateVocabulary(vocabulary.id, payload)
        : await createVocabulary(payload)
      await onSaved(result)
      toast.success(
        editing ? t("vocabulary.updateSuccess") : t("vocabulary.createSuccess")
      )
    } catch (saveError: unknown) {
      setError(getErrorMessage(saveError, t("vocabulary.saveFailed")))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(next) => !saving && onOpenChange(next)}>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit} className="grid gap-4">
          <DialogHeader>
            <DialogTitle>
              {editing
                ? t("vocabulary.editTitle")
                : t("vocabulary.createTitle")}
            </DialogTitle>
            <DialogDescription>
              {editing
                ? t("vocabulary.editDescription")
                : t("vocabulary.createDescription")}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-2">
            <Label
              htmlFor={
                editing ? "edit-vocabulary-name" : "create-vocabulary-name"
              }
            >
              {t("vocabulary.name")}
            </Label>
            <Input
              id={editing ? "edit-vocabulary-name" : "create-vocabulary-name"}
              value={name}
              onChange={(event) => setName(event.target.value)}
              disabled={saving}
              autoFocus
            />
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <div className="grid gap-2">
              <Label
                htmlFor={
                  editing ? "edit-target-language" : "create-target-language"
                }
              >
                {t("settings.targetLanguage")}
              </Label>
              <Input
                id={editing ? "edit-target-language" : "create-target-language"}
                value={targetLanguage}
                onChange={(event) => setTargetLanguage(event.target.value)}
                placeholder="en-US"
                disabled={saving}
              />
            </div>
            <div className="grid gap-2">
              <Label
                htmlFor={
                  editing ? "edit-native-language" : "create-native-language"
                }
              >
                {t("settings.nativeLanguage")}
              </Label>
              <Input
                id={editing ? "edit-native-language" : "create-native-language"}
                value={nativeLanguage}
                onChange={(event) => setNativeLanguage(event.target.value)}
                placeholder="zh-CN"
                disabled={saving}
              />
            </div>
          </div>

          {!vocabulary?.is_default ? (
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={isDefault}
                onChange={(event) => setIsDefault(event.target.checked)}
                disabled={saving}
                className="size-4 rounded border-input accent-primary"
              />
              {t("vocabulary.setDefault")}
            </label>
          ) : null}

          {error ? (
            <div className="border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          ) : null}

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" disabled={saving}>
                {t("common.cancel")}
              </Button>
            </DialogClose>
            <Button type="submit" disabled={saving}>
              {saving ? t("common.saving") : t("common.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function ImportDialog({
  open,
  onOpenChange,
  vocabulary,
  onImported,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  vocabulary: VocabularySummary | null
  onImported: (result: VocabularyImportResult) => Promise<void>
}) {
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
  actionLabel,
  actionIcon,
  onAction,
}: {
  title: string
  description: string
  actionLabel: string
  actionIcon: "create" | "import"
  onAction: () => void
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
      <Button className="mt-4" onClick={onAction}>
        {actionIcon === "create" ? <Plus /> : <Upload />}
        {actionLabel}
      </Button>
    </div>
  )
}
