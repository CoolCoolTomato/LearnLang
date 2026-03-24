import * as React from "react"
import { useNavigate } from "react-router-dom"
import { ArrowLeft, ImagePlus, Save, Upload, UserRound } from "lucide-react"
import { toast } from "sonner"
import { useTranslation } from "react-i18next"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import {
  getProfile,
  resolveAvatarUrl,
  updateProfile,
  updateProfileAvatar,
  uploadProfileAvatar,
} from "@/api/profile"
import { getErrorMessage } from "@/lib/error"
import { useAuth } from "@/contexts/auth-context"

interface ProfileFormData {
  username: string
  avatar_url: string
  email: string
  phone: string
}

export default function ProfilePage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { setUser } = useAuth()

  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [uploading, setUploading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [avatarFile, setAvatarFile] = React.useState<File | null>(null)
  const [localAvatarPreview, setLocalAvatarPreview] = React.useState("")
  const fileInputRef = React.useRef<HTMLInputElement | null>(null)

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

  React.useEffect(() => {
    return () => {
      if (localAvatarPreview) {
        URL.revokeObjectURL(localAvatarPreview)
      }
    }
  }, [localAvatarPreview])

  const handleChange = (patch: Partial<ProfileFormData>) => {
    setFormData((prev) => ({ ...prev, ...patch }))
  }

  const handleAvatarPick = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    if (!file.type.startsWith("image/")) {
      toast.error(t("profile.avatarInvalidType", "Please choose an image file"))
      return
    }

    if (file.size > 5 * 1024 * 1024) {
      toast.error(t("profile.avatarTooLarge", "Image size must be less than 5MB"))
      return
    }

    if (localAvatarPreview) {
      URL.revokeObjectURL(localAvatarPreview)
    }

    setAvatarFile(file)
    setLocalAvatarPreview(URL.createObjectURL(file))
  }

  const handleClearAvatarSelection = () => {
    setAvatarFile(null)
    if (localAvatarPreview) {
      URL.revokeObjectURL(localAvatarPreview)
      setLocalAvatarPreview("")
    }
    if (fileInputRef.current) {
      fileInputRef.current.value = ""
    }
  }

  const formatFileSize = (size: number) => {
    const mb = size / (1024 * 1024)
    if (mb >= 1) return `${mb.toFixed(2)} MB`
    const kb = size / 1024
    return `${Math.max(1, Math.round(kb))} KB`
  }

  const handleUploadAvatar = async () => {
    if (!avatarFile) return

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
      handleClearAvatarSelection()
      toast.success(t("profile.avatarUpdated", "Avatar updated"))
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("userSettings.updateFailed")))
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

  if (loading) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center">
        <div className="text-sm text-muted-foreground">{t("common.loading")}</div>
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

  const avatarSrc = localAvatarPreview || resolveAvatarUrl(formData.avatar_url)

  return (
    <div className="relative min-h-full bg-background">
      <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6 md:px-6 md:py-8">
        <div className="rounded-2xl border border-border/60 bg-background/80 p-5 shadow-sm backdrop-blur">
          <div className="grid gap-4">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => navigate("/chat")}
              className="h-8 w-15 text-muted-foreground"
              aria-label={t("common.back")}
              title={t("common.back")}
            >
              <ArrowLeft className="h-5 w-5" />
              <span className="ml-0.5">{t("common.back")}</span>
            </Button>
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <div className="inline-flex h-12 w-12 items-center justify-center rounded-xl border border-border/60 bg-muted/40">
                  <UserRound className="h-6 w-6 text-muted-foreground" />
                </div>
                <div>
                  <h1 className="text-2xl font-semibold tracking-tight">
                    {t("profile.title", "Profile")}
                  </h1>
                  <p className="text-sm text-muted-foreground">
                    {t("profile.description", "Manage your account information.")}
                  </p>
                </div>
              </div>
            </div>

            <div className="grid gap-5 md:grid-cols-[220px_1fr] mt-1">
              <div className="rounded-xl border border-border/60 bg-background/80 p-4">
                <div className="flex flex-col items-center gap-3">
                  <Avatar className="h-20 w-20">
                    <AvatarImage src={avatarSrc} alt={formData.username || "User avatar"} />
                    <AvatarFallback>
                      <UserRound className="h-6 w-6" />
                    </AvatarFallback>
                  </Avatar>
                  <div className="text-center">
                    <div className="text-sm font-medium">
                      {formData.username || t("profile.title", "Profile")}
                    </div>
                    <div className="mt-1 text-xs text-muted-foreground">
                      {t("profile.avatarTip", "JPG/PNG up to 5MB")}
                    </div>
                  </div>
                </div>
              </div>

              <div className="grid gap-3">
                <div>
                  <div className="text-sm font-medium">
                    {t("profile.avatarTitle", "Profile Avatar")}
                  </div>
                  <div className="mt-1 text-xs text-muted-foreground">
                    {t("profile.avatarDesc", "Upload and save a new avatar to update your account image.")}
                  </div>
                </div>

                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  onChange={handleAvatarPick}
                  className="hidden"
                />

                <div className="flex flex-wrap items-center gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    className="h-10 rounded-xl px-4"
                    onClick={() => fileInputRef.current?.click()}
                    disabled={uploading}
                  >
                    <ImagePlus className="mr-2 h-4 w-4" />
                    {t("profile.selectAvatar", "Select Image")}
                  </Button>

                  <Button
                    type="button"
                    onClick={handleUploadAvatar}
                    disabled={!avatarFile || uploading}
                    className="h-10 rounded-xl px-4"
                  >
                    <Upload className="mr-2 h-4 w-4" />
                    {uploading ? t("common.saving", "Saving...") : t("profile.updateAvatar", "update Avatar")}
                  </Button>

                  {avatarFile ? (
                    <Button
                      type="button"
                      variant="ghost"
                      onClick={handleClearAvatarSelection}
                      className="h-10 rounded-xl px-3"
                      disabled={uploading}
                    >
                      {t("common.cancel", "Cancel")}
                    </Button>
                  ) : null}
                </div>

                {avatarFile ? (
                  <div className="rounded-lg border border-border/60 bg-background/80 px-3 py-2 text-xs text-muted-foreground">
                    <span className="font-medium text-foreground">{avatarFile.name}</span>
                    <span className="ml-2">{formatFileSize(avatarFile.size)}</span>
                  </div>
                ) : (
                  <div className="rounded-lg border border-dashed border-border/60 bg-background/60 px-3 py-3 text-xs text-muted-foreground">
                    {t("profile.avatarEmpty", "No image selected")}
                  </div>
                )}
              </div>
            </div>

            <div className="grid gap-2">
              <label className="text-sm text-muted-foreground">
                {t("profile.username", "Username")}
              </label>
              <Input
                value={formData.username}
                onChange={(e) => handleChange({ username: e.target.value })}
                className="h-11 rounded-xl"
                placeholder="tomato"
              />
            </div>

            <div className="grid gap-2 md:grid-cols-2 md:gap-4">
              <div className="grid gap-2">
                <label className="text-sm text-muted-foreground">
                  {t("profile.email", "Email")}
                </label>
                <Input
                  value={formData.email}
                  onChange={(e) => handleChange({ email: e.target.value })}
                  className="h-11 rounded-xl"
                  placeholder="user@example.com"
                />
              </div>
              <div className="grid gap-2">
                <label className="text-sm text-muted-foreground">
                  {t("profile.phone", "Phone")}
                </label>
                <Input
                  value={formData.phone}
                  onChange={(e) => handleChange({ phone: e.target.value })}
                  className="h-11 rounded-xl"
                  placeholder=""
                />
              </div>
            </div>
          </div>

          <div className="flex items-center justify-end mt-8">
              <Button
                onClick={handleSave}
                disabled={saving}
                className="h-11 rounded-xl px-5"
              >
                <Save className="mr-2 h-4 w-4" />
                {saving ? t("common.saving", "Saving...") : t("common.save")}
              </Button>
            </div>
        </div>
      </div>
    </div>
  )
}
