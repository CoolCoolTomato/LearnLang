import { useCallback, useEffect, useMemo, useState } from "react"
import { ChevronLeft, ChevronRight, Pencil, Plus, RefreshCw, Trash2 } from "lucide-react"
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
  updateDeveloperRecord,
} from "@/api/developer"
import { getErrorMessage } from "@/lib/error"
import { developerResources, type DeveloperResource } from "@/types/developer"
import { DeveloperLayout } from "./developer-layout"

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

  const loadRecords = useCallback(async () => {
    if (!resource) return
    try {
      setLoading(true)
      const response = await listDeveloperRecords(resource, page, pageSize)
      setRecords(response.data || [])
      setTotal(response.total)
      setSelectedIds([])
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to load developer records"))
    } finally {
      setLoading(false)
    }
  }, [page, resource])

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
      toast.error("Enter a valid JSON object")
      return
    }

    try {
      setSaving(true)
      if (editingId === null) {
        await createDeveloperRecord(resource, values)
        toast.success("Record created")
      } else {
        await updateDeveloperRecord(resource, editingId, values)
        toast.success("Record updated")
      }
      setEditorOpen(false)
      await loadRecords()
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to save record"))
    } finally {
      setSaving(false)
    }
  }

  const deleteSelected = async () => {
    if (selectedIds.length === 0) return
    if (!window.confirm(`Delete ${selectedIds.length} selected record(s)?`)) return
    try {
      await deleteDeveloperRecords(resource, selectedIds)
      toast.success("Selected records deleted")
      if (records.length === selectedIds.length && page > 1) setPage((current) => current - 1)
      else await loadRecords()
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to delete records"))
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
