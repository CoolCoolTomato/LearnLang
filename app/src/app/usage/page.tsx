import * as React from "react"
import { BarChart3, RefreshCw } from "lucide-react"
import { useTranslation } from "react-i18next"
import { listAIUsage, summarizeAIUsage } from "@/api/ai-usage"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { getErrorMessage } from "@/lib/error"
import type {
  AIUsageEvent,
  AIUsageOperation,
  AIUsageSummary,
} from "@/types/ai-usage"
import { Pagination } from "../vocabulary/components/pagination"

const operations: Array<AIUsageOperation | "all"> = [
  "all",
  "chat",
  "tts",
  "stt",
  "embedding",
  "translation",
]

function formatUsage(event: Pick<AIUsageEvent, "usage" | "unit">) {
  const value = Number.isInteger(event.usage)
    ? event.usage.toLocaleString()
    : event.usage.toFixed(2)
  return `${value} ${event.unit}`
}

export default function UsagePage() {
  const { t } = useTranslation()
  const [items, setItems] = React.useState<AIUsageEvent[]>([])
  const [summary, setSummary] = React.useState<AIUsageSummary[]>([])
  const [operation, setOperation] = React.useState<AIUsageOperation | "all">(
    "all"
  )
  const [page, setPage] = React.useState(1)
  const [pageSize, setPageSize] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const filter = operation === "all" ? "" : operation
      const [events, totals] = await Promise.all([
        listAIUsage(page, pageSize, filter),
        summarizeAIUsage(filter),
      ])
      setItems(events.items)
      setTotal(events.total)
      setSummary(totals)
      setError("")
    } catch (cause: unknown) {
      setError(getErrorMessage(cause, t("usage.loadFailed")))
    } finally {
      setLoading(false)
    }
  }, [operation, page, t])

  React.useEffect(() => {
    void load()
  }, [load])
  const pageCount = Math.max(1, Math.ceil(total / pageSize))

  return (
    <div className="h-full overflow-auto p-4 md:p-6">
      <div className="mx-auto max-w-5xl space-y-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <div className="mb-2 flex items-center gap-2 text-primary">
              <BarChart3 className="size-5" />
              <span className="text-xs font-medium tracking-[0.18em] uppercase">
                {t("usage.eyebrow")}
              </span>
            </div>
            <h2 className="text-2xl font-semibold tracking-tight">
              {t("usage.title")}
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("usage.description")}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Select
              value={operation}
              onValueChange={(value) => {
                setOperation(value as AIUsageOperation | "all")
                setPage(1)
              }}
            >
              <SelectTrigger className="w-36">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {operations.map((value) => (
                  <SelectItem key={value} value={value}>
                    {t(`usage.operations.${value}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              size="icon"
              onClick={() => void load()}
              title={t("usage.refresh")}
              aria-label={t("usage.refresh")}
            >
              <RefreshCw className={loading ? "animate-spin" : ""} />
            </Button>
          </div>
        </div>

        {summary.length > 0 ? (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {summary.map((item) => (
              <Card key={`${item.operation}-${item.unit}`} size="sm">
                <CardHeader>
                  <CardTitle>
                    {t(`usage.operations.${item.operation}`)}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-xl font-semibold">{formatUsage(item)}</p>
                  <p className="text-xs text-muted-foreground">
                    {t("usage.requests", { count: item.request_count })}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : null}

        <Card>
          <CardContent className="p-0">
            {error ? (
              <p className="p-6 text-sm text-destructive">{error}</p>
            ) : loading && items.length === 0 ? (
              <p className="p-6 text-sm text-muted-foreground">
                {t("common.loading")}
              </p>
            ) : items.length === 0 ? (
              <p className="p-6 text-sm text-muted-foreground">
                {t("usage.empty")}
              </p>
            ) : (
              <div className="divide-y">
                {items.map((item, index) => (
                  <div
                    className="grid gap-2 px-4 py-3 sm:grid-cols-[1fr_1fr_auto_auto] sm:items-center"
                    key={`${item.created_at}-${index}`}
                  >
                    <div>
                      <p className="font-medium">
                        {t(`usage.operations.${item.operation}`)}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {item.model}
                      </p>
                    </div>
                    <p className="text-sm tabular-nums">{formatUsage(item)}</p>
                    <p
                      className={
                        item.status === "succeeded"
                          ? "text-xs text-emerald-600 dark:text-emerald-400"
                          : "text-xs text-destructive"
                      }
                    >
                      {t(`usage.status.${item.status}`)}
                    </p>
                    <time
                      className="text-xs text-muted-foreground sm:text-right"
                      dateTime={item.created_at}
                    >
                      {new Date(item.created_at).toLocaleString()}
                    </time>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {total > 0 ? (
          <Pagination
            page={page}
            pageSize={pageSize}
            totalPages={pageCount}
            total={total}
            onPageChange={setPage}
            onPageSizeChange={(nextPageSize) => {
              setPageSize(nextPageSize)
              setPage(1)
            }}
          />
        ) : null}
      </div>
    </div>
  )
}
