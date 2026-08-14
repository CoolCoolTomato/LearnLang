import * as React from "react"
import { ArrowRight, ChevronLeft, ChevronRight } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

const PAGE_SIZE_OPTIONS = [5, 10, 20, 50]

interface PaginationProps {
  page: number
  pageSize: number
  totalPages: number
  total: number
  onPageChange: (page: number) => void
  onPageSizeChange: (pageSize: number) => void
}

export function Pagination({
  page,
  pageSize,
  totalPages,
  total,
  onPageChange,
  onPageSizeChange,
}: PaginationProps) {
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
    <div className="flex flex-col gap-3 border-t border-border/70 py-4 sm:flex-row sm:items-center sm:justify-between">
      <span className="text-xs text-muted-foreground sm:shrink-0">
        {t("vocabulary.pagination", { page, totalPages, total })}
      </span>
      <div className="flex min-w-0 flex-wrap items-center justify-between gap-2 sm:justify-end">
        <div className="flex items-center gap-1.5">
          <span className="text-xs text-muted-foreground max-[380px]:sr-only">
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
