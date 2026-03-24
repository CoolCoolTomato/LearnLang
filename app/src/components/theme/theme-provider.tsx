"use client"

import * as React from "react"
import { ThemeProviderContext, type ThemeColor } from "@/contexts/theme-context"

type Theme = "dark" | "light" | "system"

type ThemeProviderProps = {
  children: React.ReactNode
  defaultTheme?: Theme
  storageKey?: string
  defaultThemeColor?: ThemeColor
  colorStorageKey?: string
}

const THEME_COLOR_CLASSES = [
  "theme-green",
  "theme-purple",
  "theme-orange",
  "theme-red",
  "theme-teal",
  "theme-blue",
]

export function ThemeProvider({
  children,
  defaultTheme = "system",
  storageKey = "vite-ui-theme",
  defaultThemeColor = "default",
  colorStorageKey = "vite-ui-theme-color",
  ...props
}: ThemeProviderProps) {
  const [theme, setTheme] = React.useState<Theme>(
    () => (localStorage.getItem(storageKey) as Theme) || defaultTheme
  )
  const [themeColor, setThemeColor] = React.useState<ThemeColor>(
    () => (localStorage.getItem(colorStorageKey) as ThemeColor) || defaultThemeColor
  )

  React.useEffect(() => {
    const root = window.document.documentElement

    root.classList.remove("light", "dark")

    if (theme === "system") {
      const systemTheme = window.matchMedia("(prefers-color-scheme: dark)")
        .matches
        ? "dark"
        : "light"

      root.classList.add(systemTheme)
      return
    }

    root.classList.add(theme)
  }, [theme])

  React.useEffect(() => {
    const root = window.document.documentElement
    root.classList.remove(...THEME_COLOR_CLASSES)

    if (themeColor !== "default") {
      root.classList.add(`theme-${themeColor}`)
    }
  }, [themeColor])

  const value = {
    theme,
    themeColor,
    setTheme: (theme: Theme) => {
      localStorage.setItem(storageKey, theme)
      setTheme(theme)
    },
    setThemeColor: (nextThemeColor: ThemeColor) => {
      localStorage.setItem(colorStorageKey, nextThemeColor)
      setThemeColor(nextThemeColor)
    },
  }

  return (
    <ThemeProviderContext.Provider {...props} value={value}>
      {children}
    </ThemeProviderContext.Provider>
  )
}
