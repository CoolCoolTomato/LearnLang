import { useCallback, useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { ChevronLeft, ChevronRight, LoaderCircle, Pencil, Plus, RefreshCw, Search, Trash2 } from "lucide-react"
import { Navigate, useLocation } from "react-router-dom"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  createDeveloperRecord,
  deleteDeveloperRecords,
  listDeveloperRecords,
  searchDeveloperArchives,
  updateDeveloperRecord,
} from "@/api/developer"
import { getErrorMessage } from "@/lib/error"
import { developerResources, type DeveloperArchiveSearchResult, type DeveloperResource } from "@/types/developer"
import { DeveloperLayout } from "./developer-layout"
import { Input } from "@/components/ui/input"

const pageSize = 20

function isDeveloperResource(value: string | undefined): value is DeveloperResource {
  return developerResources.some((resource) => resource.slug === value)
}

function editableValues(record: Record<string, unknown>) {
  return Object.fromEntries(
    Object.entries(record).filter(([key]) => !["id", "created_at", "updated_at", "voice_file"].includes(key))
  )
}

function formatValue(value: unknown) {
  if (value === null || value === undefined) return "-"
  if (typeof value === "object") return JSON.stringify(value)
  return String(value)
}

export default function DeveloperResourcePage() {
  const { t } = useTranslation()
  const location = useLocation()
  const resourceParam = location.pathname.split("/").filter(Boolean).at(-1)
  const resource = isDeveloperResource(resourceParam) ? resourceParam : null
  const resourceInfo = developerResources.find((item) => item.slug === resource)
  const [records, setRecords] = useState<Record<string, unknown>[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [editorOpen, setEditorOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editorValue, setEditorValue] = useState("{}")
  const [saving, setSaving] = useState(false)
  const [searchQuery, setSearchQuery] = useState("")
  const [searchLimit, setSearchLimit] = useState(5)
  const [searchResults, setSearchResults] = useState<DeveloperArchiveSearchResult[]>([])
  const [searchLoading, setSearchLoading] = useState(false)
  const [searchSubmitted, setSearchSubmitted] = useState(false)

  const loadRecords = useCallback(async () => {
    if (!resource) return
    try {
      setLoading(true)
      const response = await listDeveloperRecords(resource, page, pageSize)
      setRecords(response.data || [])
      setTotal(response.total)
      setSelectedIds([])
    } catch (error) {
      toast.error(getErrorMessage(error, t("developer.loadRecordsFailed")))
    } finally {
      setLoading(false)
    }
  }, [page, resource, t])

  const searchArchives = async () => {
    const query = searchQuery.trim()
    if (!query) {
      toast.error(t("developer.searchQueryRequired"))
      return
    }
    try {
      setSearchLoading(true)
      const response = await searchDeveloperArchives(query, searchLimit)
      setSearchResults(response.results || [])
      setSearchSubmitted(true)
    } catch (error) {
      toast.error(getErrorMessage(error, t("developer.searchFailed")))
    } finally {
      setSearchLoading(false)
    }
  }

  useEffect(() => {
    loadRecords()
  }, [loadRecords])

  const columns = useMemo(() => {
    const keys = new Set<string>(["id"])
    records.forEach((record) => Object.keys(record).forEach((key) => keys.add(key)))
    return Array.from(keys).filter((key) => key !== "voice_file")
  }, [records])

  if (!resource || !resourceInfo) return <Navigate to="/developer" replace />

  const allSelected = records.length > 0 && selectedIds.length === records.length
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const toggleRecord = (id: number) => {
    setSelectedIds((current) => current.includes(id) ? current.filter((value) => value !== id) : [...current, id])
  }

  const openCreate = () => {
    setEditingId(null)
    setEditorValue("{}")
    setEditorOpen(true)
  }

  const openEdit = (record: Record<string, unknown>) => {
    setEditingId(Number(record.id))
    setEditorValue(JSON.stringify(editableValues(record), null, 2))
    setEditorOpen(true)
  }

  const saveRecord = async () => {
    let values: Record<string, unknown>
    try {
      const parsed: unknown = JSON.parse(editorValue)
      if (!parsed || Array.isArray(parsed) || typeof parsed !== "object") throw new Error("invalid")
      values = parsed as Record<string, unknown>
    } catch {
      toast.error(t("developer.invalidJson"))
      return
    }

    try {
      setSaving(true)
      if (editingId === null) {
        await createDeveloperRecord(resource, values)
        toast.success(t("developer.createSuccess"))
      } else {
        await updateDeveloperRecord(resource, editingId, values)
        toast.success(t("developer.updateSuccess"))
      }
      setEditorOpen(false)
      await loadRecords()
    } catch (error) {
      toast.error(getErrorMessage(error, t("developer.saveFailed")))
    } finally {
      setSaving(false)
    }
  }

  const deleteSelected = async () => {
    if (selectedIds.length === 0) return
    if (!window.confirm(`Delete ${selectedIds.length} selected record(s)?`)) return
    try {
      await deleteDeveloperRecords(resource, selectedIds)
      toast.success(t("developer.deleteSuccess"))
      if (records.length === selectedIds.length && page > 1) setPage((current) => current - 1)
      else await loadRecords()
    } catch (error) {
      toast.error(getErrorMessage(error, t("developer.deleteFailed")))
    }
  }

  return (
    <>
    <DeveloperLayout
      title={resourceInfo.label}
      description={`${total} total records`}
      actions={<>
        <Button variant="outline" size="icon" onClick={loadRecords} disabled={loading} title="Refresh"><RefreshCw className="h-4 w-4" /></Button>
        <Button variant="destructive" onClick={deleteSelected} disabled={selectedIds.length === 0}><Trash2 className="mr-2 h-4 w-4" /> Delete ({selectedIds.length})</Button>
        <Button onClick={openCreate}><Plus className="mr-2 h-4 w-4" /> New record</Button>
      </>}
    >

      {resource === "conversation-archives" && <section className="mt-5 border p-4">
        <div className="flex flex-col gap-3 md:flex-row md:items-end">
          <label className="min-w-0 flex-1 text-sm font-medium">
            RAG retrieval query
            <Input
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              onKeyDown={(event) => { if (event.key === "Enter") void searchArchives() }}
              placeholder="Search archived memory"
              className="mt-1"
              disabled={searchLoading}
            />
          </label>
          <label className="w-full text-sm font-medium md:w-24">
            Top K
            <Input
              type="number"
              min={1}
              max={20}
              value={searchLimit}
              onChange={(event) => setSearchLimit(Math.min(20, Math.max(1, Number(event.target.value) || 1)))}
              className="mt-1"
              disabled={searchLoading}
            />
          </label>
          <Button onClick={() => void searchArchives()} disabled={searchLoading || !searchQuery.trim()}>
            {searchLoading ? <LoaderCircle className="mr-2 h-4 w-4 animate-spin" /> : <Search className="mr-2 h-4 w-4" />}
            Search
          </Button>
        </div>

        {searchSubmitted && <div className="mt-4 overflow-x-auto border">
          <table className="w-full text-left text-sm">
            <thead className="border-b bg-muted/50 text-xs text-muted-foreground">
              <tr>
                <th className="px-3 py-3 font-medium">Score</th>
                <th className="px-3 py-3 font-medium">Archive</th>
                <th className="px-3 py-3 font-medium">Summary</th>
                <th className="px-3 py-3 font-medium">Messages</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {searchResults.length === 0 ? (
                <tr><td colSpan={4} className="px-3 py-8 text-center text-muted-foreground">No matching memories</td></tr>
              ) : searchResults.map((result) => (
                <tr key={result.embedding_id} className="hover:bg-muted/30">
                  <td className="whitespace-nowrap px-3 py-3 align-top font-mono text-xs">{result.score.toFixed(4)}</td>
                  <td className="whitespace-nowrap px-3 py-3 align-top">#{result.archive_id || "-"}</td>
                  <td className="min-w-72 max-w-xl px-3 py-3 align-top">{result.summary}</td>
                  <td className="max-w-xl px-3 py-3 align-top text-xs text-muted-foreground">
                    <div className="mb-1 font-mono">IDs: {result.message_ids.join(", ") || "-"}</div>
                    {result.messages.map((message) => <div key={String(message.id)} className="truncate" title={formatValue(message.text_content)}>{String(message.role || "message")}: {formatValue(message.text_content)}</div>)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>}
      </section>}

        <div className="mt-5 overflow-x-auto border">
          <table className="w-full text-left text-sm">
            <thead className="border-b bg-muted/50 text-xs text-muted-foreground">
              <tr>
                <th className="w-10 px-3 py-3">
                  <input
                    type="checkbox"
                    checked={allSelected}
                    onChange={() => setSelectedIds(allSelected ? [] : records.map((record) => Number(record.id)))}
                    aria-label="Select all records"
                  />
                </th>
                {columns.map((column) => <th key={column} className="whitespace-nowrap px-3 py-3 font-medium">{column}</th>)}
                <th className="w-12 px-3 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y">
              {loading ? (
                <tr><td colSpan={columns.length + 2} className="px-3 py-10 text-center text-muted-foreground">Loading...</td></tr>
              ) : records.length === 0 ? (
                <tr><td colSpan={columns.length + 2} className="px-3 py-10 text-center text-muted-foreground">No records found</td></tr>
              ) : records.map((record) => {
                const id = Number(record.id)
                return <tr key={id} className="hover:bg-muted/30">
                  <td className="px-3 py-3"><input type="checkbox" checked={selectedIds.includes(id)} onChange={() => toggleRecord(id)} aria-label={`Select record ${id}`} /></td>
                  {columns.map((column) => <td key={column} className="max-w-72 truncate px-3 py-3 align-top" title={formatValue(record[column])}>{formatValue(record[column])}</td>)}
                  <td className="px-3 py-3"><Button variant="ghost" size="icon" onClick={() => openEdit(record)} title="Edit record"><Pencil className="h-4 w-4" /></Button></td>
                </tr>
              })}
            </tbody>
          </table>
        </div>

        <div className="mt-4 flex items-center justify-between gap-3 text-sm text-muted-foreground">
          <span>Page {page} of {totalPages}</span>
          <div className="flex gap-2">
            <Button variant="outline" size="icon" disabled={page <= 1 || loading} onClick={() => setPage((current) => current - 1)} title="Previous page"><ChevronLeft className="h-4 w-4" /></Button>
            <Button variant="outline" size="icon" disabled={page >= totalPages || loading} onClick={() => setPage((current) => current + 1)} title="Next page"><ChevronRight className="h-4 w-4" /></Button>
          </div>
        </div>
    </DeveloperLayout>

      <Dialog open={editorOpen} onOpenChange={setEditorOpen}>
        <DialogContent className="sm:max-w-4xl">
          <DialogHeader>
            <DialogTitle>{editingId === null ? "Create record" : `Edit record #${editingId}`}</DialogTitle>
            <DialogDescription>Provide writable fields as a JSON object. Server-managed IDs and timestamps are ignored.</DialogDescription>
          </DialogHeader>
          <Textarea value={editorValue} onChange={(event) => setEditorValue(event.target.value)} className="h-64 max-h-[40vh] min-h-0 resize-y font-mono text-xs" spellCheck={false} />
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditorOpen(false)} disabled={saving}>Cancel</Button>
            <Button onClick={saveRecord} disabled={saving}>{saving ? "Saving..." : "Save"}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
