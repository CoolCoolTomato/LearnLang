package aiusage

import (
	"context"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"log"
)

type ChatModel struct {
	model                model.ToolCallingChatModel
	recorder             Recorder
	userID               int64
	operation, modelName string
}

func NewChatModel(chatModel model.ToolCallingChatModel, recorder Recorder, userID int64, operation, modelName string) model.ToolCallingChatModel {
	if chatModel == nil || recorder == nil || userID <= 0 {
		return chatModel
	}
	return &ChatModel{model: chatModel, recorder: recorder, userID: userID, operation: operation, modelName: modelName}
}
func (m *ChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	response, err := m.model.Generate(ctx, input, opts...)
	usage := float64(0)
	if response != nil && response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
		usage = float64(response.ResponseMeta.Usage.TotalTokens)
	}
	status := "succeeded"
	if err != nil {
		status = "failed"
	}
	if recordErr := m.recorder.RecordAIUsage(context.WithoutCancel(ctx), Record{UserID: m.userID, Operation: m.operation, Model: m.modelName, Usage: usage, Unit: "tokens", Status: status}); recordErr != nil {
		log.Printf("record AI usage failed for user %d: %v", m.userID, recordErr)
	}
	return response, err
}
func (m *ChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return m.model.Stream(ctx, input, opts...)
}
func (m *ChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	bound, err := m.model.WithTools(tools)
	if err != nil {
		return nil, err
	}
	return NewChatModel(bound, m.recorder, m.userID, m.operation, m.modelName), nil
}
