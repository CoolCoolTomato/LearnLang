import * as React from "react"
import { BarChart3, RefreshCw } from "lucide-react"
import { useTranslation } from "react-i18next"
import { listAIUsage, summarizeAIUsage } from "@/api/ai-usage"
import { Button } from "@/components/ui/button"
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
  const [pageSize, setPageSize] = React.useState(5)
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
  }, [operation, page, pageSize, t])

  React.useEffect(() => {
    void load()
  }, [load])
  const pageCount = Math.max(1, Math.ceil(total / pageSize))

  return (
    <ScrollArea className="h-full">
      <div className="mx-auto flex w-full max-w-6xl flex-col px-4 py-5 md:px-6 md:py-7">
        <section className="flex flex-col gap-4 border-b border-border/70 pb-5 sm:flex-row sm:items-end sm:justify-between">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <BarChart3 className="size-5 text-muted-foreground" />
              <h2 className="text-xl font-semibold">{t("usage.title")}</h2>
            </div>
            <p className="mt-2 text-sm text-muted-foreground">
              {t("usage.description")}
            </p>
          </div>
          <div className="flex shrink-0 items-center gap-1.5">
            <Select
              value={operation}
              onValueChange={(value) => {
                setOperation(value as AIUsageOperation | "all")
                setPage(1)
              }}
            >
              <SelectTrigger className="w-40">
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
        </section>

        {summary.length > 0 ? (
          <div className="mt-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {summary.map((item) => (
              <div
                className="rounded-lg border border-border/70 bg-muted/15 px-4 py-3"
                key={`${item.operation}-${item.unit}`}
              >
                <p className="text-xs font-medium text-muted-foreground">
                  {t(`usage.operations.${item.operation}`)}
                </p>
                <div className="mt-1 flex items-baseline justify-between gap-3">
                  <p className="text-lg font-semibold tabular-nums">
                    {formatUsage(item)}
                  </p>
                  <p className="shrink-0 text-xs text-muted-foreground">
                    {t("usage.requests", { count: item.request_count })}
                  </p>
                </div>
              </div>
            ))}
          </div>
        ) : null}

        {error ? (
          <p className="flex min-h-[40vh] items-center justify-center px-4 text-sm text-destructive">
            {error}
          </p>
        ) : loading && items.length === 0 ? (
          <p className="flex min-h-[40vh] items-center justify-center px-4 text-sm text-muted-foreground">
            {t("common.loading")}
          </p>
        ) : items.length === 0 ? (
          <p className="flex min-h-[40vh] items-center justify-center px-4 text-sm text-muted-foreground">
            {t("usage.empty")}
          </p>
        ) : (
          <>
            <div className="mt-5 overflow-hidden rounded-lg border border-border/70">
              <div className="hidden h-9 grid-cols-[minmax(140px,1fr)_minmax(160px,1.2fr)_130px_100px_170px] items-center gap-4 border-b bg-muted/40 px-4 text-xs font-medium text-muted-foreground md:grid">
                <span>{t("usage.columns.operation")}</span>
                <span>{t("usage.columns.model")}</span>
                <span>{t("usage.columns.usage")}</span>
                <span>{t("usage.columns.status")}</span>
                <span className="text-right">{t("usage.columns.time")}</span>
              </div>
              <div className="divide-y divide-border/70">
                {items.map((item, index) => (
                  <article
                    className="grid min-h-20 grid-cols-[minmax(0,1fr)_auto] items-center gap-x-3 gap-y-2 bg-background px-4 py-3 md:grid-cols-[minmax(140px,1fr)_minmax(160px,1.2fr)_130px_100px_170px] md:gap-4"
                    key={`${item.created_at}-${index}`}
                  >
                    <p className="min-w-0 text-base font-semibold break-words">
                      {t(`usage.operations.${item.operation}`)}
                    </p>
                    <p className="col-span-2 min-w-0 truncate text-sm text-muted-foreground md:col-span-1 md:text-foreground">
                      {item.model}
                    </p>
                    <p className="text-sm font-medium tabular-nums">
                      {formatUsage(item)}
                    </p>
                    <span
                      className={cn(
                        "w-fit rounded border px-1.5 py-0.5 text-xs font-medium",
                        item.status === "succeeded"
                          ? "border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400"
                          : "border-destructive/25 bg-destructive/10 text-destructive"
                      )}
                    >
                      {t(`usage.status.${item.status}`)}
                    </span>
                    <time
                      className="col-span-2 text-xs text-muted-foreground md:col-span-1 md:text-right"
                      dateTime={item.created_at}
                    >
                      {new Date(item.created_at).toLocaleString()}
                    </time>
                  </article>
                ))}
              </div>
            </div>
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
          </>
        )}
      </div>
    </ScrollArea>
  )
}
