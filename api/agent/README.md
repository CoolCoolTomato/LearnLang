# Agent Directory Plan

This package is the backend Agent subsystem. It should stay isolated from HTTP
controllers and expose a small service-facing API that can be called from
`services/`, `controllers/`, or websocket handlers.

The implementation is planned around `github.com/tmc/langchaingo`. Keep
langchaingo-specific details inside this package so the rest of the API does not
depend on agent internals.

## Layout

```text
api/agent/
  README.md
  agent.go              # Public agent service interface and factory entrypoint.
  config/               # Agent, model, provider, timeout, and tool settings.
  llm/                  # LangChainGo LLM provider construction and adapters.
  tools/                # LangChainGo tools exposed to the agent.
  memory/               # Conversation memory and summarization adapters.
  prompts/              # System prompts, templates, and prompt builders.
  runtime/              # Executor setup, callbacks, tracing, retries, limits.
  sessions/             # Per-user/per-chat agent session lifecycle.
  usecases/             # LearnLang-specific agent workflows.
```

## Responsibilities

`agent.go`

- Defines the stable public API for the rest of LearnLang.
- Owns top-level construction, for example `NewService(...)`.
- Should return domain-oriented request/response types instead of langchaingo
  types where possible.

`config/`

- Maps app config into agent runtime configuration.
- Keeps provider/model names, API keys, base URLs, temperature, max tokens, and
  timeout defaults in one place.
- Should not create LLM clients directly.

`llm/`

- Builds concrete langchaingo LLM clients, for example OpenAI-compatible,
  Ollama, or other providers.
- Hides provider-specific options behind local constructors.
- Should be the only package that imports langchaingo provider packages such as
  `llms/openai` or `llms/ollama`.

`tools/`

- Contains tools the agent can call.
- Group tools by domain as the feature grows, for example vocabulary, speech,
  user progress, dictionary, or course lookup.
- Tool implementations can call existing `services/` through interfaces to
  avoid coupling agent logic directly to storage details.

`memory/`

- Provides chat history and long-term memory adapters.
- Starts with in-process/session memory if needed, then can add Redis/Postgres
  backed memory later.
- Keeps summarization and trimming policies close to memory storage.

`prompts/`

- Stores prompt templates and prompt builder functions.
- Keeps prompts versioned and named by workflow.
- Avoids hardcoding large prompts inside runtime or tool code.

`runtime/`

- Wires langchaingo `agents.Executor`, callbacks, retry policy, tool whitelist,
  and execution limits.
- Owns cross-cutting concerns such as logging, cancellation, and token limits.
- Should not contain LearnLang business workflows.

`sessions/`

- Manages per-user or per-conversation agent state.
- Converts LearnLang user/chat identifiers into memory/runtime state.
- Keeps session lifecycle separate from HTTP/WebSocket transport code.

`usecases/`

- Contains product-facing workflows, for example tutoring, sentence correction,
  vocabulary explanation, role play, or study plan generation.
- Depends on `runtime/`, `tools/`, and `memory/`, but exposes simple methods to
  backend services/controllers.

## Suggested Implementation Order

1. Add `github.com/tmc/langchaingo` to `api/go.mod`; while using the local clone,
   add a temporary `replace github.com/tmc/langchaingo => ../../langchaingo`.
2. Define the public service interface in `agent.go` and a minimal config model in
   `config/`.
3. Build one LLM provider in `llm/`.
4. Create a minimal runtime executor in `runtime/` with no custom tools.
5. Add the first LearnLang workflow under `usecases/`.
6. Add tools one at a time, with tests around service-facing behavior.
7. Wire the workflow into `services/` or a controller only after the package can
   be built and tested independently.

## Dependency Direction

Preferred dependency flow:

```text
controllers/routes/websocket
        -> services
        -> agent
        -> agent/usecases
        -> agent/runtime
        -> agent/{llm,tools,memory,prompts,config,sessions}
```

Avoid importing controllers, routes, or Gin from `api/agent`.
