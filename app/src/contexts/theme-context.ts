import * as React from "react"

type Theme = "dark" | "light" | "system"
export type ThemeColor = "default" | "green" | "purple" | "orange" | "red" | "teal" | "blue"

export type ThemeProviderState = {
  theme: Theme
  themeColor: ThemeColor
  setTheme: (theme: Theme) => void
  setThemeColor: (themeColor: ThemeColor) => void
}

const initialState: ThemeProviderState = {
  theme: "system",
  themeColor: "default",
  setTheme: () => null,
  setThemeColor: () => null,
}

export const ThemeProviderContext = React.createContext<ThemeProviderState>(initialState)
