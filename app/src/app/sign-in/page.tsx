import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { MessageSquare, Sparkles, ShieldCheck, Globe2 } from "lucide-react"
import { LoginForm } from "./components/login-form"
import { Logo } from "@/components/logo"
import { useAuth } from "@/contexts/auth-context"
import { ThemeToggle } from "@/components/theme/theme-toggle"

export default function Page() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { isAuthenticated, isLoading } = useAuth()

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      navigate("/chat", { replace: true })
    }
  }, [isAuthenticated, isLoading, navigate])

  if (isLoading) {
    return null
  }

  return (
    <div className="relative min-h-svh overflow-hidden bg-background text-foreground">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(120,120,120,0.12),transparent_35%),radial-gradient(circle_at_bottom_right,rgba(120,120,120,0.08),transparent_30%)]" />
      <div className="absolute inset-0 bg-grid-black/[0.02] dark:bg-grid-white/[0.02]" />

      <header className="relative z-10">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 md:px-6">
          <a
            href="/"
            className="group flex items-center gap-3 transition-opacity hover:opacity-90"
          >
            <div className="flex h-10 w-10 items-center justify-center rounded-2xl border border-border/60 bg-muted/60 shadow-sm backdrop-blur">
              <Logo size={22} />
            </div>
            <div className="flex flex-col leading-none">
              <span className="text-sm font-semibold tracking-tight">LearnLang</span>
              <span className="text-xs text-muted-foreground">
                {t('landing.subtitle')}
              </span>
            </div>
          </a>

          <ThemeToggle />
        </div>
      </header>

      <main className="relative z-10 mx-auto flex min-h-[calc(100svh-4rem)] max-w-7xl items-center px-4 py-8 md:px-6 md:py-12">
        <div className="grid w-full items-center gap-8 lg:grid-cols-[1.1fr_0.9fr] xl:gap-14">
          <section className="hidden lg:block">
            <div className="max-w-xl">
              <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-border/60 bg-background/70 px-3 py-1.5 text-xs text-muted-foreground backdrop-blur">
                <Sparkles className="h-3.5 w-3.5" />
                {t('landing.tagline')}
              </div>

              <h1 className="text-4xl font-semibold leading-tight tracking-tight xl:text-5xl">
                {t('landing.heroTitle')}
                <span className="block text-foreground/80">
                  {t('landing.heroTitleHighlight')}
                </span>
              </h1>

              <p className="mt-5 max-w-lg text-base leading-7 text-muted-foreground xl:text-lg">
                {t('landing.heroDescription')}
              </p>

              <div className="mt-8 grid gap-4 sm:grid-cols-3">
                <div className="rounded-2xl border border-border/60 bg-background/70 p-4 backdrop-blur">
                  <MessageSquare className="mb-3 h-5 w-5 text-foreground/80" />
                  <h3 className="text-sm font-medium">{t('landing.feature1Title')}</h3>
                  <p className="mt-1 text-xs leading-5 text-muted-foreground">
                    {t('landing.feature1Desc')}
                  </p>
                </div>

                <div className="rounded-2xl border border-border/60 bg-background/70 p-4 backdrop-blur">
                  <Globe2 className="mb-3 h-5 w-5 text-foreground/80" />
                  <h3 className="text-sm font-medium">{t('landing.feature2Title')}</h3>
                  <p className="mt-1 text-xs leading-5 text-muted-foreground">
                    {t('landing.feature2Desc')}
                  </p>
                </div>

                <div className="rounded-2xl border border-border/60 bg-background/70 p-4 backdrop-blur">
                  <ShieldCheck className="mb-3 h-5 w-5 text-foreground/80" />
                  <h3 className="text-sm font-medium">{t('landing.feature3Title')}</h3>
                  <p className="mt-1 text-xs leading-5 text-muted-foreground">
                    {t('landing.feature3Desc')}
                  </p>
                </div>
              </div>
            </div>
          </section>

          <section className="mx-auto w-full max-w-md lg:max-w-none">
            <div className="rounded-[calc(1.5rem-0.35rem)] border border-border/50 bg-background/90 p-6 md:p-8">
              <LoginForm />
            </div>
          </section>
        </div>
      </main>
    </div>
  )
}