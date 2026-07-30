package tools

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
)

type fakeCallableTool struct {
	input string
}

func (t *fakeCallableTool) Name() string {
	return "fake_tool"
}

func (t *fakeCallableTool) Description() string {
	return "A test tool."
}

func (t *fakeCallableTool) Call(_ context.Context, input string) (string, error) {
	t.input = input
	return "ok", nil
}

func TestEinoToolAdaptsExistingCallableTool(t *testing.T) {
	callable := &fakeCallableTool{}
	einoTool := NewEinoTool(callable, schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"query": {Type: schema.String, Required: true},
	}), JSONStringField("query"))

	info, err := einoTool.Info(context.Background())
	if err != nil {
		t.Fatalf("Info returned an error: %v", err)
	}
	if info.Name != callable.Name() || info.Desc != callable.Description() {
		t.Fatalf("unexpected tool info: %#v", info)
	}

	result, err := einoTool.InvokableRun(context.Background(), `{"query":"search this memory"}`)
	if err != nil {
		t.Fatalf("InvokableRun returned an error: %v", err)
	}
	if result != "ok" || callable.input != "search this memory" {
		t.Fatalf("unexpected adapted invocation: result=%q input=%q", result, callable.input)
	}
}

func TestJSONStringFieldRejectsInvalidArguments(t *testing.T) {
	_, err := JSONStringField("query")(`{"query":3}`)
	if err == nil {
		t.Fatal("expected non-string query to be rejected")
	}
}
