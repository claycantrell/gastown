package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/tui/feed"
	"github.com/steveyegge/gastown/internal/web"
	"github.com/steveyegge/gastown/internal/websocket"
	"github.com/steveyegge/gastown/internal/workspace"
)

var (
	dashboardPort int
	dashboardOpen bool
)

var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	GroupID: GroupDiag,
	Short:   "Start the convoy tracking web dashboard",
	Long: `Start a web server that displays the convoy tracking dashboard.

The dashboard shows real-time convoy status with:
- Convoy list with status indicators
- Progress tracking for each convoy
- Last activity indicator (green/yellow/red)
- Auto-refresh every 30 seconds via htmx

Example:
  gt dashboard              # Start on default port 8080
  gt dashboard --port 3000  # Start on port 3000
  gt dashboard --open       # Start and open browser`,
	RunE: runDashboard,
}

func init() {
	dashboardCmd.Flags().IntVar(&dashboardPort, "port", 8080, "HTTP port to listen on")
	dashboardCmd.Flags().BoolVar(&dashboardOpen, "open", false, "Open browser automatically")
	rootCmd.AddCommand(dashboardCmd)
}

func runDashboard(cmd *cobra.Command, args []string) error {
	// Verify we're in a workspace
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	// Create the live convoy fetcher
	fetcher, err := web.NewLiveConvoyFetcher()
	if err != nil {
		return fmt.Errorf("creating convoy fetcher: %w", err)
	}

	// Create event sources for WebSocket
	var eventSource feed.EventSource

	// Try to create BdActivitySource (use town root as working directory)
	bdSource, err := feed.NewBdActivitySource(townRoot)
	if err != nil {
		log.Printf("Warning: Could not create bd activity source: %v", err)
		// Continue without bd activity source
	}

	// Try to create GtEventsSource
	gtSource, err := feed.NewGtEventsSource(townRoot)
	if err != nil {
		log.Printf("Warning: Could not create gt events source: %v", err)
		// Continue without gt events source
	}

	// Combine sources
	if bdSource != nil && gtSource != nil {
		eventSource = feed.NewCombinedSource(bdSource, gtSource)
	} else if bdSource != nil {
		eventSource = bdSource
	} else if gtSource != nil {
		eventSource = gtSource
	}

	// Create WebSocket hub if we have an event source
	var wsHub *websocket.Hub
	if eventSource != nil {
		wsHub = websocket.NewHub(eventSource)
		go wsHub.Run()
		log.Printf("WebSocket hub started")
	} else {
		log.Printf("Warning: No event sources available, WebSocket disabled")
	}

	// Create the handler
	handler, err := web.NewConvoyHandler(fetcher, wsHub)
	if err != nil {
		return fmt.Errorf("creating convoy handler: %w", err)
	}

	// Create HTTP mux and register handlers
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	if wsHub != nil {
		mux.HandleFunc("/ws", handler.HandleWebSocket)
	}

	// Build the URL
	url := fmt.Sprintf("http://localhost:%d", dashboardPort)

	// Open browser if requested
	if dashboardOpen {
		go openBrowser(url)
	}

	// Start the server with timeouts
	fmt.Printf("🚚 Gas Town Dashboard starting at %s\n", url)
	if wsHub != nil {
		fmt.Printf("   WebSocket endpoint: ws://localhost:%d/ws\n", dashboardPort)
	}
	fmt.Printf("   Press Ctrl+C to stop\n")

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", dashboardPort),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return server.ListenAndServe()
}

// openBrowser opens the specified URL in the default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	_ = cmd.Start()
}
