package main

import (
	"context"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zarlcorp/core/pkg/zapp"
	"github.com/zarlcorp/zvault/internal/cli"
	"github.com/zarlcorp/zvault/internal/tui"
)

// version is set at build time via ldflags.
var version = "dev"

func main() {
	app := zapp.New(zapp.WithName("zvault"))

	ctx, cancel := zapp.SignalContext(context.Background())
	defer cancel()
	_ = ctx // reserved for future use

	if len(os.Args) > 1 {
		cli.Run(os.Args[1:], version)
		_ = app.Close()
		return
	}

	if err := runTUI(); err != nil {
		slog.Error("tui", "err", err)
		_ = app.Close()
		os.Exit(1)
	}

	if err := app.Close(); err != nil {
		slog.Error("shutdown", "err", err)
		os.Exit(1)
	}
}

func runTUI() error {
	p := tea.NewProgram(tui.New(version))
	_, err := p.Run()
	return err
}
