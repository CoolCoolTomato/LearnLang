import * as React from "react"
import { KeyRound, LogOut, Pencil, Save, UserRound } from "lucide-react"
import { toast } from "sonner"
import { useTranslation } from "react-i18next"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  getProfile,
  resolveAvatarUrl,
  updateProfile,
  updateProfileAvatar,
  uploadProfileAvatar,
} from "@/api/profile"
import { getErrorMessage } from "@/lib/error"
import { useAuth } from "@/contexts/auth-context"
import { changePassword, logout } from "@/api/auth"
import { AvatarUploadDialog } from "./components/avatar-upload-dialog"

interface ProfileFormData {
  username: string
  avatar_url: string
  email: string
  phone: string
}

interface PasswordFormData {
  newPassword: string
  confirmPassword: string
}

const emptyPasswordForm: PasswordFormData = {
  newPassword: "",
  confirmPassword: "",
}

export default function ProfilePage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { clearAuth, setUser } = useAuth()

  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [uploading, setUploading] = React.useState(false)
  const [avatarDialogOpen, setAvatarDialogOpen] = React.useState(false)
  const [changingPassword, setChangingPassword] = React.useState(false)
  const [loggingOut, setLoggingOut] = React.useState(false)
  const [passwordDialogOpen, setPasswordDialogOpen] = React.useState(false)
  const [passwordError, setPasswordError] = React.useState("")
  const [passwordForm, setPasswordForm] =
    React.useState<PasswordFormData>(emptyPasswordForm)
  const [error, setError] = React.useState<string | null>(null)

  const [formData, setFormData] = React.useState<ProfileFormData>({
    username: "",
    avatar_url: "",
    email: "",
    phone: "",
  })

  React.useEffect(() => {
    const loadProfile = async () => {
      try {
        setLoading(true)
        const profile = await getProfile()
        setFormData({
          username: profile.username || "",
          avatar_url: profile.avatar_url || "",
          email: profile.email || "",
          phone: profile.phone || "",
        })
        setError(null)
      } catch (err: unknown) {
        setError(getErrorMessage(err, t("user.loadFailed")))
      } finally {
        setLoading(false)
      }
    }

    loadProfile()
  }, [t])

  const handleChange = (patch: Partial<ProfileFormData>) => {
    setFormData((prev) => ({ ...prev, ...patch }))
  }

  const handleUploadAvatar = async (avatarFile: File) => {
    try {
      setUploading(true)
      const uploadResult = await uploadProfileAvatar(avatarFile)
      await updateProfileAvatar(uploadResult.filename)
      const profile = await getProfile()

      setFormData((prev) => ({
        ...prev,
        avatar_url: profile.avatar_url || prev.avatar_url,
      }))
      setUser(profile)
      toast.success(t("profile.avatarUpdated", "Avatar updated"))
      return true
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("userSettings.updateFailed")))
      return false
    } finally {
      setUploading(false)
    }
  }

  const handleSave = async () => {
    try {
      setSaving(true)
      await updateProfile({
        username: formData.username.trim(),
        email: formData.email.trim(),
        phone: formData.phone.trim(),
      })
      const updated = await getProfile()
      setUser(updated)
      setFormData((prev) => ({
        ...prev,
        username: updated.username || prev.username,
        email: updated.email || prev.email,
        phone: updated.phone || prev.phone,
        avatar_url: updated.avatar_url || prev.avatar_url,
      }))
      toast.success(t("settings.updateSuccess", "Updated successfully"))
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("userSettings.updateFailed")))
    } finally {
      setSaving(false)
    }
  }

  const handleLogout = async () => {
    if (loggingOut) return
    try {
      setLoggingOut(true)
      await logout()
    } finally {
      clearAuth()
      navigate("/sign-in")
    }
  }

  const resetPasswordForm = () => {
    setPasswordForm(emptyPasswordForm)
    setPasswordError("")
  }

  const handlePasswordDialogChange = (open: boolean) => {
    if (changingPassword) return
    setPasswordDialogOpen(open)
    if (!open) resetPasswordForm()
  }

  const handleChangePassword = async (
    event: React.FormEvent<HTMLFormElement>
  ) => {
    event.preventDefault()
    if (changingPassword) return

    setPasswordError("")

    if (!passwordForm.newPassword || !passwordForm.confirmPassword) {
      setPasswordError(t("auth.fillAllFields"))
      return
    }
    if (passwordForm.newPassword.length < 6) {
      setPasswordError(t("auth.passwordMinLength"))
      return
    }
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      setPasswordError(t("auth.passwordMismatch"))
      return
    }
    try {
      setChangingPassword(true)
      await changePassword({
        new_password: passwordForm.newPassword,
      })
      setPasswordDialogOpen(false)
      resetPasswordForm()
      toast.success(t("profile.passwordChanged"))
    } catch (err: unknown) {
      setPasswordError(getErrorMessage(err, t("profile.passwordChangeFailed")))
    } finally {
      setChangingPassword(false)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center">
        <div className="text-sm text-muted-foreground">
          {t("common.loading")}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="mx-auto mt-6 max-w-3xl rounded-2xl border border-destructive/20 bg-destructive/10 p-4 text-sm text-destructive">
        {error}
      </div>
    )
  }

  const avatarSrc = resolveAvatarUrl(formData.avatar_url)

  return (
    <div className="min-h-full bg-background">
      <div className="mx-auto w-full max-w-4xl py-2 sm:px-4 sm:py-6 md:px-6 md:py-10">
        <main className="overflow-hidden bg-background sm:rounded-2xl sm:border sm:border-border/70 sm:shadow-sm">
          <section className="px-4 py-5 sm:p-7">
            <div className="flex min-w-0 items-center gap-4 sm:gap-5">
              <button
                type="button"
                className="group relative shrink-0 rounded-full outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
                onClick={() => setAvatarDialogOpen(true)}
                aria-label={t("profile.editAvatar")}
                title={t("profile.editAvatar")}
              >
                <Avatar className="size-20 ring-1 ring-border sm:size-24">
                  <AvatarImage
                    src={avatarSrc}
                    alt={formData.username || "User avatar"}
                  />
                  <AvatarFallback>
                    <UserRound className="size-7 text-muted-foreground" />
                  </AvatarFallback>
                </Avatar>
                <span className="pointer-events-none absolute inset-0 flex items-center justify-center rounded-full bg-black/45 text-white opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100">
                  <Pencil className="size-5" />
                </span>
              </button>

              <div className="min-w-0 flex-1">
                <h2 className="truncate text-xl font-semibold sm:text-2xl">
                  {formData.username || t("profile.title", "Profile")}
                </h2>
                <p className="mt-1 truncate text-sm text-muted-foreground">
                  {formData.email || t("profile.emailNotSet")}
                </p>
              </div>

              <Button
                type="button"
                variant="outline"
                size="icon-lg"
                className="shrink-0 text-destructive hover:bg-destructive/10 hover:text-destructive sm:w-auto sm:px-4"
                onClick={() => void handleLogout()}
                disabled={loggingOut}
                aria-label={t("auth.logout", "Logout")}
                title={t("auth.logout", "Logout")}
              >
                <LogOut />
                <span className="hidden sm:inline">
                  {t("auth.logout", "Logout")}
                </span>
              </Button>
            </div>
          </section>

          <section className="border-t border-border/70 px-4 py-5 sm:p-7">
            <div className="grid gap-5">
              <div className="grid gap-2">
                <label
                  htmlFor="profile-username"
                  className="text-sm font-medium"
                >
                  {t("profile.username", "Username")}
                </label>
                <Input
                  id="profile-username"
                  value={formData.username}
                  onChange={(e) => handleChange({ username: e.target.value })}
                  className="h-10"
                  placeholder="tomato"
                />
              </div>

              <div className="grid gap-5 md:grid-cols-2 md:gap-4">
                <div className="grid gap-2">
                  <label
                    htmlFor="profile-email"
                    className="text-sm font-medium"
                  >
                    {t("profile.email", "Email")}
                  </label>
                  <Input
                    id="profile-email"
                    value={formData.email}
                    onChange={(e) => handleChange({ email: e.target.value })}
                    className="h-10"
                    placeholder="user@example.com"
                  />
                </div>

                <div className="grid gap-2">
                  <label
                    htmlFor="profile-phone"
                    className="text-sm font-medium"
                  >
                    {t("profile.phone", "Phone")}
                  </label>
                  <Input
                    id="profile-phone"
                    value={formData.phone}
                    onChange={(e) => handleChange({ phone: e.target.value })}
                    className="h-10"
                    placeholder=""
                  />
                </div>
              </div>
              <div className="flex justify-end">
                <Button
                  onClick={handleSave}
                  disabled={saving}
                  className="h-10 px-5"
                >
                  <Save className="mr-2 h-4 w-4" />
                  {saving ? t("common.saving", "Saving...") : t("common.save")}
                </Button>
              </div>
            </div>
          </section>

          <section className="border-t border-border/70 px-4 py-5 sm:p-7">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex min-w-0 items-start gap-3">
                <div className="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-muted">
                  <KeyRound className="h-4 w-4 text-muted-foreground" />
                </div>
                <div className="min-w-0">
                  <div className="text-sm font-semibold">
                    {t("profile.accountSecurity")}
                  </div>
                  <div className="mt-1 text-sm leading-6 text-muted-foreground">
                    {t("profile.passwordDescription")}
                  </div>
                </div>
              </div>

              <Button
                type="button"
                variant="outline"
                className="h-10 shrink-0 px-4"
                onClick={() => setPasswordDialogOpen(true)}
              >
                <KeyRound className="mr-2 h-4 w-4" />
                {t("profile.changePassword")}
              </Button>
            </div>
          </section>
        </main>
      </div>

      <AvatarUploadDialog
        open={avatarDialogOpen}
        uploading={uploading}
        onOpenChange={setAvatarDialogOpen}
        onUpload={handleUploadAvatar}
      />

      <Dialog
        open={passwordDialogOpen}
        onOpenChange={handlePasswordDialogChange}
      >
        <DialogContent>
          <form onSubmit={handleChangePassword} className="grid gap-4">
            <DialogHeader>
              <DialogTitle>{t("profile.changePassword")}</DialogTitle>
              <DialogDescription>
                {t("profile.passwordDialogDescription")}
              </DialogDescription>
            </DialogHeader>

            {passwordError ? (
              <div className="rounded-lg border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {passwordError}
              </div>
            ) : null}

            <div className="grid gap-2">
              <label htmlFor="new-password" className="text-sm font-medium">
                {t("profile.newPassword")}
              </label>
              <Input
                id="new-password"
                type="password"
                autoComplete="new-password"
                value={passwordForm.newPassword}
                onChange={(event) =>
                  setPasswordForm((prev) => ({
                    ...prev,
                    newPassword: event.target.value,
                  }))
                }
                disabled={changingPassword}
                autoFocus
              />
            </div>

            <div className="grid gap-2">
              <label
                htmlFor="confirm-new-password"
                className="text-sm font-medium"
              >
                {t("profile.confirmNewPassword")}
              </label>
              <Input
                id="confirm-new-password"
                type="password"
                autoComplete="new-password"
                value={passwordForm.confirmPassword}
                onChange={(event) =>
                  setPasswordForm((prev) => ({
                    ...prev,
                    confirmPassword: event.target.value,
                  }))
                }
                disabled={changingPassword}
              />
            </div>

            <DialogFooter>
              <DialogClose asChild>
                <Button
                  type="button"
                  variant="outline"
                  disabled={changingPassword}
                >
                  {t("common.cancel")}
                </Button>
              </DialogClose>
              <Button type="submit" disabled={changingPassword}>
                {changingPassword
                  ? t("profile.changingPassword")
                  : t("profile.confirmChange")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
