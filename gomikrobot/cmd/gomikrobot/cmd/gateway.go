package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/kamir/gomikrobot/internal/agent"
	"github.com/kamir/gomikrobot/internal/bus"
	"github.com/kamir/gomikrobot/internal/channels"
	"github.com/kamir/gomikrobot/internal/config"
	"github.com/kamir/gomikrobot/internal/provider"
	"github.com/kamir/gomikrobot/internal/timeline"
	"github.com/spf13/cobra"
)

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Start the agent gateway (WhatsApp, etc)",
	Run:   runGateway,
}

func runGateway(cmd *cobra.Command, args []string) {
	fmt.Println("Starting GoMikroBot Gateway...")

	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Config error: %v\n", err)
		os.Exit(1)
	}

	// 2. Setup Bus
	msgBus := bus.NewMessageBus()

	// 3. Setup Providers
	oaProv := provider.NewOpenAIProvider(cfg.Providers.OpenAI.APIKey, cfg.Providers.OpenAI.APIBase, cfg.Agents.Defaults.Model)
	var prov provider.LLMProvider = oaProv

	if cfg.Providers.LocalWhisper.Enabled {
		prov = provider.NewLocalWhisperProvider(cfg.Providers.LocalWhisper, oaProv)
	}

	// 4. Setup Loop
	loop := agent.NewLoop(agent.LoopOptions{
		Bus:           msgBus,
		Provider:      prov,
		Workspace:     cfg.Agents.Defaults.Workspace,
		Model:         cfg.Agents.Defaults.Model,
		MaxIterations: cfg.Agents.Defaults.MaxToolIterations,
	})

	// 5. Setup Timeline (QMD)
	home, _ := os.UserHomeDir()
	timelinePath := fmt.Sprintf("%s/.gomikrobot/timeline.db", home)
	timeSvc, err := timeline.NewTimelineService(timelinePath)
	if err != nil {
		fmt.Printf("Failed to init timeline: %v\n", err)
		os.Exit(1)
	}

	// 6. Setup Channels
	// WhatsApp
	wa := channels.NewWhatsAppChannel(cfg.Channels.WhatsApp, msgBus, prov, timeSvc)

	// 7. Start Everything
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start Channels
	if err := wa.Start(ctx); err != nil {
		fmt.Printf("Failed to start WhatsApp: %v\n", err)
	}

	// Start Bus Dispatcher
	go msgBus.DispatchOutbound(ctx)

	// Start Local HTTP Server for Local Network access
	// Start Local HTTP Server for Local Network access (API)
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			msg := r.URL.Query().Get("message")
			if msg == "" {
				http.Error(w, "Missing message parameter", http.StatusBadRequest)
				return
			}

			session := r.URL.Query().Get("session")
			if session == "" {
				session = "local:default"
			}

			fmt.Printf("🌐 Local Network Request: %s\n", msg)
			resp, err := loop.ProcessDirect(ctx, msg, session)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, resp)
		})

		addr := fmt.Sprintf("%s:%d", cfg.Gateway.Host, cfg.Gateway.Port)
		fmt.Printf("📡 API Server listening on http://%s\n", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			fmt.Printf("API Server Error: %v\n", err)
		}
	}()

	// Start Dashboard Server
	go func() {
		mux := http.NewServeMux()

		// API: Timeline
		mux.HandleFunc("/api/v1/timeline", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")

			limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			if limit == 0 {
				limit = 100
			}
			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
			sender := r.URL.Query().Get("sender")

			events, err := timeSvc.GetEvents(timeline.FilterArgs{
				Limit:    limit,
				Offset:   offset,
				SenderID: sender,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(events)
		})

		// API: Settings (GET/POST)
		mux.HandleFunc("/api/v1/settings", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Content-Type", "application/json")

			if r.Method == "OPTIONS" {
				return
			}

			if r.Method == "POST" {
				var body struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					http.Error(w, "invalid body", http.StatusBadRequest)
					return
				}
				if err := timeSvc.SetSetting(body.Key, body.Value); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				fmt.Printf("⚙️ Setting changed: %s = %s\n", body.Key, body.Value)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
				return
			}

			// GET: return all requested settings
			key := r.URL.Query().Get("key")
			if key != "" {
				val, err := timeSvc.GetSetting(key)
				if err != nil {
					json.NewEncoder(w).Encode(map[string]string{"key": key, "value": ""})
					return
				}
				json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
				return
			}
			// Return silent_mode by default
			json.NewEncoder(w).Encode(map[string]bool{"silent_mode": timeSvc.IsSilentMode()})
		})

		// Static: Media
		mediaDir := filepath.Join(cfg.Agents.Defaults.Workspace, "media")
		fs := http.FileServer(http.Dir(mediaDir))
		mux.Handle("/media/", http.StripPrefix("/media/", fs))

		// SPA: Timeline
		mux.HandleFunc("/timeline", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "web/timeline.html")
		})

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.Redirect(w, r, "/timeline", http.StatusFound)
			}
		})

		if cfg.Gateway.DashboardPort == 0 {
			cfg.Gateway.DashboardPort = 18791
		}
		addr := fmt.Sprintf("%s:%d", cfg.Gateway.Host, cfg.Gateway.DashboardPort)
		fmt.Printf("🖥️  Dashboard listening on http://%s\n", addr)

		// Startup Probe
		if err := http.ListenAndServe(addr, mux); err != nil {
			fmt.Printf("❌ Dashboard Server FAILED to start: %v\n", err)
			cancel() // Stop the whole gateway if dashboard fails
		}
	}()

	// Start Agent Loop in background
	go func() {
		if err := loop.Run(ctx); err != nil {
			fmt.Printf("Agent loop crashed: %v\n", err)
			cancel()
		}
	}()

	fmt.Println("Gateway running. Press Ctrl+C to stop.")
	<-sigChan

	fmt.Println("Shutting down...")
	wa.Stop()
	loop.Stop()
	timeSvc.Close()
}
