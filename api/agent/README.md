# Agent Package

This package owns LearnLang's backend chat-agent workflow. It is transport
independent: HTTP controllers persist a user message and call `ChatService`,
while the Agent writes assistant messages through service-facing dependencies.

## Runtime

The runtime is implemented with [CloudWeGo Eino](https://github.com/cloudwego/eino):

- `agent.go` creates an Eino ReAct agent for one chat turn.
- `llm/` builds OpenAI-compatible and Claude Eino `ToolCallingChatModel`
  implementations from user settings.
- `tools/` contains LearnLang tools and the small Eino adapter that preserves
  their existing JSON validation and business side effects.
- `archive/` uses an Eino tool-calling model to segment old conversation into
  retrieval-oriented archive records.
- `memory/` stores and searches archive vectors in Milvus.
- `prompts/` builds the chat system prompt and short-term-memory input.

## Chat Flow

```text
controller -> ChatService -> per-user run coordinator
                         -> Eino ReAct agent
                         -> LearnLang tools
                         -> services -> database / WebSocket / task scheduler
```

One user has at most one active Agent run. A later user message cancels the
current run and retries with the unprocessed messages merged in order. The
tool implementations check the context before side effects so a cancelled run
does not persist a late reply or scheduled task.

## Dependency Direction

```text
controllers/routes/websocket
        -> services
        -> agent
        -> agent/{llm,tools,archive,memory,prompts}
```

Keep Gin, controllers, routes, and WebSocket transport code out of this
package. Tools may depend on service-facing APIs, but should not reach directly
into HTTP transport concerns.
