package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/kamir/gomikrobot/internal/agent"
	"github.com/kamir/gomikrobot/internal/bus"
	"github.com/kamir/gomikrobot/internal/channels"
	"github.com/kamir/gomikrobot/internal/config"
	"github.com/kamir/gomikrobot/internal/httpmw"
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

	// Defaults (enterprise hardening)
	if cfg.Gateway.MaxBodyBytes <= 0 {
		cfg.Gateway.MaxBodyBytes = 10 << 20 // 10 MiB
	}
	if cfg.Gateway.RateLimitRPS <= 0 {
		cfg.Gateway.RateLimitRPS = 5
	}
	if cfg.Gateway.RateLimitBurst <= 0 {
		cfg.Gateway.RateLimitBurst = 10
	}
	if cfg.Gateway.ShutdownTimeout <= 0 {
		cfg.Gateway.ShutdownTimeout = 10 * time.Second
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
	wa := channels.NewWhatsAppChannel(cfg.Channels.WhatsApp, msgBus, prov, timeSvc)

	// 7. Start Everything
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Readiness flag for /ready.
	var ready atomic.Int32

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start Channels
	if err := wa.Start(ctx); err != nil {
		fmt.Printf("Failed to start WhatsApp: %v\n", err)
	} else {
		ready.Store(1)
	}

	// Start Bus Dispatcher
	go msgBus.DispatchOutbound(ctx)

	// Shared middleware
	rl := httpmw.NewRateLimiter(cfg.Gateway.RateLimitRPS, cfg.Gateway.RateLimitBurst)
	commonMW := []httpmw.Middleware{
		httpmw.Recoverer(),
		httpmw.MaxBodyBytes(cfg.Gateway.MaxBodyBytes),
		rl.Middleware(),
	}

	// API server
	apiAddr := fmt.Sprintf("%s:%d", cfg.Gateway.Host, cfg.Gateway.Port)
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	apiMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if ready.Load() == 1 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))
	})

	apiMux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		msg := r.URL.Query().Get("message")
		if msg == "" {
			http.Error(w, "missing message parameter", http.StatusBadRequest)
			return
		}

		session := r.URL.Query().Get("session")
		if session == "" {
			session = "local:default"
		}

		fmt.Printf("🌐 Local Network Request: %s\n", msg)
		resp, err := loop.ProcessDirect(ctx, msg, session)
		if err != nil {
			// Avoid leaking internal errors to clients.
			fmt.Printf("❌ /chat failed: %v\n", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		_, _ = fmt.Fprint(w, resp)
	})

	apiServer := &http.Server{
		Addr:    apiAddr,
		Handler: httpmw.Chain(apiMux, commonMW...),
	}

	go func() {
		fmt.Printf("📡 API Server listening on http://%s\n", apiAddr)
		err := apiServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("API Server Error: %v\n", err)
			cancel()
		}
	}()

	// Dashboard server
	if cfg.Gateway.DashboardPort == 0 {
		cfg.Gateway.DashboardPort = 18791
	}
	dashAddr := fmt.Sprintf("%s:%d", cfg.Gateway.Host, cfg.Gateway.DashboardPort)
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if ready.Load() == 1 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))
	})

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
			fmt.Printf("❌ /api/v1/timeline failed: %v\n", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(events)
	})

	// API: Settings (GET/POST)
	mux.HandleFunc("/api/v1/settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			return
		}

		if r.Method == http.MethodPost {
			var body struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			if err := timeSvc.SetSetting(body.Key, body.Value); err != nil {
				fmt.Printf("❌ /api/v1/settings POST failed: %v\n", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			fmt.Printf("⚙️ Setting changed: %s = %s\n", body.Key, body.Value)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		// GET: return all requested settings
		key := r.URL.Query().Get("key")
		if key != "" {
			val, err := timeSvc.GetSetting(key)
			if err != nil {
				_ = json.NewEncoder(w).Encode(map[string]string{"key": key, "value": ""})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
			return
		}
		// Return silent_mode by default
		_ = json.NewEncoder(w).Encode(map[string]bool{"silent_mode": timeSvc.IsSilentMode()})
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

	dashServer := &http.Server{
		Addr:    dashAddr,
		Handler: httpmw.Chain(mux, commonMW...),
	}

	go func() {
		fmt.Printf("🖥️  Dashboard listening on http://%s\n", dashAddr)
		err := dashServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	select {
	case <-sigChan:
		fmt.Println("Shutting down...")
	case <-ctx.Done():
		fmt.Println("Shutting down (context cancelled)...")
	}

	// Stop accepting new requests; allow in-flight to drain.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Gateway.ShutdownTimeout)
	defer shutdownCancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("API server shutdown error: %v\n", err)
	}
	if err := dashServer.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Dashboard server shutdown error: %v\n", err)
	}

	wa.Stop()
	loop.Stop()
	timeSvc.Close()
}
