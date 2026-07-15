package agent

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestAgentRunCoordinatorInterruptsAndRetriesMergedBatch(t *testing.T) {
	var coordinator agentRunCoordinator
	firstStarted := make(chan struct{})
	firstCanceled := make(chan struct{})
	secondFinished := make(chan struct{})

	var mu sync.Mutex
	var batches [][]queuedUserInput
	active := 0
	maxActive := 0
	runner := func(ctx context.Context, inputs []queuedUserInput) {
		mu.Lock()
		active++
		if active > maxActive {
			maxActive = active
		}
		batches = append(batches, append([]queuedUserInput(nil), inputs...))
		call := len(batches)
		mu.Unlock()
		defer func() {
			mu.Lock()
			active--
			mu.Unlock()
		}()

		switch call {
		case 1:
			close(firstStarted)
			<-ctx.Done()
			close(firstCanceled)
		case 2:
			close(secondFinished)
		}
	}

	first := queuedUserInput{MessageID: 101, Text: "I wanted to say"}
	second := queuedUserInput{MessageID: 102, Text: "that I will arrive tomorrow."}
	coordinator.Submit(7, first, runner)
	waitForSignal(t, firstStarted, "first Agent run to start")

	coordinator.Submit(7, second, runner)
	waitForSignal(t, firstCanceled, "first Agent run to be canceled")
	waitForSignal(t, secondFinished, "merged Agent run to finish")
	waitForCoordinatorIdle(t, &coordinator, 7)

	mu.Lock()
	defer mu.Unlock()
	want := [][]queuedUserInput{{first}, {first, second}}
	if !reflect.DeepEqual(batches, want) {
		t.Fatalf("unexpected Agent batches: got %#v, want %#v", batches, want)
	}
	if maxActive != 1 {
		t.Fatalf("same-user Agent runs overlapped: max active = %d", maxActive)
	}
}

func TestAgentRunCoordinatorStartsNewBatchAfterCompletion(t *testing.T) {
	var coordinator agentRunCoordinator
	finished := make(chan struct{}, 2)

	var mu sync.Mutex
	var batches [][]queuedUserInput
	runner := func(_ context.Context, inputs []queuedUserInput) {
		mu.Lock()
		batches = append(batches, append([]queuedUserInput(nil), inputs...))
		mu.Unlock()
		finished <- struct{}{}
	}

	first := queuedUserInput{MessageID: 201, Text: "first turn"}
	second := queuedUserInput{MessageID: 202, Text: "second turn"}
	coordinator.Submit(8, first, runner)
	waitForSignal(t, finished, "first Agent run to finish")
	waitForCoordinatorIdle(t, &coordinator, 8)

	coordinator.Submit(8, second, runner)
	waitForSignal(t, finished, "second Agent run to finish")
	waitForCoordinatorIdle(t, &coordinator, 8)

	mu.Lock()
	defer mu.Unlock()
	want := [][]queuedUserInput{{first}, {second}}
	if !reflect.DeepEqual(batches, want) {
		t.Fatalf("unexpected Agent batches: got %#v, want %#v", batches, want)
	}
}

func TestMergeQueuedUserInputs(t *testing.T) {
	inputs := []queuedUserInput{
		{MessageID: 302, Text: "we should leave now."},
		{MessageID: 301, Text: "I think"},
	}
	sortQueuedUserInputs(inputs)

	contextBeforeMessageID, userInput := mergeQueuedUserInputs(inputs)
	if contextBeforeMessageID != 301 {
		t.Fatalf("unexpected context cutoff: got %d, want 301", contextBeforeMessageID)
	}
	if userInput != "I think\nwe should leave now." {
		t.Fatalf("unexpected merged input: %q", userInput)
	}
}

func TestAgentRunCoordinatorAllowsDifferentUsersToRunConcurrently(t *testing.T) {
	var coordinator agentRunCoordinator
	started := make(chan int64, 2)
	release := make(chan struct{})
	finished := make(chan struct{}, 2)
	var releaseOnce sync.Once
	releaseRuns := func() {
		releaseOnce.Do(func() { close(release) })
	}
	defer releaseRuns()

	runner := func(userID int64) agentBatchRunner {
		return func(_ context.Context, _ []queuedUserInput) {
			started <- userID
			<-release
			finished <- struct{}{}
		}
	}

	coordinator.Submit(11, queuedUserInput{MessageID: 401, Text: "first user"}, runner(11))
	coordinator.Submit(12, queuedUserInput{MessageID: 402, Text: "second user"}, runner(12))

	seen := map[int64]bool{}
	for range 2 {
		select {
		case userID := <-started:
			seen[userID] = true
		case <-time.After(2 * time.Second):
			t.Fatal("different users did not start independently")
		}
	}
	if !seen[11] || !seen[12] {
		t.Fatalf("unexpected users started: %#v", seen)
	}

	releaseRuns()
	waitForSignal(t, finished, "first user Agent run to finish")
	waitForSignal(t, finished, "second user Agent run to finish")
}

func waitForSignal(t *testing.T, signal <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func waitForCoordinatorIdle(t *testing.T, coordinator *agentRunCoordinator, userID int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		coordinator.mu.Lock()
		state, ok := coordinator.runs[userID]
		if ok {
			state.mu.Lock()
			coordinator.mu.Unlock()
			idle := !state.workerActive
			state.mu.Unlock()
			if idle {
				return
			}
		} else {
			coordinator.mu.Unlock()
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for Agent coordinator to become idle")
}
