import { useCallback, useEffect, useState } from "react"
import {
  BarChart3,
  BookOpenText,
  MessageCircle,
  PanelLeft,
  Settings,
  UserRound,
  X,
  type LucideIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { NavLink, Outlet, useLocation } from "react-router-dom"
import { resolveAvatarUrl } from "@/api/profile"
import { Logo } from "@/components/logo"
import { ThemeToggle } from "@/components/theme/theme-toggle"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/contexts/auth-context"
import { cn } from "@/lib/utils"

const SIDEBAR_STORAGE_KEY = "learnlang_sidebar_open"

interface NavigationItem {
  path: string
  label: string
  icon: LucideIcon
}

const navigationGroups: Array<{
  label: string
  items: NavigationItem[]
}> = [
  {
    label: "navigation.groups.start",
    items: [
      { path: "/chat", label: "navigation.chat", icon: MessageCircle },
      {
        path: "/vocabulary",
        label: "navigation.vocabulary",
        icon: BookOpenText,
      },
    ],
  },
  {
    label: "navigation.groups.data",
    items: [{ path: "/usage", label: "navigation.usage", icon: BarChart3 }],
  },
  {
    label: "navigation.groups.system",
    items: [{ path: "/setting", label: "settings.title", icon: Settings }],
  },
]

export interface AppLayoutOutletContext {
  setChatConnected: (connected: boolean | null) => void
  setChatResponding: (responding: boolean) => void
}

export function AppLayout() {
  const { t } = useTranslation()
  const location = useLocation()
  const [desktopOpen, setDesktopOpen] = useState(
    () => localStorage.getItem(SIDEBAR_STORAGE_KEY) === "true"
  )
  const [mobileOpen, setMobileOpen] = useState(false)
  const [chatConnected, setChatConnected] = useState<boolean | null>(null)
  const [chatResponding, setChatResponding] = useState(false)

  const toggleDesktop = useCallback(() => {
    setDesktopOpen((open) => {
      localStorage.setItem(SIDEBAR_STORAGE_KEY, String(!open))
      return !open
    })
  }, [])

  const toggleSidebar = useCallback(() => {
    if (window.matchMedia("(max-width: 767px)").matches) {
      setMobileOpen((open) => !open)
      return
    }
    toggleDesktop()
  }, [toggleDesktop])

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key.toLowerCase() === "b" && (event.metaKey || event.ctrlKey)) {
        event.preventDefault()
        toggleSidebar()
      }
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [toggleSidebar])

  const pageTitle =
    location.pathname === "/profile"
      ? t("profile.title", "Profile")
      : location.pathname === "/vocabulary"
        ? t("vocabulary.title", "Vocabulary")
        : location.pathname === "/usage"
          ? t("usage.title", "Usage")
          : location.pathname === "/setting"
            ? t("settings.title", "Settings")
            : "LearnLang"

  return (
    <div className="flex h-dvh w-full overflow-hidden bg-sidebar">
      <button
        type="button"
        className={cn(
          "fixed inset-0 z-40 bg-black/35 transition-opacity duration-200 ease-out md:hidden",
          mobileOpen ? "opacity-100" : "pointer-events-none opacity-0"
        )}
        onClick={() => setMobileOpen(false)}
        aria-label={t("navigation.close", "Close sidebar")}
        aria-hidden={!mobileOpen}
        tabIndex={mobileOpen ? 0 : -1}
      />

      <AppSidebar
        desktopOpen={desktopOpen}
        mobileOpen={mobileOpen}
        closeMobile={() => setMobileOpen(false)}
        toggleDesktop={toggleDesktop}
      />

      <main
        className={cn(
          "relative flex min-w-0 flex-1 flex-col overflow-hidden bg-background transition-[margin,border-radius] duration-200 ease-linear",
          "md:m-2 md:ml-0 md:rounded-xl md:shadow-sm",
          !desktopOpen && "md:ml-2"
        )}
      >
        <header className="flex h-14 shrink-0 items-center gap-2 border-b border-border/60 px-4">
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="-ml-1"
            onClick={toggleSidebar}
            aria-label={t("navigation.toggle", "Toggle sidebar")}
            title={t("navigation.toggle", "Toggle sidebar")}
          >
            <PanelLeft />
          </Button>
          <div className="mx-1 h-4 w-px bg-border" />
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <h1 className="truncate text-sm font-semibold">{pageTitle}</h1>
              {location.pathname === "/chat" && chatConnected !== null ? (
                <span
                  className={cn(
                    "inline-flex size-5 items-center justify-center rounded-md",
                    chatConnected
                      ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                      : "bg-muted text-muted-foreground"
                  )}
                  title={
                    chatConnected
                      ? t("chat.connected", "Online")
                      : t("chat.disconnected", "Offline")
                  }
                >
                  {chatConnected ? (
                    <span className="size-2 rounded-full bg-emerald-500" />
                  ) : (
                    <X className="size-3 text-red-500" />
                  )}
                </span>
              ) : null}
            </div>
            {location.pathname === "/chat" ? (
              <p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                {chatResponding ? (
                  <>
                    <span>{t("chat.aiThinking")}</span>
                    <span className="flex gap-0.5" aria-hidden="true">
                      <span className="size-1 animate-bounce rounded-full bg-primary [animation-delay:-0.3s]" />
                      <span className="size-1 animate-bounce rounded-full bg-primary [animation-delay:-0.15s]" />
                      <span className="size-1 animate-bounce rounded-full bg-primary" />
                    </span>
                  </>
                ) : (
                  t("chat.headerSubtitle", "Your AI language partner")
                )}
              </p>
            ) : null}
          </div>
          <div className="ml-auto shrink-0">
            <ThemeToggle variant="ghost" />
          </div>
        </header>

        <div
          className={cn(
            "min-h-0 flex-1",
            location.pathname === "/chat" ||
              location.pathname === "/vocabulary" ||
              location.pathname === "/usage"
              ? "overflow-hidden"
              : "overflow-auto"
          )}
        >
          <Outlet context={{ setChatConnected, setChatResponding }} />
        </div>
      </main>
    </div>
  )
}

interface AppSidebarProps {
  desktopOpen: boolean
  mobileOpen: boolean
  closeMobile: () => void
  toggleDesktop: () => void
}

function AppSidebar({
  desktopOpen,
  mobileOpen,
  closeMobile,
  toggleDesktop,
}: AppSidebarProps) {
  const { t } = useTranslation()
  const { user } = useAuth()

  const sidebarContent = (expanded: boolean, mobile = false) => (
    <>
      <div className="p-2">
        <NavLink
          to="/chat"
          onClick={mobile ? closeMobile : undefined}
          className={cn(
            "flex h-12 items-center overflow-hidden rounded-md p-2 hover:bg-sidebar-accent",
            expanded ? "gap-2" : "justify-center"
          )}
          title={!expanded ? "LearnLang" : undefined}
        >
          <span className="flex size-8 shrink-0 items-center justify-center text-sidebar-primary">
            <Logo size={24} />
          </span>
          {expanded ? (
            <div className="min-w-0 text-left leading-tight">
              <div className="truncate text-sm font-medium">LearnLang</div>
              <div className="truncate text-xs text-muted-foreground">
                {t("chat.headerSubtitle", "Your AI language partner")}
              </div>
            </div>
          ) : null}
        </NavLink>
      </div>

      <div className="min-h-0 flex-1 overflow-x-hidden overflow-y-auto p-2">
        <nav
          className="flex flex-col gap-3"
          aria-label={t("navigation.main", "Main navigation")}
        >
          {navigationGroups.map((group) => (
            <div key={group.label}>
              {expanded ? (
                <div className="flex h-7 items-center px-2 text-xs font-medium text-sidebar-foreground/60">
                  {t(group.label)}
                </div>
              ) : null}
              <div className="flex flex-col gap-1">
                {group.items.map((item) => (
                  <SidebarLink
                    key={item.path}
                    item={item}
                    expanded={expanded}
                    onNavigate={mobile ? closeMobile : undefined}
                  />
                ))}
              </div>
            </div>
          ))}
        </nav>
      </div>

      <div className="p-2">
        <NavLink
          to="/profile"
          onClick={mobile ? closeMobile : undefined}
          className={({ isActive }) =>
            cn(
              "flex h-12 items-center overflow-hidden rounded-md p-2 transition-colors outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring",
              expanded ? "gap-2" : "justify-center",
              isActive
                ? "bg-sidebar-accent text-sidebar-accent-foreground"
                : "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
            )
          }
          title={
            !expanded
              ? user?.username || t("profile.title", "Profile")
              : undefined
          }
        >
          <Avatar className="size-8 shrink-0 rounded-lg">
            <AvatarImage
              src={resolveAvatarUrl(user?.avatar_url || "")}
              alt={user?.username || t("profile.title", "Profile")}
            />
            <AvatarFallback className="rounded-lg">
              <UserRound className="size-4" />
            </AvatarFallback>
          </Avatar>
          {expanded ? (
            <div className="min-w-0 flex-1 text-left leading-tight">
              <div className="truncate text-sm font-medium">
                {user?.username || t("profile.title", "Profile")}
              </div>
              <div className="truncate text-xs text-muted-foreground">
                {user?.email || user?.phone}
              </div>
            </div>
          ) : null}
        </NavLink>
      </div>
    </>
  )

  return (
    <>
      <div
        className={cn(
          "relative hidden h-full shrink-0 transition-[width] duration-200 ease-linear md:block",
          desktopOpen ? "w-64" : "w-0"
        )}
        data-state={desktopOpen ? "expanded" : "collapsed"}
      >
        <aside
          className={cn(
            "fixed inset-y-0 z-30 hidden h-svh w-64 p-2 text-sidebar-foreground transition-[left] duration-200 ease-linear md:flex",
            desktopOpen ? "left-0" : "-left-64"
          )}
        >
          <div className="relative flex h-full w-full flex-col bg-sidebar">
            {sidebarContent(true)}
            <button
              type="button"
              tabIndex={-1}
              className="absolute inset-y-0 -right-4 z-20 hidden w-4 cursor-w-resize after:absolute after:inset-y-0 after:left-1/2 after:w-px hover:after:bg-sidebar-border sm:block"
              onClick={toggleDesktop}
              aria-label={t("navigation.toggle", "Toggle sidebar")}
              title={t("navigation.toggle", "Toggle sidebar")}
            />
          </div>
        </aside>
      </div>

      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-50 flex h-dvh w-72 flex-col bg-sidebar text-sidebar-foreground shadow-xl transition-transform duration-200 ease-out md:hidden",
          mobileOpen
            ? "translate-x-0"
            : "pointer-events-none -translate-x-full"
        )}
      >
        {sidebarContent(true, true)}
      </aside>
    </>
  )
}

function SidebarLink({
  item,
  expanded,
  onNavigate,
}: {
  item: NavigationItem
  expanded: boolean
  onNavigate?: () => void
}) {
  const { t } = useTranslation()
  const Icon = item.icon

  return (
    <NavLink
      to={item.path}
      onClick={onNavigate}
      title={!expanded ? t(item.label) : undefined}
      className={({ isActive }) =>
        cn(
          "flex h-8 items-center overflow-hidden rounded-md p-2 text-left text-sm transition-[width,height,padding] outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring",
          expanded ? "gap-2" : "justify-center",
          isActive
            ? "bg-sidebar-accent font-medium text-sidebar-accent-foreground"
            : "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
        )
      }
    >
      <Icon className="size-4 shrink-0" />
      {expanded ? <span className="truncate">{t(item.label)}</span> : null}
    </NavLink>
  )
}
