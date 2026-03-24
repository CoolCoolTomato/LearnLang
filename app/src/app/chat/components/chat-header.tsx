"use client"

import { useState } from "react"
import { LogOut, Menu, Settings, UserRound, X } from "lucide-react"
import { useTranslation } from "react-i18next"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { logout } from "@/api/auth"
import { useAuth } from "@/contexts/auth-context"
import { cn } from "@/lib/utils"

interface ChatHeaderProps {
  title?: string
  subtitle?: string
  connected?: boolean
}

export function ChatHeader({
  title = "LearnLang",
  subtitle,
  connected = false,
}: ChatHeaderProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { clearAuth } = useAuth()
  const [menuOpen, setMenuOpen] = useState(false)

  const handleGoProfile = () => {
    setMenuOpen(false)
    navigate("/profile")
  }

  const handleGoSetting = () => {
    setMenuOpen(false)
    navigate("/setting")
  }

  const handleLogout = async () => {
    try {
      await logout()
    } finally {
      clearAuth()
      setMenuOpen(false)
      navigate("/sign-in")
    }
  }

  return (
    <header className="sticky top-0 z-30 w-full border-b border-border/40 bg-gradient-to-b from-background/95 to-background/80 backdrop-blur">
      <div className="flex h-16 w-full items-start px-3 pt-2.5 md:px-4 md:pt-2">
        <div className="flex min-w-0 items-start gap-2.5">
          <Popover open={menuOpen} onOpenChange={setMenuOpen}>
            <PopoverTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-9 w-9 shrink-0 rounded-xl bg-background/50 text-foreground/80 hover:bg-muted"
                aria-label={t("common.menu", "Menu")}
                title={t("common.menu", "Menu")}
              >
                <Menu className="h-4 w-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-44 p-1.5" align="start" sideOffset={10}>
              <div className="flex flex-col gap-1">
                <Button
                  type="button"
                  variant="ghost"
                  className="h-9 justify-start rounded-lg px-2.5"
                  onClick={handleGoProfile}
                >
                  <UserRound className="mr-2 h-4 w-4" />
                  {t("profile.title", "Profile")}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  className="h-9 justify-start rounded-lg px-2.5"
                  onClick={handleGoSetting}
                >
                  <Settings className="mr-2 h-4 w-4" />
                  {t("settings.title", "Settings")}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  className="h-9 justify-start rounded-lg px-2.5 text-destructive hover:text-destructive"
                  onClick={handleLogout}
                >
                  <LogOut className="mr-2 h-4 w-4" />
                  {t("auth.logout", "Logout")}
                </Button>
              </div>
            </PopoverContent>
          </Popover>

          <div className="min-w-0 pt-0.5">
            <div className="flex items-center gap-1.5">
              <h1 className="truncate text-sm font-semibold tracking-tight">{title}</h1>
              <div
                className={cn(
                  "inline-flex items-center rounded-full px-1.5 py-0.5 text-[10px] font-medium",
                  connected
                    ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                    : "bg-muted/70 text-muted-foreground"
                )}
                title={connected ? t("chat.connected", "在线") : t("chat.disconnected", "离线")}
                aria-label={connected ? t("chat.connected", "在线") : t("chat.disconnected", "离线")}
              >
                {connected ? (
                  <span className="h-2 w-2 rounded-full bg-emerald-500" />
                ) : (
                  <X className="h-3 w-3 text-red-500" />
                )}
              </div>
            </div>

            <p className="truncate text-[11px] text-muted-foreground">
              {subtitle || t("chat.headerSubtitle", "Your AI language partner")}
            </p>
          </div>
        </div>
      </div>
    </header>
  )
}
