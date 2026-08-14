import { useTranslation } from "react-i18next"
import { Bot, Brain, Mic, Settings2, SettingsIcon, Volume2 } from "lucide-react"

import { ThemeColorSelect } from "@/components/theme/theme-color-select"
import { ThemeToggle } from "@/components/theme/theme-toggle"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { TabsList, TabsTrigger } from "@/components/ui/tabs"
import type { Language } from "@/types/settings"

interface DesktopSettingsProps {
  language: Language
  languageOptions: Language[]
  onLanguageChange: (language: Language) => void
}

export function DesktopSettings({
  language,
  languageOptions,
  onLanguageChange,
}: DesktopSettingsProps) {
  const { t } = useTranslation()

  return (
    <div className="hidden md:block">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div className="flex items-center gap-3">
          <div className="inline-flex h-12 w-12 items-center justify-center rounded-xl border border-border/60 bg-muted/40">
            <SettingsIcon className="h-6 w-6 text-muted-foreground" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">
              {t("settings.title", "Settings")}
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {t(
                "settings.description",
                "Manage your language preferences, theme, and AI model providers."
              )}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Select
            value={language}
            onValueChange={(value: Language) => onLanguageChange(value)}
          >
            <SelectTrigger className="h-10 w-30">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {languageOptions.map((item) => (
                <SelectItem key={item} value={item}>
                  {item === "zh-CN"
                    ? t("settings.language_zhCN")
                    : t("settings.language_enUS")}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <ThemeColorSelect />
          <ThemeToggle />
        </div>
      </div>

      <div className="mt-8 flex h-13 w-full items-center justify-center rounded-3xl bg-muted/50 p-2">
        <TabsList className="flex w-full items-center justify-start! gap-2 overflow-x-auto overflow-y-hidden rounded-none bg-transparent p-0.5 whitespace-nowrap group-data-horizontal/tabs:h-10">
          <TabsTrigger value="general" className="shrink-0 rounded-xl px-4">
            <Settings2 className="mr-2 h-4 w-4" />
            {t("settings.generalTitle", "General")}
          </TabsTrigger>
          <TabsTrigger value="chat" className="shrink-0 rounded-xl px-4">
            <Bot className="mr-2 h-4 w-4" />
            Chat
          </TabsTrigger>
          <TabsTrigger value="embedding" className="shrink-0 rounded-xl px-4">
            <Brain className="mr-2 h-4 w-4" />
            Embedding
          </TabsTrigger>
          <TabsTrigger value="stt" className="shrink-0 rounded-xl px-4">
            <Mic className="mr-2 h-4 w-4" />
            STT
          </TabsTrigger>
          <TabsTrigger value="tts" className="shrink-0 rounded-xl px-4">
            <Volume2 className="mr-2 h-4 w-4" />
            TTS
          </TabsTrigger>
        </TabsList>
      </div>
    </div>
  )
}
