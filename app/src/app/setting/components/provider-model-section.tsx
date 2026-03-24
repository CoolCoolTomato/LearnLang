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
  apiKeyLabel: string
  modelLabel: string

  onApiBaseUrlChange: (value: string) => void
  onApiKeyChange: (value: string) => void
  onModelChange: (value: string) => void
  onLoadModels: () => void

  extra?: React.ReactNode
}

export function ProviderModelSection({
  apiBaseUrl,
  apiKey,
  model,
  models,
  loadingModels = false,
  apiBaseUrlLabel,
  apiKeyLabel,
  modelLabel,
  onApiBaseUrlChange,
  onApiKeyChange,
  onModelChange,
  onLoadModels,
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
            placeholder="https://api.openai.com/v1"
            className="h-11 rounded-xl"
          />
        </Field>

        <Field label={apiKeyLabel}>
          <Input
            type="password"
            value={apiKey}
            onChange={(e) => onApiKeyChange(e.target.value)}
            className="h-11 rounded-xl"
          />
        </Field>
      </div>

      {canPickModel ? (
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
              className="flex-1 h-11"
            />
            <Button
              type="button"
              variant="outline"
              onClick={onLoadModels}
              disabled={loadingModels}
              className="h-11 rounded-xl px-4"
            >
              {loadingModels
                ? t("systemSettings.loadingModels")
                : t("systemSettings.loadModels")}
            </Button>
          </div>
        </Field>
      ) : null}

      {extra}
    </div>
  )
}