import { useState, type ReactNode } from "react"
import { ChevronDown, ChevronUp, Volume2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import type {
  VocabularyLookupEntry,
  VocabularyLookupMatch,
} from "@/types/vocabulary"

const RELATED_PHRASE_PREVIEW_COUNT = 10

interface VocabularyEntryDialogProps {
  match: VocabularyLookupMatch | null
  onOpenChange: (open: boolean) => void
}

export function VocabularyEntryDialog({
  match,
  onOpenChange,
}: VocabularyEntryDialogProps) {
  const { t } = useTranslation()
  const entries = match?.entries ?? []

  return (
    <Dialog open={match !== null} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[calc(100svh-2rem)] grid-rows-[auto_minmax(0,1fr)] overflow-hidden sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{match?.text}</DialogTitle>
          <DialogDescription>
            {t("chat.vocabularyDetails", { count: entries.length })}
          </DialogDescription>
        </DialogHeader>

        {entries.length === 1 ? (
          <ScrollArea className="h-full max-h-[65svh] pr-3">
            <VocabularyEntryContent lookupEntry={entries[0]} />
          </ScrollArea>
        ) : entries.length > 1 ? (
          <Tabs
            defaultValue={String(entries[0].entry.id)}
            className="min-h-0 overflow-hidden"
          >
            <ScrollArea scrollbars="horizontal" className="h-9 w-full">
              <TabsList className="w-max">
                {entries.map((item) => (
                  <TabsTrigger
                    key={item.entry.id}
                    value={String(item.entry.id)}
                  >
                    {item.entry.target_text}
                  </TabsTrigger>
                ))}
              </TabsList>
            </ScrollArea>
            {entries.map((item) => (
              <TabsContent
                key={item.entry.id}
                value={String(item.entry.id)}
                className="min-h-0 overflow-hidden"
              >
                <ScrollArea className="h-full max-h-[58svh] pr-3">
                  <VocabularyEntryContent lookupEntry={item} />
                </ScrollArea>
              </TabsContent>
            ))}
          </Tabs>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}

function VocabularyEntryContent({
  lookupEntry,
}: {
  lookupEntry: VocabularyLookupEntry
}) {
  const { t } = useTranslation()
  const { entry } = lookupEntry
  const pronunciations = entry.pronunciations ?? []
  const meanings = entry.meanings ?? []
  const examples = entry.examples ?? []
  const relations = entry.relations ?? []
  const tags = entry.tags ?? []
  const [showAllRelations, setShowAllRelations] = useState(false)
  const visibleRelations = showAllRelations
    ? relations
    : relations.slice(0, RELATED_PHRASE_PREVIEW_COUNT)

  const playAudio = (audioURL: string) => {
    void new Audio(audioURL).play().catch(() => {
      toast.error(t("vocabulary.audioFailed"))
    })
  }

  return (
    <div className="space-y-5 pb-1">
      <div>
        <div className="flex flex-wrap items-center gap-2">
          <h3 className="text-lg font-semibold">{entry.target_text}</h3>
          <span className="rounded border border-border bg-muted/50 px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground uppercase">
            {entry.entry_type === "phrase"
              ? t("vocabulary.phrase")
              : t("chat.word")}
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          {lookupEntry.vocabulary_name}
        </p>
      </div>

      {pronunciations.length > 0 ? (
        <DetailSection title={t("chat.pronunciation")}>
          <div className="flex flex-wrap gap-x-4 gap-y-2">
            {pronunciations.map((item) => (
              <div key={item.id} className="inline-flex items-center gap-1.5">
                {item.region ? (
                  <span className="text-xs font-medium uppercase">
                    {item.region}
                  </span>
                ) : null}
                <span>/{item.pronunciation}/</span>
                {item.audio_url ? (
                  <button
                    type="button"
                    className="inline-flex size-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
                    onClick={() => playAudio(item.audio_url!)}
                    title={t("vocabulary.playAudio")}
                  >
                    <Volume2 className="size-3.5" />
                  </button>
                ) : null}
              </div>
            ))}
          </div>
        </DetailSection>
      ) : null}

      <DetailSection title={t("vocabulary.meaning")}>
        <div className="space-y-1.5">
          {meanings.map((meaning) => (
            <div key={meaning.id} className="flex gap-2">
              {meaning.part_of_speech ? (
                <span className="shrink-0 font-mono text-xs text-muted-foreground">
                  {meaning.part_of_speech}
                </span>
              ) : null}
              <span>{meaning.native_text}</span>
            </div>
          ))}
        </div>
      </DetailSection>

      {relations.length > 0 ? (
        <DetailSection title={t("vocabulary.relatedPhrases")}>
          <div className="space-y-2.5">
            {showAllRelations ? (
              <ScrollArea className="h-52 overscroll-contain pr-3">
                <RelatedPhraseList relations={visibleRelations} />
              </ScrollArea>
            ) : (
              <RelatedPhraseList relations={visibleRelations} />
            )}
            {relations.length > RELATED_PHRASE_PREVIEW_COUNT ? (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="px-0 text-muted-foreground"
                onClick={() => setShowAllRelations((current) => !current)}
              >
                {showAllRelations ? <ChevronUp /> : <ChevronDown />}
                {showAllRelations
                  ? t("chat.collapseRelatedPhrases")
                  : t("chat.showAllRelatedPhrases", {
                      count: relations.length,
                    })}
              </Button>
            ) : null}
          </div>
        </DetailSection>
      ) : null}

      {examples.length > 0 ? (
        <DetailSection title={t("vocabulary.examples")}>
          <div className="space-y-3">
            {examples.map((example) => (
              <div key={example.id}>
                <p>{example.target_text}</p>
                {example.native_text ? (
                  <p className="mt-0.5 text-muted-foreground">
                    {example.native_text}
                  </p>
                ) : null}
              </div>
            ))}
          </div>
        </DetailSection>
      ) : null}

      {tags.length > 0 ? (
        <DetailSection title={t("vocabulary.tags")}>
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
        <DetailSection title={t("vocabulary.notes")}>
          <p className="whitespace-pre-wrap text-muted-foreground">
            {entry.notes}
          </p>
        </DetailSection>
      ) : null}
    </div>
  )
}

function RelatedPhraseList({
  relations,
}: {
  relations: VocabularyLookupEntry["entry"]["relations"]
}) {
  return (
    <div className="space-y-2">
      {(relations ?? []).map((relation) => (
        <div key={relation.id} className="min-w-0 break-words">
          <span className="font-medium">
            {relation.related_entry?.target_text}
          </span>
          {relation.related_entry?.meanings?.[0]?.native_text ? (
            <span className="ml-2 text-muted-foreground">
              {relation.related_entry.meanings[0].native_text}
            </span>
          ) : null}
        </div>
      ))}
    </div>
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
    <section>
      <h4 className="mb-2 text-xs font-semibold text-muted-foreground uppercase">
        {title}
      </h4>
      <div className="text-sm">{children}</div>
    </section>
  )
}
