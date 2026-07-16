import { useCallback, useEffect, useState } from "react"
import {
  BookOpenText,
  LogOut,
  MessageCircle,
  PanelLeft,
  Settings,
  UserRound,
  X,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom"
import { logout } from "@/api/auth"
import { resolveAvatarUrl } from "@/api/profile"
import { Logo } from "@/components/logo"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/contexts/auth-context"
import { cn } from "@/lib/utils"

const SIDEBAR_STORAGE_KEY = "learnlang_sidebar_open"

const navigation = [
  { path: "/chat", label: "navigation.chat", icon: MessageCircle },
  { path: "/profile", label: "profile.title", icon: UserRound },
  { path: "/setting", label: "settings.title", icon: Settings },
]

export interface AppLayoutOutletContext {
  setChatConnected: (connected: boolean | null) => void
}

export function AppLayout() {
  const { t } = useTranslation()
  const location = useLocation()
  const navigate = useNavigate()
  const { clearAuth } = useAuth()
  const [desktopOpen, setDesktopOpen] = useState(
    () => localStorage.getItem(SIDEBAR_STORAGE_KEY) !== "false"
  )
  const [mobileOpen, setMobileOpen] = useState(false)
  const [chatConnected, setChatConnected] = useState<boolean | null>(null)

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

  const handleLogout = async () => {
    try {
      await logout()
    } finally {
      clearAuth()
      navigate("/sign-in")
    }
  }

  const pageTitle =
    location.pathname === "/profile"
      ? t("profile.title", "Profile")
      : location.pathname === "/setting"
        ? t("settings.title", "Settings")
        : "LearnLang"

  return (
    <div className="flex h-dvh w-full overflow-hidden bg-sidebar">
      {mobileOpen ? (
        <button
          type="button"
          className="fixed inset-0 z-40 bg-black/35 animate-in fade-in duration-200 md:hidden"
          onClick={() => setMobileOpen(false)}
          aria-label={t("navigation.close", "Close sidebar")}
        />
      ) : null}

      <AppSidebar
        desktopOpen={desktopOpen}
        mobileOpen={mobileOpen}
        closeMobile={() => setMobileOpen(false)}
        toggleDesktop={toggleDesktop}
        onLogout={handleLogout}
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
              <p className="truncate text-[11px] text-muted-foreground">
                {t("chat.headerSubtitle", "Your AI language partner")}
              </p>
            ) : null}
          </div>
        </header>

        <div className="min-h-0 flex-1 overflow-auto">
          <Outlet context={{ setChatConnected }} />
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
  onLogout: () => void
}

function AppSidebar({
  desktopOpen,
  mobileOpen,
  closeMobile,
  toggleDesktop,
  onLogout,
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
        {expanded ? (
          <div className="flex h-8 items-center px-2 text-xs font-medium text-sidebar-foreground/70">
            {t("navigation.main", "Main navigation")}
          </div>
        ) : null}
        <nav
          className="flex flex-col gap-1"
          aria-label={t("navigation.main", "Main navigation")}
        >
          <SidebarLink
            item={navigation[0]}
            expanded={expanded}
            onNavigate={mobile ? closeMobile : undefined}
          />
          <button
            type="button"
            disabled
            title={
              !expanded ? t("navigation.vocabulary", "Vocabulary") : undefined
            }
            className={cn(
              "flex h-8 w-full items-center overflow-hidden rounded-md p-2 text-left text-sm text-sidebar-foreground/45",
              expanded ? "gap-2" : "justify-center"
            )}
          >
            <BookOpenText className="size-4 shrink-0" />
            {expanded ? (
              <>
                <span className="truncate">
                  {t("navigation.vocabulary", "Vocabulary")}
                </span>
                <span className="ml-auto text-[10px] text-muted-foreground">
                  {t("navigation.comingSoon", "Soon")}
                </span>
              </>
            ) : null}
          </button>
          {navigation.slice(1).map((item) => (
            <SidebarLink
              key={item.path}
              item={item}
              expanded={expanded}
              onNavigate={mobile ? closeMobile : undefined}
            />
          ))}
        </nav>
      </div>

      <div className="border-t border-sidebar-border p-2">
        <div
          className={cn(
            "flex h-12 items-center overflow-hidden rounded-md p-2",
            expanded ? "gap-2" : "justify-center"
          )}
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
        </div>
        <Button
          type="button"
          variant="ghost"
          className={cn(
            "h-8 w-full text-destructive hover:text-destructive",
            expanded ? "justify-start px-2" : "px-0"
          )}
          onClick={onLogout}
          title={!expanded ? t("auth.logout", "Logout") : undefined}
        >
          <LogOut className={cn("size-4", expanded && "mr-2")} />
          {expanded ? t("auth.logout", "Logout") : null}
        </Button>
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

      {mobileOpen ? (
        <aside className="fixed inset-y-0 left-0 z-50 flex h-svh w-72 flex-col bg-sidebar text-sidebar-foreground shadow-xl animate-in slide-in-from-left duration-200 md:hidden">
          {sidebarContent(true, true)}
        </aside>
      ) : null}
    </>
  )
}

function SidebarLink({
  item,
  expanded,
  onNavigate,
}: {
  item: (typeof navigation)[number]
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
