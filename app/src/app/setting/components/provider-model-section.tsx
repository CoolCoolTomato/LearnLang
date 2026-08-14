import * as React from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ModelCombobox } from "@/components/ui/model-combobox"
import { Field } from "./field"
import type { Model } from "@/types/model"
import { useTranslation } from "react-i18next"

interface ProviderModelSectionProps {
  apiBaseUrl: string
  apiKey: string
  model: string

  models: Model[]
  loadingModels?: boolean

  apiBaseUrlLabel: string
  apiBaseUrlPlaceholder?: string
  apiKeyLabel: string
  modelLabel: string

  onApiBaseUrlChange: (value: string) => void
  onApiKeyChange: (value: string) => void
  onModelChange: (value: string) => void
  onLoadModels: () => void
  manualModelEntry?: boolean

  extra?: React.ReactNode
}

export function ProviderModelSection({
  apiBaseUrl,
  apiKey,
  model,
  models,
  loadingModels = false,
  apiBaseUrlLabel,
  apiBaseUrlPlaceholder = "https://api.openai.com/v1",
  apiKeyLabel,
  modelLabel,
  onApiBaseUrlChange,
  onApiKeyChange,
  onModelChange,
  onLoadModels,
  manualModelEntry = false,
  extra,
}: ProviderModelSectionProps) {
  const { t } = useTranslation()

  const canPickModel = Boolean(apiBaseUrl && apiKey)

  return (
    <div className="grid gap-4">
      <div className="grid gap-4 md:grid-cols-2">
        <Field label={apiBaseUrlLabel}>
          <Input
            value={apiBaseUrl}
            onChange={(e) => onApiBaseUrlChange(e.target.value)}
            placeholder={apiBaseUrlPlaceholder}
            className="h-10 rounded-xl md:h-11"
          />
        </Field>

        <Field label={apiKeyLabel}>
          <Input
            type="password"
            value={apiKey}
            onChange={(e) => onApiKeyChange(e.target.value)}
            className="h-10 rounded-xl md:h-11"
          />
        </Field>
      </div>

      {canPickModel ? (
        manualModelEntry ? (
          <Field label={modelLabel}>
            <Input
              value={model}
              onChange={(event) => onModelChange(event.target.value)}
              placeholder={t("settings.enterModel")}
              className="h-10 rounded-xl md:h-11"
            />
          </Field>
        ) : (
          <Field label={modelLabel}>
            <div className="flex flex-col gap-2 md:flex-row">
              <ModelCombobox
                value={model}
                onValueChange={onModelChange}
                models={models}
                placeholder={
                  models.length === 0
                    ? t("systemSettings.noModels")
                    : t("settings.selectModel")
                }
                disabled={loadingModels || models.length === 0}
                className="h-10 md:h-11 md:flex-1"
              />
              <Button
                type="button"
                variant="outline"
                onClick={onLoadModels}
                disabled={loadingModels}
                className="h-10 rounded-xl px-4 md:h-11"
              >
                {loadingModels
                  ? t("systemSettings.loadingModels")
                  : t("systemSettings.loadModels")}
              </Button>
            </div>
          </Field>
        )
      ) : null}

      {extra}
    </div>
  )
}
