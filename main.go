package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/unicoooorn/pingr/cmd"
)

// @title		pingr
// @version		0.0.1

func main() {
	var code int
	defer func() {
		os.Exit(code)
	}()

	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancelFunc()
	if err := cmd.NewRootCmd().ExecuteContext(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Some error occurred during execute app. Error: %v\n", err)
		code = 1
	}
}
