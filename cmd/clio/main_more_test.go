package main

import (
	"errors"
	"testing"
)

// TestMainSuccess verifies that main does not call exit on success.
// This ensures normal startup does not terminate the process.
func TestMainSuccess(t *testing.T) {
	oldRun := runApp
	oldExit := exitFn
	defer func() {
		runApp = oldRun
		exitFn = oldExit
	}()

	exited := false
	runApp = func() error { return nil }
	exitFn = func(code int) { exited = true }

	main()
	if exited {
		t.Fatalf("did not expect exit on success")
	}
}

// TestMainFailure verifies that main calls exit on failure.
// This ensures fatal startup errors terminate the process.
func TestMainFailure(t *testing.T) {
	oldRun := runApp
	oldExit := exitFn
	defer func() {
		runApp = oldRun
		exitFn = oldExit
	}()

	exitCode := 0
	runApp = func() error { return errors.New("boom") }
	exitFn = func(code int) { exitCode = code }

	main()
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}
