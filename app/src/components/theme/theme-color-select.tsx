"use client"

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useTheme } from "@/hooks/use-theme"
import type { ThemeColor } from "@/contexts/theme-context"
import { useTranslation } from "react-i18next"
import * as React from "react"

type ThemeColorOption = {
  key: ThemeColor
  lightPrimary: string
  darkPrimary: string
}

const THEME_COLORS: ThemeColorOption[] = [
  { key: "default", lightPrimary: "oklch(0.205 0 0)", darkPrimary: "oklch(0.922 0 0)" },
  { key: "green", lightPrimary: "oklch(0.74 0.14 142)", darkPrimary: "oklch(0.72 0.13 142)" },
  { key: "blue", lightPrimary: "oklch(0.72 0.14 248)", darkPrimary: "oklch(0.70 0.13 248)" },
  { key: "purple", lightPrimary: "oklch(0.76 0.17 300)", darkPrimary: "oklch(0.72 0.16 300)" },
  { key: "orange", lightPrimary: "oklch(0.82 0.14 62)", darkPrimary: "oklch(0.78 0.14 60)" },
  { key: "red", lightPrimary: "oklch(0.74 0.18 25)", darkPrimary: "oklch(0.72 0.17 25)" },
  { key: "teal", lightPrimary: "oklch(0.74 0.13 200)", darkPrimary: "oklch(0.72 0.12 200)" },
]

export function ThemeColorSelect() {
  const { t } = useTranslation()
  const { theme, themeColor, setThemeColor } = useTheme()
  const [isDarkMode, setIsDarkMode] = React.useState(false)

  React.useEffect(() => {
    const updateMode = () => {
      if (theme === "dark") {
        setIsDarkMode(true)
      } else if (theme === "light") {
        setIsDarkMode(false)
      } else {
        setIsDarkMode(window.matchMedia("(prefers-color-scheme: dark)").matches)
      }
    }

    updateMode()

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)")
    mediaQuery.addEventListener("change", updateMode)
    return () => mediaQuery.removeEventListener("change", updateMode)
  }, [theme])

  return (
    <Select
      value={themeColor}
      onValueChange={(value: ThemeColor) => {
        setThemeColor(value)
      }}
    >
      <SelectTrigger
        className="h-10 w-30"
        aria-label={t("settings.themeColor")}
        title={t("settings.themeColor")}
      >
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {THEME_COLORS.map((color) => (
          <SelectItem key={color.key} value={color.key}>
            <div className="flex w-full items-center justify-between gap-3">
              <span>{t(`settings.themeColor_${color.key}`)}</span>
              <span
                className="h-3.5 w-3.5 rounded-full border border-border/80"
                style={{
                  backgroundColor: isDarkMode ? color.darkPrimary : color.lightPrimary,
                }}
              />
            </div>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
