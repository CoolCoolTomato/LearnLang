"use client"

import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowRight, LockKeyhole, Mail, Phone, User2 } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { register } from "@/api/auth"
import { useAuth } from "@/contexts/auth-context"
import { Logo } from "@/components/logo"
import { getErrorMessage } from "@/lib/error"

export function RegisterForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { setUser } = useAuth()

  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState("")
  const [formData, setFormData] = useState({
    username: "",
    email: "",
    phone: "",
    password: "",
    confirmPassword: "",
  })

  const handleRegister = async () => {
    if (isLoading) return

    const username = formData.username.trim()
    const email = formData.email.trim()
    const phone = formData.phone.trim()

    setIsLoading(true)
    setError("")

    if (!username || !formData.password || !formData.confirmPassword) {
      setError(t("auth.fillAllFields"))
      setIsLoading(false)
      return
    }

    if (!email && !phone) {
      setError(t("auth.emailOrPhoneRequired"))
      setIsLoading(false)
      return
    }

    if (formData.password.length < 6) {
      setError(t("auth.passwordMinLength"))
      setIsLoading(false)
      return
    }

    if (formData.password !== formData.confirmPassword) {
      setError(t("auth.passwordMismatch"))
      setIsLoading(false)
      return
    }

    try {
      const response = await register({
        username,
        password: formData.password,
        ...(email ? { email } : {}),
        ...(phone ? { phone } : {}),
      })
      setUser(response.user)
      navigate("/chat")
    } catch (err: unknown) {
      setError(getErrorMessage(err, t("auth.registerFailed")))
    } finally {
      setIsLoading(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleRegister()
    }
  }

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <div className="flex flex-col items-center gap-4 text-center lg:hidden">
        <div className="flex h-12 w-12 items-center justify-center rounded-2xl border border-border/60 bg-muted/60 shadow-sm">
          <Logo size={24} />
        </div>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            {t("auth.signUpTitle")}
          </h1>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            {t("auth.signUpDescription")}
          </p>
        </div>
      </div>

      <div className="hidden lg:block">
        <div className="mb-2 text-sm font-medium text-muted-foreground">
          {t("auth.welcomeBack")}
        </div>
        <h1 className="text-3xl font-semibold tracking-tight">
          {t("auth.signUpTitle")}
        </h1>
        <p className="mt-3 text-sm leading-6 text-muted-foreground">
          {t("auth.signUpDescription")}
        </p>
      </div>

      {error && (
        <div className="rounded-2xl border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="grid gap-5">
        <div className="grid gap-2.5">
          <Label htmlFor="username" className="text-sm">
            {t("auth.usernameLabel")}
          </Label>
          <div className="relative">
            <User2 className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="username"
              type="text"
              placeholder={t("auth.usernamePlaceholder")}
              value={formData.username}
              onChange={(e) =>
                setFormData({ ...formData, username: e.target.value })
              }
              onKeyDown={handleKeyDown}
              disabled={isLoading}
              className="h-11 rounded-xl border-border/60 bg-background pl-10 shadow-none transition-[border,box-shadow] focus-visible:ring-2"
            />
          </div>
        </div>

        <div className="grid gap-2.5">
          <Label htmlFor="email" className="text-sm">
            {t("auth.email")}
          </Label>
          <div className="relative">
            <Mail className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="email"
              type="email"
              placeholder={t("auth.emailPlaceholder")}
              value={formData.email}
              onChange={(e) =>
                setFormData({ ...formData, email: e.target.value })
              }
              onKeyDown={handleKeyDown}
              disabled={isLoading}
              className="h-11 rounded-xl border-border/60 bg-background pl-10 shadow-none transition-[border,box-shadow] focus-visible:ring-2"
            />
          </div>
        </div>

        <div className="grid gap-2.5">
          <Label htmlFor="phone" className="text-sm">
            {t("auth.phoneLabel")}
          </Label>
          <div className="relative">
            <Phone className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="phone"
              type="text"
              placeholder={t("auth.phonePlaceholder")}
              value={formData.phone}
              onChange={(e) =>
                setFormData({ ...formData, phone: e.target.value })
              }
              onKeyDown={handleKeyDown}
              disabled={isLoading}
              className="h-11 rounded-xl border-border/60 bg-background pl-10 shadow-none transition-[border,box-shadow] focus-visible:ring-2"
            />
          </div>
        </div>

        <div className="grid gap-2.5">
          <Label htmlFor="password" className="text-sm">
            {t("auth.passwordLabel")}
          </Label>

          <div className="relative">
            <LockKeyhole className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="password"
              type="password"
              placeholder={t("auth.passwordPlaceholder")}
              value={formData.password}
              onChange={(e) =>
                setFormData({ ...formData, password: e.target.value })
              }
              onKeyDown={handleKeyDown}
              disabled={isLoading}
              className="h-11 rounded-xl border-border/60 bg-background pl-10 shadow-none transition-[border,box-shadow] focus-visible:ring-2"
            />
          </div>
        </div>

        <div className="grid gap-2.5">
          <Label htmlFor="confirmPassword" className="text-sm">
            {t("auth.confirmPasswordLabel")}
          </Label>

          <div className="relative">
            <LockKeyhole className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="confirmPassword"
              type="password"
              placeholder={t("auth.confirmPasswordPlaceholder")}
              value={formData.confirmPassword}
              onChange={(e) =>
                setFormData({ ...formData, confirmPassword: e.target.value })
              }
              onKeyDown={handleKeyDown}
              disabled={isLoading}
              className="h-11 rounded-xl border-border/60 bg-background pl-10 shadow-none transition-[border,box-shadow] focus-visible:ring-2"
            />
          </div>
        </div>

        <Button
          type="button"
          className="group h-11 w-full cursor-pointer rounded-xl"
          disabled={isLoading}
          onClick={handleRegister}
        >
          <span>{isLoading ? t("auth.registering") : t("auth.createAccount")}</span>
          <ArrowRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-0.5" />
        </Button>
      </div>

      <div className="flex items-center gap-3 text-xs text-muted-foreground">
        <div className="h-px flex-1 bg-border" />
      </div>

      <div className="text-center text-sm text-muted-foreground">
        {t("auth.alreadyHaveAccount")}{" "}
        <a
          href="/sign-in"
          className="font-medium text-foreground underline underline-offset-4"
        >
          {t("auth.signIn")}
        </a>
      </div>
    </div>
  )
}
