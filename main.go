package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/naglezhang/fingersaver/internal/agent"
	"github.com/naglezhang/fingersaver/internal/config"
	"github.com/naglezhang/fingersaver/internal/llm"
	"github.com/naglezhang/fingersaver/internal/tmux"
	"github.com/naglezhang/fingersaver/internal/tui"
)

var (
	showHelp    = flag.Bool("help", false, "Show help")
	showVersion = flag.Bool("version", false, "Show version")
	showConfig  = flag.Bool("config", false, "Show current configuration and exit")
)

const version = "0.1.0"

func main() {
	flag.BoolVar(showHelp, "h", false, "Show help")
	flag.Parse()

	if *showHelp {
		fmt.Print(helpText())
		return
	}
	if *showVersion {
		fmt.Printf("fingersaver %s\n", version)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if *showConfig {
		fmt.Print(cfg.Summary())
		return
	}

	// Start tmux client.
	tc := tmux.NewClient(cfg.TmuxSocketPath)
	if err := tc.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting tmux: %v\n", err)
		os.Exit(1)
	}
	defer tc.Stop()

	// Create LLM provider.
	provider, err := llm.NewProvider(cfg.LLMProvider, cfg.LLMAPIKey, cfg.LLMBaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating LLM provider: %v\n", err)
		os.Exit(1)
	}

	// Create orchestrator.
	hm := agent.NewHookManager()
	orch := agent.NewOrchestrator(provider, tc, hm, agent.AllTools(tc))
	orch.SetCommandRegistry(agent.NewCommandRegistry(tc))
	orch.SetModel(cfg.LLMModel)

	// Create and run TUI.
	app := tui.NewAppModel(orch, tc)
	app.SetConfigInfo(cfg.Summary())

	// Set up chat history persistence.
	if cfg.ChatHistoryPath != "" {
		history, err := tui.NewChatHistory(cfg.ChatHistoryPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not open chat history: %v\n", err)
		} else {
			defer history.Close()
			app.SetChatHistory(history)
		}
	}

	p := tea.NewProgram(app)

	app.SetSendFn(func(msg tea.Msg) {
		p.Send(msg)
	})

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func helpText() string {
	return `fingersaver - AI coding agent orchestrator

USAGE
  fingersaver [flags]

FLAGS
  -h, --help      Show this help
  --version       Show version
  --config        Show current configuration and exit

CONFIGURATION
  FingerSaver reads from ~/.claude/settings.json automatically:
  - ANTHROPIC_AUTH_TOKEN  -> API key
  - ANTHROPIC_BASE_URL   -> Custom API endpoint
  - ANTHROPIC_DEFAULT_SONNET_MODEL -> Model name

  Override with environment variables:
  - FINGERSAVER_LLM_PROVIDER  (anthropic|openai)
  - FINGERSAVER_LLM_API_KEY
  - FINGERSAVER_LLM_MODEL
  - ANTHROPIC_API_KEY / OPENAI_API_KEY

  Or create ~/.fingersaver/config.json for persistent settings.

KEY BINDINGS
  Tab           Switch between Chat and Viewer panes
  [ / ]         Switch between tmux sessions (in Viewer)
  Up/Down       Navigate input history (in Chat)
  Enter         Send message
  Ctrl+C        Exit

CHAT COMMANDS
  @session text   Send text to a tmux session
  /help           Show available slash commands
`
}
