package agent

import (
	"context"
	"sort"
	"strings"
	"sync"
)

type queuedUserInput struct {
	MessageID int64
	Text      string
}

type agentBatchRunner func(context.Context, []queuedUserInput)

type agentRunState struct {
	mu           sync.Mutex
	inputs       []queuedUserInput
	version      uint64
	cancel       context.CancelFunc
	workerActive bool
}

// agentRunCoordinator keeps one Agent worker per user. A new input cancels the
// current run; the worker then retries with every input accumulated in its batch.
type agentRunCoordinator struct {
	mu   sync.Mutex
	runs map[int64]*agentRunState
}

func mergeQueuedUserInputs(inputs []queuedUserInput) (int64, string) {
	if len(inputs) == 0 {
		return 0, ""
	}

	parts := make([]string, 0, len(inputs))
	for _, input := range inputs {
		parts = append(parts, input.Text)
	}
	return inputs[0].MessageID, strings.Join(parts, "\n")
}

func sortQueuedUserInputs(inputs []queuedUserInput) {
	sort.SliceStable(inputs, func(i, j int) bool {
		return inputs[i].MessageID < inputs[j].MessageID
	})
}

func (c *agentRunCoordinator) Submit(userID int64, input queuedUserInput, runner agentBatchRunner) {
	c.mu.Lock()
	if c.runs == nil {
		c.runs = make(map[int64]*agentRunState)
	}
	state := c.runs[userID]
	if state == nil {
		state = &agentRunState{}
		c.runs[userID] = state
	}
	state.mu.Lock()
	c.mu.Unlock()

	state.inputs = append(state.inputs, input)
	sortQueuedUserInputs(state.inputs)
	state.version++
	if state.cancel != nil {
		state.cancel()
	}
	startWorker := !state.workerActive
	if startWorker {
		state.workerActive = true
	}
	state.mu.Unlock()
	if !startWorker {
		return
	}

	go c.run(userID, state, runner)
}

func (c *agentRunCoordinator) run(userID int64, state *agentRunState, runner agentBatchRunner) {
	for {
		state.mu.Lock()
		version := state.version
		inputs := append([]queuedUserInput(nil), state.inputs...)
		ctx, cancel := context.WithCancel(context.Background())
		state.cancel = cancel
		state.mu.Unlock()

		runner(ctx, inputs)
		cancel()

		c.mu.Lock()
		state.mu.Lock()
		if state.version == version {
			state.inputs = nil
			state.cancel = nil
			state.workerActive = false
			if c.runs[userID] == state {
				delete(c.runs, userID)
			}
			state.mu.Unlock()
			c.mu.Unlock()
			return
		}
		state.cancel = nil
		state.mu.Unlock()
		c.mu.Unlock()
	}
}
