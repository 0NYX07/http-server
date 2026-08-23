package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/0NYX07/http-server/internal/config"
	"github.com/0NYX07/http-server/internal/logger"
	"github.com/0NYX07/http-server/internal/server"
)

func main() {
	// Print banner
	printBanner()

	// Load configuration
	cfg, err := config.Load(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Logging.Level, cfg.Logging.Colored, cfg.Logging.TimestampFormat)
	log := logger.GetLogger()

	// Create and start server
	srv, err := server.New(cfg)
	if err != nil {
		log.Error("Failed to create server: %v", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Error("Server error: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Info("Shutdown signal received, gracefully stopping server...")
	srv.Stop()
	log.Info("Server stopped successfully")
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║   Multi-threaded HTTP Server                                ║
║   Built with Go - Socket Programming Assignment             ║
║                                                              ║
║   Features:                                                  ║
║   • Low-level socket programming                            ║
║   • Multi-threaded with goroutines                          ║
║   • Persistent connections (keep-alive)                     ║
║   • Binary file transfer support                            ║
║   • Security features (path/host validation)                ║
║   • Real-time metrics and monitoring                        ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

