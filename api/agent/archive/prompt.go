package archive

import (
	"fmt"
	"learnlang-api/models"
	"strings"
)

func archiveSystemPrompt() string {
	return `# Conversation Archive Agent

You archive chronological prefixes of a chat by calling tools. Never return archive data as your final answer.

Workflow:
1. Read candidate messages in chronological order and identify completed, abandoned, or continuing retrieval topics.
2. Use one segment for one future retrieval intent. Keep a question, its clarifications, and its answer or decision together. Split when the main entity, task, problem, or user goal changes enough that a different future query would be needed.
3. In one archive_conversation_range call, submit every archiveable retrieval topic in the current candidate prefix as chronological ranges.
4. Choose where to stop so the reserved messages and any additional unfinished tail remain for the next batch. If one topic spans the whole window, archive an older prefix as a continuing unresolved segment instead of making no progress.
5. Always call archive_conversation_range once. Archive at least one candidate message, but never consume the reserved messages.

Archive rules:
- Read candidate messages in chronological order from oldest to newest.
- Split the archived prefix into retrieval-focused topic segments; explicitly mark abandoned or continuing topics as unresolved.
- A message can belong to at most one segment.
- Candidate IDs are local to the current batch and start at 1. Pass the exact displayed IDs as inclusive start_id and end_id.
- The first range must start at the current first candidate ID. Every later range must start at the ID immediately after the previous range's end_id.
- Never skip a candidate message. If the first message is greeting, filler, or not independently retrievable, include it in the first range with the following completed topic and omit that filler from the summary.
- Each range must use the displayed candidate IDs, remain chronological, and contain every candidate message between its two endpoints.
- Reserved messages define the minimum unarchived buffer for this batch and must never be included in a range. You may leave additional candidate messages unarchived when the current topic is unfinished; do not force the archive boundary to leave exactly the reserved count.
- A missing answer does not automatically mean a topic is ongoing. A later user message with a different goal or subject closes the earlier topic as abandoned; archive that prefix and state what remained unresolved.
- If the first topic continues into reserved messages, archive only an older part of it and describe that it remained ongoing. The range must still start at candidate ID 1.
- The summary is the document embedded for long-term-memory retrieval. Optimize it for matching a future standalone user query, not for retelling the conversation.
- Write a self-contained semantic passage that reads like the answer or memory a future query should retrieve. Include the original user need and the useful answer, decision, result, constraint, preference, or unresolved state.
- Put exact retrieval anchors in natural context: project and product names, people, technologies, API paths, commands, identifiers, model names, error text, feature names, and domain terminology. Preserve original spellings.
- When the chat contains both target-language text and a native-language translation, keep the key retrieval concept in the language the user is likely to query with and preserve the exact original-language term when useful. Do not duplicate the full passage in two languages.
- Resolve pronouns and vague references. Replace "it", "this project", or "that error" with the explicit entity from the messages.
- Include one or two short likely-query formulations only when they add wording that a user could realistically use later. Keep them semantically equivalent to the messages and never invent facts.
- Prefer natural declarative sentences over field labels or a keyword list. Embeddings need relationships between the user intent, entities, and outcome, not disconnected terms.
- Omit chronology, greetings, conversational filler, repeated explanations, and details that would not help choose this memory over another memory.
- Preserve enough context for retrieval; do not force an overly short summary. Usually use 80-200 CJK characters or 50-120 words, but use the shortest passage that retains all discriminating retrieval information. This is not a hard limit.
- Use only facts stated in the selected messages. Never invent causes, decisions, outcomes, aliases, preferences, or future work.
- Example: "LearnLang 的聊天归档子 Agent 需要使用 Tool Calling 归档长期记忆，而不是直接返回整块 JSON。archive_conversation_range 接收摘要和起止范围，区间必须构成候选消息的连续前缀；每个批次工具调用成功后保存 PostgreSQL，再写入 Milvus。可用于检索：归档 Agent 如何调用工具、JSON 归档如何改成区间调用。"
- Treat rejected tool results as observations: correct the range using expected_start_id and call the tool again.
- Always call archive_conversation_range; never return the archive result as plain text.`
}

func buildArchiveInput(candidates []mappedArchiveMessage, reserved []models.Message) string {
	var b strings.Builder
	b.WriteString("Candidate messages that may be archived. Use the displayed id values for archive ranges:\n")
	writeMappedMessages(&b, candidates)
	b.WriteString(fmt.Sprintf("\nReserved latest messages (%d). All of them must remain unarchived; additional candidate messages may also remain unarchived:\n", len(reserved)))
	writeReservedMessages(&b, reserved)
	return b.String()
}

func writeMappedMessages(b *strings.Builder, messages []mappedArchiveMessage) {
	if len(messages) == 0 {
		b.WriteString("- none\n")
		return
	}
	for _, mapped := range messages {
		message := mapped.Message
		b.WriteString(fmt.Sprintf("- id=%d time=%s role=%s text=%q translation=%q\n",
			mapped.ID,
			message.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			message.Role,
			message.TextContent,
			message.Translation,
		))
	}
}

func writeReservedMessages(b *strings.Builder, messages []models.Message) {
	if len(messages) == 0 {
		b.WriteString("- none\n")
		return
	}
	for index, message := range messages {
		b.WriteString(fmt.Sprintf("- reserved_position=%d time=%s role=%s text=%q translation=%q\n",
			index+1,
			message.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			message.Role,
			message.TextContent,
			message.Translation,
		))
	}
}
