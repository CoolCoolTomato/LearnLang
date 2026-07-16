import type { ReactNode } from "react"
import { ChevronDown, Volume2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import type { VocabularyEntry } from "@/types/vocabulary"

interface VocabularyRowProps {
  entry: VocabularyEntry
  expanded: boolean
  onToggle: () => void
}

export function VocabularyRow({
  entry,
  expanded,
  onToggle,
}: VocabularyRowProps) {
  const { t } = useTranslation()
  const pronunciations = entry.pronunciations ?? []
  const meanings = entry.meanings ?? []
  const examples = entry.examples ?? []
  const relations = entry.relations ?? []
  const tags = entry.tags ?? []
  const hasDetails =
    examples.length > 0 ||
    relations.length > 0 ||
    tags.length > 0 ||
    Boolean(entry.notes)

  const playAudio = (audioURL: string) => {
    void new Audio(audioURL).play().catch(() => {
      toast.error(t("vocabulary.audioFailed", "Unable to play audio"))
    })
  }

  return (
    <article className="bg-background">
      <div className="relative grid min-h-20 grid-cols-[minmax(0,1fr)_36px] gap-3 px-4 py-3 md:grid-cols-[minmax(180px,1.2fr)_minmax(220px,1.6fr)_120px_40px] md:items-center md:gap-4">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-base font-semibold break-words">
              {entry.target_text}
            </span>
            {entry.entry_type === "phrase" ? (
              <span className="rounded border border-border bg-muted/50 px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground uppercase">
                {t("vocabulary.phrase", "Phrase")}
              </span>
            ) : null}
          </div>
          {pronunciations.length > 0 ? (
            <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
              {pronunciations.map((item) => (
                <span key={item.id} className="inline-flex items-center gap-1">
                  {item.region ? (
                    <span className="font-medium uppercase">{item.region}</span>
                  ) : null}
                  <span>/{item.pronunciation}/</span>
                  {item.audio_url ? (
                    <button
                      type="button"
                      className="inline-flex size-5 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
                      onClick={() => playAudio(item.audio_url!)}
                      title={t("vocabulary.playAudio", "Play audio")}
                      aria-label={t("vocabulary.playAudio", "Play audio")}
                    >
                      <Volume2 className="size-3" />
                    </button>
                  ) : null}
                </span>
              ))}
            </div>
          ) : null}
        </div>

        <div className="col-span-2 min-w-0 md:col-span-1">
          {meanings.length > 0 ? (
            <div className="space-y-1">
              {meanings.map((meaning) => (
                <div key={meaning.id} className="flex min-w-0 gap-2 text-sm">
                  {meaning.part_of_speech ? (
                    <span className="shrink-0 font-mono text-xs text-muted-foreground">
                      {meaning.part_of_speech}
                    </span>
                  ) : null}
                  <span className="break-words">{meaning.native_text}</span>
                </div>
              ))}
            </div>
          ) : (
            <span className="text-sm text-muted-foreground">—</span>
          )}
        </div>

        <div className="col-span-2 flex items-center gap-2 md:col-span-1">
          <span
            className={cn(
              "inline-flex rounded border px-1.5 py-0.5 text-xs font-medium",
              entry.encountered
                ? "border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400"
                : "border-border bg-muted/40 text-muted-foreground"
            )}
          >
            {entry.encountered
              ? t("vocabulary.encountered", "Encountered")
              : t("vocabulary.notEncountered", "Not encountered")}
          </span>
          {entry.encounter_count > 0 ? (
            <span className="text-xs text-muted-foreground">
              ×{entry.encounter_count}
            </span>
          ) : null}
        </div>

        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="absolute right-3 mt-0 md:static"
          onClick={onToggle}
          disabled={!hasDetails}
          aria-expanded={expanded}
          title={t("vocabulary.details", "Details")}
        >
          <ChevronDown
            className={cn("transition-transform", expanded && "rotate-180")}
          />
        </Button>
      </div>

      {expanded && hasDetails ? (
        <div className="grid gap-5 border-t border-dashed border-border/70 bg-muted/20 px-4 py-4 md:grid-cols-2">
          {relations.length > 0 ? (
            <DetailSection title={t("vocabulary.relatedPhrases")}>
              <div className="flex flex-wrap gap-2">
                {relations.map((relation) => (
                  <span
                    key={relation.id}
                    className="inline-flex max-w-full items-baseline gap-1.5 rounded border border-border bg-background px-2 py-1 text-sm"
                  >
                    <span className="font-medium break-words">
                      {relation.related_entry?.target_text}
                    </span>
                    <span className="break-words text-muted-foreground">
                      {relation.related_entry?.meanings?.[0]?.native_text}
                    </span>
                  </span>
                ))}
              </div>
            </DetailSection>
          ) : null}

          {examples.length > 0 ? (
            <DetailSection title={t("vocabulary.examples", "Examples")}>
              <div className="space-y-3">
                {examples.map((example) => (
                  <div key={example.id} className="text-sm">
                    <p className="break-words text-foreground">
                      {example.target_text}
                    </p>
                    {example.native_text ? (
                      <p className="mt-0.5 break-words text-muted-foreground">
                        {example.native_text}
                      </p>
                    ) : null}
                  </div>
                ))}
              </div>
            </DetailSection>
          ) : null}

          {tags.length > 0 ? (
            <DetailSection title={t("vocabulary.tags", "Tags")}>
              <div className="flex flex-wrap gap-1.5">
                {tags.map((tag) => (
                  <span
                    key={tag}
                    className="rounded bg-muted px-2 py-1 text-xs text-muted-foreground"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </DetailSection>
          ) : null}

          {entry.notes ? (
            <DetailSection title={t("vocabulary.notes", "Notes")}>
              <p className="text-sm break-words whitespace-pre-wrap text-muted-foreground">
                {entry.notes}
              </p>
            </DetailSection>
          ) : null}
        </div>
      ) : null}
    </article>
  )
}

function DetailSection({
  title,
  children,
}: {
  title: string
  children: ReactNode
}) {
  return (
    <section className="min-w-0">
      <h3 className="mb-2 text-xs font-semibold text-muted-foreground uppercase">
        {title}
      </h3>
      {children}
    </section>
  )
}
