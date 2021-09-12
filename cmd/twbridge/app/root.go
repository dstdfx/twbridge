package app

import (
	"context"
	"os/signal"
	"runtime"
	"syscall"
)

// Variables that are injected in build time.
var (
	buildGitCommit string
	buildGitTag    string
	buildDate      string
	buildCompiler  = runtime.Version()
)

func Start() {
	// Handle interrupt signals
	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	<-rootCtx.Done()
}
