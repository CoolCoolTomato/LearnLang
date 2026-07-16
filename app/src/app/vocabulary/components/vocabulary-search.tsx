import { Search, X } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

interface VocabularySearchProps {
  value: string
  resultCount: number
  onChange: (value: string) => void
}

export function VocabularySearch({
  value,
  resultCount,
  onChange,
}: VocabularySearchProps) {
  const { t } = useTranslation()
  const hasQuery = value.trim().length > 0

  return (
    <div className="mt-5 flex flex-wrap items-center justify-between gap-2">
      <div className="relative w-full sm:max-w-sm">
        <Search className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="search"
          value={value}
          onChange={(event) => onChange(event.target.value)}
          placeholder={t(
            "vocabulary.searchPlaceholder",
            "Search words and meanings"
          )}
          aria-label={t("vocabulary.search", "Search vocabulary")}
          className="pr-9 pl-8 [&::-webkit-search-cancel-button]:hidden"
        />
        {hasQuery ? (
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="absolute top-1/2 right-0.5 -translate-y-1/2"
            onClick={() => onChange("")}
            title={t("vocabulary.clearSearch", "Clear search")}
          >
            <X />
          </Button>
        ) : null}
      </div>
      {hasQuery ? (
        <span className="text-xs text-muted-foreground">
          {t("vocabulary.searchResults", { count: resultCount })}
        </span>
      ) : null}
    </div>
  )
}
