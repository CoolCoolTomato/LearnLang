# LearnLang

<div align="center">

[English](README.md) | 中文

</div>

LearnLang 是一个面向语言学习场景的 AI 聊天应用，支持文本 / 语音输入、WebSocket 实时回复、长期记忆检索以及定时消息能力。

它不仅是一个 AI 助手，更是一个**具备记忆能力的语言学习伙伴**。

---

## 功能特性

- 文本聊天与语音聊天
- AI 双语回复（每句话自动附带翻译）
- 基于 WebSocket 的实时流式回复
- 基于 `pgvector` 的长期记忆存储与语义检索
- 会话摘要与上下文压缩
- 定时消息 / 提醒能力（自动转换为 UTC 执行）
- 语音生成与语音识别（TTS / STT）
- OpenAI 兼容 API，支持多模型接入
- 支持 Docker 一键部署
- 
<p align="center">
  <img src="./assets/learnlang-desktop.png" width="90%" />
</p>

---

## 快速开始

### Docker 启动

克隆项目仓库

```bash
git clone https://github.com/your-repo/learnlang.git
cd learnlang
```

复制配置文件：

```bash
cd docker
cp .env.example .env
```

修改 `api.config.yaml` 和 `.env` 后，使用 `docker compose` 一键启动

```bash
docker compose -f docker-compose.yml up -d
```

### 本地构建镜像

复制配置文件：

```bash
cd docker
cp .env.example .env
```

修改 `api.config.yaml` 和 `.env` 后，使用 `docker compose` 一键启动

```bash
docker compose -f docker-compose.local.yml build
docker compose -f docker-compose.local.yml up -d
```

### 源码启动

复制配置文件：

```bash
cd docker
cp .env.example .env
```

修改 `api.config.yaml` 和 `.env` 后，使用 `docker compose` 启动数据库

```bash
docker compose -f docker-compose.dev.yml up -d
```

启动后端服务

```bash
cd api
go mod tidy
go run main.go
```

启动app服务

```bash
cd app
pnpm install
pnpm run dev
```

运行桌面端

```bash
pnpm tauri dev
```

### 配置模型

在使用前，需要配置聊天，嵌入，语言转文字，文本转语音四种模型

<p align="center">
  <img src="./assets/settings-desktop.png" width="90%" />
</p>