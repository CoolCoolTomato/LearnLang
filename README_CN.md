# LearnLang

<div align="center">

**让对话沉淀为词汇与记忆的 AI 语言学习伙伴。**

[English](README.md)

</div>

LearnLang 是一款围绕真实对话设计的语言学习聊天应用。你可以通过文本或语音练习，阅读 AI 的双语回复，从消息中积累词汇，并让长期记忆保存持续学习所需的上下文。

## 产品演示

<p align="center">
  <img src="./assets/LearnLang.png" alt="LearnLang 产品概览" width="92%" />
</p>

### 在聊天中直接翻译

不知道如何表达？先翻译再发给AI。

<p align="center">
  <img src="./assets/translate-demo.gif" alt="在聊天中翻译 AI 消息" width="92%" />
</p>

### 点击 AI 头像查看译文

点击 AI 头像，即可在目标语言原文和对应译文之间切换。

<p align="center">
  <img src="./assets/avatar-translation-demo.gif" alt="点击 AI 头像查看译文" width="92%" />
</p>

### 从消息中查询词汇

在聊天消息中选择单词或短语，即可在保留上下文的同时查询个人词库。

<p align="center">
  <img src="./assets/vocabulary-demo.gif" alt="从聊天消息中查询词汇" width="92%" />
</p>

### 让每次对话都能积累

AI 回复可包含译文与语音。近期消息保留为对话上下文，较早的会话可被摘要并写入语义记忆，供后续检索。学习者可以从消息中查询词汇、保存到个人词库，再在同一段对话中继续练习。

### 使用自己的模型配置

LearnLang 支持 OpenAI 兼容的聊天、向量、语音识别和语音合成服务。你可以在设置页配置供应商和模型，并根据目标语言、母语与时区调整学习体验。

## 词库来源

LearnLang 支持个人词库和 JSON 导入。英文词表可参考 [KyleBing/english-vocabulary](https://github.com/KyleBing/english-vocabulary)，该项目提供四级、六级、考研、SAT 等分级英文词汇数据。

1. 在上游项目中选择适合学习阶段的词表。
2. 将内容转换或映射到 LearnLang 的导入格式。本仓库中的 `demo_vocabulary.json` 和 `demo_vocabulary_learnlang.json` 可作为示例。
3. 在应用中创建或选择词库，然后使用 **导入** 功能写入数据。

上游词表的内容和许可证仍归其项目所有。若要重新分发由其衍生的词库，请先确认上游许可条款。

## 快速开始

### 前置条件

- 容器化部署需要 Docker Engine 和 Docker Compose v2
- 源码运行需要 Go 与 pnpm
- 需要准备聊天、向量、语音识别和语音合成的 AI 服务凭据

### 使用 Docker 启动

```bash
git clone https://github.com/CoolCoolTomato/LearnLang.git
cd LearnLang/docker
cp .env.example .env
```

在 `.env` 中配置数据库密码和公开 API 地址，再检查 `api.config.yaml` 中的数据库、Redis、Milvus 和服务端配置。

```bash
docker compose -f docker-compose.yml up -d
docker compose -f docker-compose.yml ps
```

该部署会启动 PostgreSQL、Redis、Milvus、MinIO、API、后台 Worker 和 Web 应用。

### 本地构建镜像

```bash
cd docker
cp .env.example .env
docker compose -f docker-compose.local.yml build
docker compose -f docker-compose.local.yml up -d
```

### 源码运行

先启动本地依赖：

```bash
cd docker
cp .env.example .env
docker compose -f docker-compose.dev.yml up -d
```

随后在独立终端运行 API 和 Web 应用：

```bash
cd api
go run main.go
```

```bash
cd app
pnpm install
pnpm run dev
```

运行桌面端：

```bash
cd app
pnpm tauri dev
```

## 首次配置

1. 登录后打开设置页。
2. 设置母语、目标语言和时区。
3. 为聊天、向量、语音识别和语音合成配置供应商及模型。
4. 打开对话页发送消息，并随着学习导入或积累词汇。

密钥只应保存在本地 `.env` 和配置文件中。不要提交 API 密钥、密码或存储凭据。

## 架构概览

```text
Web 应用 / Tauri 桌面端
           |
         Go API ---- Redis（会话、通知、定时任务）
           |  \
           |   \--- PostgreSQL（用户、消息、词库）
           |
           +------ Milvus（会话语义记忆）
           +------ MinIO（语音资源）
           +------ AI 供应商（聊天、向量、STT、TTS）
```

API 负责消息与词库持久化、Agent 回复协调，并通过 WebSocket 推送 AI 消息。后台任务负责定时消息和会话归档。

## 开发与验证

```bash
cd api && go test ./... && go build ./...
cd app && pnpm run lint && pnpm run build
```

目录说明：

- `api/`：Go API、Agent、持久化与后台任务
- `app/`：React Web 应用和 Tauri 桌面端
- `docker/`：Compose 文件和本地服务配置
- `assets/`：README 媒体与项目资源

## 致谢

- 英文词表可参考 [english-vocabulary](https://github.com/KyleBing/english-vocabulary)。
