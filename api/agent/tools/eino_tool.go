package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// CallableTool is the existing LearnLang tool contract. EinoTool adapts it to
// Eino without moving validation or business side effects out of the tools.
type CallableTool interface {
	Name() string
	Description() string
	Call(context.Context, string) (string, error)
}

type InputTransformer func(string) (string, error)

type EinoTool struct {
	Tool      CallableTool
	Params    *schema.ParamsOneOf
	Transform InputTransformer
}

var _ tool.InvokableTool = (*EinoTool)(nil)

func NewEinoTool(callable CallableTool, params *schema.ParamsOneOf, transform InputTransformer) *EinoTool {
	return &EinoTool{Tool: callable, Params: params, Transform: transform}
}

func (t *EinoTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	if t == nil || t.Tool == nil {
		return nil, fmt.Errorf("tool is not configured")
	}
	return &schema.ToolInfo{
		Name:        t.Tool.Name(),
		Desc:        t.Tool.Description(),
		ParamsOneOf: t.Params,
	}, nil
}

func (t *EinoTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	if t == nil || t.Tool == nil {
		return "", fmt.Errorf("tool is not configured")
	}
	input := argumentsInJSON
	if t.Transform != nil {
		var err error
		input, err = t.Transform(argumentsInJSON)
		if err != nil {
			return "", err
		}
	}
	return t.Tool.Call(ctx, input)
}

// JSONStringField converts Eino's function-call object into the raw string
// expected by a legacy free-form input tool.
func JSONStringField(field string) InputTransformer {
	return func(arguments string) (string, error) {
		var values map[string]json.RawMessage
		if err := json.Unmarshal([]byte(arguments), &values); err != nil {
			return "", fmt.Errorf("tool input must be a JSON object: %w", err)
		}
		raw, ok := values[field]
		if !ok {
			return "", fmt.Errorf("tool input field %q is required", field)
		}
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", fmt.Errorf("tool input field %q must be a string: %w", field, err)
		}
		return value, nil
	}
}
