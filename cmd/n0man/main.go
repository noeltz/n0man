package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/noeltz/n0man/internal/cmd"
)

func main() {
	// Setup context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT and SIGTERM for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived signal %v, shutting down...\n", sig)
		cancel() // Signal all operations to stop

		// Give operations time to cleanup (max 5 seconds)
		<-time.After(5 * time.Second)
		os.Exit(130) // Standard exit code for SIGINT
	}()

	// Pass context to commands
	cmd.SetContext(ctx)
	cmd.Execute()
}
