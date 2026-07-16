import * as React from "react"
import { BookSearch, LoaderCircle } from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { lookupMessageVocabulary } from "@/api/vocabulary"
import { getErrorMessage } from "@/lib/error"
import type { ChatMessage } from "@/types/chat"
import type {
  VocabularyLookupMatch,
  VocabularyLookupResult,
} from "@/types/vocabulary"
import { VocabularyEntryDialog } from "./vocabulary-entry-dialog"

interface MessageVocabularyTextProps {
  message: ChatMessage
}

interface MenuPosition {
  x: number
  y: number
}

export function MessageVocabularyText({ message }: MessageVocabularyTextProps) {
  const { t } = useTranslation()
  const [menu, setMenu] = React.useState<MenuPosition | null>(null)
  const [querying, setQuerying] = React.useState(false)
  const [lookup, setLookup] = React.useState<VocabularyLookupResult | null>(
    null
  )
  const [selectedMatch, setSelectedMatch] =
    React.useState<VocabularyLookupMatch | null>(null)
  const lastTouchRef = React.useRef<{
    time: number
    x: number
    y: number
  } | null>(null)

  React.useEffect(() => {
    setLookup(null)
    setSelectedMatch(null)
  }, [message.id, message.text_content])

  React.useEffect(() => {
    if (!menu) return
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !querying) setMenu(null)
    }
    window.addEventListener("keydown", closeOnEscape)
    return () => window.removeEventListener("keydown", closeOnEscape)
  }, [menu, querying])

  const openMenu = (clientX: number, clientY: number) => {
    const width = 190
    const height = 44
    setMenu({
      x: Math.max(8, Math.min(clientX, window.innerWidth - width - 8)),
      y: Math.max(8, Math.min(clientY, window.innerHeight - height - 8)),
    })
  }

  const handleLookup = async () => {
    try {
      setQuerying(true)
      const result = await lookupMessageVocabulary(message.id)
      setLookup(result)
      setMenu(null)
      if (result.matches.length === 0) {
        toast.info(t("chat.vocabularyNoMatches"))
      }
    } catch (error: unknown) {
      toast.error(getErrorMessage(error, t("chat.vocabularyLookupFailed")))
    } finally {
      setQuerying(false)
    }
  }

  return (
    <>
      <div
        className="touch-manipulation text-sm break-words whitespace-pre-wrap"
        onContextMenu={(event) => {
          event.preventDefault()
          openMenu(event.clientX, event.clientY)
        }}
        onDoubleClick={(event) => {
          event.preventDefault()
          openMenu(event.clientX, event.clientY)
        }}
        onPointerUp={(event) => {
          if (event.pointerType !== "touch") return
          if ((event.target as HTMLElement).closest("button")) return
          const previous = lastTouchRef.current
          const current = {
            time: Date.now(),
            x: event.clientX,
            y: event.clientY,
          }
          lastTouchRef.current = current
          if (
            previous &&
            current.time - previous.time <= 350 &&
            Math.hypot(current.x - previous.x, current.y - previous.y) <= 24
          ) {
            lastTouchRef.current = null
            openMenu(current.x, current.y)
          }
        }}
      >
        <HighlightedMessageText
          text={message.text_content}
          lookup={lookup}
          onSelectMatch={setSelectedMatch}
        />
      </div>

      {menu ? (
        <div
          className="fixed inset-0 z-50"
          onPointerDown={() => !querying && setMenu(null)}
        >
          <div
            role="menu"
            className="fixed w-[190px] rounded-md bg-popover p-1 text-popover-foreground shadow-md ring-1 ring-foreground/10"
            style={{ left: menu.x, top: menu.y }}
            onPointerDown={(event) => event.stopPropagation()}
          >
            <button
              type="button"
              role="menuitem"
              className="flex h-9 w-full items-center gap-2 rounded px-2.5 text-left text-sm hover:bg-muted focus-visible:bg-muted focus-visible:outline-none disabled:opacity-60"
              disabled={querying}
              onClick={() => void handleLookup()}
            >
              {querying ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : (
                <BookSearch className="size-4" />
              )}
              {t("chat.lookupVocabulary")}
            </button>
          </div>
        </div>
      ) : null}

      <VocabularyEntryDialog
        match={selectedMatch}
        onOpenChange={(open) => !open && setSelectedMatch(null)}
      />
    </>
  )
}

function HighlightedMessageText({
  text,
  lookup,
  onSelectMatch,
}: {
  text: string
  lookup: VocabularyLookupResult | null
  onSelectMatch: (match: VocabularyLookupMatch) => void
}) {
  if (!lookup || lookup.text !== text || lookup.matches.length === 0) {
    return text
  }

  const content: React.ReactNode[] = []
  let cursor = 0
  for (const match of lookup.matches) {
    if (match.start > cursor) {
      content.push(text.slice(cursor, match.start))
    }
    content.push(
      <button
        key={`${match.start}:${match.end}`}
        type="button"
        className="inline rounded-sm bg-amber-300/55 px-0.5 text-inherit underline decoration-amber-600/60 underline-offset-2 hover:bg-amber-300/80 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none dark:bg-amber-400/25 dark:decoration-amber-300/70 dark:hover:bg-amber-400/40"
        onClick={() => onSelectMatch(match)}
        onDoubleClick={(event) => event.stopPropagation()}
      >
        {text.slice(match.start, match.end)}
      </button>
    )
    cursor = match.end
  }
  if (cursor < text.length) {
    content.push(text.slice(cursor))
  }
  return content
}
