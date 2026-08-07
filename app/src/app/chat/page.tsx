"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { toast } from "sonner"
import { useTranslation } from "react-i18next"
import { MessageList } from "./components/message-list"
import { MessageInput } from "./components/message-input"
import { getChatHistory, sendChatMessage } from "@/api/chat"
import type {
  AgentErrorEvent,
  ChatMessage,
  ChatWebSocketEvent,
} from "@/types/chat"
import { API_CONFIG, TOKEN_KEY } from "@/api/config"
import { getErrorMessage } from "@/lib/error"
import { getChatDisplaySettings } from "@/lib/chat-display-settings"
import { useOutletContext } from "react-router-dom"
import type { AppLayoutOutletContext } from "@/components/layout/app-layout"

export default function ChatPage() {
  const { t } = useTranslation()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [sending, setSending] = useState(false)
  const [isResponding, setIsResponding] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [wsConnected, setWsConnected] = useState(false)
  const [chatDisplaySettings] = useState(getChatDisplaySettings)
  const { setChatConnected, setChatResponding } =
    useOutletContext<AppLayoutOutletContext>()

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const isManualCloseRef = useRef(false)

  const connectWebSocket = useCallback(() => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) return

    const wsUrl =
      API_CONFIG.BASE_URL.replace("http", "ws") +
      API_CONFIG.API_PREFIX +
      `/ws/chat?token=${token}`

    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      reconnectAttemptsRef.current = 0
      setWsConnected(true)
      console.log("WebSocket connected")
    }

    ws.onmessage = (event) => {
      try {
        const message: ChatWebSocketEvent = JSON.parse(event.data)
        if ((message as AgentErrorEvent).type === "agent_error") {
          setIsResponding(false)
          toast.error(t("chat.agentFailed"))
          return
        }

        if ("type" in message && message.type === "agent_turn_completed") {
          setIsResponding(false)
          return
        }

        setMessages((prev) => [...prev, message as ChatMessage])
        setIsResponding(false)
      } catch (err) {
        console.error("Failed to parse WebSocket message:", err)
      }
    }

    ws.onerror = (error) => {
      console.error("WebSocket error:", error)
      setWsConnected(false)
    }

    ws.onclose = (event) => {
      console.log("WebSocket closed:", event.code, event.reason)
      setWsConnected(false)

      if (!isManualCloseRef.current && event.code !== 1000) {
        if (reconnectAttemptsRef.current >= 3) {
          toast.error(t("chat.wsConnectionFailed"))
          reconnectAttemptsRef.current = 0
          setWsConnected(false)
          return
        }

        reconnectAttemptsRef.current += 1
        reconnectTimeoutRef.current = setTimeout(() => {
          connectWebSocket()
        }, 2000)
      }
    }

    wsRef.current = ws
  }, [t])

  const loadInitialMessages = useCallback(async () => {
    try {
      setLoading(true)
      const response = await getChatHistory()
      setMessages(response.data || [])
      setHasMore((response.data?.length || 0) === 20)
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("chat.loadMessagesFailed")))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadInitialMessages()
    connectWebSocket()

    return () => {
      isManualCloseRef.current = true
      setWsConnected(false)

      if (wsRef.current) {
        wsRef.current.close()
      }

      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
    }
  }, [connectWebSocket, loadInitialMessages])

  useEffect(() => {
    setChatConnected(wsConnected)
    return () => setChatConnected(null)
  }, [setChatConnected, wsConnected])

  useEffect(() => {
    setChatResponding(isResponding)
    return () => setChatResponding(false)
  }, [isResponding, setChatResponding])

  const loadMoreMessages = async () => {
    if (loadingMore || !hasMore || messages.length === 0) return

    try {
      setLoadingMore(true)
      const oldestMessageId = messages[0].id
      const response = await getChatHistory({ before_id: oldestMessageId })
      setMessages((prev) => [...(response.data || []), ...prev])
      setHasMore((response.data?.length || 0) === 20)
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("chat.loadMessagesFailed")))
    } finally {
      setLoadingMore(false)
    }
  }

  const handleSend = async (message: string, voiceFileID?: number) => {
    try {
      setSending(true)
      const userMessage = await sendChatMessage({
        message,
        voice_file_id: voiceFileID,
      })
      setMessages((prev) => [...prev, userMessage])
      setIsResponding(true)
      return true
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("chat.sendFailed")))
      return false
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="flex h-full flex-col overflow-hidden bg-background">
      <div className="min-h-0 flex-1">
        <MessageList
          messages={messages}
          showVoiceChat={chatDisplaySettings.showVoiceChat}
          showTextChat={chatDisplaySettings.showTextChat}
          loading={loading}
          loadingMore={loadingMore}
          hasMore={hasMore}
          onLoadMore={loadMoreMessages}
        />
      </div>

      <div className="border-t border-border/40 bg-gradient-to-t from-background via-background/95 to-background/70">
        <MessageInput
          onSend={handleSend}
          disabled={sending}
        />
      </div>
    </div>
  )
}
