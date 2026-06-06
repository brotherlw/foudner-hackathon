package main

import (
	"context"
	"log"
	"os"

	"github.com/agentic-paywall/agentic-paywall/internal/config"
	"github.com/agentic-paywall/agentic-paywall/internal/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	log.SetOutput(os.Stderr)

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	wallet := mcp.NewWallet(cfg)
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "agentic-paywall-wallet",
		Version: "1.0.0",
	}, nil)
	mcp.RegisterTools(server, wallet)

	log.Printf("payment-mcp starting (gateway=%s provider=%s)", cfg.Gateway.BaseURL, cfg.Provider)
	if err := server.Run(context.Background(), &mcpsdk.StdioTransport{}); err != nil {
		log.Fatalf("mcp server: %v", err)
	}
}

func loadConfig() (*config.Config, error) {
	path := "config.json"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		path = p
	}
	if _, err := os.Stat(path); err == nil {
		return config.Load(path)
	}
	return config.Default(), nil
}
