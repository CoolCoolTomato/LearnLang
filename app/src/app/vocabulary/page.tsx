import * as React from "react"
import {
  BookOpenText,
  Eraser,
  LibraryBig,
  Languages,
  PanelRightClose,
  PanelRightOpen,
  Pencil,
  Plus,
  Search,
  Star,
  Trash2,
  Upload,
  X,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import {
  clearVocabulary,
  deleteVocabulary,
  getVocabularyEntries,
  listVocabularies,
  setDefaultVocabulary,
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
import { ScrollArea } from "@/components/ui/scroll-area"
import { getErrorMessage } from "@/lib/error"
import { cn } from "@/lib/utils"
import type {
  Vocabulary,
  VocabularyImportResult,
  VocabularyPage as VocabularyPageData,
  VocabularySummary,
} from "@/types/vocabulary"
import { ImportDialog } from "./components/import-dialog"
import { Pagination } from "./components/pagination"
import { EmptyState, ErrorState, LoadingState } from "./components/page-state"
import { VocabularyFormDialog } from "./components/vocabulary-form-dialog"
import { VocabularyRow } from "./components/vocabulary-row"
import { VocabularySearch } from "./components/vocabulary-search"

const DEFAULT_PAGE_SIZE = 5
const LIBRARIES_PANEL_STORAGE_KEY = "learnlang_vocabulary_libraries_open"

export default function VocabularyPage() {
  const { t } = useTranslation()
  const [vocabularies, setVocabularies] = React.useState<VocabularySummary[]>(
    []
  )
  const [selectedID, setSelectedID] = React.useState<number | null>(null)
  const [page, setPage] = React.useState(1)
  const [pageSize, setPageSize] = React.useState(DEFAULT_PAGE_SIZE)
  const [librariesOpen, setLibrariesOpen] = React.useState(
    () => localStorage.getItem(LIBRARIES_PANEL_STORAGE_KEY) !== "false"
  )
  const [mobileLibrariesOpen, setMobileLibrariesOpen] = React.useState(false)
  const [searchQuery, setSearchQuery] = React.useState("")
  const [debouncedSearchQuery, setDebouncedSearchQuery] = React.useState("")
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
  const [settingDefault, setSettingDefault] = React.useState(false)
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
      const result = await getVocabularyEntries(
        selectedID,
        page,
        pageSize,
        debouncedSearchQuery
      )
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
  }, [debouncedSearchQuery, page, pageSize, selectedID, t])

  React.useEffect(() => {
    void entriesVersion
    void loadEntries()
  }, [entriesVersion, loadEntries])

  React.useEffect(() => {
    setExpanded(new Set())
    setPage(1)
    setData(null)
    setSearchQuery("")
    setDebouncedSearchQuery("")
  }, [selectedID])

  React.useEffect(() => {
    const timeout = window.setTimeout(() => {
      setDebouncedSearchQuery(searchQuery.trim())
    }, 300)
    return () => window.clearTimeout(timeout)
  }, [searchQuery])

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

  const handleSetDefault = async () => {
    if (!selectedID || selected?.is_default) return
    try {
      setSettingDefault(true)
      await setDefaultVocabulary(selectedID)
      await loadVocabularies(selectedID)
      toast.success(t("vocabulary.setDefaultSuccess"))
    } catch (setDefaultError: unknown) {
      toast.error(
        getErrorMessage(
          setDefaultError,
          t("vocabulary.setDefaultFailed", "Failed to set default vocabulary")
        )
      )
    } finally {
      setSettingDefault(false)
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

  const toggleLibrariesPanel = () => {
    setLibrariesOpen((open) => {
      localStorage.setItem(LIBRARIES_PANEL_STORAGE_KEY, String(!open))
      return !open
    })
  }

  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const vocabularyEntryCount = selected?.entry_count ?? 0
  const hasSearchQuery = debouncedSearchQuery.length > 0

  return (
    <div
      className={cn(
        "grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] bg-background transition-[grid-template-columns] duration-200 lg:grid-rows-1",
        librariesOpen
          ? "lg:grid-cols-[minmax(0,1fr)_280px]"
          : "lg:grid-cols-[minmax(0,1fr)_48px]"
      )}
    >
      <main className="order-last min-h-0 min-w-0 lg:order-first">
        <ScrollArea className="h-full">
          <div className="mx-auto flex w-full max-w-6xl flex-col px-4 py-5 md:px-6 md:py-7">
            <div className="mb-3 flex justify-end lg:hidden">
              <Button
                size="icon"
                variant="outline"
                onClick={() => setMobileLibrariesOpen(true)}
                aria-label={t("vocabulary.openLibraries")}
                title={t("vocabulary.openLibraries")}
              >
                <PanelRightOpen />
              </Button>
            </div>
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
                    <span>
                      {t("vocabulary.entryCount", {
                        count: vocabularyEntryCount,
                      })}
                    </span>
                  </div>
                </div>

                <div className="flex shrink-0 flex-wrap items-center gap-1.5">
                  {!selected.is_default ? (
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => void handleSetDefault()}
                      disabled={settingDefault}
                      title={t("vocabulary.setDefault")}
                    >
                      <Star />
                    </Button>
                  ) : null}
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
                    disabled={vocabularyEntryCount === 0}
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

            {selected && vocabularyEntryCount > 0 ? (
              <VocabularySearch
                value={searchQuery}
                resultCount={total}
                onChange={(value) => {
                  setSearchQuery(value)
                  setPage(1)
                  setExpanded(new Set())
                }}
              />
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
            ) : vocabularyEntryCount === 0 ? (
              <EmptyState
                title={t("vocabulary.emptyTitle", "No vocabulary yet")}
                description={t("vocabulary.emptyDescription")}
                actionLabel={t("vocabulary.import", "Import")}
                actionIcon="import"
                onAction={() => setImportOpen(true)}
              />
            ) : hasSearchQuery && total === 0 ? (
              <div className="flex min-h-[40vh] flex-col items-center justify-center px-4 text-center">
                <Search className="mb-3 size-6 text-muted-foreground" />
                <h2 className="text-sm font-semibold">
                  {t("vocabulary.noSearchResultsTitle", "No matching entries")}
                </h2>
                <p className="mt-1 text-sm text-muted-foreground">
                  {t(
                    "vocabulary.noSearchResultsDescription",
                    "Try another word or meaning."
                  )}
                </p>
              </div>
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

      {mobileLibrariesOpen ? (
        <button
          type="button"
          className="fixed inset-0 z-40 bg-black/30 lg:hidden"
          onClick={() => setMobileLibrariesOpen(false)}
          aria-label={t("vocabulary.closeLibraries")}
        />
      ) : null}

      <aside
        className={cn(
          "order-first min-h-0 border-b border-border/70 bg-background lg:order-last lg:border-b-0 lg:border-l",
          mobileLibrariesOpen
            ? "fixed inset-y-0 right-0 z-50 flex h-svh w-[min(86vw,22rem)] flex-col border-l shadow-xl lg:static lg:h-auto lg:w-auto lg:shadow-none"
            : "hidden lg:block"
        )}
      >
        <div
          className={cn(
            "flex h-14 shrink-0 items-center justify-between gap-2 border-b border-border/70 px-4",
            !librariesOpen && "lg:justify-center lg:px-2"
          )}
        >
          <div
            className={cn(
              "flex items-center gap-2 text-sm font-semibold",
              !librariesOpen && "lg:hidden"
            )}
          >
            <LibraryBig className="size-4 text-muted-foreground" />
            {t("vocabulary.libraries")}
          </div>
          <div className="flex items-center gap-1">
            <Button
              size="icon-sm"
              variant="ghost"
              className={cn(!librariesOpen && "lg:hidden")}
              onClick={() => setCreateOpen(true)}
              title={t("vocabulary.create")}
            >
              <Plus />
            </Button>
            <Button
              size="icon-sm"
              variant="ghost"
              className="hidden lg:inline-flex"
              onClick={toggleLibrariesPanel}
              aria-expanded={librariesOpen}
              title={
                librariesOpen
                  ? t("vocabulary.collapseLibraries")
                  : t("vocabulary.expandLibraries")
              }
            >
              {librariesOpen ? <PanelRightClose /> : <PanelRightOpen />}
            </Button>
            <Button
              size="icon-sm"
              variant="ghost"
              className="lg:hidden"
              onClick={() => setMobileLibrariesOpen(false)}
              aria-label={t("vocabulary.closeLibraries")}
              title={t("vocabulary.closeLibraries")}
            >
              <X />
            </Button>
          </div>
        </div>
        <ScrollArea
          className={cn(
            "h-[calc(100%-3.5rem)] lg:h-[calc(100%-3.5rem)]",
            !librariesOpen && "lg:hidden"
          )}
          scrollbars="both"
        >
          <div className="flex w-full flex-col gap-2 p-3 lg:w-auto lg:p-2">
            {vocabularies.map((vocabulary) => (
              <button
                key={vocabulary.id}
                type="button"
                className={cn(
                  "w-full rounded-md border px-3 py-2 text-left transition-colors lg:w-auto",
                  vocabulary.id === selectedID
                    ? "border-primary/30 bg-background shadow-sm"
                    : "border-transparent hover:bg-muted"
                )}
                onClick={() => {
                  setSelectedID(vocabulary.id)
                  setMobileLibrariesOpen(false)
                }}
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
              {t("vocabulary.clearDescription", {
                count: vocabularyEntryCount,
              })}
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
