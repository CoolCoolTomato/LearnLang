import { ArrowLeft, ChevronRight, type LucideIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { useTranslation } from "react-i18next"

export type MobileSettingsView = "list" | "detail"

export interface MobileSettingItem {
  id: string
  title: string
  description?: string
  icon: LucideIcon
}

interface MobileSettingsProps {
  items: MobileSettingItem[]
  view: MobileSettingsView
  onViewChange: (view: MobileSettingsView) => void
  onSettingSelect: (id: string) => void
}

export function MobileSettings({
  items,
  view,
  onViewChange,
  onSettingSelect,
}: MobileSettingsProps) {
  const { t } = useTranslation()

  return (
    <>
      <div
        className={
          view === "list"
            ? "border-b border-border/60 md:hidden"
            : "hidden md:hidden"
        }
      >
        {items.map(({ id, title, description, icon: Icon }) => (
          <button
            key={id}
            type="button"
            className="flex min-h-18 w-full items-center gap-3 border-b border-border/60 px-4 py-3 text-left transition-colors last:border-b-0 hover:bg-muted/40 active:bg-muted/60"
            onClick={() => {
              onSettingSelect(id)
              onViewChange("detail")
            }}
          >
            <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
              <Icon className="size-4" />
            </span>
            <span className="min-w-0 flex-1">
              <span className="block text-sm font-medium">{title}</span>
              {description ? (
                <span className="mt-0.5 block truncate text-xs text-muted-foreground">
                  {description}
                </span>
              ) : null}
            </span>
            <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
          </button>
        ))}
      </div>

      <div
        className={
          view === "detail"
            ? "flex items-center gap-2 border-b border-border/60 px-4 py-3 md:hidden"
            : "hidden md:hidden"
        }
      >
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="-ml-2"
          onClick={() => onViewChange("list")}
          aria-label={t("common.back", "Back")}
          title={t("common.back", "Back")}
        >
          <ArrowLeft />
        </Button>
      </div>
    </>
  )
}
